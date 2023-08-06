package storage_node

import (
	"fmt"
	"go.uber.org/zap"
	"log"
	"net"
	"src/file"
	messagesClient "src/messages/client_storage"
	messages "src/messages/controller_storage"
	messagesStorage "src/messages/storage_storage"
	"sync"

	proto3Client "src/proto/client_storage"
	proto3 "src/proto/controller_storage"
	proto3Storage "src/proto/storage_storage"
	"time"
)

var sem = make(chan struct{}, 1) // allow up to 1 goroutines at once
type StorageNode struct {
	dir               string
	nodeID            string
	nodeStatus        messages.StorageNodeMessage_NodeStatus
	interval          int32
	conn              net.Conn
	msgHandlerStorage *messagesStorage.MessageHandler
	msgHandler        *messages.MessageHandler
	proto             *proto3.ProtoHandler
	protoStorage      *proto3Storage.ProtoHandler
	logger            *zap.Logger
	fileInfo          FileInfo
	mutex             *sync.Mutex
	networkInterfaces NetworkInterfaces
}

func (s *StorageNode) SetDir(dir string) {
	s.dir = dir
}

func (s *StorageNode) Dir() string {
	return s.dir
}

func (s *StorageNode) MsgHandlerStorage() *messagesStorage.MessageHandler {
	return s.msgHandlerStorage
}

func (s *StorageNode) SetMsgHandlerStorage(msgHandlerStorage *messagesStorage.MessageHandler) {
	s.msgHandlerStorage = msgHandlerStorage
}

func (s *StorageNode) ProtoStorage() *proto3Storage.ProtoHandler {
	return s.protoStorage
}

func (s *StorageNode) SetProtoStorage(protoStorage *proto3Storage.ProtoHandler) {
	s.protoStorage = protoStorage
}

func NewStorageNode(nodeID string, interfaces NetworkInterfaces, logger *zap.Logger) *StorageNode {
	newStorageNode := &StorageNode{
		nodeID:            nodeID,
		networkInterfaces: interfaces,
		logger:            logger,
		mutex:             &sync.Mutex{},
	}
	//start a go routine to listen for messages from the client
	return newStorageNode
}

func (s *StorageNode) ConcurrentChecksumCheck() {
	go func() {

		for {
			time.Sleep(30 * time.Second)
			s.ChecksumCheck()
		}
	}()
}

func (s *StorageNode) ChecksumCheck() {

	//files := s.fileInfo.AllFiles
	files := s.GetAllFiles()

	for _, f := range files {
		if !s.isFileFragment(f) {
			continue
		}
		if s.potentiallyCorrupt(f) {
			fileHandler2 := file.NewFileHandler(f)
			fileHandler2.SetDir(s.Dir())
			_, err := fileHandler2.ReadFile()
			if err != nil {
				s.logger.Error("There was an error reading the file.")
			}
			fileHandler2.FindAndSetCheckSum()
			valid, err := fileHandler2.ValidateChecksumFromFile(f)
			if err != nil {
				s.logger.Error("There was an error validating the checksum.")
			}

			if !valid {
				s.logger.Info("Checksum is invalid")
				s.HandleCorruptedFile(f)

			}
		}

	}

}

func (s *StorageNode) ConcurrentListen() {
	go s.ListenForClients()
	go s.ListenForOtherNodes()
}

func (s *StorageNode) ListenForOtherNodes() {

	host := s.networkInterfaces.NodeInterface.Host
	port := "23100"

	ln, err := net.Listen("tcp", host+":"+port)

	if err != nil {
		s.logger.Sugar().Errorf("There was an error listening on the port: %s", err)
		s.logger.Error("There was an error listening on the port.")
		return
	}

	s.logger.Info("Listening for other nodes on port " + port)
	s.logger.Info("Listening for other nodes on host " + host)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("There was an error accepting the connection.")
			return
		}
		msgHandler := messagesStorage.NewMessageHandler(conn)
		go s.handleOtherNode(msgHandler) //TODO: Race condition here
	}

}

