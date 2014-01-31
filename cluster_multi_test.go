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

	for i := 1; i <= serverCount; i += 1 {
		// create custom map for each receiver
		record := make([]uint32, count)
		go receive(servers[i].Inbox(), record, count, serverCount, done)
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

// NumberOfBitsSet uses Kerninghan's method to count number of
// bits set Given a uint32 number, it returns number of bits set
func NumberOfBitsSet(l uint32) int {
	count := 0
	for l != 0 {
		l = l & (l - 1)
		count++
	}
	return count
}
