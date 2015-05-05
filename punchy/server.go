package punchy

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

type ChatServer interface {
	Serve()
	Encrypted() bool
	AddUserToRoom(string *net.UDPAddr)
}

type Server struct {
	Port  int
	Conn  *net.UDPConn
	Rooms map[string]*ChatRoom
	//	ActiveClients []ClientConnection
}

type Uptime struct {
	lastSeen   time.Time
	checkCount int
}

type ChatRoom struct {
	clients     map[*net.UDPAddr]*Uptime
	upTimeQueue chan *net.UDPAddr
	pongQueue   chan *net.UDPAddr
}

func NewServer(port *int) Server {
	return Server{*port, nil, make(map[string]*ChatRoom)}
}
func (s *Server) AddToRoom(room *ChatRoom, client *net.UDPAddr) {
	room.clients[client] = &Uptime{time.Now(), 0}
	room.upTimeQueue <- client
	fmt.Printf("Handshake begins\n")
	for otherAddr, _ := range room.clients {
		if otherAddr != client {
			returnBuffer := bytes.NewBuffer(make([]byte, 0))
			otherBuffer := bytes.NewBuffer(make([]byte, 0))
			returnEncoder := gob.NewEncoder(returnBuffer)
			otherEncoder := gob.NewEncoder(otherBuffer)
			err := returnEncoder.Encode(client)
			if err != nil {
				panic(err)
			}
			err = otherEncoder.Encode(&otherAddr)
			if err != nil {
				panic(err)
			}
			s.Conn.WriteToUDP(otherBuffer.Bytes(), client)
			s.Conn.WriteToUDP(returnBuffer.Bytes(), otherAddr)
			fmt.Printf("Addresses Sent\n")
		}
	}
}
func (s *Server) Ping(client *net.UDPAddr) {
	m := &Message{RawMessage{nil, make([]byte, 0)}, PING, false, 0}
	data, err := m.EncodeMessage()
	if err != nil {
		panic(err)
	}
	s.Conn.WriteToUDP(data, client)
}

func (s *Server) RoomWatcher(room *ChatRoom) {
	for {
		select {
		case checkMe := <-room.upTimeQueue:
			fmt.Println("Pinging ", checkMe)
			if room.clients[checkMe].lastSeen.Before(time.Now().Add(-60*time.Second)) ||
				room.clients[checkMe].checkCount > 5 {
				fmt.Println(checkMe, " disconnected")
				delete(room.clients, checkMe)
			} else {
				fmt.Println("Fire ping ", checkMe)
				fmt.Println("Last seen ", room.clients[checkMe].lastSeen)
				s.Ping(checkMe)
				room.clients[checkMe].checkCount += 1
				go func() {
					t := time.NewTimer(10 * time.Second)
					<-t.C
					room.upTimeQueue <- checkMe
				}()
			}
		case checkMe := <-room.pongQueue:
			if room.clients[checkMe] != nil {
				room.clients[checkMe].lastSeen = time.Now()
				room.clients[checkMe].checkCount = 0
			}
		}
	}
}

func (s *Server) ClientConnectToRoom(message Message) {
	var room RoomMessage
	err := room.DecodeMessage(message.RawData())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Request for room %s\n", room.Room)
	if s.Rooms[room.Room] == nil {
		s.Rooms[room.Room] = &ChatRoom{make(map[*net.UDPAddr]*Uptime), make(chan *net.UDPAddr, 10), make(chan *net.UDPAddr, 10)}
		go s.RoomWatcher(s.Rooms[room.Room])
		s.AddToRoom(s.Rooms[room.Room], message.Sender())
	} else {
		s.AddToRoom(s.Rooms[room.Room], message.Sender())
	}
}

func (s *Server) Serve() {
	addressString := fmt.Sprintf("%v:%v", "", s.Port)
	ServerAddr, err := net.ResolveUDPAddr("udp", addressString)
	if err != nil {
		panic(err)
	}
	s.Conn, err = net.ListenUDP("udp", ServerAddr)
	if err != nil {
		panic(err)
	}
	defer s.Conn.Close()

	buf := bytes.NewBuffer(make([]byte, MAX_UDP_DATAGRAM))
	for {
		_, clientAddr, err := s.Conn.ReadFromUDP(buf.Bytes())
		var message Message
		message.DecodeMessage(clientAddr, buf.Bytes())
		switch message.Type() {
		case CONNECT_TO_ROOM:
			s.ClientConnectToRoom(message)
			break
		case PONG:
			for _, room := range s.Rooms {
				room.pongQueue <- clientAddr
			}
			break
		}
		fmt.Println("Received ", string(buf.Bytes()), " from ", clientAddr)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
