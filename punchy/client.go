package punchy

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strings"
)

type ClientInter interface {
	ConnectToRoom(string)
	ConnectToServer(*net.UDPAddr)
}

type Client struct {
	inputChannel  chan string
	outputChannel chan TempMessage
	errorChannel  chan error
	middleMan     *net.UDPAddr
	conn          *net.UDPConn
	rooms         map[string][]*net.UDPAddr
}

func NewClient(port *int) *Client {
	addressString := fmt.Sprintf("%v:%v", "", *port)
	s, err := net.ResolveUDPAddr("udp", addressString)
	if err != nil {
		panic(err)
	}
	cAddr, err := net.ResolveUDPAddr("udp", ":")
	if err != nil {
		panic(err)
	}
	c, err := net.ListenUDP("udp", cAddr)
	if err != nil {
		panic(err)
	}
	return &Client{make(chan string), make(chan TempMessage), make(chan error), s, c, make(map[string][]*net.UDPAddr)}
}

func (c *Client) ConnectToRoom(roomName string) {
	// Continous Read & Writes.
	c.conn.WriteTo([]byte(roomName), c.middleMan)
	fmt.Println("Join room")
	partnerDecoder := gob.NewDecoder(c.conn)
	partner := net.UDPAddr{}
	fmt.Println("Waiting for another member")
	err := partnerDecoder.Decode(&partner)
	fmt.Printf("Attempting to connect to %v\n", partner)
	if err != nil {
		panic(err)
	}
	room := c.rooms[roomName]
	if room != nil {
		room = make([]*net.UDPAddr, 10)
	}
	room = append(room, &partner)
	c.rooms[roomName] = room
	fmt.Printf("Listening on...%v\n", c.conn.LocalAddr())
	//	go protocol.ContiniousRead(c.conn, c.middleMan, c.errorChannel)
	//	go protocol.ContiniousWrite(c.conn, &partner, c.errorChannel)
	go c.ClientContiniousRead()
	go c.ClientContiniousWrite(roomName)
	go c.Display()

	panic(<-c.errorChannel)
}
func (c *Client) Display() {
	for {
		message := <-c.outputChannel
		fmt.Printf("%v says \"%v\"", message.Sender, message.Message)
	}
}

func (c *Client) ClientContiniousRead() {
	buf := make([]byte, MAX_UDP_DATAGRAM)
	for {
		n, sender, err := c.conn.ReadFromUDP(buf)
		ReadPacket(sender)
		if n > 0 && err == nil && sender != c.middleMan {
			c.outputChannel <- TempMessage{sender, string(buf)}
		} else if n > 0 && err == nil && sender == c.middleMan {

		} else if err != nil {
			c.errorChannel <- err
		}
	}
}

func (c *Client) ClientContiniousWrite(roomName string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		for _, client := range c.rooms[roomName] {
			n, err := c.conn.WriteToUDP([]byte(text), client)
			if n > 0 && err == nil {
				fmt.Printf("Sent to %v\n", client)
			} else if err != nil {
				c.errorChannel <- err
			}
		}
		reader.Reset(os.Stdin)
	}
}

func (c *Client) MakeRoomMessage(roomName, message string) Message {
	var sending Message
	sending.Encrypted = false
	sending.Type = ROOM_MESSAGE
	sendMe := &RoomMessage{roomName, message}
	writeToMe := bytes.NewBuffer(make([]byte, 0))
	err := binary.Write(writeToMe, binary.LittleEndian, sendMe)
	if err != nil {
		panic(err)
	}
	return sending
}
