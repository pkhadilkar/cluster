package cluster

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

/*
TestMultiSendReceive creates serverCount = 5 servers. Each server broadcasts
msgCount = 10000 messages. Receivers check whether they receive
(serverCount - 1) * msgCount messages. The msg sending works fast, verification
of received messages takes time and memory
*/
func TestMutliBroadcast(t *testing.T) {
	fmt.Println("Starting multibroadcast test. Be patient, the test can run for several minutes.")
	fmt.Println("MultiBroadcastTest broadcasts 10k messages from each of the five servers launched.")
	fmt.Println("Each server checks that it has received each of the 40k messages it expects exactly once.")
	conf := Config{MemberRegSocket: "127.0.0.1:9999", PeerSocket: "127.0.0.1:9009"}

	// launch proxy
	go acceptClusterMember(9999)
	go sendClusterMembers(9009)

	serverCount := 5
	servers := make([]Server, 10)
	for i := 1; i <= serverCount; i += 1 {
		s, err := NewWithConfig(i, "127.0.0.1", 5001+i, &conf)
		if err != nil {
			t.Errorf("Error in creating server ", err.Error())
		}
		servers[i] = s
	}
	// start receive message loop for each server first
	count := 10000
	done := make(chan bool, serverCount)

	// master map contains all possible messages
	masterMap := make(map[string]bool, count)

	for i := 1; i <= 5; i += 1 {
		for id := 0; id < count; id += 1 {
			masterMap[strconv.Itoa(i)+":"+strconv.Itoa(id)] = true
		}
	}

	for i := 1; i <= serverCount; i += 1 {
		// create custom map for each receiver
		messages := make(map[string]bool, (serverCount-1)*count)
		base := strconv.Itoa(i) + ":"
		for key, value := range masterMap {
			if strings.HasPrefix(key, base) != true {
				messages[key] = value
			}
		}
		go receive(servers[i].Inbox(), messages, done)
	}

	for i := 1; i <= serverCount; i += 1 {
		go sendMessages(servers[i].Outbox(), count, BROADCAST, i)
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
		case <-time.After(9 * time.Minute):
			t.Errorf("Did not receive all broadcast messages. Number of servers who received all messages were " + strconv.Itoa(serverCount-completed))
			break
		}
	}
	fmt.Println("TestMultiBroadcast passed successfully")
}
