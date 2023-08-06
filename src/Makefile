# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Main program paths
CONTROLLER_SRC=controller/controller.go controller/client_conn.go controller/storage_conn.go
CONTROLLER_BIN=controllerExec

CLIENT_SRC=client/client_main.go client/dispatch.go client/client.go client/fetch.go
CLIENT_BIN=clientExec

SPAWN_SRC=spawn/spawn.go
SPAWN_BIN=spawnExec




all: clean build

build:
	$(GOBUILD) -o $(CONTROLLER_BIN) $(CONTROLLER_SRC)
	$(GOBUILD) -o $(CLIENT_BIN) $(CLIENT_SRC)
	$(GOBUILD) -o $(SPAWN_BIN) $(SPAWN_SRC)


clean:
	$(GOCLEAN)
	rm -f $(CONTROLLER_BIN) $(CLIENT_BIN) $(SPAWN_BIN)

run_controller:
	$(GOBUILD) -o $(CONTROLLER_BIN) $(CONTROLLER_SRC)
	./$(CONTROLLER_BIN) 23000 23003

run_client:
	$(GOBUILD) -o $(CLIENT_BIN) $(CLIENT_SRC)
	./$(CLIENT_BIN)

run_spawn:
	$(GOBUILD) -o $(SPAWN_BIN) $(SPAWN_SRC)
	./$(SPAWN_BIN)
