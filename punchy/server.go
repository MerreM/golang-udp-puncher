package punchy

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
)

func Serve(port *int) {
	addressString := fmt.Sprintf("%v:%v", "", *port)
	ServerAddr, err := net.ResolveUDPAddr("udp", addressString)
	if err != nil {
		panic(err)
	}
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		panic(err)
	}
	defer ServerConn.Close()

	buf := make([]byte, 32)
	pools := make(map[string]*net.UDPAddr)

	for {
		n, returnAddr, err := ServerConn.ReadFromUDP(buf)
		poolString := string(buf)
		fmt.Printf("Request for pool %s\n", poolString)
		if pools[poolString] == nil {
			pools[poolString] = returnAddr
		} else {
			fmt.Printf("Handshake begins\n")
			otherAddr := pools[poolString]
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
			delete(pools, poolString)

		}
		fmt.Println("Received ", string(buf[0:n]), " from ", returnAddr)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
