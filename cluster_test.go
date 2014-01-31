package cluster

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

const delayInMillis = 50

// TODO: Add error handling to tests

// sendMessage sends "count" number of messages on channel
// outbox. The id of server to whom the messages should be
// sent is given by "to"
func sendMessages(outbox chan *Envelope, count int, to int, from int) {
	for i := 0; i < count; i += 1 {
		outbox <- &Envelope{Pid: to, Msg: strconv.Itoa(from) + ":" + strconv.Itoa(i)}
		if i%10 == 0 {
			//			fmt.Println("Sender sleeping ", i)
			time.Sleep(delayInMillis * time.Millisecond)
		}
	}
	fmt.Println("Sender done sending")
}

// receive receives messages on inbox. Count indicates expected number of
// messages. The verification uses following idea.
// record slice contains one slot for each of the messages. Messages
// received are of the form "senderPid:msgId". Every server sends same
// set of msgIds but prefixes its Pid. We extract msgId and senderPid
// and set a bit corresponding to server in slot for msgId. For each
// server, we check that for each slot, the number of bits set is
// equal to (serverCount - 1) (since server does not receive its
// own broadcast.
// Parameters:
// inbox         : Server inbox channel
// record        : Temp array to store information about received messages
// count         : Number of messages expected from each peer server
// serverCount   : Total number of servers in the cluster
// done          : Channel to write to when the server receives all expected
//                 messages
func receive(inbox chan *Envelope, record []uint32, count int, serverCount int, done chan bool) {
	receivedCount := count
	smallestCount := uint32((1 << (uint(serverCount) - 1)) - 1)
	for {
		envelope := <-inbox

		if msg, ok := envelope.Msg.(string); ok {
			sender := uint(envelope.Pid) - 1 // sender's Pid
			msgId, _ := strconv.ParseInt(strings.Split(msg, ":")[1], 10, 0)
			record[msgId] = record[int(msgId)] | (1 << sender)
			// record[msgId] is all 1's except for one bit. Thus, the smallest
			// possible valid record[msgId] is (1 << (serverCount - 1)) - 1
			//fmt.Println(record[msgId] , "\t" , smallestCount)
			if record[msgId] >= smallestCount && NumberOfBitsSet(record[msgId]) == serverCount-1 {
				//fmt.Println(record[msgId])
				receivedCount--
			}
		}

		if receivedCount == 0 {
			break
		}
	}
	done <- true
}

// TestOnwaySend creates two servers, sends large number of messages
// from one server to another and checks that messages are received
func TestOneWaySend(t *testing.T) {
	fmt.Println("Starting one way send test. Be patient, the test may run for several minutes")
	fmt.Println("One way send test case launches two servers and sends 10k messages from one ")
	fmt.Println("server to the other. Test checks that exactly one copy of every message is received")
	conf := Config{MemberRegSocket: "127.0.0.1:9999", PeerSocket: "127.0.0.1:9009"}

	// launch proxy server
	go acceptClusterMember(9999)
	go sendClusterMembers(9009)

	time.Sleep(100 * time.Millisecond)

	sender, err := NewWithConfig(1, "127.0.0.1", 5011, &conf)
	if err != nil {
		t.Errorf("Error in creating server ", err.Error())
	}

	receiver, err := NewWithConfig(2, "127.0.0.1", 5012, &conf)
	if err != nil {
		t.Errorf("Error in creating server ", err.Error())
	}

	done := make(chan bool, 1)
	count := 10000

	record := make([]uint32, count)

	go receive(receiver.Inbox(), record, count, 2, done)
	go sendMessages(sender.Outbox(), count, 2, 1)
	select {
	case <-done:
		fmt.Println("TestOneWaySend passed successfully")
		break
	case <-time.After(5 * time.Minute):
		t.Errorf("Could not send ", strconv.Itoa(count), " messages in 5 minute")
		break
	}
}

const messageSize = 10000000 // message size in bytes
// TestHugeMessages passes 1000 messages, each of
// size 10000000 bytes in length. Receiver counts
// number of messages and the length of each message
func _TestHugeMessages(t *testing.T) {
	fmt.Println("Starting large message send test. Be patient, the test may run for several minutes")
	fmt.Println("TestHugeMessages sends 1k messages each of 10^7 bytes in length")
	conf := Config{MemberRegSocket: "127.0.0.1:9999", PeerSocket: "127.0.0.1:9009"}

	// launch proxy server
	go acceptClusterMember(9999)
	go sendClusterMembers(9009)

	time.Sleep(100 * time.Millisecond)

	sender, err := NewWithConfig(1, "127.0.0.1", 5021, &conf)
	if err != nil {
		t.Errorf("Error in creating server ", err.Error())
	}

	receiver, err := NewWithConfig(2, "127.0.0.1", 5022, &conf)
	if err != nil {
		t.Errorf("Error in creating server ", err.Error())
	}
	errorChannel := make(chan error, 1)
	count := 1000

	var buf bytes.Buffer
	for i := 0; i < messageSize; i += 1 {
		buf.WriteString("1")
	}
	done := make(chan bool, 1)
	go sendHugeMessages(sender.Outbox(), errorChannel, buf.String(), count, 2, 1)
	go recvHugeMessages(receiver.Inbox(), errorChannel, count, done)
	select {
	case <-done:
		fmt.Println("TestOneWaySend passed successfully")
		break
	case err := <-errorChannel:
		t.Errorf("Error in sending large message.\n" + err.Error())
	case <-time.After(5 * time.Minute):
		t.Errorf("Could not send ", strconv.Itoa(count), " messages in 5 minute")
		break
	}
}

func sendHugeMessages(outbox chan *Envelope, errorChannel chan error, msg string, count int, to int, from int) {
	for i := 0; i < count; i += 1 {
		outbox <- &Envelope{Pid: to, Msg: msg}
		if i%10 == 0 {
			//			fmt.Println("Sender sleeping ", i)
			time.Sleep(delayInMillis * time.Millisecond)
		}
	}
	fmt.Println("Sender done sending")
}

func recvHugeMessages(inbox chan *Envelope, errorChannel chan error, count int, done chan bool) {
	for {
		envelope := <-inbox
		msg := envelope.Msg
		if msg1, ok := msg.(string); ok && len(msg1) != messageSize {
			errorChannel <- errors.New("Received a message with incorrect length")
			return
		}
		//		fmt.Println(count)
		count--
		if count == 0 {
			done <- true
			return
		}
	}
}
