ClusterTalk
=========

ClusterTalk is a go library that provides an API that allows a bunch of servers to talk to each other. Servers represent any entity that needs to communicates. The servers are logically assumed to form a cluster. Servers can talk to each other through special type of messages. Both point to point and broadcast messages are supported. Servers **auto-join** the cluster. They do not need to know about their peers in advance.
(*There is an older version which does not use master and uses hard-coded, fixed set of peers from config file. Please see old-decentralized branch*)

Servers expose their services through channels (See interface.go). The messages sent to server are objects of type Envelope (see interface.go). Envelope packages message with additional information such as pid of the server to which the message is to be sent and optional message id. Type of actual message is opaque to server.

Installation
--------------
```
$ go get github.com/pkhadilkar/cluster
$ go get build github.com/pkhadilkar/cluster
```
ClusterTalk uses [zmq4](https://github.com/pebbe/zmq4) by pebbe. Zmq4 requires installation of [ZeroMQ](http://zeromq.org/). Installation instruction can be found [here](http://zeromq.org/intro:get-the-software). 

Test
--------------
```
$ go test github.com/pkhadilkar/cluster
```

Test cases:

+ **One way message send** :
 Send 10k messages from one server to the other and check the number and contents of messages received

+ **Multiple server broadcast** : 
Launches 5 servers each of which broadcasts 10k messages. Each servers confirms that it has received (serverCount - 1) * numMsg messages from remaining servers.
*Note that the tests are known to take upto several minutes as they check whether each message that is sent is received exactly as it was sent. If you want to just confirm that behaviour under cases tested is correct faster, reduce the count variable in approrpiate method in test file*. 
If you write additional test cases, ensure that port numbers are not same as the ones used in other test cases. The servers are service endpoints and keep listening on their endpoints untill the main process that started them completes.

+ **Large message send**:
This test case sends 5000 messages, each of size 10^7 (~10 MB) from one server to another.

+ **Ring Order Messages** :
 This test case creates 1000 servers and connects them logically in the form of a ring. Server i forwards messages it receives to Server (i + 1). The first server sends 1000 messages and confirms that it receives all the messages .

Architecture
---------------
ClusterTalk consists of several components

![components.jpg] (https://github.com/pkhadilkar/cluster/blob/master/images/components.jpg?raw=true)

+ **Master / Metadata Server** :
 Master node is a server that accepts member registration requests from servers in cluster. It also sends a list of all current members (peer list) to servers upon request. Main purpose of master node is to allow auto-joining of nodes to cluster. Master server does not process actual message data and hence it does not become a bottlneck with increased cluster size / number of messages. It acts as a metadata server. Current version has only one master node. The idea is to add a secondary metadata server (like Secondary name node in Hadoop) which should be in sync with the master and should be able to take over in case master is not available .

+ **Servers** :
 The servers in ClusterTalk are peers. Every server can send/receive messages to/from other servers. Servers can also send a broadcast message to all peers in the cluster. When a new server is launched, it reads address of master from config file and contacts master to register itself. Servers also request master to get a list of peers for broadcast messages. This list is cached in each server. Server also contacts the master when it gets a message with id of a server that is not in its cache. 

*Note that master server should be launched before other servers. See cluster_test.go for details*

Config File
---------------
Configuration file is in JSON format. Configuration file stores socket endpoints for member registration and peer information servers.


Data Types
------------
Envelope struct accepts any type of message. Type of Msg is interface{}. ClusterTalk internally uses GOB encoding/ decoding to send messages across servers in cluster. Interface encoding/ decoding in GOB requires "register"ing types with gob before Encoder and Decoder are created. Since we want to allow arbitrary message types, it is not possible to register types in advance. Thus, users of the library should encode their data in a format that provides string representation and then use that as Msg field in Envelope.

~Pushkar Khadilkar