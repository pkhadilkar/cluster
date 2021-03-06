/*
Package cluster provides API in the form of a server on cluster
that can talk to other servers on cluster. Both point to point
and broadcast messages are supported.
*/
package cluster

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"strconv"
	"strings"
	"sync"
)

// buffer size for inbox and outbox channels
const bufferSize = 1000

// serverImpl is a type that implements server interface
type serverImpl struct {
	IP              string         // server's IP
	port            int            // server's input port
	pid             int            // own id
	peers           []int          // ids of peers
	outbox          chan *Envelope // channel for outbox messages
	inbox           chan *Envelope // channel for inbox messages
	addressOf       map[int]string // map to get hostname/ IP address given PID
	memberRegSocket string         //socket to connect to , to register new server
	peerSocket      string         // socket to contact to get list of peers
	sync.RWMutex // Mutex to protect addressOf map
}

const retryCount = 100 //retry count to contact registration server

// register contacts member registrar server and registers itself
// It keeps on retrying to connect to registrar server, because
// unless that server responds, the Server is not in cluster
func (s *serverImpl) register() error {
	var err error
	member := ClusterMember{Pid: s.pid, IP: s.IP, Port: s.port}
	buf := ClusterMemberToBytes(&member)
	for i := 0; i < retryCount; i += 1 {
		requester, err := zmq.NewSocket(zmq.REQ)
		if err != nil {
			fmt.Println("register(): Error in creating new socket.", err.Error())
			continue
		}
		requester.Connect("tcp://" + s.memberRegSocket)
		requester.SendBytes(buf, 0)

		_, err = requester.Recv(0)
		// server registered successfully
		if err == nil {
			requester.Close()
			return err
		}
		fmt.Println("Error in connecting member registration server. ", err.Error())
		requester.Close()
	}
	return err
}

func (s *serverImpl) Pid() int {
	return s.pid
}

func (s *serverImpl) Peers() []int {
	err := s.refreshPeerCache(BROADCAST)
	if err != nil {
		fmt.Println("Could not refresh peer cache." + err.Error())
		return nil
	}
	s.Lock()
	s.peers = make([]int, len(s.addressOf))
	s.Unlock()
	i := 0
	s.RLock()
	for key, _ := range s.addressOf {
		s.peers[i] = key
		i++
	}
	s.RUnlock()
	return s.peers
}
// Outbox returns channel that can be used to send messages to other servers
func (s *serverImpl) Outbox() chan *Envelope {
	return s.outbox
}
// Inbox returns channel that receives messages from other servers
func (s *serverImpl) Inbox() chan *Envelope {
	return s.inbox
}

// setAddressOf sets the map between pids and hostnames / ip addresses
func (s *serverImpl) setAddressOf(newmap map[int]string) {
	s.Lock()
	s.addressOf = newmap
	s.Unlock()
}

func (s *serverImpl) address(pid int) (string, bool) {
	s.RLock()
	defer s.RUnlock()
	address, ok := s.addressOf[pid]
	return address, ok
}

func initializeServer(selfId int, selfIP string, selfPort int, conf *Config) (Server, error) {
	s := serverImpl{pid: selfId, IP: selfIP, port: selfPort}
	s.outbox = make(chan *Envelope, bufferSize)
	s.inbox = make(chan *Envelope, bufferSize)
	s.memberRegSocket = conf.MemberRegSocket
	s.peerSocket = conf.PeerSocket
	//fmt.Println("Registering server with proxy")
	//fmt.Println("server : ", s)
	// register with proxy
	s.register()
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
func New(selfId int, selfIP string, selfPort int, configFilePath string) (Server, error) {
	fmt.Println("Launching new server")
	conf, err := ReadConfig(configFilePath)
	if err != nil {
		return nil, err
	}
	s, err := initializeServer(selfId, selfIP, selfPort, conf)
	return s, err
}

// NewWithConfi accepts a config object and returns
// server with appropriate behavior. This method is
// useful in test code.
func NewWithConfig(selfId int, selfIP string, selfPort int, c *Config) (Server, error) {
	s, err := initializeServer(selfId, selfIP, selfPort, c)
	return s, err
}

// NewProxy launches a new Proxy that handles member registration
// and accepts requests to get a list of current members. This
// function must be called on the server where we want
func NewProxy(configPath string) error {
	conf, err := ReadConfig(configPath)
	if err != nil {
		return err
	}
	// TODO: Should call NewProxyWithConfig from here
	memberPort := getPort(conf.MemberRegSocket)
	peerPort := getPort(conf.PeerSocket)
	// current implementation requires both acceptor and
	// peer server to be on the same machine
	go acceptClusterMember(memberPort)
	go sendClusterMembers(peerPort)
	return err
}

// getPort returns a port given a socket of the form "IP:port"
// TODO: Add error check here
func getPort(socket string) int {
	s, _ := strconv.ParseInt(strings.Split(socket, ":")[1], 10, 0)
	return int(s)
}

// NewProxyWithConfig provides a helper method
// for creating proxies programmatically
func NewProxyWithConfig(conf *Config) {
	memberPort := getPort(conf.MemberRegSocket)
	peerPort := getPort(conf.PeerSocket)
	// current implementation requires both acceptor and
	// peer server to be on the same machine
	go acceptClusterMember(memberPort)
	go sendClusterMembers(peerPort)
}
