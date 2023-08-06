package main

import (
	"go.uber.org/zap"
	"net"
	"src/controller/file_distributor"
	"src/controller/storage_handler"
	clientMessages "src/messages/controller_client"
	clientProto3 "src/proto/controller_client"
)

func acceptClientConnections(listener2 net.Listener, spokeHandler *storage_handler.StorageNodeHandler, logger *zap.Logger) {

	for {
		if conn, err := listener2.Accept(); err == nil {

			msgHandler := clientMessages.NewMessageHandler(conn)
			go handleClient(msgHandler, spokeHandler, logger)

		} else {

			logger.Error(err.Error())
		}
	}
}

func handleClient(msgHandler *clientMessages.MessageHandler, spokeHandler *storage_handler.StorageNodeHandler, logger *zap.Logger) {
	defer msgHandler.Close()
	proto := clientProto3.NewProtoHandler(msgHandler, logger)

	for {
		wrapper, _ := proto.MsgHandler().ClientRequestReceive()

		switch wrapper.ClientMessage.(type) {

		default:
			//TODO: handle this
			req := proto.HandleClientRequest(wrapper)
			switch req.GetReqType() {
			case "PUT":
				logger.Info("Processing PUT request")

				var fragMap map[*file_distributor.Fragment][]*storage_handler.Node

				//TODO: FindFiles might be a little slow here. Find a better way to do this
				FileMap := spokeHandler.FindFiles(req.GetFileName(), logger)
				if FileMap != nil {
					logger.Info("File exists.")
					fragMap = nil
				} else {
					logger.Info("File doesn't Exist.")
					var fileDistributor file_distributor.FileDistributorInterface
					fileDistributor = file_distributor.NewFileDistributor(req.GetFileName(), req.GetFileSize(), req.GetChunkSize(), spokeHandler)

					var err error
					fragMap, err = fileDistributor.DistributeFile()
					if err != nil {
						logger.Error(err.Error())
					}
					//spokeHandler.AddFile(req.GetFileName())
				}

				proto.HandlePlanResponse(fragMap, req)

			case "GET":
				logger.Info("Processing GET request")

				FileMap := spokeHandler.FindFiles(req.GetFileName(), logger)
				if FileMap == nil {
					logger.Info("File doesn't exists.")
					proto.HandleGetResponse(nil, req)
				} else {
					logger.Info("File exists.")

					logger.Sugar().Info("FileMap length: ", len(FileMap))
					proto.HandleGetResponse(FileMap, req)
				}

			case "DELETE":
			case "LIST":
				logger.Info("Processing LIST request")
				fileFragments := spokeHandler.FindAllFiles(logger)

				fileSet := spokeHandler.ExtractFiles(fileFragments, logger)
				proto.HandleListResponse(fileSet, req)

			case "NODE_INFO":
				logger.Info("Processing NODE_INFO request")
				nodeInfo := spokeHandler.GetNodeInfo()
				extractedMap, ok := nodeInfo.(map[string]storage_handler.Node)
				if !ok {
					logger.Error("Error in type assertion")
				}

				for k, v := range extractedMap {
					logger.Info("Key: " + k)
					logger.Sugar().Info("Value: ", v)
				}

				proto.HandleNodeInfoResponse(extractedMap, req)

			}

			return

		case nil:
			return
		}

	}

}
