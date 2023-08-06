package storage_node

type NetworkInterfaces struct {
	NodeInterface       NodeInterface       `yaml:"node_interface"`
	ControllerInterface ControllerInterface `yaml:"controller_interface"`
}

type NodeInterface struct {
	Host                string `yaml:"host"`
	ControllerCommsPort string `yaml:"controller_comms_port"`
	ClientCommsPort     string `yaml:"client_comms_port"`
}

type ControllerInterface struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}
