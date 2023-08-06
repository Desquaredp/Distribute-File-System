package file_distributor

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"src/controller/storage_handler"
)

type FileDistributor struct {
	fileName string
	fileSize int64
	//set chunk size to 128MB
	fragmentSize int64
	storageSys   *storage_handler.StorageNodeHandler
}

type Fragment struct {
	fragName string
	fragSize int64
}

func (f Fragment) GetFragmentName() string {
	return f.fragName
}

func (f Fragment) GetFragmentSize() int64 {
	return f.fragSize
}

func NewFileDistributor(fileName string, fileSize int64, chunkSize int64, storageSys *storage_handler.StorageNodeHandler) (fileDistributor *FileDistributor) {
	fileDistributor = &FileDistributor{
		fileName:   fileName,
		fileSize:   fileSize,
		storageSys: storageSys,
	}

	if chunkSize == 0 {
		//set chunk size to 128MB
		fileDistributor.fragmentSize = 128000000
	} else {
		fileDistributor.fragmentSize = chunkSize
	}
	return
}

func (fd *FileDistributor) sliceOfFragments() (fragments []*Fragment, err error) {
	if fd.fileSize == 0 {
		return nil, errors.New("file size is 0")
	}

	numFragments := int(fd.fileSize / fd.fragmentSize)
	remainder := int(fd.fileSize) % int(fd.fragmentSize)

	if remainder != 0 {
		numFragments++
	}

	fragments = make([]*Fragment, numFragments)

	for i := 0; i < numFragments-1; i++ {
		fragment := &Fragment{
			fragName: fd.fileName + "_" + fmt.Sprint(i),
			fragSize: fd.fragmentSize,
		}
		fragments[i] = fragment
	}

	// Last fragment with remaining bytes
	lastFragmentSize := fd.fragmentSize
	if remainder != 0 {
		lastFragmentSize = int64(remainder)
	}

	lastFragment := &Fragment{
		fragName: fd.fileName + "_" + fmt.Sprint(numFragments-1),
		fragSize: lastFragmentSize,
	}
	fragments[numFragments-1] = lastFragment

	return fragments, nil
}

func (fd *FileDistributor) SortNodes() (sortedNodes []*storage_handler.Node, err error) {
	storageNodes := fd.storageSys.GetStorageNodes()
	if len(storageNodes) == 0 {
		return nil, errors.New("no storage nodes available")
	}

	sortedNodes = make([]*storage_handler.Node, len(storageNodes))
	i := 0
	for _, node := range storageNodes {
		sortedNodes[i] = node
		i++
	}

	//sort the nodes by free storage
	sort.Slice(sortedNodes, func(i, j int) bool {
		return sortedNodes[i].GetFreeSpace() > sortedNodes[j].GetFreeSpace()
	})

	return sortedNodes, nil
}

func (fd *FileDistributor) DistributeCopies(chunkMap map[*Fragment][]*storage_handler.Node, nodes []*storage_handler.Node) (err error) {
	//randomize the order of the nodes
	randNodes := nodes
	rand.Shuffle(len(randNodes), func(i, j int) { randNodes[i], randNodes[j] = randNodes[j], randNodes[i] })

	for chunk, nodeIDs := range chunkMap {

		for {
			node := randNodes[rand.Intn(len(randNodes))]
			if qualifiesToHoldCopy(nodeIDs, node.GetID()) {
				nodeIDs = append(nodeIDs, node)
				chunkMap[chunk] = nodeIDs

			}
			if len(nodeIDs) == 3 {
				//randomize the order of the nodes
				rand.Shuffle(len(chunkMap[chunk]), func(i, j int) { chunkMap[chunk][i], chunkMap[chunk][j] = chunkMap[chunk][j], chunkMap[chunk][i] })
				break
			}
		}
	}
	return
}

func qualifiesToHoldCopy(nodes []*storage_handler.Node, id string) bool {

	if len(nodes) < 3 {

		for _, node := range nodes {
			if node.GetID() == id {
				return false
			}
		}
		return true

	} else {
		return false
	}

}

// create a map of storage nodes and their available storage
// sort the map by available storage
// create a map of chunks and their storage nodes
// iterate through the sorted map and assign chunks to storage nodes
// return the map of chunks and their storage nodes
func (fd *FileDistributor) DistributeFile() (chunkMap map[*Fragment][]*storage_handler.Node, err error) {

	if nodes, err := fd.SortNodes(); err != nil {
		return nil, err
	} else {

		//chunkMap to be assigned to the file
		chunkMap = make(map[*Fragment][]*storage_handler.Node)

		//get the fragments of the file
		fragments, err := fd.sliceOfFragments()
		if err != nil {
			return nil, err
		}

		//TODO: check if SetFreeSpace is creates race conditions
		for _, fragment := range fragments {
			for _, node := range nodes {
				if node.GetFreeSpace() >= fragment.fragSize {
					chunkMap[fragment] = append(chunkMap[fragment], node)
					node.SetFreeSpace(node.GetFreeSpace() - fragment.fragSize)
					break
				}
			}
		}

		if err := fd.DistributeCopies(chunkMap, nodes); err != nil {
			return nil, err
		}

		return chunkMap, nil

	}

}
