package storage_handler

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"src/proto/controller_storage"
	"strings"
	"sync"
	"time"
)

type Node struct {
	ID       string
	openPort string
	host     string

	LastHeartbeat    time.Time
	MissedHeartbeats int

	freeSpace            int64
	numRequestsProcessed int32
	newFiles             []string
	allFiles             []string
}

// create Getters and Setters for the Node struct
func (n *Node) GetID() string {
	return n.ID
}

func (n *Node) GetOpenPort() string {
	return n.openPort

}

func (n *Node) GetAddress() string {
	return n.host
}

func (n *Node) GetLastHeartbeat() time.Time {
	return n.LastHeartbeat
}

func (n *Node) GetMissedHeartbeats() int {
	return n.MissedHeartbeats
}

func (n *Node) GetFreeSpace() int64 {
	return n.freeSpace
}

func (n *Node) GetNumRequestsProcessed() int32 {
	return n.numRequestsProcessed
}

func (n *Node) GetNewFiles() []string {
	return n.newFiles
}

func (n *Node) GetAllFiles() []string {
	return n.allFiles
}

func (n *Node) SetFreeSpace(f int64) {
	n.freeSpace = f
}

type ClientSafeNode struct {
	nodeId   string
	openPort string
	host     string
}

type StorageNodeHandler struct {
	spokeMap map[string]*Node
	files    map[string]string
	Index    *Index

	totalStorage int64
	logger       *zap.Logger
	mutex        *sync.RWMutex
}

func NewStorageNodeHandler(logger *zap.Logger) (newSH *StorageNodeHandler) {
	newSH = &StorageNodeHandler{
		spokeMap: make(map[string]*Node),
		files:    make(map[string]string),
		logger:   logger,
		mutex:    &sync.RWMutex{},

		Index: NewIndex(logger),
	}

	return
}

func (sh *StorageNodeHandler) GetFileName(fragID string) string {

	lastUnderscore := strings.LastIndex(fragID, "_")
	if lastUnderscore == -1 {
		// If there is no underscore, return the whole string
		return fragID
	}
	return fragID[:lastUnderscore]
}

func (sh *StorageNodeHandler) ExtractFiles(fileFragments []string, logger *zap.Logger) (files map[string]bool) {

	files = make(map[string]bool)
	for _, f := range fileFragments {
		fileName := sh.GetFileName(f)
		files[fileName] = true
	}

	return
}
func (sh *StorageNodeHandler) FindAllFiles(logger *zap.Logger) (files []string) {

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	logger.Info("Finding all files")
	files = make([]string, 0)
	for _, node := range sh.spokeMap {
		for _, f := range node.allFiles {
			if isFileFragment(f) {
				files = append(files, f)
			}
		}
	}

	if len(files) == 0 {
		logger.Info("No files found")
		return nil
	}

	return
}
func (sh *StorageNodeHandler) FindFiles(file string, logger *zap.Logger) (fileMap map[string][]*Node) {

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	logger.Info("Finding files", zap.String("file", file))

	// map: file fragment -> list of nodes that have the file fragment
	//spokemap: nodeID -> node struct
	fileMap = make(map[string][]*Node)
	for _, node := range sh.spokeMap {
		for _, f := range node.allFiles {
			if fileFragmentFound(file, f) {
				if _, ok := fileMap[f]; !ok {
					fileMap[f] = make([]*Node, 0)
				}
				fileMap[f] = append(fileMap[f], node)
			}
		}
	}

	if len(fileMap) == 0 {
		logger.Info("No files found")
		return nil
	}

	return
}

func fileFragmentFound(file string, f string) bool {

	if len(f) < len(file) {
		return false
	} else {

		//check if the f contains the term checksum
		if strings.Contains(f, "checksum") {
			return false
		}
		return file == f[:len(file)]
	}

}

func (sh *StorageNodeHandler) UpdateNodeStats(Req *controller_storage.Request) (err error) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	nodeId := Req.GetNodeId()
	node := sh.spokeMap[nodeId]

	node.freeSpace = Req.GetNodeFreeSpace()
	sh.logger.Info("Updating node stats", zap.String("nodeId", nodeId), zap.Int64("freeSpace", node.freeSpace))
	node.numRequestsProcessed = Req.GetNodeNumRequestsProcessed()

	node.newFiles = Req.GetNewFiles()

	//TODO: sh.updateNodeFiles(files, node)
	node.allFiles = Req.GetAllFiles()

	return
}

