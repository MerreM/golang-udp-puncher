// UDPPuncher project main.go
package main

import (
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

func Serve(port *int) {
	addressString := fmt.Sprintf("%v:%v", "", *port)
	ServerAddr, err := net.ResolveUDPAddr("udp", addressString)
	CheckError(err)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)
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
	fmt.Println(ClientAddr)
	conn, err := net.DialUDP("udp", ClientAddr, ServerAddr)
	if err != nil {
		panic(err)
	}
	// Continous Read & Writes.
	conn.Write([]byte("Hello"))

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
