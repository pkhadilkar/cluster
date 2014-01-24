package main

// This file is an example of use of cluster package
// TODO: Change type arguments of cluster.New
// Add error handling

import (
	"flag"
	"fmt"
	"github.com/pkhadilkar/cluster"
	"time"
)

var (
	confFilePath = flag.String("config", "config.json", "Path to config file (in json format)")
	selfId       = flag.Int("id", 0, "Pid of current server")
)

func main() {
	flag.Parse()
	server, _ := cluster.New(*selfId, *confFilePath)
	// the returned server object obeys the Server interface above.

	// Let each server broadcast a message
	server.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, Msg: "hello there"}
	var envelope *cluster.Envelope
	select {
	case envelope = <-server.Inbox():
		fmt.Printf("Received msg from %d: '%s'\n", envelope.Pid, envelope.Msg)

	case <-time.After(10 * time.Second):
		fmt.Println("Waited and waited. Ab thak gaya\n")
	}
}
