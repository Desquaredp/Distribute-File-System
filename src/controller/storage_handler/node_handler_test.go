package storage_handler

import (
	"go.uber.org/zap"
	"src/proto/controller_storage"
	"sync"
	"testing"
)

func TestStorageNodeHandler_FillNodes(t *testing.T) {
	type fields struct {
		spokeMap     map[string]*Node
		files        map[string]string
		Index        *Index
		totalStorage int64
		logger       *zap.Logger
		mutex        *sync.RWMutex
	}
	type args struct {
		nodes []Node
		proto []*controller_storage.Node
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				spokeMap: make(map[string]*Node),
				files:    make(map[string]string),
				logger:   zap.NewNop(),
				mutex:    &sync.RWMutex{},
			},
			args: args{
				nodes: []Node{
					{
						host: "localhost",
						ID:   "1",
					},
				},
				proto: []*controller_storage.Node{
					{
						Host: "localhost",
						ID:   "1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := &StorageNodeHandler{
				spokeMap:     tt.fields.spokeMap,
				files:        tt.fields.files,
				Index:        tt.fields.Index,
				totalStorage: tt.fields.totalStorage,
				logger:       tt.fields.logger,
				mutex:        tt.fields.mutex,
			}
			sh.FillNodes(tt.args.nodes, tt.args.proto)
		})
	}
}

func Test_fileFragmentFound(t *testing.T) {
	type args struct {
		file string
		f    string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "test 1",
			args: args{
				file: "file",
				f:    "file_1",
			},
			want: true,
		},
		{
			name: "test 2",
			args: args{
				file: "file",
				f:    "file.txt_1.checksum",
			},
			want: false,
		},
		{
			name: "test 3",
			args: args{
				file: "file_2",
				f:    "file_2_2",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fileFragmentFound(tt.args.file, tt.args.f); got != tt.want {
				t.Errorf("fileFragmentFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
