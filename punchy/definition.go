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

type Message struct {
	Sender    *net.UDPAddr
	Length    int64
	Encrypted bool
	Data      []byte
}

type TempMessage struct {
	Sender  *net.UDPAddr
	Message string
}

var ProtocolReadError = errors.New("Length of message cannot be read")

func readPacket(p []byte) (*Message, error) {
	buffy := bytes.NewBuffer(p)
	length, err := binary.ReadVarint(buffy)
	if err != nil {
		return nil, ProtocolReadError
	}
	var tempEncrypted bool
	err = binary.Read(buffy, binary.LittleEndian, tempEncrypted)
	if err != nil {
		return nil, ProtocolReadError
	}
	return &Message{nil, length, tempEncrypted, buffy.Bytes()}, nil

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
