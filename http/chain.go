package http

import "net/http"

// Chain represents middleware chain.
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a new middleware chain.
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{
		middlewares: middlewares,
	}
}

// Then applies chain to handler.
func (c *Chain) Then(h http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

// Append adds middleware to end of chain.
func (c *Chain) Append(middlewares ...Middleware) *Chain {
	c.middlewares = append(c.middlewares, middlewares...)
	return c
}

// Prepend adds middleware to start of chain.
func (c *Chain) Prepend(middlewares ...Middleware) *Chain {
	c.middlewares = append(middlewares, c.middlewares...)
	return c
}
