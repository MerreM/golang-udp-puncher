package punchy

import (
	"bytes"
	"encoding/gob"
	"errors"
	"net"
)

type MessageType uint8

const (
	PING                  MessageType = 1
	PONG                  MessageType = 2
	CONNECT_TO_MIDDLE_MAN MessageType = 3
	RESPOND_TO_MIDDLE_MAN MessageType = 4
	CONNECT_TO_ROOM       MessageType = 5
	DISCONNECT_FROM_ROOM  MessageType = 6
	ROOM_LIST             MessageType = 7
	ROOM_MESSAGE          MessageType = 8
	ROOM_HISTORY          MessageType = 9
)
const MAX_UDP_DATAGRAM = 65507

type RawMessage struct {
	Sender *net.UDPAddr
	Data   []byte
}

type Message struct {
	RawMessage
	MsgType      MessageType
	EncryptedMsg bool
	Length       uint16
}

type RoomMessage struct {
	Room string
}

type ChatMessage struct {
	RoomMessage
	Message string
}

var ProtocolReadError = errors.New("Message cannot be read")
var ProtocolWriteError = errors.New("Message cannot be written")

func (m *RoomMessage) EncodeMessage() ([]byte, error) {
	w := new(bytes.Buffer)
	enc := gob.NewEncoder(w)
	err := enc.Encode(m.Room)
	if err != nil {
		panic(err)
	}
	return w.Bytes(), nil
}

func (m *RoomMessage) DecodeMessage(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&m.Room)
	if err != nil {
		panic(err)
	}
	return nil
}

func (m *ChatMessage) EncodeMessage() ([]byte, error) {
	w := new(bytes.Buffer)
	enc := gob.NewEncoder(w)
	err := enc.Encode(m.Room)
	if err != nil {
		panic(err)
	}
	err = enc.Encode(m.Message)
	if err != nil {
		panic(err)
	}
	return w.Bytes(), nil
}

func (m *ChatMessage) DecodeMessage(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&m.Room)
	if err != nil {
		panic(err)
	}
	err = decoder.Decode(&m.Message)
	if err != nil {
		panic(err)
	}
	return nil
}

func (m *Message) DecodeMessage(sender *net.UDPAddr, p []byte) error {
	buffy := bytes.NewBuffer(p)
	decod := gob.NewDecoder(buffy)
	err := decod.Decode(&m.MsgType)
	if err != nil {
		return ProtocolReadError
	}
	err = decod.Decode(&m.EncryptedMsg)
	if err != nil {
		return ProtocolReadError
	}
	err = decod.Decode(&m.Length)
	if err != nil {
		return ProtocolReadError
	}
	err = decod.Decode(&m.Data)
	if err != nil {
		return ProtocolReadError
	}
	if err != nil {
		return ProtocolReadError
	}
	m.RawMessage.Sender = sender
	return nil
}

func (m *Message) EncodeMessage() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	enc := gob.NewEncoder(buf)
	err := enc.Encode(m.MsgType)
	if err != nil {
		return nil, ProtocolWriteError
	}
	err = enc.Encode(m.EncryptedMsg)
	if err != nil {
		return nil, ProtocolWriteError
	}
	err = enc.Encode(m.Length)
	if err != nil {
		return nil, ProtocolWriteError
	}
	err = enc.Encode(m.Data)
	if err != nil {
		return nil, ProtocolWriteError
	}
	return buf.Bytes(), nil
}
