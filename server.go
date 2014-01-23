/*
Package cluster provides API in the form of a server on cluster
that can talk to other servers on cluster. Both point to point
and broadcast messages are supported.
*/
package cluster

import (
	"encoding/json"
	"errors"
	"io/ioutil"
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

// Config struct represents all config information
// required to start a server. It represents
// information in config file in structure
type Config struct {
	// List of pids of all servers
	PidList []int
	// Port used to send data out on cluster
	SendPort int
	// Port used to receive data from cluster
	ReceivePort int
}

func initializeServer(selfId int, conf *Config) (Server, error) {
	s := serverImpl{pid: selfId, peers: conf.PidList}
	s.outbox = make(chan *Envelope, bufferSize)
	s.inbox = make(chan *Envelope, bufferSize)
	// launch goroutines to handle communication
	// between channels and cluster
	return Server(&s), nil
}

func readConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var conf Config
	err = json.Unmarshal(data, &conf)
	if err != nil {
		return nil, errors.New("Incorrect format in config file.\n" + err.Error())
	}
	return &conf, err
}

// New function accepts two parameters.
// selfId: Pid of new server
// configFile: Path to a file containing configuration details.
// These include ids of all servers, port information.
func New(selfId int, configFilePath string) (Server, error) {

	conf, err := readConfig(configFilePath)
	if err != nil {
		return nil, err
	}
	s, _ := initializeServer(selfId, conf)
	return s, err
}
