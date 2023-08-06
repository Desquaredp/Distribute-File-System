package storage_storage

import (
	"go.uber.org/zap"
	"src/file"
	messages "src/messages/storage_storage"
)

type ProtoHandler struct {
	dir        string
	msgHandler *messages.MessageHandler
	logger     *zap.Logger
	file       *file.FileHandler
	nodeId     string
}

func (p *ProtoHandler) Logger() *zap.Logger {
	return p.logger
}

func (p *ProtoHandler) SetLogger(logger *zap.Logger) {
	p.logger = logger
}

func (p *ProtoHandler) MsgHandler() *messages.MessageHandler {
	return p.msgHandler
}

func NewProtoHandler(msgHandler *messages.MessageHandler, logger *zap.Logger, dir string) *ProtoHandler {
	newProtoHandler := &ProtoHandler{
		msgHandler: msgHandler,
		logger:     logger,
		dir:        dir,
	}
	return newProtoHandler
}

type Response interface {
	GetType() string
}

func (p *ProtoHandler) HandleStorageNodeResponse(wrapper *messages.StorageNodeMessage) (Res Response, err error) {

	switch msg := wrapper.StorageNodeMessage.(type) {

	case *messages.StorageNodeMessage_GetReplicaResponse:
		p.logger.Info("Got GetReplicaResponse")
		p.fetchGetReplicaResponse(msg)
	}

	return
}

func (p *ProtoHandler) HandleStorageNodeRequest(wrapper *messages.StorageNodeMessage) {

	switch msg := wrapper.StorageNodeMessage.(type) {

	case *messages.StorageNodeMessage_PutCopy:
		p.logger.Info("Handling PUTCopy request")
		p.fetchPutCopyRequest(msg)

	case *messages.StorageNodeMessage_GetReplica:
		p.logger.Info("Handling GetReplica request")
		p.fetchGetReplicaRequest(msg)

	}

	return
}

func (p *ProtoHandler) fetchPutCopyRequest(msg *messages.StorageNodeMessage_PutCopy) *interface{} {

	fileHandler := file.FileHandler{}
	fileHandler.SetDir(p.dir)
	fileHandler.SetFileName(msg.PutCopy.FileName)
	fileHandler.SetDataStream(msg.PutCopy.FileData)
	fileHandler.FindAndSetCheckSum()
	chkSumChk := fileHandler.CompareChecksum(msg.PutCopy.Checksum)
	if !chkSumChk {
		p.logger.Error("Checksums do not match")
		return nil
	}

	p.logger.Info("Checksums match")
	p.logger.Info(" Writing file to disk")

	fileHandler.ChecksumOnDisk()

	err := fileHandler.WriteFile()
	if err != nil {
		p.logger.Error("Error writing file", zap.Error(err))

	}

	return nil
}

func (p *ProtoHandler) fetchGetReplicaRequest(msg *messages.StorageNodeMessage_GetReplica) *interface{} {

	fileName := msg.GetReplica.GetFileName()

	fileHandler := file.FileHandler{}
	fileHandler.SetDir(p.dir)
	fileHandler.SetFileName(fileName)
	_, err := fileHandler.ReadFile()
	if err != nil {
		p.logger.Error("Error reading file", zap.Error(err))
		return nil
	}
	p.HandleGetReplicaResponse(&fileHandler)

	return nil
}

func (p *ProtoHandler) fetchGetReplicaResponse(msg *messages.StorageNodeMessage_GetReplicaResponse) *interface{} {
	p.logger.Info("Got replica response")
	fileHandler := file.FileHandler{}
	fileHandler.SetDir(p.dir)
	fileHandler.SetFileName(msg.GetReplicaResponse.FileName)
	fileHandler.SetDataStream(msg.GetReplicaResponse.FileData)
	fileHandler.FindAndSetCheckSum()
	p.file = &fileHandler

	p.logger.Info("Checksums match")
	p.logger.Info(" Writing file to disk")
	fileHandler.ChecksumOnDisk()
	err := fileHandler.WriteFile()
	if err != nil {
		p.logger.Error("Error writing file", zap.Error(err))

	}
	p.logger.Info("File written to disk")

	return nil

}

func (p *ProtoHandler) HandlePUTCopyRequest(id string, handler *file.FileHandler, node string) {

	req := &messages.StorageNodeMessage_PUTCopy{
		FileName: handler.FileName(),
		Checksum: handler.Checksum(),
		FileData: handler.DataStream(),
	}

	wrapper := &messages.StorageNodeMessage{
		StorageNodeMessage: &messages.StorageNodeMessage_PutCopy{
			PutCopy: req,
		},
	}

	p.MsgHandler().ClientRequestSend(wrapper)
}

func (p *ProtoHandler) HandleGetReplicationRequest(fileName string) {
	p.logger.Info("Sending Replication request")

	req := &messages.StorageNodeMessage_GETReplica{

		FileName: fileName,
	}

	wrapper := &messages.StorageNodeMessage{
		StorageNodeMessage: &messages.StorageNodeMessage_GetReplica{
			GetReplica: req,
		},
	}
	p.MsgHandler().ClientRequestSend(wrapper)

}

func (p *ProtoHandler) HandleGetReplicaResponse(f *file.FileHandler) {

	p.logger.Info("Sending Replication response")

	res := &messages.StorageNodeMessage_GETReplicaResponse{
		FileName: f.FileName(),
		FileData: f.DataStream(),
		//TODO: add checksum
		//Checksum: f.Checksum(),
	}

	wrapper := &messages.StorageNodeMessage{
		StorageNodeMessage: &messages.StorageNodeMessage_GetReplicaResponse{
			GetReplicaResponse: res,
		},
	}

	p.MsgHandler().ServerResponseSend(wrapper)
}
