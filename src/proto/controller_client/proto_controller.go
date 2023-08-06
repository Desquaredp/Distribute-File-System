package controller_client

import messages "src/messages/controller_client"

type Request struct {
	reqType   string
	fileName  string
	fileSize  int64
	chunkSize int64
}

func (r *Request) GetReqType() string {
	return r.reqType
}

func (r *Request) GetFileName() string {
	return r.fileName
}

func (r *Request) GetFileSize() int64 {
	return r.fileSize
}

func (r *Request) GetChunkSize() int64 {
	return r.chunkSize
}

func (p *ProtoHandler) fetchPutRequest(msg *messages.ClientMessage_PutRequest_) *Request {
	p.logger.Info("Received Put Request")
	putReq := &Request{
		reqType:   msg.PutRequest.RestOption.String(),
		fileName:  msg.PutRequest.Filename,
		fileSize:  int64(msg.PutRequest.Filesize),
		chunkSize: int64(msg.PutRequest.OptionalChunkSize),
	}
	p.logger.Sugar().Info("Request: ", putReq.GetReqType())
	p.logger.Sugar().Info("Request for filename: ", putReq.GetFileName())
	p.logger.Sugar().Info("size: ", putReq.GetFileSize())
	return putReq
}

func (p *ProtoHandler) fetchGetRequest(msg *messages.ClientMessage_GetRequest_) *Request {
	p.logger.Info("Received Get Request")
	getReq := &Request{
		reqType:  msg.GetRequest.RestOption.String(),
		fileName: msg.GetRequest.FileName,
	}
	p.logger.Sugar().Info("Request: ", getReq.GetReqType())
	p.logger.Sugar().Info("Request for filename: ", getReq.GetFileName())
	return getReq
}

func (p *ProtoHandler) fetchDeleteRequest(msg *messages.ClientMessage_DeleteRequest_) {

}

func (p *ProtoHandler) fetchLsRequest(msg *messages.ClientMessage_LsRequest_) *Request {

	p.logger.Info("Received Ls Request")
	lsReq := &Request{
		reqType: "LIST",
	}

	p.logger.Sugar().Info("Request: ", lsReq.GetReqType())

	return lsReq

}
func (p *ProtoHandler) fetchNodeStatsRequest(msg *messages.ClientMessage_NodeStatsRequest_) *Request {

	p.logger.Info("Received NodeStats Request")
	nodeStatsReq := &Request{
		reqType: "NODE_INFO",
	}

	p.logger.Sugar().Info("Request: ", nodeStatsReq.GetReqType())

	return nodeStatsReq

}
