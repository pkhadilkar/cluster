package cluster

import (
	"testing"
	"time"
	"strconv"
	"fmt"
)

const delayInMillis = 500

// TODO: Add error handling to tests

// sendMessage sends "count" number of messages on channel
// outbox. The id of server to whom the messages should be
// sent is given by "to"
func sendMessages(outbox chan *Envelope, count int, to int){
	for i := 0 ; i < count; i += 1 {
		outbox <- &Envelope{Pid: to, Msg: "hello there" + strconv.Itoa(i)}
		if i % 10 == 0 {
			fmt.Println("Sender sleeping " , i)
			time.Sleep(delayInMillis * time.Millisecond)
		}
	}
	fmt.Println("Sender done sending")
}

// receiveMessage receives messages on inbox. Count indicates
// expected number of messages. If this routine receives 
// "count" messages on inbox, it writes true to done channel
func receiveMessages(inbox chan *Envelope, count int, done chan bool){
	for {
		<- inbox
		count--;
		fmt.Println(count)
		if count == 0 {
			done <- true
			break
		}
	}
}


// TestOnwaySend creates two servers, sends large number of messages
// from one server to another and checks that messages are received
func TestOneWaySend(t *testing.T){
	conf := Config{PidList : []int{1, 2}, Servers : map[string]string{"1" : "127.0.0.1:5001", "2" : "127.0.0.1:5002"}}
	sender, err := NewWithConfig(1, &conf)
	if err != nil {
		t.Errorf("Error in creating server ", err.Error())
	}

	receiver, err := NewWithConfig(2, &conf)
	if err != nil {
		t.Errorf("Error in creating server ", err.Error())
	}

	done := make(chan bool, 1)
	count := 324
	go receiveMessages(receiver.Inbox(), count, done)
	go sendMessages(sender.Outbox(), count, 2)
	<- done
}
