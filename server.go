/*
Package cluster provides API in the form of a server on cluster
that can talk to other servers on cluster. Both point to point
and broadcast messages are supported. 
*/
package cluster

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