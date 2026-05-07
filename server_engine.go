package engine

import (
	"context"
	"io"
)

type ServerEngine struct {
	transport Transport
	handler   Handler
}

type Transport interface {
	Accept(ctx context.Context) (io.ReadWriteCloser, error)
	Close() error
}

type Handler interface {
	Handle(ctx context.Context, stream io.ReadWriteCloser) error
}

func NewServerEngine(transport Transport, handler Handler) *ServerEngine {
	return &ServerEngine{
		transport: transport,
		handler:   handler,
	}
}

func (s *ServerEngine) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			stream, err := s.transport.Accept(ctx)
			if err != nil {
				return err
			}

			go func() {
				defer stream.Close()
				if err := s.handler.Handle(ctx, stream); err != nil {
					// Log error
				}
			}()
		}
	}
}

func (s *ServerEngine) Close() error {
	return s.transport.Close()
}
