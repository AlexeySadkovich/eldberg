package discover

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
)

type messageType int

const (
	Ping messageType = iota
	Pong
	FindNode
	Nodes
)

type message struct {
	ID   uint64
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
	if err := enc.Encode(msg); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	length := uint64(buf.Len())
	var lengthBytes [8]byte
	binary.PutUvarint(lengthBytes[:], length)

	result := make([]byte, 0, 8+length)
	result = append(result, lengthBytes[:]...)
	result = append(result, buf.Bytes()...)

	return result, nil
}

func deserializeMessage(data io.Reader) (*message, error) {
	gob.Register([]*netnode{})

	lengthBytes := make([]byte, 8)
	_, err := data.Read(lengthBytes)
	if err != nil {
		return nil, fmt.Errorf("read length bytes: %w", err)
	}

	lengthReader := bytes.NewBuffer(lengthBytes)
	length, err := binary.ReadUvarint(lengthReader)
	if err != nil {
		return nil, fmt.Errorf("read length: %w", err)
	}

	msgBytes := make([]byte, length)
	_, err = data.Read(msgBytes)
	if err != nil {
		return nil, fmt.Errorf("read message bytes: %w", err)
	}

	msg := &message{}
	buf := bytes.NewBuffer(msgBytes)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(msg); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return msg, nil
}
