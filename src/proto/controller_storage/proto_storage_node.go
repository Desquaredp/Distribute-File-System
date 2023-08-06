package controller_storage

import (
	"fmt"
	messages "src/messages/controller_storage"
)

func (p *ProtoHandler) HandleIntroRequest(nodeID, openPort, address string) (err error) {

	msg := &messages.StorageNodeMessage{
		StorageNodeMessage: &messages.StorageNodeMessage_Intro_{
			Intro: &messages.StorageNodeMessage_Intro{
				NodeId:     nodeID,
				NodeStatus: messages.StorageNodeMessage_NEW,
				OpenPort:   openPort,
				Host:       address,
			},
		},
	}

	p.msgHandler.ClientRequestSend(msg)
	return
}

func (p *ProtoHandler) HandleCorruptedFile(id string, f string) {

	msg := &messages.StorageNodeMessage{
		StorageNodeMessage: &messages.StorageNodeMessage_FileCorruption_{
			FileCorruption: &messages.StorageNodeMessage_FileCorruption{
				NodeId:   id,
				FileName: f,
			},
		},
	}

	p.msgHandler.ClientRequestSend(msg)
	return
}

type Response interface {
	ResponseType() string
}

type AcceptNewNode struct {
	responseType string
	Interval     int32
	StatusCode   messages.ControllerMessage_StatusCode
}

func (a AcceptNewNode) ResponseType() string {
	return a.responseType
}

type MissedHeartbeats struct {
	responseType string
	StatusCode   messages.ControllerMessage_StatusCode
}

func (m MissedHeartbeats) ResponseType() string {
	return m.responseType
}

type FileCorruption struct {
	responseType string
	FileName     string
	StorageNodes []*StorageNodes

	statusCode messages.ControllerMessage_StatusCode
}

type ReplicationRequest struct {
	responseType    string
	StatusCode      messages.ControllerMessage_StatusCode
	ReplicationInfo []*ReplicationInfo
	//FileName     string
	//StorageNodes []*StorageNodes
}

type ReplicationInfo struct {
	FileName     string
	StorageNodes []*StorageNodes
}

func (r ReplicationRequest) ResponseType() string {
	return r.responseType
}

func (f FileCorruption) ResponseType() string {
	return f.responseType
}

type StorageNodes struct {
	nodeId string
	host   string
	port   string
}

func (s StorageNodes) NodeId() string {
	return s.nodeId
}

func (s StorageNodes) Host() string {
	return s.host
}

func (s StorageNodes) Port() string {
	return s.port
}

func (p *ProtoHandler) fetchAcceptNewNodeResponse(msg *messages.ControllerMessage_AcceptNewNode_) (res Response) {

	fmt.Println(msg.AcceptNewNode.StatusCode)
	fmt.Println(msg.AcceptNewNode.ExpectedHeartbeatInterval)

	if msg.AcceptNewNode.StatusCode == messages.ControllerMessage_OK {
		res = &AcceptNewNode{
			responseType: "AcceptNewNode",
			Interval:     msg.AcceptNewNode.ExpectedHeartbeatInterval,
			StatusCode:   msg.AcceptNewNode.StatusCode,
		}

	} else {
		res = &AcceptNewNode{
			responseType: "AcceptNewNode",
			Interval:     0,
			StatusCode:   msg.AcceptNewNode.StatusCode,
		}
	}
	return
}

func (p *ProtoHandler) fetchMissedHeartbeatsResponse(msg *messages.ControllerMessage_MissedHeartbeats_) (res Response) {

	if msg.MissedHeartbeats.StatusCode == messages.ControllerMessage_OK {
		res = &MissedHeartbeats{
			responseType: "MissedHeartbeats",
			StatusCode:   msg.MissedHeartbeats.StatusCode,
		}
	} else {
		res = &MissedHeartbeats{
			responseType: "MissedHeartbeats",
			StatusCode:   msg.MissedHeartbeats.StatusCode,
		}

	}
	return
}

func (p *ProtoHandler) fetchFileCorruptionResponse(msg *messages.ControllerMessage_FileCorruptionResponse_) (res Response) {

	if msg.FileCorruptionResponse.StatusCode == messages.ControllerMessage_OK {

		res = &FileCorruption{
			responseType: "FileCorruption",
			FileName:     msg.FileCorruptionResponse.FileName,
			StorageNodes: make([]*StorageNodes, len(msg.FileCorruptionResponse.StorageNodes)),
			statusCode:   msg.FileCorruptionResponse.StatusCode,
		}

		for i, node := range msg.FileCorruptionResponse.StorageNodes {
			res.(*FileCorruption).StorageNodes[i] = &StorageNodes{
				nodeId: node.StorageNodeId,
				host:   node.Host,
				port:   node.Port,
			}
		}

	} else {
		res = &FileCorruption{
			responseType: "FileCorruption",
			FileName:     msg.FileCorruptionResponse.FileName,
			StorageNodes: nil,
			statusCode:   msg.FileCorruptionResponse.StatusCode,
		}
	}

	return

}

func (p *ProtoHandler) fetchReplicationRequest(msg *messages.ControllerMessage_ReplicationRequest_) (res Response) {

	res = &ReplicationRequest{
		responseType:    "ReplicationRequest",
		StatusCode:      msg.ReplicationRequest.StatusCode,
		ReplicationInfo: make([]*ReplicationInfo, len(msg.ReplicationRequest.ReplicationInfo)),
	}

	for i, node := range msg.ReplicationRequest.ReplicationInfo {

		res.(*ReplicationRequest).ReplicationInfo[i] = &ReplicationInfo{
			FileName:     node.FileName,
			StorageNodes: make([]*StorageNodes, len(node.StorageNodes)),
		}

		for j, node := range node.StorageNodes {
			res.(*ReplicationRequest).ReplicationInfo[i].StorageNodes[j] = &StorageNodes{
				nodeId: node.StorageNodeId,
				host:   node.Host,
				//TODO: port is not being sent by the controller, fix this
			}
		}
	}

	return

}

func (p *ProtoHandler) SendHeartbeatRequest(nodeID string, freeSpace int64, numRequestsProcessed int32, newFiles []string, allFiles []string) {
	msg := &messages.StorageNodeMessage{
		StorageNodeMessage: &messages.StorageNodeMessage_Heartbeat_{
			Heartbeat: &messages.StorageNodeMessage_Heartbeat{
				NodeId:               nodeID,
				NodeStatus:           messages.StorageNodeMessage_ACTIVE,
				FreeSpace:            freeSpace,
				NumRequestsProcessed: numRequestsProcessed,
				NewFiles:             newFiles,
				AllFiles:             allFiles,
			},
		},
	}

	p.msgHandler.ClientRequestSend(msg)

}
