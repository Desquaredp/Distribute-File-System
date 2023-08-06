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


## Retrospective Questions
### How many extra days did you use for the project?

### Given the same goals, how would you complete the project differently if you didn’t have any restrictions imposed by the instructor? This could involve using a particular library, programming language, etc. Be sure to provide sufficient detail to justify your response.

- I think using gRPC for a small part of the project would be a great way to learn gRPC.
### Let’s imagine that your next project was to improve and extend P1. What are the features/functionality you would add, use cases you would support, etc? Are there any weaknesses in your current implementation that you would like to improve upon? This should include at least three areas you would improve/extend.

#### Features I would add:
- A computation engine :-)
- Definitely a better user interface
- I would like to change the topology and build a consensus based system. That would be fun.
There's a lot more I would like to add as an extension to this project.


#### Changes I would make to rectify the weaknesses in my current implementation:
##### Slow file distribution from client-storage nodes
I chose to not use go routines for file distribution, and instead chose to use a single thread. This is obviously not ideal, and I would like to change that. Using a controlled number of go routines would be a better way to do this.

##### Better error handling
  There are several instances where I do not handle errors. I would like to handle errors better, and not just print them out to the console.

##### Controller oblivious to the fragments of a file in the DFS
In trying to make the controller stateless, I made it oblivious to the fragments of a file in the DFS. Meaning, if a node goes down, the controller does not know which fragments were on that node. It relies on the nodes to tell it which fragments they have. 
This has an obvious drawback. Say if A and B are the two storage nodes, and A has fragments 1, 2, 3, and B has fragments 4, 5, 6. If A goes down, the controller will not receive any heartbeats from A, and will not know that A has fragments 1, 2, 3. It will only know that B has fragments 4, 5, 6.
If a client wants to GET the file back, the controller will send the client to B, and B will only have fragments 4, 5, 6. The client will not be able to GET the file back.

There are a couple of ways to fix this:
1. Don't make the controller stateless. This is the easiest way to fix this. The controller will have to keep track of the fragments of a file in the DFS. This will make the controller a single point of failure, but it will be able to handle the above scenario.
2. Client can provide the number of fragments it dispatched and the controller checks if there are enough fragments to GET the file back. This is a bit dumb, but it will keep the controller stateless.

##### Better algorithm for load balancing
The current algorithm for load balancing is very simple. Start with the node with the most free space, and choose the other two nodes randomly. A sophisticated algorithm would be to take into account the number of tasks a node is currently executing, and the number of tasks a node has executed in the past.
Plus, factors like network latency, and the distance b/w the nodes should probably play a role in deciding which nodes to choose. This could be a lot of fun to implement.
  
##### Better testing
  One can never have enough tests. I would like to write more tests to test the different components of the system. I would also like to write tests to test the system as a whole. Implementing the 'chaos monkey' would be a great way to test the system as a whole.

##### Code cleanliness
  Simply put, the code is awful. I would like to clean up the code, and make it more readable. I would also like to add more comments to the code.



### Give a rough estimate of how long you spent completing this assignment. Additionally, what part of the assignment took the most time?

I had an error that I couldn't resolve for about 3 weeks! I spent a lot of time trying to figure out what was wrong. I spent a lot of time trying to figure out what was wrong.
I probably spent more time figuring out what was wrong than actually implementing the project. It had to do with a port on orion11 being reserved for an application. The port was one of the ports I was using for the project. This application was actually responding to the requests I was sending to the port. 
However, since the recv method had no clue what the response was, it broke and gave me a "makeslice: len out of range error". 

### What did you learn from completing this project? Is there anything you would change about the project?
#### I learned a lot about go and distributed systems. To be specific I learnt about the following:
##### Go's concurrency model
This was one of the positive things that came out of the error I mentioned above. In trying to figure out what exactly was wrong, I learnt a lot about go's concurrency model. I learnt about go routines, channels, and how to use them effectively. I also learnt about the different ways to synchronize go routines.
Channels are something I used quite a bit for project 2. Plus, it helped me avoid using mutexes for project 2. The book Concurrency in Go by Katherine Cox-Buday helped me quite a bit.
##### ...

#### Changes I would make to the project:
- Since orionxx are all on the same cluster, it'd have been nice to work with two clusters that are far apart(or some way to simulate that). It'd be interesting to see how the system behaves when the nodes are far apart. Plus I wonder if load balancing would need to change because of the distance between the nodes.
