package storage_handler

import (
	"errors"
	"go.uber.org/zap"
	"math/rand"
	"regexp"
	"src/proto/controller_storage"
)

type Index struct {
	//Big brain data structure
	//fileMap: file name -> (fragMap : file fragment -> slice of node ids)
	fileMap map[string]map[string][]string

	//map: node id -> slice of file fragments that need to be replicated by it
	replicasNeeded map[string][]string

	logger *zap.Logger
	//IndexMutex *sync.RWMutex
}

func NewIndex(logger *zap.Logger) (index *Index) {
	index = &Index{
		fileMap: make(map[string]map[string][]string),
		//IndexMutex:     &sync.RWMutex{},
		logger:         logger,
		replicasNeeded: make(map[string][]string),
	}
	return
}

func (sh *StorageNodeHandler) ReplicaRequired(nodeId string) (needed bool) {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()
	if _, ok := sh.Index.replicasNeeded[nodeId]; ok {
		return true
	}

	return false
}

func (sh *StorageNodeHandler) FillFilesToReplicate(id string, p []*controller_storage.FragmentDistribution) (proto []*controller_storage.FragmentDistribution, err error) {

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	proto = make([]*controller_storage.FragmentDistribution, 0)
	index := 0
	for _, f := range sh.Index.replicasNeeded[id] {

		proto = append(proto, &controller_storage.FragmentDistribution{
			Fragment: f,
			Nodes:    make([]*controller_storage.Node, 0),
		})

		fileName := regexp.MustCompile(`_\d+$`).ReplaceAllString(f, "")

		nodesWithFile := sh.Index.GetFileMap()[fileName][f]
		if len(sh.spokeMap) < 3 {
			err = errors.New("not enough nodes to replicate")
			return
		}

		spokes := make([]string, 0)
		for k := range sh.spokeMap {
			spokes = append(spokes, k)
		}

		//shuffle the spokes

		rand.Shuffle(len(spokes), func(i, j int) { spokes[i], spokes[j] = spokes[j], spokes[i] })

		count := 0
		for i := 0; i < len(spokes); i++ {
			if !contains(nodesWithFile, spokes[i]) {
				//TODO: Error here. might need to restructure the code
				proto[index].Nodes = append(proto[index].Nodes, &controller_storage.Node{
					ID:   spokes[i],
					Host: sh.spokeMap[spokes[i]].host,
				})

				count++
				if len(nodesWithFile)+count == 3 {
					break
				}
			}
		}

		index++
	}

	sh.logger.Sugar().Info("The files to replicate are", proto)
	return

}

func contains(file []string, s string) bool {
	for _, v := range file {
		if v == s {
			return true
		}
	}
	return false

}

func (i *Index) GetFileMap() map[string]map[string][]string {
	return i.fileMap
}

func (sh *StorageNodeHandler) ConcurrentIndexing() {

	sh.mutex.Lock()
	defer sh.mutex.Unlock()
	//TODO: might cause nil pointer dereference
	sh.Index.fileMap = nil
	sh.Index.fileMap = make(map[string]map[string][]string)

	sh.Index.replicasNeeded = nil
	sh.Index.replicasNeeded = make(map[string][]string)

	newFiles := make(map[string]bool)

	for _, node := range sh.spokeMap {
		//if the file name ends with _x, then it is a file fragment
		for _, f := range node.allFiles {

			if isFileFragment(f) {
				sh.updateFileMap(f, node.ID)
			}
		}

		for _, newFile := range node.newFiles {
			newFiles[newFile] = true
		}

	}

	//print the file map
	sh.logger.Info("The new files are:", zap.Any("fileMap", newFiles))

	//check if the replica count is met
	//TODO: Uncomment this when replica count is to be used
	sh.replicaCountCheck(newFiles)

}

func (sh *StorageNodeHandler) replicaCountCheck(newFiles map[string]bool) {

	if len(sh.Index.fileMap) == 0 {
		return
	}
	for fileName, fragMap := range sh.Index.fileMap {
		for fragment, nodeIDs := range fragMap {

			if len(nodeIDs) < 3 && !newFiles[fragment] {

				sh.logger.Warn("Replica count not met", zap.String("fragment: ", fragment))
				//TODO: send a message to the node to replicate the file
				if _, ok := sh.Index.replicasNeeded[nodeIDs[0]]; !ok {
					sh.Index.replicasNeeded[nodeIDs[0]] = make([]string, 0)
				}
				sh.Index.replicasNeeded[nodeIDs[0]] = append(sh.Index.replicasNeeded[nodeIDs[0]], fragment)

			}

			sh.logger.Info("Replica count met", zap.String("file", fileName))

		}
	}
}

func (sh *StorageNodeHandler) updateFileMap(f string, nodeID string) {

	pattern := regexp.MustCompile(`_\d+$`)

	//find the file name
	fileName := pattern.ReplaceAllString(f, "")

	if _, ok := sh.Index.fileMap[fileName]; !ok {
		sh.Index.fileMap[fileName] = make(map[string][]string)
	}

	if _, ok := sh.Index.fileMap[fileName][f]; !ok {
		sh.Index.fileMap[fileName][f] = make([]string, 0)
	}

	sh.Index.fileMap[fileName][f] = append(sh.Index.fileMap[fileName][f], nodeID)

}

func isFileFragment(f string) bool {

	//check if the file name ends with _x, where x is a number
	//if it does, then it is a file fragment
	//if it doesn't, then it is a file

	pattern := regexp.MustCompile(`_\d+$`)

	if pattern.MatchString(f) {
		return true
	}

	return false

}
