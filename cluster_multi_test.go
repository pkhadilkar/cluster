package cluster

/*
import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)
*/
/*
TestMultiSendReceive creates serverCount = 5 servers. Each server broadcasts
msgCount = 10000 messages. Receivers check whether they receive
(serverCount - 1) * msgCount messages. The msg sending works fast, verification
of received messages takes time and memory
*/ /*
func _TestMutliBroadcast(t *testing.T) {
	fmt.Println("Starting multibroadcast test. Be patient, the test can run for several minutes.")
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
*/
