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
func sendMessages(outbox chan *Envelope, count int, to int) {
	for i := 0; i < count; i += 1 {
		outbox <- &Envelope{Pid: to, Msg: "hello there" + strconv.Itoa(i)}
		if i%10 == 0 {
			//			fmt.Println("Sender sleeping ", i)
			time.Sleep(delayInMillis * time.Millisecond)
		}
	}
	fmt.Println("Sender done sending")
}

// receiveMessage receives messages on inbox. Count indicates
// expected number of messages
func receiveMessages(inbox chan *Envelope, count int) {
	for {
		<-inbox
		count--
		//		fmt.Println(count)
		if count == 0 {
			break
		}
	}
}

// receive acts as a wrapper for receiveMessages and adds
// channel synchronization to ensure that caller waits
// till receiveMessages completes
func receive(inbox chan *Envelope, count int, done chan bool) {
	//	fmt.Println("Receiver count is ", strconv.Itoa(count))
	receiveMessages(inbox, count)
	//	fmt.Println("Receiver done")
	done <- true
	// keep receiving messages otherwise other servers might block
	/*	for {
		receiveMessages(inbox, count)
	}*/
}

// TestOnwaySend creates two servers, sends large number of messages
// from one server to another and checks that messages are received
func TestOneWaySend(t *testing.T) {
	conf := Config{PidList: []int{1, 2}, Servers: map[string]string{"1": "127.0.0.1:5001", "2": "127.0.0.1:5002"}}
	sender, err := NewWithConfig(1, &conf)
	if err != nil {
		t.Errorf("Error in creating server ", err.Error())
	}

	receiver, err := NewWithConfig(2, &conf)
	if err != nil {
		t.Errorf("Error in creating server ", err.Error())
	}

	done := make(chan bool, 1)
	count := 10000
	go receive(receiver.Inbox(), count, done)
	go sendMessages(sender.Outbox(), count, 2)
	select {
	case <-done:
		fmt.Println("TestOneWaySend passed successfully")
		break
	case <-time.After(5 * time.Minute):
		t.Errorf("Could not send ", strconv.Itoa(count), " messages in 1 minute")
		break
	}
}
