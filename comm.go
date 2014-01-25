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
	responder.Bind("tcp://" + s.addressOf[s.pid])
	for {
		//     	 fmt.Println("Waiting to receive message from socket")
		msg, err := responder.RecvBytes(0)
		if err != nil {
			fmt.Println("Error in handleInPort ", err.Error())
		}
		s.inbox <- BytesToEnvelope(msg)
		responder.Send("1", 0)
	}
}

// handleOutPort handles messages sent from this server
func (s *serverImpl) handleOutPort() {
	//    fmt.Println("Waiting for message on outbox")

	for {
		msg := <-s.outbox
		requester, err := zmq.NewSocket(zmq.REQ)
		// do not use defer
		if err != nil {
			fmt.Println("Error in creating request socket. ", err.Error())
			return
		}

		receivers := list.New()

		if msg.Pid != BROADCAST {
			receivers.PushBack(s.addressOf[msg.Pid])
		} else {
			for key, value := range s.addressOf {
				if key != s.pid {
					receivers.PushBack(value)
				}
			}
		}
		// send message to receivers
		for socket := receivers.Front(); socket != nil; socket = socket.Next() {
			if socketStr, ok := socket.Value.(string); ok {
				requester.Connect("tcp://" + string(socketStr))
				requester.SendBytes(EnvelopeToBytes(msg), 0)
				_, err := requester.Recv(0)
				if err != nil {
					fmt.Println("error in send", err.Error())
					break
				}
			}
		}
		requester.Close()
	}
}
