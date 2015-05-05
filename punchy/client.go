package punchy

import (
	"bufio"
	"bytes"
	"encoding/binary"
	//	"encoding/gob"
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
	clientChannel chan InboundMessage
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
	return &Client{make(chan string), make(chan InboundMessage), make(chan error), s, c, make(map[string][]*net.UDPAddr)}
}

func (c *Client) ConnectToRoom(roomName string) {
	// Continous Read & Writes.
	roomMessage := RoomMessage{roomName}
	raw, err := roomMessage.RawMessage()
	if err != nil {
		panic(err)
	}
	message := &Message{raw, CONNECT_TO_ROOM, false, uint16(len(raw.Data))}
	data, err := message.EncodeMessage()
	if err != nil {
		panic(err)
	}
	c.conn.WriteTo(data, c.middleMan)
	fmt.Println("Join room")
	//	partnerDecoder := gob.NewDecoder(c.conn)
	partner := net.UDPAddr{}
	fmt.Println("Waiting for another member")
	//	err = partnerDecoder.Decode(&partner)
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

	go c.ClientContiniousWrite(roomName)
	panic(<-c.errorChannel)
}

func (c *Client) StartUp() {
	go c.ClientContiniousRead()
	go c.Display()
}

func (c *Client) Display() {
	for {
		message := <-c.clientChannel
		if message.Type() == ROOM_MESSAGE {
			var chatMessage ChatMessage
			chatMessage.DecodeMessage(message.RawData())
			fmt.Printf("%v says \"%v\" to room %v", message.Sender(), chatMessage.Message, chatMessage.Room)
		}
	}
}

func (c *Client) ClientContiniousRead() {
	buf := make([]byte, MAX_UDP_DATAGRAM)
	for {
		n, sender, err := c.conn.ReadFromUDP(buf)
		fmt.Println(string(buf[:n]))
		var message Message
		message.RawMessage.Sender = sender
		message.DecodeMessage(sender, buf)
		fmt.Println(message)
		fmt.Println(message.Type() == PING)
		fmt.Println(sender.String())
		fmt.Println(c.middleMan.String())
		if n > 0 && err == nil {
			if message.Type() == PING {
				c.Pong()
			} else if message.Type() == ROOM_MESSAGE {
				c.clientChannel <- &message
			}

		} else if err != nil {
			c.errorChannel <- err
		}
	}
}

func (c *Client) Pong() {
	fmt.Println("Ponging...")
	m := &Message{RawMessage{nil, make([]byte, 0)}, PONG, false, 0}
	data, err := m.EncodeMessage()
	if err != nil {
		panic(err)
	}
	c.conn.WriteToUDP(data, c.middleMan)
}

func (c *Client) ClientContiniousWrite(roomName string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		for _, client := range c.rooms[roomName] {
			roomMes := &ChatMessage{RoomMessage{roomName}, text}
			roomData, err := roomMes.EncodeMessage()
			if err != nil {
				panic(err)
			}
			sendMe := Message{RawMessage{nil, roomData}, ROOM_MESSAGE, false, uint16(len(roomData))}
			data, err := sendMe.EncodeMessage()
			if err != nil {
				panic(err)
			}
			n, err := c.conn.WriteToUDP(data, client)
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
	sending.EncryptedMsg = false
	sending.MsgType = ROOM_MESSAGE
	sendMe := &ChatMessage{RoomMessage{roomName}, message}
	writeToMe := bytes.NewBuffer(make([]byte, 0))
	err := binary.Write(writeToMe, binary.LittleEndian, sendMe)
	if err != nil {
		panic(err)
	}
	return sending
}