func (sh *StorageNodeHandler) Add(Req *controller_storage.Request) (err error) {

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	node := &Node{
		ID:               Req.GetNodeId(),
		openPort:         Req.GetNodePort(),
		host:             Req.GetNodeHost(),
		LastHeartbeat:    time.Now(),
		MissedHeartbeats: 0,
	}

	sh.spokeMap[Req.GetNodeId()] = node

	return
}

func (sh *StorageNodeHandler) FileExists(fileName string) (found bool, err error) {

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	if _, ok := sh.files[fileName]; ok {
		return true, nil
	}
	return
}

func (sh *StorageNodeHandler) AddFile(fileName string) {

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	sh.files[fileName] = fileName
}

/*
func (sh *StorageNodeHandler) AddFile(fileName string, distributorMap map[*file_distributor.Fragment][]*Node) (err error) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	sh.files[fileName] = distributorMap

	return
}
*/

func (sh *StorageNodeHandler) GetVal(nodeId string) (val *Node, err error) {

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	val, exists := sh.spokeMap[nodeId]
	if !exists {
		return nil, err
	}
	return
}

func (sh *StorageNodeHandler) Exists(nodeId string) (found bool, err error) {

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	_, found = sh.spokeMap[nodeId]
	return
}

func (sh *StorageNodeHandler) ResetTimer(nodeId string) (err error) {

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	node, ok := sh.spokeMap[nodeId]
	if !ok {
		return errors.New("node doesn't exist")
	}

	node.LastHeartbeat = time.Now()
	node.MissedHeartbeats = 0
	return
}

func (sh *StorageNodeHandler) Delete(nodeId string) (err error) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	_, ok := sh.spokeMap[nodeId]
	if !ok {
		return errors.New("Node doesn't exist.")
	}

	delete(sh.spokeMap, nodeId)
	fmt.Println("Map size shrunk to: ", len(sh.spokeMap))
	return

}

func (sh *StorageNodeHandler) IsStale(nodeId string, acceptedDelay int) (stale bool, err error) {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	if _, ok := sh.spokeMap[nodeId]; ok {
		lastBeat := sh.spokeMap[nodeId].LastHeartbeat

		if time.Since(lastBeat) > time.Duration(acceptedDelay)*time.Second {
			stale = true
			return
		}

	}
	return
}

func (sh *StorageNodeHandler) GetStorageNodes() (nodes []*Node) {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	for _, node := range sh.spokeMap {
		nodes = append(nodes, node)
	}

	return

}

func (sh *StorageNodeHandler) ConcurrentStaleNodeRemoval(delay int, logger *zap.Logger) {

	sh.mutex.Lock()
	defer sh.mutex.Unlock()
	for nodeId, node := range sh.spokeMap {
		if time.Since(node.LastHeartbeat) > time.Duration(delay)*time.Second {
			logger.Info("Removing stale node", zap.String("nodeId", nodeId))
			delete(sh.spokeMap, nodeId)
		}
	}
}

func (sh *StorageNodeHandler) HasFile(name string, id string) (nodes []Node) {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	nodes = make([]Node, 0)
	for _, node := range sh.spokeMap {
		if node.ID != id {

			for _, file := range node.allFiles {
				if file == name {
					nodes = append(nodes, *node)
				}
			}
		}
	}
	return

}

func (sh *StorageNodeHandler) FillNodes(nodes []Node, proto []*controller_storage.Node) {

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	if len(nodes) == 0 {
		return
	}

	for i := 0; i < len(nodes); i++ {

		node := &controller_storage.Node{
			ID:   nodes[i].ID,
			Host: nodes[i].host,
		}

		proto[i] = node
	}

	fmt.Println("Filling nodes: ", proto[0].ID)

}

func (sh *StorageNodeHandler) RemoveReplicaRequired(id string) {

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	delete(sh.Index.replicasNeeded, id)

}

func (sh *StorageNodeHandler) GetNodeInfo() interface{} {

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	// return a copy of the map
	mapCopy := make(map[string]Node)
	for k, v := range sh.spokeMap {
		mapCopy[k] = Node{
			ID:        v.ID,
			freeSpace: v.freeSpace,
		}
	}

	return mapCopy

}
