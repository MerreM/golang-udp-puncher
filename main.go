// UDPPuncher project main.go
package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"net"
	"os"
)

/* A Simple function to verify error */
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

type UDPLinker struct {
	server *net.UDPConn
	pools  map[string]*net.UDPAddr
}

func Serve(port *int) {
	addressString := fmt.Sprintf("%v:%v", "", *port)
	ServerAddr, err := net.ResolveUDPAddr("udp", addressString)
	CheckError(err)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buf := make([]byte, 32)
	serverInternal := &UDPLinker{ServerConn, make(map[string]*net.UDPAddr)}

	for {
		n, returnAddr, err := ServerConn.ReadFromUDP(buf)
		poolString := string(buf)
		fmt.Printf("Request for pool %s\n", poolString)
		if serverInternal.pools[poolString] == nil {
			serverInternal.pools[poolString] = returnAddr
		} else {
			fmt.Printf("Handshake begins\n")
			otherAddr := serverInternal.pools[poolString]
			returnBuffer := bytes.NewBuffer(make([]byte, 0))
			otherBuffer := bytes.NewBuffer(make([]byte, 0))
			returnEncoder := gob.NewEncoder(returnBuffer)
			otherEncoder := gob.NewEncoder(otherBuffer)
			err = returnEncoder.Encode(&returnAddr)
			if err != nil {
				panic(err)
			}
			err = otherEncoder.Encode(&otherAddr)
			if err != nil {
				panic(err)
			}
			ServerConn.WriteToUDP(otherBuffer.Bytes(), returnAddr)
			ServerConn.WriteToUDP(returnBuffer.Bytes(), otherAddr)
			fmt.Printf("Addresses Sent\n")
			delete(serverInternal.pools, poolString)

		}
		fmt.Println("Received ", string(buf[0:n]), " from ", returnAddr)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func clientContiniousRead(conn *net.UDPConn, errorChan chan error) {
	buf := bytes.NewBuffer(make([]byte, bytes.MinRead))
	for {
		n, sender, err := conn.ReadFromUDP(buf.Bytes())
		if n > 0 && err == nil {
			fmt.Printf("%v says \"%v\"", sender, buf.String())
		} else if err != nil {
			errorChan <- err
		}
		buf.Reset()
	}
}

func clientContiniousWrite(conn *net.UDPConn, partner *net.UDPAddr, errorChan chan error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')
		n, err := conn.WriteToUDP([]byte(text), partner)
		if n > 0 && err == nil {
			fmt.Printf("Sent to %v", partner)
		} else if err != nil {
			errorChan <- err
		}
		reader.Reset(os.Stdin)
	}
}

func Client(port *int) {
	addressString := fmt.Sprintf("%v:%v", "", *port)
	ServerAddr, err := net.ResolveUDPAddr("udp", addressString)
	CheckError(err)
	ClientAddr, err := net.ResolveUDPAddr("udp", ":")
	CheckError(err)
	conn, err := net.ListenUDP("udp", ClientAddr)
	fmt.Println(conn.LocalAddr())
	if err != nil {
		panic(err)
	}
	// Continous Read & Writes.
	conn.WriteTo([]byte("Hello"), ServerAddr)
	fmt.Println("Join room")
	partnerDecoder := gob.NewDecoder(conn)
	partner := net.UDPAddr{}
	fmt.Println("Reading")
	partnerDecoder.Decode(&partner)
	fmt.Printf("Attempting to connect to %v\n", partner)
	errorChannel := make(chan error)
	if err != nil {
		panic(err)
	}
	//	conn, err = net.ListenUDP("udp", ClientAddr)
	fmt.Printf("Listening on...%v\n", conn.LocalAddr())
	go clientContiniousRead(conn, errorChannel)
	go clientContiniousWrite(conn, &partner, errorChannel)
	panic(<-errorChannel)
}

func main() {
	serverPort := flag.Int("s", 0, "Listen mode. Specify port")
	clientConnect := flag.Int("c", 0, "Send mode. Specify port")
	flag.Parse()
	if serverPort != nil && *serverPort != 0 {
		Serve(serverPort)
		return
	} else if clientConnect != nil && *clientConnect != 0 {
		Client(clientConnect)
		return
	}
	flag.Usage()
}
