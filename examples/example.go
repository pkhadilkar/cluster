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
	selfIP       = flag.String("address", "127.0.0.1", "IP address of the current server")
	selfPort     = flag.Int("port", 9090, "Input port for the current server")
	launchProxy  = flag.String("launchProxy", "n", "Launch proxy server before launching the server")
)

func main() {
	flag.Parse()
	// launch Proxy
	fmt.Println("example: ", *launchProxy)
	if *launchProxy == "y" {
		fmt.Println("Launching proxy ....")
		cluster.NewProxy(*confFilePath)
		time.Sleep(10 * time.Second)
	}
	// add a small delay to let Proxy server start and the other server join
	time.Sleep(1 * time.Minute)
	server, _ := cluster.New(*selfId, *selfIP, *selfPort, *confFilePath)
	// the returned server object obeys the Server interface above.
	fmt.Println(selfId, " sending message")
	// Let each server broadcast a message
	server.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, Msg: "hello there"}
	// add explicit delay to let the other server join and notify the proxy
	time.Sleep(1 * time.Minute)
	var envelope *cluster.Envelope
	select {
	case envelope = <-server.Inbox():
		fmt.Printf("Received msg from %d: '%s'\n", envelope.Pid, envelope.Msg)

	case <-time.After(10 * time.Second):
		fmt.Println("Waited and waited. Ab thak gaya\n")
	}
	// time out to ensure that server started first does not
	// exit immediately after receiving message
	time.Sleep(50 * time.Second)
}