func (s *StorageNode) ListenForClients() {

	//listen on the port specified in the config file

	host := s.networkInterfaces.NodeInterface.Host
	port := s.networkInterfaces.NodeInterface.ClientCommsPort

	ln, err := net.Listen("tcp", host+":"+port)

	if err != nil {
		s.logger.Sugar().Errorf("There was an error listening on the port: %s", err)
		s.logger.Error("There was an error listening on the port.")

		//TODO: check if there is a better way to handle this error
		panic(err)

	}

	s.logger.Info("Listening for clients on port " + port)
	s.logger.Info("Listening for clients on host " + host)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("There was an error accepting the connection.")
			continue
		}
		msgHandler := messagesClient.NewMessageHandler(conn)
		go s.handleClient(msgHandler)
	}

}
func (s *StorageNode) DialOtherNode(host string) (proto *proto3Storage.ProtoHandler, err error) {
	port := "23100"

	conn, err := net.Dial("tcp", host+":"+port)

	if err != nil {
		s.logger.Sugar().Errorf("There was an error connecting to the Host: %s", err)
		return
	}
	msgHandler := messagesStorage.NewMessageHandler(conn)
	//s.SetMsgHandlerStorage(msgHandler)
	proto = proto3Storage.NewProtoHandler(msgHandler, s.logger, s.dir)
	//s.SetProtoStorage(proto)
	return proto, nil
}

func (s *StorageNode) handleClient(msgHandler *messagesClient.MessageHandler) {

	defer msgHandler.Close()

	proto := proto3Client.NewProtoHandler(msgHandler, s.logger, s.dir)
	for {
		wrapper, _ := proto.MsgHandler().ClientRequestReceive()

		switch wrapper.Request.(type) {
		default:
			res, err := proto.HandleRequest(wrapper)
			if err != nil {
				s.logger.Sugar().Errorf("There was an error handling the request: %s", err)
				return
			}
			switch res.Operation {
			case "PUT":
			case "DATA":

				//TODO: handle the race condition before
				if res.Success {

					nodes := res.Result.(*file.FileHandler).Location()

					for _, node := range nodes {
						//stream the data to the nodes
						protoStorage, err := s.DialOtherNode(node) //TODO: Race condition here
						if err != nil {
							s.logger.Sugar().Errorf("There was an error connecting to the Host: %s", err)
							continue
						} else {
							s.logger.Info("Connected to node")
							s.StreamData(protoStorage, res.Result.(*file.FileHandler), node)
						}
					}

				} else {
					//TODO
				}

				return

			case "GET":
				return

			}

		case nil:
			log.Println("Received an empty message, terminating client")
			return

		}

	}

	//TODO
}

func (s *StorageNode) handleOtherNode(handler *messagesStorage.MessageHandler) {

	s.logger.Info("New node connected")
	defer handler.Close()
	proto := proto3Storage.NewProtoHandler(handler, s.logger, s.dir)

	for {
		wrapper, _ := proto.MsgHandler().ServerResponseReceive()

		switch wrapper.StorageNodeMessage.(type) {
		default:
			proto.HandleStorageNodeRequest(wrapper)

			return

		case nil:
			return
		}
	}

}

func (s *StorageNode) Interval(interval int32) {
	s.interval = interval
}

func (s *StorageNode) GetNodeId() string {
	return s.nodeID
}

func (s *StorageNode) GetNodeStatus() messages.StorageNodeMessage_NodeStatus {

	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.nodeStatus
}

func (s *StorageNode) GetInterval() int32 {
	return s.interval
}

func (s *StorageNode) ReConnect() (err error) {
	server := s.networkInterfaces.ControllerInterface.Port

	conn, err := net.Dial("tcp", server)

	if err != nil {
		fmt.Println("There was an error connecting to the Host.")
		return
	}
	msgHandler := messages.NewMessageHandler(conn)
	s.MsgHandler(msgHandler)
	proto := proto3.NewProtoHandler(msgHandler)
	s.Proto(proto)
	s.Conn(conn)

	return
}

func (s *StorageNode) Dial() (*proto3.ProtoHandler, net.Conn, error) {
	host := s.networkInterfaces.ControllerInterface.Host
	server := s.networkInterfaces.ControllerInterface.Port

	conn, err := net.Dial("tcp", host+":"+server)

	if err != nil {
		fmt.Println("There was an error connecting to the Host.")
		return nil, nil, err
	}
	msgHandler := messages.NewMessageHandler(conn)
	//s.MsgHandler(msgHandler)
	proto := proto3.NewProtoHandler(msgHandler)
	//s.Proto(proto)
	//s.Conn(conn)

	return proto, conn, err
}

func (s *StorageNode) handlePUTCopyRequest() (err error) {
	//TODO
	return
}

func (s *StorageNode) Disconnect(conn net.Conn) {
	conn.Close()
}

func (s *StorageNode) Conn(conn net.Conn) {
	s.conn = conn
}

func (s *StorageNode) MsgHandler(handler *messages.MessageHandler) {
	s.msgHandler = handler
}

func (s *StorageNode) Proto(proto *proto3.ProtoHandler) {
	s.proto = proto
}

