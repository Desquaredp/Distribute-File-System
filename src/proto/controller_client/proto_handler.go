package controller_client

import (
	"go.uber.org/zap"
	"src/controller/file_distributor"
	"src/controller/storage_handler"
	messages "src/messages/controller_client"
)

type ProtoHandler struct {
	msgHandler *messages.MessageHandler
	logger     *zap.Logger
	nodeId     string
}

func (p *ProtoHandler) MsgHandler() *messages.MessageHandler {
	return p.msgHandler
}

func NewProtoHandler(msgHandler *messages.MessageHandler, logger *zap.Logger) *ProtoHandler {
	newProtoHandler := &ProtoHandler{
		msgHandler: msgHandler,
		logger:     logger,
	}
	return newProtoHandler
}

func (p *ProtoHandler) HandleControllerResponse(wrapper *messages.ControllerMessage) (res ResponseInterface) {

	switch msg := wrapper.ControllerMessage.(type) {

	case *messages.ControllerMessage_PlanResponse_:
		p.logger.Info("Received plan response from the Controller.")
		res = p.fetchPlanResponse(msg)
		return
	case *messages.ControllerMessage_FragLayoutResponse_:
		res = p.fetchLayoutResponse(msg)
	case *messages.ControllerMessage_DeleteResponse_:
		p.fetchDeleteResponse(msg)
	case *messages.ControllerMessage_LsResponse_:
		res = p.fetchLsResponse(msg)

	case *messages.ControllerMessage_NodeStats_:
		res = p.fetchNodeStats(msg)
	}

	return
}

func (p *ProtoHandler) HandleClientRequest(wrapper *messages.ClientMessage) (req *Request) {

	switch msg := wrapper.ClientMessage.(type) {

	case *messages.ClientMessage_PutRequest_:
		req = p.fetchPutRequest(msg)

	case *messages.ClientMessage_GetRequest_:
		req = p.fetchGetRequest(msg)

	case *messages.ClientMessage_DeleteRequest_:
		p.fetchDeleteRequest(msg)

	case *messages.ClientMessage_LsRequest_:
		req = p.fetchLsRequest(msg)

	case *messages.ClientMessage_NodeStatsRequest_:
		req = p.fetchNodeStatsRequest(msg)

	}

	return
}

func (p *ProtoHandler) HandlePlanResponse(fragMap map[*file_distributor.Fragment][]*storage_handler.Node, req *Request) {

	p.logger.Info("Sending plan response to the Controller.")

	if len(fragMap) == 0 {
		p.logger.Info("No fragments to send.")

		res := &messages.ControllerMessage_PlanResponse_{
			PlanResponse: &messages.ControllerMessage_PlanResponse{
				StatusCode: messages.ControllerMessage_FILE_ALREADY_EXISTS,
			},
		}

		wrapper := &messages.ControllerMessage{
			ControllerMessage: res,
		}

		p.msgHandler.ControllerResponseSend(wrapper)
		return
	}
	res := &messages.ControllerMessage_PlanResponse_{
		PlanResponse: &messages.ControllerMessage_PlanResponse{
			StatusCode:        messages.ControllerMessage_OK,
			TotalNumFragments: uint32(len(fragMap)),
			//repeated fragments
			FragmentLayout: []*messages.ControllerMessage_PlanResponse_FragmentInfo{},
		},
	}

	for frag, nodes := range fragMap {
		fragInfo := &messages.ControllerMessage_PlanResponse_FragmentInfo{
			FragmentId:     frag.GetFragmentName(),
			Size:           frag.GetFragmentSize(),
			StorageNodeIds: []*messages.ControllerMessage_PlanResponse_StorageNodeInfo{},
		}

		for _, node := range nodes {
			nodeInfo := &messages.ControllerMessage_PlanResponse_StorageNodeInfo{
				StorageNodeId: node.GetID(),
				Host:          node.GetAddress(),
				Port:          node.GetOpenPort(),
			}
			fragInfo.StorageNodeIds = append(fragInfo.StorageNodeIds, nodeInfo)
		}

		res.PlanResponse.FragmentLayout = append(res.PlanResponse.FragmentLayout, fragInfo)
	}

	wrapper := &messages.ControllerMessage{
		ControllerMessage: res,
	}

	p.msgHandler.ControllerResponseSend(wrapper)

}

func (p *ProtoHandler) HandleGetRequest(file string) {

	p.logger.Info("Sending Get request to the Controller.")

	req := &messages.ClientMessage_GetRequest_{
		GetRequest: &messages.ClientMessage_GetRequest{
			RestOption: messages.ClientMessage_GET,
			FileName:   file,
		},
	}

	wrapper := &messages.ClientMessage{
		ClientMessage: req,
	}

	p.msgHandler.ClientRequestSend(wrapper)

}

