package main

import (
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	messages "src/messages/controller_storage"
	"src/storage_node"
	"time"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println("Usage: go run spawn.go <Directory for the files>")
		fmt.Println("Note the config.yaml file should be in the same directory as the location of the storage node")
		return
	}

	dir := os.Args[1]

	logger := appendLogger(dir)

	//file, err := os.OpenFile("../../../../../../bigdata/students/dmistry4/logfileS.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	yfile, err := ioutil.ReadFile(dir + "config.yaml")
	if err != nil {
		logger.Error("Error reading YAML file: ", zap.Error(err))
		log.Fatal(err)
	}

	// Parse YAML file
	var networkInterfaces storage_node.NetworkInterfaces
	err = yaml.Unmarshal(yfile, &networkInterfaces)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("Config: ", zap.Any("config", networkInterfaces))

	id := uuid.New()
	newStorageNode := storage_node.NewStorageNode(id.String(), networkInterfaces, logger)
	newStorageNode.SetDir(dir)
	proto, conn, err := newStorageNode.Dial()
	if err != nil {
		logger.Error("Error dialing: ", zap.Error(err))
		return
	}
	newStorageNode.ConcurrentChecksumCheck()
	newStorageNode.HandleIntroduction(proto)
	newStorageNode.ConcurrentListen()
	newStorageNode.HandleConnection(proto)
	newStorageNode.Disconnect(conn)
	for {
		timer := time.Duration(newStorageNode.GetInterval()) * time.Second
		time.Sleep(timer / 2)
		proto, conn, err = newStorageNode.Dial()
		if err != nil {
			logger.Error("Error dialing: ", zap.Error(err))
			//TODO: Add a retry mechanism or return, continue for now
			continue
		}

		time.Sleep(timer / 2)

		status := newStorageNode.GetNodeStatus()
		if status == messages.StorageNodeMessage_ACTIVE {
			newStorageNode.HandleHeartbeats(proto)
		} else {
			newStorageNode.HandleIntroduction(proto)
		}

		newStorageNode.HandleConnection(proto)
		newStorageNode.Disconnect(conn)
	}

	select {}

}
func appendLogger(dir string) *zap.Logger {

	file, err := os.OpenFile(dir+"logfileS.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// Create a logger that writes to the file
	fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	fileWriter := zapcore.AddSync(file)
	fileLevel := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, fileLevel)

	// Create a logger that writes to the console
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	consoleWriter := zapcore.Lock(os.Stdout)
	consoleLevel := zap.NewAtomicLevelAt(zapcore.DebugLevel)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, consoleLevel)

	// Create a final logger that writes to both the file and console
	logger := zap.New(zapcore.NewTee(fileCore, consoleCore))
	return logger
}
