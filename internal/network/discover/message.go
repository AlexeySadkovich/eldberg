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
	FindNode
	Nodes
)

type message struct {
	From *netnode
	Type messageType
	Data interface{}
}

func (m *message) isRequest() bool {
	switch m.Type {
	case Ping:
	case FindNode:
	default:
		return false
	}

	return true
}

func (m *message) isResponse() bool {
	switch m.Type {
	case Pong:
	case Nodes:
	default:
		return false
	}

	return true
}

func serializeMessage(msg *message) ([]byte, error) {
	gob.Register([]*netnode{})

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
