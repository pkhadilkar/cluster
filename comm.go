package cluster

// This file contains functions that deal with
// sending and receiving messages to/from
// network

// TODO: Handle errors from handleInPort and handleOutPort
// TODO: Handle blocking that might occur when sender is not up (Use select)

import (
	"container/list"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"strconv"
)

// Handleinport listens on port inPort and forwards received messages
// to channel inbox. It assumes that messages received are gob
// encoded objects of the type Envelope
func (s *serverImpl) handleInPort() {
	responder, err := zmq.NewSocket(zmq.REP)
	if err != nil {
		fmt.Println("Error in creating socket.", err)
		return
	}
	defer responder.Close()
	// addressOf map stores mapping from pid to socket
	responder.Bind("tcp://" + s.IP + ":" + strconv.Itoa(s.port))
	for {
		fmt.Println("handleInPort: Waiting to receive message from socket for pid", s.pid, "on", s.port)
		msg, err := responder.RecvBytes(0)
		fmt.Println("handleInPort: Received message")
		if err != nil {
			fmt.Println("Error in handleInPort ", err.Error())
		}
		s.inbox <- BytesToEnvelope(msg)
		fmt.Println("handleInPort: Message ", BytesToEnvelope(msg))
		responder.Send("1", 0)
		fmt.Println("handleInPort: sent reply")
	}
}

// handleOutPort handles messages sent from this server
func (s *serverImpl) handleOutPort() {
	fmt.Println("Waiting for message on outbox")
	// initial cache refresh
	s.addressOf = make(map[int]string)
	s.refreshPeerCache(BROADCAST)

	for {
		msg := <-s.outbox

		receivers := list.New()
		// BROADCAST packet or packet for new server
		if _, ok := s.addressOf[msg.Pid]; !ok {
			s.refreshPeerCache(msg.Pid)
			// packet is dropped silently
			if _, ok := s.addressOf[msg.Pid]; !ok && msg.Pid != BROADCAST {
				fmt.Println("Address of ", msg.Pid, " could not be found")
				continue
			}
		}

		if msg.Pid != BROADCAST {
			receivers.PushBack(s.addressOf[msg.Pid])
		} else {
			fmt.Println("Sending broadcast to ", s.addressOf)
			for key, value := range s.addressOf {
				if key != s.pid {
					receivers.PushBack(value)
				}
			}
		}
		// change msg.Pid to match this server
		// receiving server will then find correct Pid
		msg.Pid = s.Pid()
		// send message to receivers
		for socket := receivers.Front(); socket != nil; socket = socket.Next() {
			requester, err := zmq.NewSocket(zmq.REQ)
			if err != nil {
				fmt.Println("Error in creating request socket. ", err.Error())
				return
			}
			if socketStr, ok := socket.Value.(string); ok {
				fmt.Println("handleOutPort: ", s.pid , " sending message to ", string(socketStr))
				requester.Connect("tcp://" + string(socketStr))
				requester.SendBytes(EnvelopeToBytes(msg), 0)
				fmt.Println("handleOutPort: ", s.pid, " message sent")
				_, err := requester.Recv(0)
				fmt.Println("handleOutPort: ", s.pid, " received reply")
				if err != nil {
					fmt.Println("error in send", err.Error())
					requester.Close()
					break
				}
			}
			requester.Close()
		}

	}
}

// refreshPeerCache tries to contact peer server to get
// details about new server. The server returns a list of
// all peers. Server updates its cache with new list
func (s *serverImpl) refreshPeerCache(pid int) error {
	// retry three times
	var err error
	err = nil
	fmt.Println("refreshPeerCache: refreshing cache")
	for retryCount := 0; retryCount < 3; retryCount += 1 {
		requester, err := zmq.NewSocket(zmq.REQ)
		if err != nil {
			return err
		}

		err = requester.Connect("tcp://" + s.peerSocket)

		// actual content does not matter here
		_, err = requester.Send("1", 0)

		catBytes, err := requester.RecvBytes(0)

		if err != nil {
			fmt.Println("Error in receiving data from Peer registry server. Retrying ....\n", err.Error())
			requester.Close()
			continue
		}

		var catalog *Catalog
		catalog, err = BytesToCatalog(catBytes)

		if err != nil {
			fmt.Println("Error in decoding catalog data. Retrying ....\n", err.Error())
			requester.Close()
			continue
		}
		fmt.Println("Received catalog is ", catalog)
		// found required entry
		if _, ok := catalog.Records[pid]; ok || pid == BROADCAST {
			s.addressOf = catalog.Records
			requester.Close()
			return nil
		}
		requester.Close()
	}
	return err
}
