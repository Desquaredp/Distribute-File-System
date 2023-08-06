package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	FileHandler "src/file"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	//	FileHandler "src/file"
)

func main() {

	inputType, err := parseArgs(os.Args)
	if err != nil {
		fmt.Println(err)
		return
	}

	logger := appendLogger()
	defer logger.Sync()

	switch inputType.(type) {
	case *inputPUTYaml:

		fmt.Println("PUT")
		putInput := inputType.(*inputPUTYaml)
		chunkSize := putInput.ChunkSize
		_ = putInput.FileDir

		addr := putInput.Controller.Host + ":" + putInput.Controller.Port
		client := NewClient(addr, logger)
		client.Dial()

		fileName := putInput.InputFile
		fileHandler := FileHandler.NewFileHandler(fileName)
		fileHandler.SetDir(putInput.FileDir)
		fileHandler.CalcFileSize()
		if err != nil {
			logger.Error("Error reading file", zap.Error(err))
			return
		}
		client.HandlePUT(fileHandler, chunkSize)
		client.HandleConnection()

	case *inputGETYaml:
		fmt.Println("GET")
		getInput := inputType.(*inputGETYaml)
		_ = getInput.FileDir

		addr := getInput.Controller.Host + ":" + getInput.Controller.Port
		client := NewClient(addr, logger)
		client.Dial()
		fileHandler := FileHandler.NewFileHandler(getInput.InputFile)
		fileHandler.SetDir(getInput.FileDir)
		fileName := getInput.InputFile
		client.SetFileHandler(fileHandler)
		client.HandleGET(fileName)
		client.HandleConnection()

	case *inputListFilesYaml:
		fmt.Println("List Files")
		listFilesInput := inputType.(*inputListFilesYaml)

		addr := listFilesInput.Controller.Host + ":" + listFilesInput.Controller.Port
		client := NewClient(addr, logger)
		client.Dial()
		client.HandleListFiles()
		client.HandleConnection()

	case *inputNodeStatsYaml:
		fmt.Println("Node Stats")
		nodeStatsInput := inputType.(*inputNodeStatsYaml)

		addr := nodeStatsInput.Controller.Host + ":" + nodeStatsInput.Controller.Port
		client := NewClient(addr, logger)
		client.Dial()
		client.HandleNodeStats()
		client.HandleConnection()

	case nil:
		fmt.Println("No input type specified")
		return
	}

	select {}

}

type InputInterface interface {
	Type() string
}
type Address struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Job struct {
	InputFile    string `yaml:"input_file"`
	OutputFile   string `yaml:"output_file"`
	ReducerCount int32  `yaml:"reducer_count"`
	JobBinary    string `yaml:"job_binary"`
}

type inputGETYaml struct {
	Controller Address `yaml:"controller"`
	InputFile  string  `yaml:"input_file"`
	FileDir    string  `yaml:"file_dir"`
}

func (i *inputGETYaml) Type() string {
	return "mr"
}

type inputPUTYaml struct {
	Controller Address `yaml:"controller"`
	InputFile  string  `yaml:"input_file"`
	FileDir    string  `yaml:"file_dir"`
	ChunkSize  int64   `yaml:"chunk_size"`
}

func (i *inputPUTYaml) Type() string {
	return "dfs"
}

type inputListFilesYaml struct {
	Controller Address `yaml:"controller"`
}

func (i *inputListFilesYaml) Type() string {
	return "list_files"
}

type inputNodeStatsYaml struct {
	Controller Address `yaml:"controller"`
}

func (i *inputNodeStatsYaml) Type() string {
	return "node_stats"
}

