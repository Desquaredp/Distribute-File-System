# Project 1: Distributed File System 



## Building
To build the project, run the following command from the root of the project:

```make build```

This will build the project and place the binaries in the ```src``` directory.

## Running
Since there are many components in the project, each will have its own set of instructions on how to run it.

### Controller
To run the Controller, run the following command from the ```src``` directory:

```make run_controller```

Note: The controller has default ports it listens on and runs on. The Makefile has these ports baked in. If you want to change the ports, you will have to change them in the Makefile or pass them as arguments to the controller.

To run the controller using the binary, run the following command from the ```src``` directory:

```./controllerExec <Storage Nodes facing port port> <Client facing port>```


### Storage Node
(Tentative)

To run the Storage Node, it is recommended to use the script provided in the ```scripts``` directory. To run the script, run the following command from the ```scripts``` directory:

```bash startStorageNodes```


Note: The script is meant to run on orion. If you want to run it on a different machine, you will have to change the script.


### Client
To run the Client, run the following command from the ```src``` directory:

#### To get list of commands:

```./clientExec -h```

#### To populate the DFS configuration file:

```./clientExec --populate-config <PUT or GET> <config file>```

This will create a config file and populate it with the default values. You can then edit the config file to change the values.

#### To load the DFS configuration file:

```./clientExec --load-config <PUT or GET> <config file>```


#### To list all files in DFS:

```./clientExec --list-files <host:port>```


#### To get a list of nodes:

```./clientExec --list-nodes <host:port>```





## Design

The design document can be found in the design directory.

## Components
### Client
The client is responsible for sending a request to the Controller and splitting the files into fragments. It is also responsible for sending the fragments to the storage nodes.
### Controller
The Controller is the entry point of the DFS, and is responsible for indexing the files in the DFS. It also handles the communication between the client and the DFS.
### Storage Node
The Storage Node is responsible for storing the files in the DFS. It is also responsible for sending the files to the nodes that need them. It has to send periodic heartbeats to the Controller to let it know that it is still alive.
It communicates with the Client and sends and receives files. Additionally, it handles corruption checks, and transfers files to other nodes.
