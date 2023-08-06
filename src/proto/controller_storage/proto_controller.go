package controller_storage

import (
	messages "src/messages/controller_storage"
)

type Request struct {
	requestType          string
	nodeId               string
	openPort             string
	host                 string
	corruptedFile        string
	nodeStatus           messages.StorageNodeMessage_NodeStatus
	freeSpace            int64
	numRequestsProcessed int32
	newFiles             []string
	allFiles             []string
	err                  error
}

func (r *Request) RequestType() string {
	return r.requestType
}

func (r *Request) CorruptedFile() string {
	return r.corruptedFile
}

func (r *Request) GetNodeId() string {
	return r.nodeId
}

func (r *Request) GetNodePort() string {
	return r.openPort
}

func (r *Request) GetNodeStatus() string {
	return r.nodeStatus.String()
}

func (r *Request) GetNodeFreeSpace() int64 {
	return r.freeSpace
}

func (r *Request) GetNodeNumRequestsProcessed() int32 {
	return r.numRequestsProcessed
}

func (r *Request) GetNewFiles() []string {
	return r.newFiles
}

func (r *Request) GetAllFiles() []string {
	return r.allFiles
}

func (r *Request) GetNodeHost() string {
	return r.host
}

type RequestHandler interface {
	RequestType() string
}

func (p *ProtoHandler) fetchIntroRequest(msg *messages.StorageNodeMessage_Intro_, interval int) RequestHandler {

	//fmt.Println("The newly added node is: ", msg.Intro.NodeId)
	//TODO send an ACK after processing the metadata.

	introRequest := &Request{
		requestType: "intro",
		nodeId:      msg.Intro.NodeId,
		nodeStatus:  msg.Intro.NodeStatus,
		openPort:    msg.Intro.OpenPort,
		host:        msg.Intro.Host,
	}

	p.handleIntroResponse(interval)

	return introRequest
}

func (p *ProtoHandler) fetchHeartbeatRequest(msg *messages.StorageNodeMessage_Heartbeat_) RequestHandler {
	/*
		string node_id = 1;
		    NodeStatus node_status = 2;
		    float free_space = 3;
		    int32 num_requests_processed = 4;
		    repeated string new_files = 5;
		    repeated string all_files = 6;
	*/

	HeartbeatRequest := &Request{
		requestType:          "heartbeat",
		nodeId:               msg.Heartbeat.NodeId,
		nodeStatus:           msg.Heartbeat.NodeStatus,
		freeSpace:            msg.Heartbeat.FreeSpace,
		numRequestsProcessed: msg.Heartbeat.NumRequestsProcessed,
	}

	for _, files := range msg.Heartbeat.NewFiles {
		HeartbeatRequest.newFiles = append(HeartbeatRequest.newFiles, files)
	}

	for _, files := range msg.Heartbeat.AllFiles {
		HeartbeatRequest.allFiles = append(HeartbeatRequest.allFiles, files)
	}

	//TODO: verify if the storage_node is valid, and take actions accordingly
	//p.HandleHeartbeatMiss(HeartbeatRequest.nodeId)
	return HeartbeatRequest
}

func (p *ProtoHandler) fetchFileCorruptionRequest(msg *messages.StorageNodeMessage_FileCorruption_) RequestHandler {

	FileCorruptionRequest := &Request{
		requestType:   "fileCorruption",
		nodeId:        msg.FileCorruption.NodeId,
		corruptedFile: msg.FileCorruption.FileName,
	}

	return FileCorruptionRequest

}

func (p *ProtoHandler) handleIntroResponse(interval int) {

	//TODO:check if the ACK is OK

	res := &messages.ControllerMessage{ControllerMessage: &messages.ControllerMessage_AcceptNewNode_{

		AcceptNewNode: &messages.ControllerMessage_AcceptNewNode{
			StatusCode: messages.ControllerMessage_OK,
			//TODO: get rid of hardcoding
			ExpectedHeartbeatInterval: int32(interval),
		},
	}}

	p.msgHandler.ServerResponseSend(res)
}

func (p *ProtoHandler) HandleHeartbeatMiss(nodeId string) {
	res := &messages.ControllerMessage{ControllerMessage: &messages.ControllerMessage_MissedHeartbeats_{

		MissedHeartbeats: &messages.ControllerMessage_MissedHeartbeats{

			StatusCode: messages.ControllerMessage_UNKNOWN_NODE,
			NodeId:     nodeId,
			CanReinit:  true,
		},
	},
	}

	p.msgHandler.ServerResponseSend(res)
}

type Node struct {
	ID   string
	Host string
	Port string
}

type FragmentDistribution struct {
	Fragment string
	Nodes    []*Node
}

func (p *ProtoHandler) HandleFileCorruptionResponse(nodes []*Node, req *Request) {

	res := &messages.ControllerMessage{
		ControllerMessage: &messages.ControllerMessage_FileCorruptionResponse_{

			FileCorruptionResponse: &messages.ControllerMessage_FileCorruptionResponse{
				StatusCode:   messages.ControllerMessage_OK,
				StorageNodes: make([]*messages.ControllerMessage_StorageNodeInfo, len(nodes)),
				FileName:     req.CorruptedFile(),
			},
		},
	}

	for i, node := range nodes {
		res.ControllerMessage.(*messages.ControllerMessage_FileCorruptionResponse_).FileCorruptionResponse.StorageNodes[i] = &messages.ControllerMessage_StorageNodeInfo{
			StorageNodeId: node.ID,
			Host:          node.Host,
			//TODO: add the port
		}
	}

	p.msgHandler.ServerResponseSend(res)

}

func (p *ProtoHandler) HandleReplicationRequest(proto []*FragmentDistribution, id string) {

	var res *messages.ControllerMessage

	res = &messages.ControllerMessage{
		ControllerMessage: &messages.ControllerMessage_ReplicationRequest_{
			ReplicationRequest: &messages.ControllerMessage_ReplicationRequest{
				StatusCode:      messages.ControllerMessage_OK,
				ReplicationInfo: make([]*messages.ControllerMessage_ReplicationInfo, len(proto)),
			},
		},
	}
	for i, frag := range proto {
		//Set the filename for each fragment to be replicated, and create a list of nodes to replicate to.

		res.ControllerMessage.(*messages.ControllerMessage_ReplicationRequest_).ReplicationRequest.ReplicationInfo[i] = &messages.ControllerMessage_ReplicationInfo{
			FileName:     frag.Fragment,
			StorageNodes: make([]*messages.ControllerMessage_StorageNodeInfo, len(frag.Nodes)),
		}

		for j, node := range frag.Nodes {
			res.ControllerMessage.(*messages.ControllerMessage_ReplicationRequest_).ReplicationRequest.ReplicationInfo[i].StorageNodes[j] = &messages.ControllerMessage_StorageNodeInfo{
				StorageNodeId: node.ID,
				Host:          node.Host,
				//TODO: add the port when it is implemented
			}
		}

	}

	p.msgHandler.ServerResponseSend(res)
}
