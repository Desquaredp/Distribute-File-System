package main

import (
	"go.uber.org/zap"
	"net"
	proto3Storage "src/proto/client_storage"
	"testing"
)

func TestClient_handleStorageResponse(t *testing.T) {
	type fields struct {
		serverPort string
		logger     *zap.Logger
		conn       net.Conn
	}
	type args struct {
		proto *proto3Storage.ProtoHandler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				serverPort: tt.fields.serverPort,
				logger:     tt.fields.logger,
				conn:       tt.fields.conn,
			}
			c.handleStorageResponse(tt.args.proto)
		})
	}
}
