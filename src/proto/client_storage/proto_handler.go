package proto

import (
	"fmt"
	"go.uber.org/zap"
	"src/file"
	messages "src/messages/client_storage"
)

type ProtoHandler struct {
	fileHandler *file.FileHandler
	msgHandler  *messages.MessageHandler
	logger      *zap.Logger
	dir         string
}

func (p *ProtoHandler) FileHandler() *file.FileHandler {
	return p.fileHandler
}

func (p *ProtoHandler) SetFileHandler(fileHandler *file.FileHandler) {
	p.fileHandler = fileHandler
}

func (p *ProtoHandler) SetMsgHandler(msgHandler *messages.MessageHandler) {
	p.msgHandler = msgHandler
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

//func (p *ProtoHandler) FFileHandler() *file.FileHandler {
//	return p.FileHandler
//}

func NewProtoHandler(msgHandler *messages.MessageHandler, logger *zap.Logger, dir string) *ProtoHandler {
	newProtoHandler := &ProtoHandler{
		msgHandler: msgHandler,
		logger:     logger,
		dir:        dir,
	}

	return newProtoHandler
}

func (p *ProtoHandler) HandleResponse(wrapper *messages.ServerResponse) {

	switch msg := wrapper.Response.(type) {

	case *messages.ServerResponse_FilePutResponse:

		p.fetchFilePutResponse(msg)

	case *messages.ServerResponse_FileDataResponse:

		fmt.Println("Received: ", msg.FileDataResponse.Success)
		fmt.Println("Received: ", msg.FileDataResponse.ErrorCode)

	case *messages.ServerResponse_FileGetResponse:
		p.logger.Info("Received FileGetResponse")
		p.fetchFileGetResponse(msg)
		return

	case nil:
		return
	}

}

type Req struct {
	Operation string
	Success   bool
	Result    interface{}
	DataType  string
}

func (p *ProtoHandler) HandleRequest(wrapper *messages.ClientRequest) (req *Req, err error) {
	var result interface{}
	//var dataType string

	switch msg := wrapper.Request.(type) {

	case *messages.ClientRequest_FilePutRequest:
		p.logger.Info("Received PutRequest")
		p.fileHandler = &file.FileHandler{}
		p.fileHandler.SetDir(p.dir)
		err = p.fetchFilePutRequest(msg)
		if err != nil {
			return
		} else {
			p.logger.Info("PutRequest Success")
			req = &Req{
				Operation: "PUT",
				Success:   true,
			}

		}

	case *messages.ClientRequest_FileDataRequest:
		p.logger.Info("Received FileDataRequest")
		result, err = p.fetchFileDataRequest(msg)
		if err != nil {
			p.logger.Error("FileDataRequest Failed", zap.Error(err))
			return
		} else {
			p.logger.Info("FileDataRequest Success")
			req = &Req{
				Operation: "DATA",
				Success:   true,
				Result:    result,
			}
			return
		}

	case *messages.ClientRequest_FileGetRequest:

		p.fileHandler = &file.FileHandler{}
		err = p.fetchFileGetRequest(msg)
		if err != nil {
			p.logger.Error("FileGetRequest Failed", zap.Error(err))
			return
		}
		p.logger.Info("FileGetRequest Success")
		req = &Req{
			Operation: "GET",
			Success:   true,
		}

	}
	return
}
