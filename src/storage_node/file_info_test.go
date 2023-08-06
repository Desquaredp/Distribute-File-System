package storage_node

import (
	"go.uber.org/zap"
	"net"
	"src/messages/controller_storage"
	"sync"
	"testing"
)

func TestStorageNode_isFileFragment(t *testing.T) {
	type fields struct {
		nodeID     string
		nodeStatus controller_storage.StorageNodeMessage_NodeStatus
		interval   int32
		conn       net.Conn

		logger            *zap.Logger
		fileInfo          FileInfo
		mutex             *sync.Mutex
		networkInterfaces NetworkInterfaces
	}
	type args struct {
		f string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
		{
			name: "Test isFileFragment",
			args: args{
				f: "file1",
			},
			want: false,
		},
		{
			name: "Test isFileFragment",
			args: args{
				f: "file1_1",
			},
			want: true,
		},
		{
			name: "Test isFileFragment",
			args: args{
				f: "file1_1.checksum",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StorageNode{
				nodeID:     tt.fields.nodeID,
				nodeStatus: tt.fields.nodeStatus,
				interval:   tt.fields.interval,
				conn:       tt.fields.conn,

				logger:            tt.fields.logger,
				fileInfo:          tt.fields.fileInfo,
				mutex:             tt.fields.mutex,
				networkInterfaces: tt.fields.networkInterfaces,
			}
			if got := s.isFileFragment(tt.args.f); got != tt.want {
				t.Errorf("isFileFragment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorageNode_isNewFile(t *testing.T) {
	type fields struct {
		nodeID     string
		nodeStatus controller_storage.StorageNodeMessage_NodeStatus
		interval   int32
		conn       net.Conn

		logger            *zap.Logger
		fileInfo          FileInfo
		mutex             *sync.Mutex
		networkInterfaces NetworkInterfaces
	}
	type args struct {
		f string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.

		{
			name: "Test isNewFile",
			args: args{
				f: "/home/deep/cs677/projects/P1-pickle-rick/src/proto/client_storage/proto_client.go",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StorageNode{
				nodeID:     tt.fields.nodeID,
				nodeStatus: tt.fields.nodeStatus,
				interval:   tt.fields.interval,
				conn:       tt.fields.conn,

				logger:            tt.fields.logger,
				fileInfo:          tt.fields.fileInfo,
				mutex:             tt.fields.mutex,
				networkInterfaces: tt.fields.networkInterfaces,
			}
			if got := s.isNewFile(tt.args.f); got != tt.want {
				t.Errorf("isNewFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