func parseArgs(args []string) (inputType InputInterface, err error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("not enough arguments:\n use -h for help")
	}

	flag := args[1]

	switch flag {
	case "-h":

		fmt.Println("For help:")
		fmt.Println("./clientExec -h")

		fmt.Println("To load config file:")
		fmt.Println("./clientExec --load-config <PUT or GET> <config file>")

		fmt.Println("To populate config file with template values:")
		fmt.Println("./clientExec --populate-config <PUT or GET> <config file>")

		fmt.Println("To list all files in DFS:")
		fmt.Println("./clientExec --list-files <host:port>")

		fmt.Println("To get a list of nodes:")
		fmt.Println("./clientExec --list-nodes")

		os.Exit(0)

	case "--load-config":

		if len(args) < 4 {
			fmt.Println("not enough arguments")

			err = fmt.Errorf("not enough arguments:\n use --load-config <PUT or GET> <config file>")
			return
		}

		configType := args[2]
		configFile := args[3]

		if configType == "PUT" {
			//TODO: populate dfs config

			var readData inputPUTYaml
			readFile, err := os.Open(configFile)
			if err != nil {
				fmt.Printf("Error opening file: %s", err)
				return inputType, err
			}

			defer readFile.Close()
			decoder := yaml.NewDecoder(readFile)
			err = decoder.Decode(&readData)
			if err != nil {
				fmt.Printf("Error decoding YAML: %s", err)
				return inputType, err
			}

			fmt.Printf("Read data: %#v\n", readData)

			inputType = &readData
			return inputType, err

		} else if configType == "GET" {

			var readData inputGETYaml
			readFile, err := os.Open(configFile)
			if err != nil {
				fmt.Printf("Error opening file: %s", err)
				return inputType, err
			}
			defer readFile.Close()
			decoder := yaml.NewDecoder(readFile)
			err = decoder.Decode(&readData)
			if err != nil {
				fmt.Printf("Error decoding YAML: %s", err)
				return inputType, err
			}

			fmt.Printf("Read data: %#v\n", readData)
			inputType = &readData
			return inputType, err

		} else {
			fmt.Println("invalid config type")
			err = fmt.Errorf("invalid config type:\n use --load-config <PUT or GET> <config file>")
			return
		}

	case "--populate-config":
		if len(args) < 4 {
			fmt.Println("not enough arguments")

			err = fmt.Errorf("not enough arguments:\n use --populate-config <PUT or GET> <config file>")
			return
		}
		configType := args[2]
		configFile := args[3]

		if configType == "PUT" {
			//TODO: populate dfs config

			data := inputPUTYaml{
				Controller: Address{
					Host: "localhost",
					Port: "8080",
				},
				InputFile: "inputFile",
				ChunkSize: 128000000,
				FileDir:   "/path/to/file/dir",
			}

			file, err := os.Create(configFile)
			if err != nil {
				fmt.Printf("Error creating file: %s", err)
				return inputType, err
			}
			defer file.Close()
			encoder := yaml.NewEncoder(file)
			err = encoder.Encode(data)
			if err != nil {
				fmt.Printf("Error encoding YAML: %s", err)
				return inputType, err
			}

			fmt.Println("Config file populated successfully. Please edit the config file and provide the correct values.")
			os.Exit(0)

		} else if configType == "GET" {

			//TODO: populate dfs config

			data := inputGETYaml{
				Controller: Address{
					Host: "localhost",
					Port: "8080",
				},
				InputFile: "inputFile",
				FileDir:   "/path/to/file/dir",
			}

			file, err := os.Create(configFile)
			if err != nil {
				fmt.Printf("Error creating file: %s", err)
				return inputType, err
			}
			defer file.Close()
			encoder := yaml.NewEncoder(file)
			err = encoder.Encode(data)
			if err != nil {
				fmt.Printf("Error encoding YAML: %s", err)
				return inputType, err
			}

			fmt.Println("Config file populated successfully. Please edit the config file and provide the correct values.")
			os.Exit(0)

		} else {

			err = fmt.Errorf("invalid config type:\n use --populate-config <dfs or mr> <config file>")
			return
		}

	case "--list-files":

		if len(args) < 3 {

			err = fmt.Errorf("not enough arguments:\n use --list-files <host:port>")
			return
		}

		hostPort := args[2]

		//split host and port
		hostPortSplit := strings.Split(hostPort, ":")
		if len(hostPortSplit) != 2 {
			err = fmt.Errorf("invalid host:port format")
			return
		}

		data := inputListFilesYaml{
			Controller: Address{
				Host: hostPortSplit[0],
				Port: hostPortSplit[1],
			},
		}

		inputType = &data

	case "--list-nodes":

		if len(args) < 3 {

			err = fmt.Errorf("not enough arguments:\n use --list-nodes <host:port>")
			return
		}

		hostPort := args[2]

		//split host and port
		hostPortSplit := strings.Split(hostPort, ":")
		if len(hostPortSplit) != 2 {
			err = fmt.Errorf("invalid host:port format")
			return
		}

		data := inputNodeStatsYaml{
			Controller: Address{
				Host: hostPortSplit[0],
				Port: hostPortSplit[1],
			},
		}

		inputType = &data
	}

	return
}

func appendLogger() *zap.Logger {

	file, err := os.OpenFile("logfileC.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
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
