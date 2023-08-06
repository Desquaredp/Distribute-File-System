package proto

import (
	messages "src/messages/client_storage"
)

func (p *ProtoHandler) HandleFileGetRequest(fragID string) (err error) {

	p.logger.Info("Handling File Get Request")
	msg := messages.FileGetRequest{FileName: fragID}
	p.sendClientGetRequest(p.msgHandler, &messages.ClientRequest_FileGetRequest{FileGetRequest: &msg})

	return
}

func (p *ProtoHandler) fetchFileGetResponse(msg *messages.ServerResponse_FileGetResponse) (err error) {

	if msg.FileGetResponse.Success {

		p.logger.Info("Success in FileGetResponse")
		p.logger.Info("Fetching FileGetResponse")
		fileHandler := p.FileHandler()
		p.logger.Sugar().Infof("File Name: %s", fileHandler.FileName())

		fileHandler.SetDataStream(msg.FileGetResponse.MessageBody)
		fileHandler.FindAndSetCheckSum()

		if !fileHandler.CompareChecksum(msg.FileGetResponse.Checksum) {
			p.logger.Info("Failed in checking the checksum")
		} else {
			p.logger.Info("Success in checking the checksum")
			fileHandler.WriteFile()
		}

	} else {

		p.logger.Info("Failed in FileGetResponse")
	}
	return
}

func (p *ProtoHandler) HandleFilePutRequest(fragID string) (err error) {

	p.logger.Info("Sending File Put Request")
	FileSize := p.FileHandler().FragmentMap()[fragID].FragSize()
	msg := messages.FilePutRequest{FileName: fragID, FileSize: FileSize}
	p.sendClientRequest(p.msgHandler, &messages.ClientRequest_FilePutRequest{FilePutRequest: &msg})
	return
}

func (p *ProtoHandler) fetchFilePutResponse(msg *messages.ServerResponse_FilePutResponse) (err error) {

	if msg.FilePutResponse.Success {
		p.logger.Info("File Put Request Success")
		p.handleFileDataRequest(msg.FilePutResponse.FileName)
	} else {
		p.logger.Info("File Put Request Failed")
	}
	return
}

// will cause race condition if multi go routines are used
func (p *ProtoHandler) handleFileDataRequest(fragID string) (err error) {
	p.logger.Info("Sending File Data Request")
	fileData := p.FileHandler().FillFragmentData(fragID)
	fileChecksum := p.FileHandler().FragmentMap()[fragID].FragChecksum()
	fileStorageNodes := p.FileHandler().FragmentMap()[fragID].Location()
	msg := messages.FileDataRequest{FileName: fragID, MessageBody: fileData, Checksum: fileChecksum[:], OtherNodes: fileStorageNodes}
	p.sendClientDataRequest(p.msgHandler, &messages.ClientRequest_FileDataRequest{FileDataRequest: &msg})
	return
}

/*
func (p *ProtoHandler) handleFileDataRequest() (err error) {
	//fileSize := p.FileSize()
	//chunkSize := 1024 * 1024                                            // 1MB
	//numChunks := (fileSize + uint64(chunkSize) - 1) / uint64(chunkSize) // Round up
	//p.DataStream()
	//
	//for i := 0; i < int(numChunks); i++ {
	//	start := i * chunkSize
	//	end := (i + 1) * chunkSize
	//	if end > int(fileSize) {
	//		end = int(fileSize)
	//	}
	//
	//	chunk := p.DataStream()[start:end]
	//
	//	msg := messages.FileDataRequest{
	//		MessageBody: chunk,
	//		Checksum:    p.Checksum(),
	//		ChunkNum:    uint32(i),
	//		NumChunk:    uint32(numChunks),
	//	}
	//
	//	p.sendClientDataRequest(p.msgHandler, &messages.ClientRequest_FileDataRequest{FileDataRequest: &msg})
	//}
	//
	return
}

*/

// TODO: figure out a way to use an interface instead of individual methods for a wrapper.
func (p *ProtoHandler) sendClientRequest(msgHandler *messages.MessageHandler, request *messages.ClientRequest_FilePutRequest) {

	wrapper := &messages.ClientRequest{
		Request: request,
	}
	msgHandler.ClientRequestSend(wrapper)
}

func (p *ProtoHandler) sendClientDataRequest(msgHandler *messages.MessageHandler, request *messages.ClientRequest_FileDataRequest) {
	wrapper := &messages.ClientRequest{
		Request: request,
	}
	msgHandler.ClientRequestSend(wrapper)

}

func (p *ProtoHandler) sendClientGetRequest(msgHandler *messages.MessageHandler, request *messages.ClientRequest_FileGetRequest) {
	p.logger.Info("Sending File Get Request")
	wrapper := &messages.ClientRequest{
		Request: request,
	}
	msgHandler.ClientRequestSend(wrapper)
}
