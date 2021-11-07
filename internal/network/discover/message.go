package discover

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type messageType int

const (
	Ping messageType = iota
	Pong
	Ask
	Nodes
)

type message struct {
	ID   nodeID
	Type messageType
	Data interface{}
}

func serializeMessage(msg *message) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(msg)
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	return buf.Bytes(), nil
}

func deserializeMessage(data []byte) (*message, error) {
	msg := &message{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(msg)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return msg, nil
}
