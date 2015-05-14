package punchy

import (
	"fmt"
	"log"
	"net"
	"os"
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

type RemoteClient struct {
	address   *net.UDPAddr
	sharedKey [32]byte
	Uptime
}

type Uptime struct {
	lastSeen   time.Time
	checkCount int
}

type ChatRoom struct {
	clients     map[string]*RemoteClient
	upTimeQueue chan *net.UDPAddr
	pongQueue   chan *net.UDPAddr
}

var serverLogger = log.New(os.Stdout, "server: ", log.LstdFlags|log.Lshortfile)

func NewServer(port *int) Server {
	return Server{*port, nil, make(map[string]*ChatRoom)}
}

func (s *Server) UpdateRoomList(roomName string, room *ChatRoom, client *net.UDPAddr) {
	roomList := RoomListMessage{}
	roomList.Length = 0
	if len(room.clients) > 0 {
		roomList.Length = uint16(len(room.clients) - 1)
	}
	roomList.Room = roomName
	roomList.Addresses = make([]net.UDPAddr, roomList.Length)
	count := 0
	for _, other := range room.clients {
		if other.address.String() == client.String() {
			continue
		}
		roomList.Addresses[count] = *other.address
		count++
	}
	raw, err := roomList.RawMessage()
	if err != nil {
		panic(err)
	}
	message := &Message{raw, ROOM_LIST, false, uint16(len(raw.Data))}
	data, err := message.EncodeMessage()
	if err != nil {
		panic(err)
	}
	s.Conn.WriteTo(data, client)
	serverLogger.Println("Room list sent")
	if err != nil {
		panic(err)
	}
}

func (s *Server) AddToRoom(roomName string, room *ChatRoom, client *RemoteClient) {
	serverLogger.Println("Adding client to room", client.address.String())
	room.clients[client.address.String()] = client
	room.upTimeQueue <- client.address
	serverLogger.Printf("Handshake begins\n")
	for _, client := range room.clients {
		go s.UpdateRoomList(roomName, room, client.address)
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
			if room.clients[checkMe.String()].lastSeen.Before(time.Now().Add(-60*time.Second)) ||
				room.clients[checkMe.String()].checkCount > 5 {
				serverLogger.Println(checkMe, " disconnected")
				delete(room.clients, checkMe.String())
			} else {
				serverLogger.Println("Last seen ", room.clients[checkMe.String()].lastSeen)
				s.Ping(checkMe)
				room.clients[checkMe.String()].checkCount += 1
				go func() {
					t := time.NewTimer(10 * time.Second)
					<-t.C
					room.upTimeQueue <- checkMe
				}()
			}
		case checkMe := <-room.pongQueue:
			serverLogger.Println("Room:", room.clients)
			if room.clients[checkMe.String()] != nil {
				room.clients[checkMe.String()].lastSeen = time.Now()
				room.clients[checkMe.String()].checkCount = 0
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
	serverLogger.Printf("Request for room %s\n", room.Room)
	if s.Rooms[room.Room] == nil {
		s.Rooms[room.Room] = &ChatRoom{make(map[string]*RemoteClient), make(chan *net.UDPAddr, 10), make(chan *net.UDPAddr, 10)}
		go s.RoomWatcher(s.Rooms[room.Room])
	}
	remoteClient := RemoteClient{message.Sender(), *new([32]byte), Uptime{time.Now(), 0}}
	s.AddToRoom(room.Room, s.Rooms[room.Room], &remoteClient)
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

	for {
		buf := make([]byte, MAX_UDP_DATAGRAM)
		n, clientAddr, err := s.Conn.ReadFromUDP(buf)
		var message Message
		message.DecodeMessage(clientAddr, buf[:n])
		switch message.Type() {
		case CONNECT_TO_ROOM:
			s.ClientConnectToRoom(message)
			break
		case PONG:
			for _, room := range s.Rooms {
				serverLogger.Println("Got pong from ", clientAddr)
				room.pongQueue <- clientAddr
			}
			break
		}
		//		serverLogger.Println("Received ", string(buf[:n]), " from ", clientAddr)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