func (p *ProtoHandler) HandleNodeInfoResponse(extractedMap map[string]storage_handler.Node, req *Request) {

	p.logger.Info("Handling NodeInfo response to send.")
	var res *messages.ControllerMessage_NodeStats_

	if extractedMap == nil {

		res = &messages.ControllerMessage_NodeStats_{
			NodeStats: &messages.ControllerMessage_NodeStats{
				StatusCode: messages.ControllerMessage_ERROR,
			},
		}

	} else {

		res = &messages.ControllerMessage_NodeStats_{
			NodeStats: &messages.ControllerMessage_NodeStats{
				StatusCode:  messages.ControllerMessage_OK,
				ActiveNodes: []*messages.ControllerMessage_NodeStats_NodeInfo{},
			},
		}

		for _, node := range extractedMap {

			p.logger.Info("NodeInfo response to send.", zap.String("nodeId", node.GetID()))
			p.logger.Info("NodeInfo response to send.", zap.String("node free space", string(node.GetFreeSpace())))
			nodeInfo := &messages.ControllerMessage_NodeStats_NodeInfo{
				NodeId:    node.GetID(),
				DiskSpace: node.GetFreeSpace(),
			}
			res.NodeStats.ActiveNodes = append(res.NodeStats.ActiveNodes, nodeInfo)
		}

	}

	wrapper := &messages.ControllerMessage{
		ControllerMessage: res,
	}

	p.msgHandler.ControllerResponseSend(wrapper)

}
func (p *ProtoHandler) HandleGetResponse(fileMap map[string][]*storage_handler.Node, req *Request) {

	p.logger.Info("Handling Get response to send.")
	var res *messages.ControllerMessage_FragLayoutResponse_

	if fileMap == nil {

		res = &messages.ControllerMessage_FragLayoutResponse_{
			FragLayoutResponse: &messages.ControllerMessage_FragLayoutResponse{
				StatusCode: messages.ControllerMessage_FILE_NOT_FOUND,
			},
		}

	} else {

		res = &messages.ControllerMessage_FragLayoutResponse_{
			FragLayoutResponse: &messages.ControllerMessage_FragLayoutResponse{
				StatusCode:        messages.ControllerMessage_OK,
				TotalNumFragments: uint32(len(fileMap)),
				//repeated fragments
				FragmentLayout: []*messages.ControllerMessage_FragLayoutResponse_FragmentInfo{},
			},
		}

		for frag, nodes := range fileMap {
			fragInfo := &messages.ControllerMessage_FragLayoutResponse_FragmentInfo{
				FragmentId:     frag,
				StorageNodeIds: []*messages.ControllerMessage_FragLayoutResponse_StorageNodeInfo{},
			}

			for _, node := range nodes {
				nodeInfo := &messages.ControllerMessage_FragLayoutResponse_StorageNodeInfo{
					StorageNodeId: node.GetID(),
					Host:          node.GetAddress(),
					Port:          node.GetOpenPort(),
				}
				fragInfo.StorageNodeIds = append(fragInfo.StorageNodeIds, nodeInfo)
			}

			res.FragLayoutResponse.FragmentLayout = append(res.FragLayoutResponse.FragmentLayout, fragInfo)
		}

	}
	wrapper := &messages.ControllerMessage{
		ControllerMessage: res,
	}

	p.msgHandler.ControllerResponseSend(wrapper)

}

func (p *ProtoHandler) HandleListResponse(set map[string]bool, req *Request) {

	p.logger.Info("Handling List response to send.")
	var res *messages.ControllerMessage_LsResponse_

	if set == nil {

		res = &messages.ControllerMessage_LsResponse_{
			LsResponse: &messages.ControllerMessage_LsResponse{
				StatusCode: messages.ControllerMessage_FILE_NOT_FOUND,
			},
		}

	} else {

		res = &messages.ControllerMessage_LsResponse_{
			LsResponse: &messages.ControllerMessage_LsResponse{
				StatusCode: messages.ControllerMessage_OK,
				//repeated fragments
				FileNames: []string{},
			},
		}

		for file := range set {
			res.LsResponse.FileNames = append(res.LsResponse.FileNames, file)
		}

	}
	wrapper := &messages.ControllerMessage{
		ControllerMessage: res,
	}

	p.msgHandler.ControllerResponseSend(wrapper)

}

func (p *ProtoHandler) HandleLsRequest() {

	p.logger.Info("Sending Ls request to the Controller.")

	req := &messages.ClientMessage_LsRequest_{
		LsRequest: &messages.ClientMessage_LsRequest{
			RestOption: messages.ClientMessage_LS,
		},
	}

	wrapper := &messages.ClientMessage{
		ClientMessage: req,
	}

	p.msgHandler.ClientRequestSend(wrapper)

}

func (p *ProtoHandler) HandleNodeStatsRequest() {

	p.logger.Info("Sending NodeStats request to the Controller.")

	req := &messages.ClientMessage_NodeStatsRequest_{
		NodeStatsRequest: &messages.ClientMessage_NodeStatsRequest{
			RestOption: messages.ClientMessage_NODE_STATS,
		},
	}

	wrapper := &messages.ClientMessage{
		ClientMessage: req,
	}

	p.msgHandler.ClientRequestSend(wrapper)

}
