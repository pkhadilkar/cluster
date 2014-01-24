package cluster

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
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
		return nil, errors.New("Incorrect format in config file.\n" + err.Error())
	}
	return &conf, err
}

// Config struct represents all config information
// required to start a server. It represents
// information in config file in structure
type Config struct {
	PidList     []int             // List of pids of all servers
	SendPort    int               // Port used to send data out on cluster
	ReceivePort int               // Port used to receive data from cluster
	Servers     map[string]string // map from string ids to socket
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
