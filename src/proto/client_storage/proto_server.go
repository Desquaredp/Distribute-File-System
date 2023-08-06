package proto

import (
	"errors"
	"fmt"
	FileHandler "src/file"
	messages "src/messages/client_storage"
)

func (p *ProtoHandler) fetchFilePutRequest(msg *messages.ClientRequest_FilePutRequest) (err error) {

	fmt.Println("> ",
		msg.FilePutRequest.FileName,
	)

	fileSize := msg.FilePutRequest.FileSize
	fileName := msg.FilePutRequest.FileName
	fileHandler := p.FileHandler()

	fileHandler.SetFileName(fileName)
	fileHandler.SetFileSize(fileSize)

	storageAvailable, _ := p.FileHandler().StorageCheck()
	fileExists, _ := p.fileHandler.FileCheck()

	err = p.handleFilePutResponse(storageAvailable, fileExists, fileName)
	return
}

func (p *ProtoHandler) handleFilePutResponse(storageAvailable bool, fileExists bool, fileName string) (err error) {

	var res messages.FilePutResponse
	if !storageAvailable || fileExists {
		if !storageAvailable {
			res = messages.FilePutResponse{Success: false, ErrorCode: messages.ErrorCode_FILE_SIZE_LIMIT_EXCEEDED, FileName: fileName}
			err = errors.New(string(messages.ErrorCode_FILE_SIZE_LIMIT_EXCEEDED))
		} else if fileExists {
			res = messages.FilePutResponse{Success: false, ErrorCode: messages.ErrorCode_FILE_ALREADY_EXISTS, FileName: fileName}
			err = errors.New(string(messages.ErrorCode_FILE_ALREADY_EXISTS))
		}

	} else {
		res = messages.FilePutResponse{Success: true, ErrorCode: messages.ErrorCode_NO_ERROR, FileName: fileName}
	}
	p.sendServerPutResponse(p.msgHandler, &messages.ServerResponse_FilePutResponse{FilePutResponse: &res})

	return
}

func (p *ProtoHandler) sendServerPutResponse(msgHandler *messages.MessageHandler, response *messages.ServerResponse_FilePutResponse) {

	wrapper := &messages.ServerResponse{
		Response: response,
	}
	err := msgHandler.ServerResponseSend(wrapper)
	if err != nil {
		return
	}

}

func (p *ProtoHandler) fetchFileDataRequest(msg *messages.ClientRequest_FileDataRequest) (res *FileHandler.FileHandler, err error) {

	p.logger.Info("Received FileDataRequest")
	data := msg.FileDataRequest.MessageBody
	fileHandler := p.FileHandler()
	fileHandler.SetDataStream(data)
	p.logger.Sugar().Infof("Received %d bytes", len(data))

	fileHandler.FindAndSetCheckSum()
	checksumCheck := p.FileHandler().CompareChecksum(msg.FileDataRequest.Checksum)
	p.logger.Sugar().Infof("Checksums match: %t", checksumCheck)
	fileHandler.SetLocation(msg.FileDataRequest.OtherNodes)

	err = p.handleFileDataResponse(checksumCheck)
	if err != nil {
		return
	}

	return fileHandler, nil

}

func (p *ProtoHandler) handleFileDataResponse(checksumCheck bool) (err error) {

	var res messages.FileDataResponse
	if !checksumCheck {
		fmt.Println("The two checksums are not equal")
		res = messages.FileDataResponse{Success: false, ErrorCode: messages.ErrorCode_CHECKSUM_MISMATCH}
		err = errors.New(string(messages.ErrorCode_CHECKSUM_MISMATCH))

	} else {
		fmt.Println("The two checksums match")
		//TODO: uncomment this when writing to disk is working
		p.logger.Info("Writing checksum to disk")
		p.FileHandler().SetDir(p.dir)
		p.FileHandler().ChecksumOnDisk()
		p.logger.Sugar().Infof("File data is %d bytes", len(p.FileHandler().DataStream()))
		err = p.FileHandler().WriteFile()
		if err != nil {
			res = messages.FileDataResponse{Success: false, ErrorCode: messages.ErrorCode_SERVER_ERROR}
			err = errors.New(string(messages.ErrorCode_SERVER_ERROR))

		} else {

		}

		res = messages.FileDataResponse{Success: true, ErrorCode: messages.ErrorCode_NO_ERROR}
	}
	p.sendServerDataResponse(p.msgHandler, &messages.ServerResponse_FileDataResponse{FileDataResponse: &res})

	return
}

func (p *ProtoHandler) sendServerDataResponse(msgHandler *messages.MessageHandler, response *messages.ServerResponse_FileDataResponse) (err error) {
	wrapper := &messages.ServerResponse{
		Response: response,
	}
	msgHandler.ServerResponseSend(wrapper)
	return
}

func (p *ProtoHandler) fetchFileGetRequest(msg *messages.ClientRequest_FileGetRequest) (err error) {
	p.FileHandler().SetFileName(msg.FileGetRequest.FileName)
	p.FileHandler().SetDir(p.dir)

	fileExists, _ := p.FileHandler().FileCheck()
	err = p.handleFileGetResponse(fileExists)
	return nil
}

func (p *ProtoHandler) handleFileGetResponse(fileExists bool) (err error) {

	var res messages.FileGetResponse
	if fileExists {
		p.logger.Info("File exists")

		fileHandler := p.FileHandler()
		p.FileHandler().SetDir(p.dir)
		_, errF := fileHandler.ReadFile()
		if errF != nil {
			p.logger.Error("Error reading file")
			p.logger.Error(errF.Error())
			//res = messages.FileGetResponse{Success: false, ErrorCode: messages.ErrorCode_SERVER_ERROR}
			//err = errors.New(string(messages.ErrorCode_SERVER_ERROR))
			//return err
		}
		fileHandler.FindAndSetCheckSum()

		p.logger.Sugar().Infof("File size: %d", fileHandler.FileSize())
		p.logger.Sugar().Infof("Checksum: %s", fileHandler.Checksum())

		res = messages.FileGetResponse{
			Success:     true,
			FileSize:    fileHandler.FileSize(),
			Checksum:    fileHandler.Checksum(),
			MessageBody: fileHandler.DataStream(),
			ErrorCode:   messages.ErrorCode_NO_ERROR,
		}
		err = errors.New(string(messages.ErrorCode_NO_ERROR))

	} else {

		res = messages.FileGetResponse{Success: false, ErrorCode: messages.ErrorCode_FILE_NOT_FOUND}
		err = errors.New(string(messages.ErrorCode_FILE_NOT_FOUND))
	}

	p.sendFileGetResponse(p.msgHandler, &messages.ServerResponse_FileGetResponse{FileGetResponse: &res})
	return
}

func (p *ProtoHandler) sendFileGetResponse(msgHandler *messages.MessageHandler, response *messages.ServerResponse_FileGetResponse) (err error) {
	wrapper := &messages.ServerResponse{
		Response: response,
	}
	msgHandler.ServerResponseSend(wrapper)
	return
}
