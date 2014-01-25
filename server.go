/*
Package cluster provides API in the form of a server on cluster
that can talk to other servers on cluster. Both point to point
and broadcast messages are supported.
*/
package cluster

import (
	"errors"
)

// buffer size for inbox and outbox channels
const bufferSize = 100

// serverImpl is a type that implements server interface
type serverImpl struct {
	// own id
	pid int
	// ids of peers
	peers []int
	// channel for outbox messages
	outbox chan *Envelope
	// channel for inbox messages
	inbox chan *Envelope
	// map to get hostname/ IP address given PID
	addressOf map[int]string
}

func (s *serverImpl) Pid() int {
	return s.pid
}

func (s *serverImpl) Peers() []int {
	return s.peers
}

func (s *serverImpl) Outbox() chan *Envelope {
	return s.outbox
}

func (s *serverImpl) Inbox() chan *Envelope {
	return s.inbox
}

func initializeServer(selfId int, conf *Config) (Server, error) {
	s := serverImpl{pid: selfId, peers: conf.PidList}
	addresses, err := conf.getServerAddressMap()
	if err != nil {
		return nil, errors.New("Error in server ip/port information" + err.Error())
	}
	s.addressOf = addresses
	s.outbox = make(chan *Envelope, bufferSize)
	s.inbox = make(chan *Envelope, bufferSize)
	go s.handleInPort()
	go s.handleOutPort()
	// launch goroutines to handle communication
	// between channels and cluster
	return Server(&s), nil
}

// New Function accepts two parameters.
// selfId: Pid of new server
// configFile: Path to a file containing configuration details.
// These include ids of all servers, port information.
func New(selfId int, configFilePath string) (Server, error) {

	conf, err := ReadConfig(configFilePath)
	if err != nil {
		return nil, err
	}
	s, _ := initializeServer(selfId, conf)
	return s, err
}

// NewWithConfi accepts a config object and returns
// server with appropriate behavior. This method is
// useful in test code.
func NewWithConfig(selfId int, c *Config) (Server, error) {
	s , err := initializeServer(selfId, c)
	return s, err
}
