package controller_client

import messages "src/messages/controller_client"

func (p *ProtoHandler) HandlePutRequest(fileName string, fileSize int64, chunkSize int64) {
	res := &messages.ClientMessage{
		ClientMessage: &messages.ClientMessage_PutRequest_{
			PutRequest: &messages.ClientMessage_PutRequest{
				RestOption:        messages.ClientMessage_PUT,
				Filename:          fileName,
				Filesize:          fileSize,
				OptionalChunkSize: chunkSize,
			},
		},
	}
	p.msgHandler.ClientRequestSend(res)

}

type StorageNodeInfo struct {
	NodeId string
	Host   string
	Port   string
}

type FragmentInfo struct {
	FragmentId   string
	Size         int64
	StorageNodes []StorageNodeInfo
}

type PlanResponse struct {
	ResponseType      string
	StatusCode        string
	TotalNumFragments uint32
	FragmentLayout    []FragmentInfo
}

type FragLayoutResponse struct {
	ResponseType      string
	StatusCode        string
	TotalNumFragments uint32
	FragmentLayout    []FragmentInfo
}

func (pr *FragLayoutResponse) GetResType() string {
	return pr.ResponseType
}

func (pr *PlanResponse) GetResType() string {
	return pr.ResponseType
}

type ResponseInterface interface {
	GetResType() string
}

func (p *ProtoHandler) fetchPlanResponse(msg *messages.ControllerMessage_PlanResponse_) (res ResponseInterface) {

	//log the plan response
	p.logger.Info("Received plan response from the Controller.")
	p.logger.Sugar().Info("Status code: ", msg.PlanResponse.StatusCode.String())
	p.logger.Sugar().Info("Num of frags: ", msg.PlanResponse.TotalNumFragments)

	if msg.PlanResponse.StatusCode != messages.ControllerMessage_OK {
		res = &PlanResponse{
			ResponseType: "PlanResponse",
			StatusCode:   msg.PlanResponse.StatusCode.String(),
		}
		return
	}

	res = &PlanResponse{
		ResponseType:      "PlanResponse",
		StatusCode:        msg.PlanResponse.StatusCode.String(),
		TotalNumFragments: msg.PlanResponse.TotalNumFragments,
		FragmentLayout:    make([]FragmentInfo, 0),
	}

	i := 0
	for _, frag := range msg.PlanResponse.FragmentLayout {

		p.logger.Sugar().Info("FragID: ", frag.FragmentId)
		p.logger.Sugar().Info("Frag Size: ", frag.Size)
		res.(*PlanResponse).FragmentLayout = append(res.(*PlanResponse).FragmentLayout, FragmentInfo{
			FragmentId:   frag.FragmentId,
			Size:         frag.Size,
			StorageNodes: make([]StorageNodeInfo, 0),
		})

		for _, node := range frag.StorageNodeIds {

			res.(*PlanResponse).FragmentLayout[i].StorageNodes = append(res.(*PlanResponse).FragmentLayout[i].StorageNodes, StorageNodeInfo{
				NodeId: node.StorageNodeId,
				Host:   node.Host,
				Port:   node.Port,
			})
		}
		i++
	}

	return
}

