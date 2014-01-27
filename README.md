ClusterTalk
=========

ClusterTalk is a go library that provides an API that allows a bunch of servers to talk to each other. Servers represent any entity that needs to communicates. The servers are logically assumed to form a cluster. Servers can talk to each other through special type of messages. Both point to point and broadcast messages are supported.

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

+ One way message send :
 Send 10k messages from one server to the other and check the number and contents of messages received

+ Multiple server broadcast : 
Launches 5 servers each of which broadcasts 10k messages. Each servers confirms that it has received (serverCount - 1) * numMsg messages from remaining servers.

Tests are known to take a few minutes due to verification. If you write additional test cases, ensure that port numbers are not same as the ones used in other test cases. The servers are service endpoints and keep listening on their endpoints untill the main process that started them completes.

Config File
---------------
Configuration file is in JSON format. Configuration file gives information about Pids of servers (which should be unique in one cluster) and logical socket endpoint for each server. Only IP address is not enought as one machine can have more than one server instances. Thus, fixed port and different IP address setup would not work.

Pushkar Khadilkar