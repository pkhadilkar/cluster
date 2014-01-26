package cluster

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

/*
TestMultiSendReceive creates serverCount = 5 servers. Each server broadcasts
msgCount = 10000 messages. Receivers check whether they receive
(serverCount - 1) * msgCount messages
*/
func TestMutliBroadcast(t *testing.T) {
	conf := Config{PidList: []int{1, 2, 3, 4, 5}, Servers: map[string]string{"1": "127.0.0.1:5004",
		"2": "127.0.0.1:5005", "3": "127.0.0.1:5006", "4": "127.0.0.1:5007",
		"5": "127.0.0.1: 5008",
	}}
	serverCount := 5
	servers := make([]Server, 10)
	for i := 1; i <= serverCount; i += 1 {
		s, err := NewWithConfig(i, &conf)
		if err != nil {
			t.Errorf("Error in creating server ", err.Error())
		}
		servers[i] = s
	}
	// start receive message loop for each server first
	count := 10000
	done := make(chan bool, serverCount)

	for i := 1; i <= serverCount; i += 1 {
		go receive(servers[i].Inbox(), count*(serverCount-1), done)
	}

	for i := 1; i <= serverCount; i += 1 {
		go sendMessages(servers[i].Outbox(), count, BROADCAST)
	}

	completed := serverCount
	for completed > 0 {
		select {
		case <-done:
			completed--
			fmt.Println("Completed server count is ", strconv.Itoa(serverCount-completed))
			if completed == 0 {
				break
			}
		case <-time.After(5 * time.Minute):
			t.Errorf("Did not receive all broadcast messages. Number of servers who received all messages were ", strconv.Itoa(serverCount-completed))
			break
		}
	}
	fmt.Println("TestMultiBroadcast passed successfully")
}
