package cluster

import (
	"fmt"
	"strconv"
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

// receiveMessage receives messages on inbox. Count indicates
// expected number of messages. Every receiveMessages gets a
// map of its own. Thus, no synchronization is needed while
// accessing the map.
func receiveMessages(inbox chan *Envelope, messages map[string]bool) {
	for {
		envelope := <-inbox

		if msg, ok := envelope.Msg.(string); ok {
			if _, ok := messages[msg]; ok {
				delete(messages, msg)
			}
		}

		if len(messages) == 0 {
			return
		}
		//		fmt.Println(count)
	}
}

// receive acts as a wrapper for receiveMessages and adds
// channel synchronization to ensure that caller waits
// till receiveMessages completes
func receive(inbox chan *Envelope, messages map[string]bool, done chan bool) {
	//	fmt.Println("Receiver count is ", strconv.Itoa(count))
	receiveMessages(inbox, messages)
	//	fmt.Println("Receiver done")
	done <- true
}

// TestOnwaySend creates two servers, sends large number of messages
// from one server to another and checks that messages are received
func _TestOneWaySend(t *testing.T) {
	fmt.Println("Starting one way send test. Be patient, the test may run for several minutes")
	conf := Config{MemberRegSocket: "127.0.0.1:9999", PeerSocket: "127.0.0.1:9009"}

	// launch proxy server
	go acceptClusterMember(9999)
	go sendClusterMembers(9009)

	time.Sleep(100 * time.Millisecond)

	sender, err := NewWithConfig(1, "127.0.0.1", 5001, &conf)
	if err != nil {
		t.Errorf("Error in creating server ", err.Error())
	}

	receiver, err := NewWithConfig(2, "127.0.0.1", 5002, &conf)
	if err != nil {
		t.Errorf("Error in creating server ", err.Error())
	}

	done := make(chan bool, 1)
	count := 10000
	// create map of expected messages
	messages := make(map[string]bool, count)
	for i := 0; i < count; i += 1 {
		messages[strconv.Itoa(1)+":"+strconv.Itoa(i)] = true
	}

	go receive(receiver.Inbox(), messages, done)
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
