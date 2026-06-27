package asynq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Client wraps asynq.Client with govern patterns.
type Client struct {
	client *asynq.Client
	logger *zap.SugaredLogger
	config *ClientConfig
}

// NewClient creates a task queue client with govern Redis integration.
//
// The returned cleanup function closes the client and should be called
// when the client is no longer needed (typically via defer).
//
// Example:
//
//	redisClient, redisCleanup, _ := redis.New("localhost:6379")
//	defer redisCleanup()
//
//	client, cleanup, _ := asynq.NewClient(redisClient)
//	defer cleanup()
func NewClient(redisClient redis.UniversalClient, opts ...ClientOption) (*Client, func(), error) {
	cfg := defaultClientConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// Convert redis.UniversalClient to asynq.RedisClientOpt
	asynqOpt := asynqRedisClientOpt(redisClient)

	client := asynq.NewClient(asynqOpt)

	c := &Client{
		client: client,
		logger: zap.NewNop().Sugar(), // Default no-op logger, can be replaced
		config: cfg,
	}

	cleanup := func() {
		if err := c.Close(); err != nil {
			if c.logger != nil {
				c.logger.Errorf("Failed to close asynq client: %v", err)
			}
		}
	}

	return c, cleanup, nil
}

// Enqueue puts a task on the queue with optional configuration.
//
// Example:
//
//	task := asynq.NewTask("email:send", payload)
//	info, err := client.Enqueue(ctx, task,
//	    asynq.WithQueue("critical"),
//	    asynq.WithMaxRetry(5),
//	)
func (c *Client) Enqueue(ctx context.Context, task *asynq.Task, opts ...EnqueueOption) (*asynq.TaskInfo, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	cfg := defaultEnqueueConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// Build asynq options from config
	asynqOpts := []asynq.Option{
		asynq.Queue(cfg.Queue),
		asynq.MaxRetry(cfg.MaxRetry),
		asynq.Timeout(cfg.Timeout),
	}

	if !cfg.Deadline.IsZero() {
		asynqOpts = append(asynqOpts, asynq.Deadline(cfg.Deadline))
	}

	if cfg.UniqueTTL > 0 {
		asynqOpts = append(asynqOpts, asynq.Unique(cfg.UniqueTTL))
	}

	info, err := c.client.Enqueue(task, asynqOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task %s: %w", task.Type(), err)
	}

	c.logger.Debugf("Enqueued task: type=%s id=%s queue=%s",
		task.Type(), info.ID, cfg.Queue)

	return info, nil
}

// EnqueueIn schedules a task to run after the specified duration.
//
// Example:
//
//	task := asynq.NewTask("email:send", payload)
//	info, err := client.EnqueueIn(ctx, task, 1*time.Hour)
func (c *Client) EnqueueIn(ctx context.Context, task *asynq.Task, duration time.Duration) (*asynq.TaskInfo, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	if duration < 0 {
		return nil, fmt.Errorf("duration cannot be negative")
	}

	info, err := c.client.Enqueue(task, asynq.ProcessIn(duration))
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task %s in %v: %w", task.Type(), duration, err)
	}

	c.logger.Debugf("Enqueued task in %v: type=%s id=%s", duration, task.Type(), info.ID)

	return info, nil
}

// EnqueueAt schedules a task to run at the specified time.
//
// Example:
//
//	task := asynq.NewTask("email:send", payload)
//	info, err := client.EnqueueAt(ctx, task, time.Date(2026, 3, 8, 9, 0, 0, time.UTC))
func (c *Client) EnqueueAt(ctx context.Context, task *asynq.Task, scheduledTime time.Time) (*asynq.TaskInfo, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	if scheduledTime.Before(time.Now()) {
		return nil, fmt.Errorf("scheduled time must be in the future")
	}

	info, err := c.client.Enqueue(task, asynq.ProcessAt(scheduledTime))
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task %s at %v: %w", task.Type(), scheduledTime, err)
	}

	c.logger.Debugf("Enqueued task at %v: type=%s id=%s", scheduledTime, task.Type(), info.ID)

	return info, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	return c.client.Close()
}

// NewTask creates a new task with the given type and payload.
// The payload is marshaled to JSON.
//
// Example:
//
//	payload := map[string]interface{}{"user_id": 42, "template": "welcome"}
//	task := asynq.NewTask("email:send", payload)
func NewTask(typ string, payload interface{}) (*asynq.Task, error) {
	var data []byte
	var err error

	if payload != nil {
		data, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	return asynq.NewTask(typ, data), nil
}

// asynqRedisClientOpt converts redis.UniversalClient to asynq.RedisClientOpt.
// asynq.RedisClientOpt has limited fields compared to go-redis options.
func asynqRedisClientOpt(redisClient redis.UniversalClient) asynq.RedisClientOpt {
	opt := asynq.RedisClientOpt{}

	switch client := redisClient.(type) {
	case *redis.Client:
		// Single-node client
		opts := client.Options()
		opt.Addr = opts.Addr
		opt.Password = opts.Password
		opt.DB = opts.DB
		opt.PoolSize = opts.PoolSize
		opt.DialTimeout = opts.DialTimeout
		opt.ReadTimeout = opts.ReadTimeout
		opt.WriteTimeout = opts.WriteTimeout
	case *redis.ClusterClient:
		// Cluster client: use first address as fallback
		// For cluster, asynq may need special handling
		opts := client.Options()
		if len(opts.Addrs) > 0 {
			opt.Addr = opts.Addrs[0]
		}
		opt.Password = opts.Password
		opt.PoolSize = opts.PoolSize
		opt.DialTimeout = opts.DialTimeout
		opt.ReadTimeout = opts.ReadTimeout
		opt.WriteTimeout = opts.WriteTimeout
	}

	return opt
}
