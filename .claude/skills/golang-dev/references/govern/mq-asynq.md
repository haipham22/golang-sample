# govern/mq/asynq

Import: `github.com/haipham22/govern/mq/asynq`

Asynq task queue wrapper with consistent options + graceful lifecycle.

## Use When

- App processes async tasks (email, reports, webhooks) via asynq + Redis.

## Typical Wiring

```go
import "github.com/haipham22/govern/mq/asynq"

// Client (enqueue)
client, cleanup, _ := asynq.NewClient(redisAddr)
defer cleanup()

// Server (consume)
srv, cleanup, _ := asynq.NewServer(redisAddr, asynq.WithLogger(logger))
defer cleanup()

srv.Handle("email:send", handleEmail)
graceful.Run(ctx, logger, 30*time.Second, srv)
```

## Rules

- ✅ Define task type strings as constants (e.g. `email:send`).
- ✅ Handlers should be idempotent — asynq may retry.
- ✅ Wire server through `graceful.Run` for drain on shutdown.
- ✅ Set per-task retry policy + max attempts.
- ✅ Move DB work to repository; keep handler thin.
- ❌ Do not enqueue inside DB transactions (use outbox pattern if needed).
- ❌ Do not block handlers on long external calls without timeout.

## Avoid

- Hand-wiring asynq client/server lifecycle.
- Asynq without graceful shutdown integration.
- Per-handler Redis connection instead of shared client.

## Reference

Source: [`mq/asynq/`](../../../../../../../mq/asynq/). Uses `github.com/hibiken/asynq`.
