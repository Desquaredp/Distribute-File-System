package main

import (
	"fmt"
	"go.uber.org/zap"
	"net"
	"os"
	"src/file"
	messages "src/messages/controller_client"
	proto3 "src/proto/controller_client"
	"strconv"
)

type Client struct {
	serverPort string
	file       *file.FileHandler
	logger     *zap.Logger
	conn       net.Conn
	proto      *proto3.ProtoHandler
	msgHandler *messages.MessageHandler
}

type File struct {
	name      string
	size      int64
	chunkSize int64
	checkSum  [32]byte
}

func NewClient(serverPort string, logger *zap.Logger) (client *Client) {

	client = &Client{
		serverPort: serverPort,
		logger:     logger,
	}

	return
}

func (c *Client) SetFileHandler(handler *file.FileHandler) {
	c.file = handler

}
func (c *Client) Disconnect() {
	c.conn.Close()
}
func (c *Client) Conn(conn net.Conn) {
	c.conn = conn
}

func (c *Client) MsgHandler(handler *messages.MessageHandler) {
	c.msgHandler = handler
}

func (c *Client) Proto(proto *proto3.ProtoHandler) {
	c.proto = proto
}

func (c *Client) Dial() (err error) {
	server := c.serverPort

	conn, err := net.Dial("tcp", server)

	if err != nil {
		c.logger.Fatal("There was an error connecting to the host.")
		panic("There was an error connecting to the host.")
		return
	}
	msgHandler := messages.NewMessageHandler(conn)
	c.MsgHandler(msgHandler)
	proto := proto3.NewProtoHandler(msgHandler, c.logger)
	c.Proto(proto)
	c.Conn(conn)

	return
}

func (c *Client) HandleConnection() (err error) {
	for {
		wrapper, _ := c.proto.MsgHandler().ControllerResponseReceive()

		switch wrapper.ControllerMessage.(type) {
		default:
			res := c.proto.HandleControllerResponse(wrapper)

			switch res.GetResType() {
			case "PlanResponse":
				if res.(*proto3.PlanResponse).StatusCode == "OK" {
					c.DispatchFile(res)
				} else {
					fmt.Println("Error: ", res.(*proto3.PlanResponse).StatusCode)
				}

			case "FragmentLayoutResponse":
				if res.(*proto3.FragLayoutResponse).StatusCode == "OK" {
					c.FetchFile(res)
				} else {
					fmt.Println("Error: ", res.(*proto3.FragLayoutResponse).StatusCode)
				}

			case "LsResponse":
				if res.(*proto3.LsResponse).StatusCode == "OK" {
					c.PrintFiles(res)
					return
				} else {
					fmt.Println("Error: ", res.(*proto3.LsResponse).StatusCode)
				}

			case "NodeStats":
				if res.(*proto3.NodeStats).StatusCode == "OK" {
					c.PrintNodeStats(res)
					return
				} else {
					fmt.Println("Error: ", res.(*proto3.NodeStats).StatusCode)
				}
			}
		case nil:
			return

		}
	}
}

func (c *Client) HandlePUT(file *file.FileHandler, fragSize int64) (err error) {

	c.file = file
	c.proto.HandlePutRequest(file.FileName(), file.FileSize(), fragSize)
	return
}

func (c *Client) HandleGET(file string) {

	c.logger.Info("Handling GET request")
	c.proto.HandleGetRequest(file)

}

func (c *Client) PrintFiles(res proto3.ResponseInterface) {

	c.logger.Info("Files present in the DFS:")
	for _, file := range res.(*proto3.LsResponse).Files {

		c.logger.Info(file)

	}

	os.Exit(0)

}

func (c *Client) HandleListFiles() {

	c.proto.HandleLsRequest()

}

func (c *Client) PrintNodeStats(res proto3.ResponseInterface) {

	c.logger.Info("Node Stats:")
	for _, node := range res.(*proto3.NodeStats).Nodes {

		c.logger.Info("Node Id:" + node.NodeId)
		c.logger.Info("Free space:" + strconv.FormatInt(node.DiskSpace, 10))

		fmt.Println()
		fmt.Println()
	}

	os.Exit(0)

}

func (c *Client) HandleNodeStats() {

	c.proto.HandleNodeStatsRequest()

}
