/*
This file contains code for primary and backup name servers. Name servers accept new
cluster members and provide information about current cluster members. They do not
transmit / receive data messages. Every server is expected to know primary and
backup servers IP address and ports.
*/

package cluster

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"strconv"
	"sync"
)

type Catalog struct {
	Records map[int]string
	mutex sync.RWMutex
}

var catalog = Catalog{}

// acceptClusterMember accepts requests about newly joined servers in the cluster
func acceptClusterMember(port int) {
	catalog.mutex.Lock()
	catalog.Records = make(map[int]string)
	catalog.mutex.Unlock()
	responder, err := zmq.NewSocket(zmq.REP)
	if err != nil {
		// TODO: replace this with error channel
		fmt.Println("Cannot create a new socket for primary / backup. ", err.Error())
		return
	}
	defer responder.Close()
	responder.Bind("tcp://*:" + strconv.Itoa(port))
	for {
		//		fmt.Println("Waiting to receive message from socket")
		msg, err := responder.RecvBytes(0)
		if err != nil {
			fmt.Println("Error in acceptClusterMember ", err.Error())
		}

		member, err := BytesToClusterMember(msg)
		if err == nil {
			recordMember(member, &catalog)
		} else {
			fmt.Println("Received object which is not of type ClusterMember")
		}

		responder.Send("1", 0)
	}
}

func recordMember(member *ClusterMember, catalog *Catalog) {
	catalog.mutex.Lock()
	catalog.Records[member.Pid] = member.IP + ":" + strconv.Itoa(member.Port)
	catalog.mutex.Unlock()
}

// send details of cluster members for every request
func sendClusterMembers(port int) {
	responder, err := zmq.NewSocket(zmq.REP)
	if err != nil {
		// TODO: replace this with error channel
		fmt.Println("Cannot create a new socket for primary / backup. ", err.Error())
		return
	}
	defer responder.Close()
	responder.Bind("tcp://*:" + strconv.Itoa(port))
	for {
		if _, err = responder.RecvBytes(0); err != nil {
			fmt.Println("Error in sendClusterMembers", err.Error())
		}

		catalog.mutex.RLock()
		buf := CatalogToBytes(&catalog)
		catalog.mutex.RUnlock()
		responder.SendBytes(buf, 0)
	}
}
