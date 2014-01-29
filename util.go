package cluster

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
)

// This file contains utility functions for cluster

// EnvelopeToBytes converts an Envelope object into its binary representation.
// It returns a byte slice.

func EnvelopeToBytes(e *Envelope) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(e)
	return buf.Bytes()
}

// BytesToEnvelope decodes gob encoded representation of Envelope
func BytesToEnvelope(gobbed []byte) *Envelope {
	buf := bytes.NewBuffer(gobbed)
	dec := gob.NewDecoder(buf)
	var ungobbed Envelope
	dec.Decode(&ungobbed)
	return &ungobbed
}

func CatalogToBytes(e *Catalog) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(e)
	return buf.Bytes()
}

func BytesToCatalog(gobbed []byte) (*Catalog, error) {
	buf := bytes.NewBuffer(gobbed)
	dec := gob.NewDecoder(buf)
	var ungobbed Catalog
	err := dec.Decode(&ungobbed)
	if err != nil {
		return nil, err
	}
	return &ungobbed, err
}

func ClusterMemberToBytes(e *ClusterMember) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(e)
	return buf.Bytes()
}

func BytesToClusterMember(gobbed []byte) (*ClusterMember, error) {
	buf := bytes.NewBuffer(gobbed)
	dec := gob.NewDecoder(buf)
	var ungobbed ClusterMember
	err := dec.Decode(&ungobbed)
	if err != nil {
		return nil, err
	}
	return &ungobbed, nil
}

// ReadConfig reads configuration file information into Config object
// parameters:
// path : Path to config file
func ReadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var conf Config
	err = json.Unmarshal(data, &conf)
	if err != nil {
		fmt.Println("Error", err.Error())
		return nil, errors.New("Incorrect format in config file.\n" + err.Error())
	}
	fmt.Println("Config is ", conf)
	return &conf, err
}

// Config struct represents all config information
// required to start a server. It represents
// information in config file in structure
type Config struct {
	PidList         []int             // List of pids of all servers
	Servers         map[string]string // map from string ids to socket
	MemberRegSocket string            // socket to connect to , to register a server
	PeerSocket      string            // socket to connect to , to get a list of peers
}

// ClusterMember is used by new cluster members in their
// message to primary about joining the cluster
type ClusterMember struct {
	// Pid of the newly joined server.
	// Unique across all cluster members
	Pid int
	// IP address of the member
	IP string
	// Inbox port for the server
	Port int
}

// Gets map from Pids to sockets from config
func (c *Config) getServerAddressMap() (map[int]string, error) {
	addressOf := make(map[int]string)
	for key, value := range c.Servers {
		// bitSize 0 for int
		i, err := strconv.ParseInt(key, 10, 0)
		if err != nil {
			return nil, err
		}
		addressOf[int(i)] = value
	}
	return addressOf, nil
}
