package discover

// TODO: fill me up

import (
	"context"
	"io"
	"sync"
)

var (
	mu              sync.Mutex
	defaultProvider Provider
)

func SetProvider(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	defaultProvider = p
}

func Register(ctx context.Context, addr string, selfDesc ServiceDesc) (Dialer, error) {
	mu.Lock()
	p := defaultProvider
	mu.Unlock()
	return p.Register(ctx, addr, selfDesc)
}

type ServiceDesc struct {
	Name       string
	RPCAddr    string
	StatusAddr string
}

type Provider interface {
	Register(ctx context.Context, addr string, selfDesc ServiceDesc) (Dialer, error)
}

type Dialer interface {
	Dial() (io.ReadWriteCloser, error)
}
