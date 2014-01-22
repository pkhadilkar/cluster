package cluster

const (BROADCAST = -1)
// Envelope represents a message on cluster network
type Envelope struct {
    // On the sender side, Pid identifies the receiving peer. If instead, Pid is
    // set to cluster.BROADCAST, the message is sent to all peers. On the receiver side, the
    // Id is always set to the original sender. If the Id is not found, the message is silently dropped
    Pid int
   
   
    // An id that globally and uniquely identifies the message, meant for duplicate detection at
    // higher levels. It is opaque to this package.
    MsgId int64
   
    // the actual message.
    Msg interface{}
}
   
type Server interface {
    // Id of this server
    Pid() int
   
    // array of other servers' ids in the same cluster
    Peers() []int
   
   
    // the channel to use to send messages to other peers
    // Note that there are no guarantees of message delivery, and messages
    // are silently dropped
    Outbox() chan *Envelope
   
   
    // the channel to receive messages from other peers.
    Inbox() chan *Envelope
}