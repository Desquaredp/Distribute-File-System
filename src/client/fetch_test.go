package main

import (
	"go.uber.org/zap"
	"net"
	"testing"
)

func TestClient_GetFileName(t *testing.T) {
	type fields struct {
		serverPort string
		logger     *zap.Logger
		conn       net.Conn
	}
	type args struct {
		fragName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "Test 1",
			fields: fields{
				serverPort: "8080",
				logger:     zap.NewExample(),
				conn:       nil,
			},
			args: args{
				fragName: "file_21",
			},
			want: "file",
		},
		{
			name: "Test 2",
			fields: fields{
				serverPort: "8080",
				logger:     zap.NewExample(),
				conn:       nil,
			},
			args: args{
				fragName: "file_name_21",
			},
			want: "file_name",
		},
		{
			name: "Test 3",
			fields: fields{
				serverPort: "8080",
				logger:     zap.NewExample(),
				conn:       nil,
			},
			args: args{
				fragName: "file_name_21_21",
			},
			want: "file_name_21",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				serverPort: tt.fields.serverPort,
				logger:     tt.fields.logger,
				conn:       tt.fields.conn,
			}
			if got := c.GetFileName(tt.args.fragName); got != tt.want {
				t.Errorf("GetFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetFileName1(t *testing.T) {
	type fields struct {
		serverPort string
		logger     *zap.Logger
		conn       net.Conn
	}
	type args struct {
		fragName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "Test 1",
			fields: fields{
				serverPort: "8080",
				logger:     zap.NewExample(),
				conn:       nil,
			},
			args: args{
				fragName: "file_21",
			},
			want: "file",
		},
		{
			name: "Test 2",
			fields: fields{
				serverPort: "8080",
				logger:     zap.NewExample(),
				conn:       nil,
			},
			args: args{
				fragName: "file_21.checksum",
			},
			want: "file",
		},

		{
			name: "Test 3",
			fields: fields{
				serverPort: "8080",
				logger:     zap.NewExample(),
				conn:       nil,
			},
			args: args{
				fragName: "file.txt_21.checksum",
			},
			want: "file.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				serverPort: tt.fields.serverPort,
				logger:     tt.fields.logger,
				conn:       tt.fields.conn,
			}
			if got := c.GetFileName(tt.args.fragName); got != tt.want {
				t.Errorf("GetFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}
