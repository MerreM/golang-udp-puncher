// UDPPuncher project main.go
package main

import (
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
			fmt.Println(otherBuffer.Bytes())
			ServerConn.WriteToUDP(returnBuffer.Bytes(), otherAddr)
			fmt.Print(returnBuffer.Bytes())
			fmt.Printf("Addresses Sent\n")

		}
		fmt.Println("Received ", string(buf[0:n]), " from ", returnAddr)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func Client(port *int) {
	addressString := fmt.Sprintf("%v:%v", "", *port)
	ServerAddr, err := net.ResolveUDPAddr("udp", addressString)
	CheckError(err)
	ClientAddr, err := net.ResolveUDPAddr("udp", ":")
	CheckError(err)
	conn, err := net.DialUDP("udp", ClientAddr, ServerAddr)
	fmt.Println(conn.LocalAddr())
	if err != nil {
		panic(err)
	}
	// Continous Read & Writes.
	conn.Write([]byte("Hello"))
	fmt.Println("Join room")
	partnerDecoder := gob.NewDecoder(conn)
	partner := net.UDPAddr{}
	fmt.Println("Reading")
	partnerDecoder.Decode(&partner)
	fmt.Println(partner)
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
	}
	flag.Usage()
}
