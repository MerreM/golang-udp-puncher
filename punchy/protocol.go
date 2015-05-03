package punchy

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

const MAX_UDP_DATAGRAM = 65507

type SenderMessage struct {
	Sender *net.UDPAddr
	Message
}

type Message struct {
	Type      MessageType
	Encrypted bool
	Length    uint16
	Data      []byte
}

type RoomMessage struct {
	room, message string
}

type MessageType uint8

const (
	PING                 MessageType = 1
	PONG                 MessageType = 2
	CONNECT_TO_ROOM      MessageType = 3
	DISCONNECT_FROM_ROOM MessageType = 4
	ROOM_LIST            MessageType = 5
	ROOM_MESSAGE         MessageType = 6
)

type TempMessage struct {
	Sender  *net.UDPAddr
	Message string
}

var ProtocolReadError = errors.New("Message cannot be read")
var ProtocolWriteError = errors.New("Message cannot be written")

func (m *RoomMessage) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	enc := gob.NewEncoder(w)
	err := enc.Encode(m.room)
	if err != nil {
		panic(err)
	}
	err = enc.Encode(m.message)
	if err != nil {
		panic(err)
	}
	return w.Bytes(), nil
}

func (m *RoomMessage) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&m.room)
	if err != nil {
		panic(err)
	}
	err = decoder.Decode(&m.message)
	if err != nil {
		panic(err)
	}
	return nil
}

func (m *Message) DecodeMessage(sender *net.UDPAddr, p []byte) error {
	buffy := bytes.NewBuffer(p)
	decod := gob.NewDecoder(buffy)
	err := decod.Decode(&m.Type)
	if err != nil {
		return ProtocolReadError
	}
	err = decod.Decode(&m.Encrypted)
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
	return nil
}

func (m *Message) EncodeMessage() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	enc := gob.NewEncoder(buf)
	err := enc.Encode(m.Type)
	if err != nil {
		return nil, ProtocolWriteError
	}
	err = enc.Encode(m.Encrypted)
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

func ContiniousRead(conn *net.UDPConn, server *net.UDPAddr, errorChan chan error) {
	buf := make([]byte, MAX_UDP_DATAGRAM)
	for {
		n, sender, err := conn.ReadFromUDP(buf)
		if n > 0 && err == nil {

		} else if n > 0 && err == nil && sender == server {

		} else if err != nil {
			errorChan <- err
		}
	}
}

func ContiniousWrite(conn *net.UDPConn, partner *net.UDPAddr, errorChan chan error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		n, err := conn.WriteToUDP([]byte(text), partner)
		if n > 0 && err == nil {
			fmt.Printf("Sent to %v\n", partner)
		} else if err != nil {
			errorChan <- err
		}
		reader.Reset(os.Stdin)
	}
}
