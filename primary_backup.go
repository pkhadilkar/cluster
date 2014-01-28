/*
This file contains code for primary and backup name servers. Name servers accept new
cluster members and provide information about current cluster members. They do not 
transmit / receive data messages. Every server is expected to know primary and 
backup servers IP address and ports.
*/

package cluster


import (
	zmq "github.com/pebbe/zmq4"
	"strconv"
	"fmt"
)

type Catalog struct {
	Records map[int]string
}

// acceptClusterMember accepts requests about newly joined servers in the cluster
func acceptClusterMember(port int){
	catalog := Catalog{make(map[int]string)}
	responder, err := zmq.NewSocket(zmq.REP)
	if err != nil {
		// TODO: replace this with error channel
		fmt.Println("Cannot create a new socket for primary / backup. ", err.Error())
		return
	}
	defer responder.Close()
	responder.Bind("tcp://*:" + strconv.Itoa(port))
	for {
		//     	 fmt.Println("Waiting to receive message from socket")
		msg, err := responder.RecvBytes(0)
		if err != nil {
			fmt.Println("Error in acceptClusterMember ", err.Error())
		}
		
		member, err := BytesToClusterMember(msg)
		if err != nil {
			recordMember(member, &catalog)
		} else {
			fmt.Println("Received object which is not of type ClusterMember")
		}
		
		responder.Send("1", 0)
	}
}

// ClusterMember is used by new cluster members in their message to primary about joining the cluster
type ClusterMember struct {

	// Pid of the newly joined server. 
	// Unique across all cluster members
	Pid int
	// IP address of the member
	IP string
	// Inbox port for the server
	Port int
}

func recordMember(member *ClusterMember, catalog *Catalog) {
	catalog.Records[member.Pid] = member.IP + strconv.Itoa(member.Port)
}
