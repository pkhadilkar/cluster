package cluster

// This file contains functions that deal with
// sending and receiving messages to/from
// network

// TODO: Handle errors from handleInPort and handleOutPort
// TODO: Find alternate method to store list of peers in handleOutport. Doing list.New() for each message is expensive

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
	responder, err := zmq.NewSocket(zmq.PULL)
	if err != nil {
		fmt.Println("Error in creating socket.", err)
		return
	}
	defer responder.Close()
	// addressOf map stores mapping from pid to socket
	responder.Bind("tcp://" + s.IP + ":" + strconv.Itoa(s.port))
	for {
		//		fmt.Println("handleInPort: Waiting to receive message from socket for pid", s.pid, "on", s.port)
		msg, err := responder.RecvBytes(0)
		//		fmt.Println("handleInPort: Received message")
		if err != nil {
			fmt.Println("Error in handleInPort ", err.Error())
		}

		//		fmt.Println("handleInPort: Message ", BytesToEnvelope(msg))
		//		fmt.Println("handleInPort: sent reply")
		s.inbox <- BytesToEnvelope(msg)
	}
}

// handleOutPort handles messages sent from this server
func (s *serverImpl) handleOutPort() {
	//	fmt.Println("handleOutPort: Waiting for message on outbox")
	// initial cache refresh
	s.setAddressOf(make(map[int]string))
	s.refreshPeerCache(BROADCAST)
	peerSocketCache := make(map[int]*zmq.Socket)

	for {
		msg := <-s.outbox

		receivers := list.New()
		// BROADCAST packet or packet for new server
		if _, ok := s.address(msg.Pid); !ok {
			s.refreshPeerCache(msg.Pid)
			// packet is dropped silently
			if _, ok := s.address(msg.Pid); !ok && msg.Pid != BROADCAST {
				fmt.Println("Address of ", msg.Pid, " could not be found")
				continue
			}
		}

		if msg.Pid != BROADCAST {
			receivers.PushBack(msg.Pid)
		} else {
			s.RLock()
			for key, _ := range s.addressOf {
				if key != s.pid {
					receivers.PushBack(key)
				}
			}
			s.RUnlock()
		}
		// change msg.Pid to match this server
		// receiving server will then find correct Pid
		msg.Pid = s.Pid()
		// send message to receivers
		gobbedMessage := EnvelopeToBytes(msg)
		for pid := receivers.Front(); pid != nil; pid = pid.Next() {
			// cache connections to peers
			var requester *zmq.Socket
			var ok bool
			var err error
			pidInt := 0
			if pidInt, ok = pid.Value.(int); ok {
			}

			if requester, ok = peerSocketCache[pidInt]; !ok {
				requester, err = zmq.NewSocket(zmq.PUSH)
				if err != nil {
					fmt.Println("Error in creating request socket. ", err.Error())
					return
				}
				peerSocketCache[pidInt] = requester
				socketStr, _ := s.address(pidInt)
				requester.Connect("tcp://" + socketStr)
				defer requester.Close()
			}

			requester.SendBytes(gobbedMessage, 0)
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
	//	fmt.Println("refreshPeerCache: refreshing cache")
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
		
		if _, ok := catalog.Records[pid]; ok || pid == BROADCAST {
			s.setAddressOf(catalog.Records)
			requester.Close()
			return nil
		}
		requester.Close()
	}
	return err
}
