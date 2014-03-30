package cluster

import (
	"fmt"
	"strings"
	"testing"
	"strconv"
	"time"
)

// TestRingSend creates a 100 servers and passes the
// messages from one server to another. The test checks
// if the first server receives all the messages it sends
// other servers simply forward the messages to next
// server in logical order
func TestRingSend(t *testing.T) {
	fmt.Println("TestRingSend creates 100 servers and connects them logically in the form of a ring.")
	fmt.Println("The first server sends 100 messages and confirms that it receives all the messages.")
	fmt.Println("Server i forwards all the messages it receives to Server (i + 1).")
	conf := Config{MemberRegSocket: "127.0.0.1:9900", PeerSocket: "127.0.0.1:9909"}

	// launch proxy
	go acceptClusterMember(9900)
	go sendClusterMembers(9909)

	serverCount := 100
	servers := make([]Server, serverCount)
	for i := 0; i < serverCount; i += 1 {
		s, err := NewWithConfig(i, "127.0.0.1", 8481+i, &conf)
		if err != nil {
			t.Errorf("Error in creating server ", err.Error())
		}
		servers[i] = s
	}
	// start receive message loop for each server first
	count := 1000
	done := make(chan bool, 1)
	errorChannel := make(chan string, 1)
	// launch initiator
	go sendMessages(servers[0].Outbox(), count, 1, 0)
	go initiatorRecv(servers[0].Inbox(), count, done, errorChannel)

	for i := 1; i < serverCount; i += 1 {
		go process(servers[i].Inbox(), servers[i].Outbox(), servers[i].Pid(), serverCount)
	}

	select {
	case <-done:
		fmt.Println("Received messages succesfully")
		break
	case <-time.After(9 * time.Minute):
		t.Errorf("Did not receive all broadcast messages. ")
		break
	}
}

func initiatorRecv(inbox chan *Envelope, count int, done chan bool, errorChannel chan string) {
	received := make([]bool, count)
	for {
		envelope := <-inbox
		msg := envelope.Msg
		var strMsg string
		if msg, ok := msg.(string); !ok {
			errorChannel <- "Received incorrect message. String msg expected."
			return
		} else {
			strMsg = msg
		}
		msgId, _ := strconv.ParseInt(strings.Split(strMsg, ":")[1], 10, 0)
		fmt.Println("from: " + strconv.Itoa(envelope.Pid) + " " + strconv.FormatInt(msgId, 10))
		if !received[int(msgId)] {
			//fmt.Println(strMsg)
			received[int(msgId)] = true
			count--
		}

		if count == 0 {
			done <- true
			return
		}
	}
}

func process(inbox chan *Envelope, outbox chan *Envelope, pid int, serverCount int) {
	for {
		envelope := <-inbox
		envelope.Pid = (pid + 1) % serverCount
		outbox <- envelope
	}
}
