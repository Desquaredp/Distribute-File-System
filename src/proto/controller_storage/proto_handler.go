package controller_storage

import (
	messages "src/messages/controller_storage"
)

type ProtoHandler struct {
	msgHandler *messages.MessageHandler
	nodeId     string
}

func (p *ProtoHandler) MsgHandler() *messages.MessageHandler {
	return p.msgHandler
}

func NewProtoHandler(msgHandler *messages.MessageHandler) *ProtoHandler {
	newProtoHandler := &ProtoHandler{
		msgHandler: msgHandler,
	}
	return newProtoHandler
}

func (p *ProtoHandler) HandleControllerResponse(wrapper *messages.ControllerMessage) (res Response) {

	switch msg := wrapper.ControllerMessage.(type) {

	case *messages.ControllerMessage_AcceptNewNode_:
		res = p.fetchAcceptNewNodeResponse(msg)
		return

	case *messages.ControllerMessage_MissedHeartbeats_:
		res = p.fetchMissedHeartbeatsResponse(msg)
		return

	case *messages.ControllerMessage_FileCorruptionResponse_:

		res = p.fetchFileCorruptionResponse(msg)
		return

	case *messages.ControllerMessage_ReplicationRequest_:

		res = p.fetchReplicationRequest(msg)
		return

	}

	return
}

func (p *ProtoHandler) HandleStorageNodeRequest(wrapper *messages.StorageNodeMessage, interval int) (Req RequestHandler) {

	switch msg := wrapper.StorageNodeMessage.(type) {

	case *messages.StorageNodeMessage_Intro_:
		Req = p.fetchIntroRequest(msg, interval)

	case *messages.StorageNodeMessage_Heartbeat_:
		Req = p.fetchHeartbeatRequest(msg)

	case *messages.StorageNodeMessage_FileCorruption_:
		Req = p.fetchFileCorruptionRequest(msg)

	}

	return
}
