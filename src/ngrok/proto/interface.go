package proto

import (
	"context"

	"github.com/inconshreveable/ngrok/src/ngrok/conn"
)

type Protocol interface {
	GetName() string
	WrapConn(context.Context, conn.Conn, interface{}) conn.Conn
}
