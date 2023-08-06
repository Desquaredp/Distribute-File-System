package storage_handler

import (
	"go.uber.org/zap"
	"reflect"
	"sync"
	"testing"
)

func TestStorageNodeHandler_ConcurrentIndexing(t *testing.T) {
	type fields struct {
		spokeMap     map[string]*Node
		files        map[string]string
		Index        *Index
		totalStorage int64
		logger       *zap.Logger
		mutex        *sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
		{
			name: "Test Concurrency",
			fields: fields{
				spokeMap:     map[string]*Node{"node1": &Node{}, "node2": &Node{}},
				files:        map[string]string{"file1": "node1", "file2": "node2"},
				Index:        &Index{},
				totalStorage: 1000,
				logger:       zap.NewNop(),
				mutex:        &sync.RWMutex{},
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
			sh.ConcurrentIndexing()
		})
	}
}

//func TestIndex_replicaCountCheck(t *testing.T) {
//	type fields struct {
//		fileMap map[string]map[string][]string
//		logger  *zap.Logger
//		mutex   *sync.RWMutex
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//		{
//			name: "Test replica count check",
//			fields: fields{
//				fileMap: map[string]map[string][]string{"file1": map[string][]string{"node1": []string{"node1", "node2"}}},
//				logger:  zap.NewNop(),
//				mutex:   &sync.RWMutex{},
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			i := &Index{
//				fileMap: tt.fields.fileMap,
//				logger:  tt.fields.logger,
//				mutex:   tt.fields.mutex,
//			}
//			i.replicaCountCheck()
//		})
//	}
//}
//
//func TestIndex_updateFileMap(t *testing.T) {
//	type fields struct {
//		fileMap map[string]map[string][]string
//		logger  *zap.Logger
//		mutex   *sync.RWMutex
//	}
//	type args struct {
//		f      string
//		nodeID string
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//	}{
//		// TODO: Add test cases.
//		{
//			name: "Test updateFileMap",
//			fields: fields{
//				fileMap: map[string]map[string][]string{},
//				logger:  zap.NewNop(),
//				mutex:   &sync.RWMutex{},
//			},
//			args: args{
//				f:      "file1",
//				nodeID: "node1",
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			i := &Index{
//				fileMap: tt.fields.fileMap,
//				logger:  tt.fields.logger,
//				mutex:   tt.fields.mutex,
//			}
//			i.updateFileMap(tt.args.f, tt.args.nodeID)
//		})
//	}
//}

func TestNewIndex(t *testing.T) {
	type args struct {
		logger *zap.Logger
	}
	tests := []struct {
		name      string
		args      args
		wantIndex *Index
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIndex := NewIndex(tt.args.logger); !reflect.DeepEqual(gotIndex, tt.wantIndex) {
				t.Errorf("NewIndex() = %v, want %v", gotIndex, tt.wantIndex)
			}
		})
	}
}

func TestStorageNodeHandler_ConcurrentIndexing1(t *testing.T) {
	type fields struct {
		spokeMap     map[string]*Node
		files        map[string]string
		Index        *Index
		totalStorage int64
		logger       *zap.Logger
		mutex        *sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
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
			sh.ConcurrentIndexing()
		})
	}
}

func Test_isFileFragment(t *testing.T) {
	type args struct {
		f string
	}
	tests := []struct {
		name string
		args args
		want bool
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFileFragment(tt.args.f); got != tt.want {
				t.Errorf("isFileFragment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIndex_GetFileMap(t *testing.T) {
	type fields struct {
		fileMap map[string]map[string][]string
		logger  *zap.Logger
		mutex   *sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]map[string][]string
	}{
		// TODO: Add test cases.
		{
			name: "Test GetFileMap",
			fields: fields{
				fileMap: map[string]map[string][]string{"file1": map[string][]string{"node1": []string{"node1", "node2"}}},
			},
			want: map[string]map[string][]string{"file1": map[string][]string{"node1": []string{"node1", "node2"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Index{
				fileMap: tt.fields.fileMap,
				logger:  tt.fields.logger,
			}
			if got := i.GetFileMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFileMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