func (s *StorageNode) HandleIntroduction(proto *proto3.ProtoHandler) (err error) {

	clientCommsPort := s.networkInterfaces.NodeInterface.ClientCommsPort
	address := s.networkInterfaces.NodeInterface.Host

	err = proto.HandleIntroRequest(s.nodeID, clientCommsPort, address)
	return
}

func (s *StorageNode) HandleOtherNodeConnection(protoStorage *proto3Storage.ProtoHandler) {
	defer protoStorage.MsgHandler().Close()

	for {
		wrapper, _ := protoStorage.MsgHandler().ServerResponseReceive()

		switch wrapper.StorageNodeMessage.(type) {
		default:
			protoStorage.HandleStorageNodeResponse(wrapper)
			return
		case nil:
			return
		}

	}

}
func (s *StorageNode) HandleConnection(proto *proto3.ProtoHandler) (err error) {
	defer proto.MsgHandler().Close()

	for {
		wrapper, _ := proto.MsgHandler().ServerResponseReceive()

		switch wrapper.ControllerMessage.(type) {
		default:
			//var status messages.ControllerMessage_StatusCode
			res := proto.HandleControllerResponse(wrapper)
			switch res.ResponseType() {
			case "AcceptNewNode":
				s.mutex.Lock()
				s.nodeStatus = messages.StorageNodeMessage_ACTIVE
				s.interval = res.(*proto3.AcceptNewNode).Interval
				s.mutex.Unlock()
			case "MissedHeartbeats":
				s.mutex.Lock()
				s.nodeStatus = messages.StorageNodeMessage_NEW
				s.mutex.Unlock()

			case "FileCorruption":
				//get the file from the node that has it
				s.logger.Sugar().Infof("File %s is corrupted, getting it from another node", res.(*proto3.FileCorruption).FileName)

				fileName := res.(*proto3.FileCorruption).FileName
				nodes := res.(*proto3.FileCorruption).StorageNodes
				for _, node := range nodes {
					protoStorage, err := s.DialOtherNode(node.Host())
					if err != nil {
						s.logger.Sugar().Errorf("There was an error connecting to the Host: %s", err)
					} else {
						s.logger.Sugar().Info("Connected to: ", node.Host())
						s.logger.Info("Connected to node")
						protoStorage.HandleGetReplicationRequest(fileName)
						go s.HandleOtherNodeConnection(protoStorage)
						break
					}

				}

			case "ReplicationRequest":

				statusCode := res.(*proto3.ReplicationRequest).StatusCode
				if statusCode == messages.ControllerMessage_OK {

					s.logger.Info("Received replication request for file")

					for _, frag := range res.(*proto3.ReplicationRequest).ReplicationInfo {
						s.logger.Sugar().Infof("Fragment %s is corrupted, getting it from another node", frag.FileName)
						nodes := frag.StorageNodes
						for _, node := range nodes {
							//stream the data to the nodes
							protoStorage, errf := s.DialOtherNode(node.Host())
							if errf != nil {
								s.logger.Sugar().Errorf("There was an error connecting to the Host: %s", err)
								return
							} else {
								s.logger.Info("Connected to node")
								fileHandler := file.FileHandler{}
								fileHandler.SetDir(s.dir)
								fileHandler.SetFileName(frag.FileName)
								fileHandler.ReadFile()
								fileHandler.FindAndSetCheckSum()

								s.StreamData(protoStorage, &fileHandler, node.Host())
							}
						}

					}

				}

			}

			return

		case nil:
			return
		}

	}

}

type HeartBeat struct {
	nodeID               string
	freeSpace            int64
	numRequestsProcessed int32
	newFiles             []string
	allFiles             []string
}

func (s *StorageNode) HandleHeartbeats(proto *proto3.ProtoHandler) (err error) {

	space := s.CheckFreeSpace()
	files := s.GetAllFiles()
	newFiles := s.GetNewFiles()

	proto.SendHeartbeatRequest(s.nodeID, space, s.fileInfo.NumRequestsProcessed, newFiles, files)

	return
}

func (s *StorageNode) StreamData(proto *proto3Storage.ProtoHandler, handler *file.FileHandler, node string) {

	proto.HandlePUTCopyRequest(s.nodeID, handler, node)

}

func (s *StorageNode) HandleCorruptedFile(f string) {

	s.logger.Info("Reporting corrupted file to Controller")
	proto, _, err := s.Dial()
	if err != nil {
		s.logger.Error("Error dialing: ", zap.Error(err))
		return
	}

	proto.HandleCorruptedFile(s.nodeID, f)
	go s.HandleConnection(proto)

}
