package punchy

import (
	"bufio"
	"bytes"
	"encoding/binary"
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

func ReadPacket(sender *net.UDPAddr, p []byte) (*SenderMessage, error) {
	buffy := bytes.NewBuffer(p)
	var newMessage SenderMessage
	newMessage.Sender = sender
	err := binary.Read(buffy, binary.LittleEndian, newMessage.Type)
	if err != nil {
		return nil, ProtocolReadError
	}
	err = binary.Read(buffy, binary.LittleEndian, newMessage.Encrypted)
	if err != nil {
		return nil, ProtocolReadError
	}
	err = binary.Read(buffy, binary.LittleEndian, newMessage.Length)
	if err != nil {
		return nil, ProtocolReadError
	}
	newMessage.Data = buffy.Bytes()[:newMessage.Length]
	if err != nil {
		return nil, ProtocolReadError
	}
	return &newMessage, nil
}

func BuildMessage(Sender *net.UDPAddr, newMessage *Message) ([]byte, error) {
	buffy := bytes.NewBuffer(make([]byte, MAX_UDP_DATAGRAM))
	err := binary.Write(buffy, binary.LittleEndian, newMessage.Type)
	if err != nil {
		return nil, ProtocolWriteError
	}
	err = binary.Write(buffy, binary.LittleEndian, newMessage.Encrypted)
	if err != nil {
		return nil, ProtocolWriteError
	}
	err = binary.Write(buffy, binary.LittleEndian, newMessage.Length)
	if err != nil {
		return nil, ProtocolWriteError
	}
	err = binary.Write(buffy, binary.LittleEndian, newMessage.Data)
	if err != nil {
		return nil, ProtocolWriteError
	}
	return buffy.Bytes(), nil
}

func ContiniousRead(conn *net.UDPConn, server *net.UDPAddr, errorChan chan error) {
	buf := make([]byte, MAX_UDP_DATAGRAM)
	for {
		n, sender, err := conn.ReadFromUDP(buf)
		if n > 0 && err == nil && sender != server {
			fmt.Printf("%v says \"%v\"", sender, string(buf[:n]))
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
