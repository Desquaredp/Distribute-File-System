package file_distributor

import "src/controller/storage_handler"

type FileDistributorInterface interface {
	DistributeFile() (map[*Fragment][]*storage_handler.Node, error) //map of chunk name to storage node IDs
	sliceOfFragments() ([]*Fragment, error)
	SortNodes() ([]*storage_handler.Node, error)
}
