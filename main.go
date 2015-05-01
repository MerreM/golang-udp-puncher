// UDPPuncher project main.go
package main

import (
	"UDPPuncher/punchy"
	"flag"
)

func main() {
	serverPort := flag.Int("s", 0, "Listen mode. Specify port")
	clientConnect := flag.Int("c", 0, "Send mode. Specify port")
	flag.Parse()
	if serverPort != nil && *serverPort != 0 {
		punchy.Serve(serverPort)
		return
	} else if clientConnect != nil && *clientConnect != 0 {
		client := punchy.NewClient(clientConnect)
		client.ConnectToRoom("Hello")
		return
	}
	flag.Usage()
}
