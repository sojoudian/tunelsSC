package proto

import (
	"context"

	"github.com/inconshreveable/ngrok/src/ngrok/conn"
)

type Tcp struct{}

func NewTcp() *Tcp {
	return new(Tcp)
}

func (h *Tcp) GetName() string { return "tcp" }

func (h *Tcp) WrapConn(ctx context.Context, c conn.Conn, connCtx interface{}) conn.Conn {
	return c
}
