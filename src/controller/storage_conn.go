package main

import (
	"fmt"
	"go.uber.org/zap"
	"net"
	"src/controller/storage_handler"
	storageNodeMessages "src/messages/controller_storage"
	storageNodeProto3 "src/proto/controller_storage"
)

func acceptStorageNodeConnections(listener1 net.Listener, spokeHandler *storage_handler.StorageNodeHandler, logger *zap.Logger) {
	for {
		if conn, err := listener1.Accept(); err == nil {

			msgHandler := storageNodeMessages.NewMessageHandler(conn)
			go handleStorageNode(msgHandler, spokeHandler, logger)

		} else {

			logger.Error(err.Error())
		}
	}

}

func handleStorageNode(msgHandler *storageNodeMessages.MessageHandler, spoke *storage_handler.StorageNodeHandler, logger *zap.Logger) {
	defer msgHandler.Close()
	proto := storageNodeProto3.NewProtoHandler(msgHandler)
	for {
		wrapper, _ := proto.MsgHandler().ClientRequestReceive()

		switch wrapper.StorageNodeMessage.(type) {

		default:
			ReqHandler := proto.HandleStorageNodeRequest(wrapper, HEARTBEAT_INTERVAL)

			switch ReqHandler.RequestType() {
			case "intro":
				//TODO: trim this down later
				Req := ReqHandler.(*storageNodeProto3.Request)
				nodeId := Req.GetNodeId()
				fmt.Println("Node ID: ", nodeId)

				if found, _ := spoke.Exists(nodeId); !found {
					fmt.Println("Status: ", Req.GetNodeStatus())
					if Req.GetNodeStatus() == "NEW" {
						logger.Sugar().Info("Node %s is new and it has been added\n", Req.GetNodeId())
						logger.Sugar().Info("Added new node %s\n", Req.GetNodeId())

						logger.Sugar().Info("Node communication port: ", Req.GetNodePort())
						logger.Sugar().Info("Node host: ", Req.GetNodeHost())
						spoke.Add(Req)
					} else {

						proto.HandleHeartbeatMiss(nodeId)
						//TODO : might not be needed
					}

				} else {
					//TODO: check if this is needed. If not, delete.
					/*
						if stale, _ := spoke.IsStale(nodeId, ACCEPTED_DELAY); stale {
							logger.Sugar().Info("Node %s missed heartbeats and it has been deleted\n", nodeId)
							spoke.Delete(nodeId)
							proto.HandleHeartbeatMiss(nodeId)
						} else {
							logger.Sugar().Info("Received hb from:", nodeId)

							go spoke.ResetTimer(nodeId)
							go spoke.UpdateNodeStats(Req)
						}

					*/
				}

				return
			case "heartbeat":
				//TODO: trim this down later
				Req := ReqHandler.(*storageNodeProto3.Request)
				nodeId := Req.GetNodeId()
				fmt.Println("Node ID: ", nodeId)

				if found, _ := spoke.Exists(nodeId); !found {
					fmt.Println("Status: ", Req.GetNodeStatus())
					if Req.GetNodeStatus() == "NEW" {
						logger.Sugar().Info("Node %s is new and it has been added\n", Req.GetNodeId())
						logger.Sugar().Info("Added new node %s\n", Req.GetNodeId())

						logger.Sugar().Info("Node communication port: ", Req.GetNodePort())
						logger.Sugar().Info("Node host: ", Req.GetNodeHost())
						spoke.Add(Req)
					} else {

						proto.HandleHeartbeatMiss(nodeId)
						//TODO : might not be needed
					}

				} else {
					if stale, _ := spoke.IsStale(nodeId, ACCEPTED_DELAY); stale {
						logger.Sugar().Info("Node %s missed heartbeats and it has been deleted\n", nodeId)
						spoke.Delete(nodeId)
						proto.HandleHeartbeatMiss(nodeId)
					} else {
						logger.Sugar().Info("Received hb from:", nodeId)
						//TODO: handle replication
						if spoke.ReplicaRequired(nodeId) {
							nodesProto := make([]*storageNodeProto3.FragmentDistribution, 0)
							p, err := spoke.FillFilesToReplicate(nodeId, nodesProto)
							nodesProto = p

							logger.Sugar().Info("Nodes to replicate: ", nodesProto)
							if err != nil {
								logger.Error(err.Error())
							} else {

								spoke.RemoveReplicaRequired(nodeId)
								logger.Sugar().Info("Sending replication request to node \n", nodeId)
								proto.HandleReplicationRequest(nodesProto, nodeId)
							}

						}

						go spoke.ResetTimer(nodeId)
						go func() {
							if Req == nil {
								logger.Error("Request is nil")
							} else {
								err := spoke.UpdateNodeStats(Req)
								if err != nil {
								}
							}

						}()
					}
				}

				return
			case "fileCorruption":
				Req := ReqHandler.(*storageNodeProto3.Request)
				nodeId := Req.GetNodeId()
				fileName := Req.CorruptedFile()
				logger.Sugar().Info("File %s is corrupted on node %s\n", fileName, nodeId)
				nodes := spoke.HasFile(fileName, nodeId)
				nodesProto := make([]*storageNodeProto3.Node, len(nodes))
				spoke.FillNodes(nodes, nodesProto)

				//print nodes
				for _, node := range nodesProto {
					logger.Sugar().Info("Node: ", node)
				}

				proto.HandleFileCorruptionResponse(nodesProto, Req)

			}

		case nil:
			return
		}

	}
}
