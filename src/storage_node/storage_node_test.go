package storage_node

import (
	"go.uber.org/zap"
	"net"
	"src/messages/controller_storage"
	"testing"
)

func TestStorageNode_ChecksumCheck(t *testing.T) {
	type fields struct {
		nodeID            string
		nodeStatus        controller_storage.StorageNodeMessage_NodeStatus
		interval          int32
		conn              net.Conn
		logger            *zap.Logger
		fileInfo          FileInfo
		networkInterfaces NetworkInterfaces
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
		{
			name: "Test Checksum Check",
			fields: fields{
				nodeID:     "node1",
				nodeStatus: controller_storage.StorageNodeMessage_ACTIVE,
				interval:   10,
				conn:       nil,
				logger:     zap.NewExample(),
				fileInfo:   FileInfo{},

				networkInterfaces: NetworkInterfaces{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StorageNode{
				nodeID:            tt.fields.nodeID,
				nodeStatus:        tt.fields.nodeStatus,
				interval:          tt.fields.interval,
				conn:              tt.fields.conn,
				logger:            tt.fields.logger,
				fileInfo:          tt.fields.fileInfo,
				networkInterfaces: tt.fields.networkInterfaces,
			}
			s.ChecksumCheck()
		})
	}
}
