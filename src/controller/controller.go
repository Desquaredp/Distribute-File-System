package main

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"os"
	"src/controller/storage_handler"
	"strconv"
	"time"
)

const HEARTBEAT_INTERVAL = 5
const ACCEPTED_DELAY = HEARTBEAT_INTERVAL * 3

func initLogger(file *os.File) *zap.Logger {

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

func main() {

	//TODO: create a config file to handle this.
	file, err := os.OpenFile("logfile.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	logger := initLogger(file)

	if len(os.Args) != 3 {
		logger.Error("Command line args not provided.")
		logger.Info("Usage: ./controller <Storage Nodes facing port port> <Client facing port>")
		logger.Info("Usage(2): go run controller/controller.go <Storage Nodes facing port port> <Client facing port>")
		logger.Fatal("Exiting.")
		os.Exit(1)
	}

	port1, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid port number:", os.Args[1])
		os.Exit(1)
	}

	logger.Sugar().Info("Listening on port for Storage node connections: ", port1)

	listener1, err := net.Listen("tcp", ":"+strconv.Itoa(port1))
	if err != nil {
		logger.Error(err.Error())
		return
	}

	port2, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Invalid port number:", os.Args[2])
		os.Exit(1)
	}

	logger.Sugar().Info("Listening on port for Client connections: ", port2)

	listener2, err := net.Listen("tcp", ":"+strconv.Itoa(port2))
	if err != nil {
		logger.Error(err.Error())
		return
	}

	spokeHandler := storage_handler.NewStorageNodeHandler(logger)

	go func() {

		for {
			//wait for a heartbeat interval
			time.Sleep(HEARTBEAT_INTERVAL * time.Second)
			spokeHandler.ConcurrentStaleNodeRemoval(ACCEPTED_DELAY, logger)
		}
	}()

	go func() {

		for {
			//wait for a accepted delay interval
			time.Sleep(HEARTBEAT_INTERVAL * ACCEPTED_DELAY * time.Second)
			spokeHandler.ConcurrentIndexing()

		}
	}()

	go acceptStorageNodeConnections(listener1, spokeHandler, logger)
	go acceptClientConnections(listener2, spokeHandler, logger)

	select {}
}