func (p *ProtoHandler) fetchLayoutResponse(msg *messages.ControllerMessage_FragLayoutResponse_) (res ResponseInterface) {
	p.logger.Info("Received fragment layout response from the Controller.")
	p.logger.Sugar().Info("Status code: ", msg.FragLayoutResponse.StatusCode.String())
	p.logger.Sugar().Info("Num of frags: ", msg.FragLayoutResponse.TotalNumFragments)

	if msg.FragLayoutResponse.StatusCode != messages.ControllerMessage_OK {
		res = &FragLayoutResponse{
			ResponseType: "FragmentLayoutResponse",
			StatusCode:   msg.FragLayoutResponse.StatusCode.String(),
		}
		return
	}

	res = &FragLayoutResponse{
		ResponseType:      "FragmentLayoutResponse",
		StatusCode:        msg.FragLayoutResponse.StatusCode.String(),
		TotalNumFragments: msg.FragLayoutResponse.TotalNumFragments,
		FragmentLayout:    make([]FragmentInfo, 0),
	}

	i := 0
	for _, frag := range msg.FragLayoutResponse.FragmentLayout {

		p.logger.Sugar().Info("FragID: ", frag.FragmentId)
		p.logger.Sugar().Info("Frag Size: ", frag.Size)
		res.(*FragLayoutResponse).FragmentLayout = append(res.(*FragLayoutResponse).FragmentLayout, FragmentInfo{
			FragmentId:   frag.FragmentId,
			Size:         frag.Size,
			StorageNodes: make([]StorageNodeInfo, 0),
		})

		for _, node := range frag.StorageNodeIds {

			res.(*FragLayoutResponse).FragmentLayout[i].StorageNodes = append(res.(*FragLayoutResponse).FragmentLayout[i].StorageNodes, StorageNodeInfo{
				NodeId: node.StorageNodeId,
				Host:   node.Host,
				Port:   node.Port,
			})
		}
		i++
	}

	return

}

func (p *ProtoHandler) fetchDeleteResponse(msg *messages.ControllerMessage_DeleteResponse_) {

}

type LsResponse struct {
	ResponseType string
	StatusCode   string
	Files        []string
}

func (pr *LsResponse) GetResType() string {
	return pr.ResponseType
}
func (p *ProtoHandler) fetchLsResponse(msg *messages.ControllerMessage_LsResponse_) (res ResponseInterface) {

	p.logger.Info("Received ls response from the Controller.")
	p.logger.Sugar().Info("Status code: ", msg.LsResponse.StatusCode.String())
	p.logger.Sugar().Info("Num of files: ", len(msg.LsResponse.FileNames))

	if msg.LsResponse.StatusCode != messages.ControllerMessage_OK {
		res = &LsResponse{
			ResponseType: "LsResponse",
			StatusCode:   msg.LsResponse.StatusCode.String(),
		}
		return
	}

	res = &LsResponse{
		ResponseType: "LsResponse",
		StatusCode:   msg.LsResponse.StatusCode.String(),
		Files:        make([]string, 0),
	}

	for _, file := range msg.LsResponse.FileNames {
		res.(*LsResponse).Files = append(res.(*LsResponse).Files, file)
	}

	return

}

type NodeInfo struct {
	NodeId    string
	DiskSpace int64
}
type NodeStats struct {
	ResponseType string
	StatusCode   string
	Nodes        []NodeInfo
}

func (pr *NodeStats) GetResType() string {
	return pr.ResponseType

}

func (p *ProtoHandler) fetchNodeStats(msg *messages.ControllerMessage_NodeStats_) (res ResponseInterface) {

	p.logger.Info("Fetching NodeStats from Controller.")

	if msg.NodeStats.StatusCode == messages.ControllerMessage_OK {

		p.logger.Sugar().Info("Status code: ", msg.NodeStats.StatusCode.String())
		p.logger.Sugar().Info("Num of nodes: ", len(msg.NodeStats.ActiveNodes))

		res = &NodeStats{
			ResponseType: "NodeStats",
			StatusCode:   msg.NodeStats.StatusCode.String(),
			Nodes:        make([]NodeInfo, 0),
		}

		for _, node := range msg.NodeStats.ActiveNodes {
			res.(*NodeStats).Nodes = append(res.(*NodeStats).Nodes, NodeInfo{
				NodeId:    node.NodeId,
				DiskSpace: node.DiskSpace,
			})
		}
	} else {
		p.logger.Sugar().Info("Status code: ", msg.NodeStats.StatusCode.String())
		res = &NodeStats{
			ResponseType: "NodeStats",
			StatusCode:   msg.NodeStats.StatusCode.String(),
		}
	}

	return res

}
