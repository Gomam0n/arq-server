package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

var IP = "127.0.0.1"
var PORT = 8888

func main() {
	fmt.Println("Hello world")
	address := IP + ":" + strconv.Itoa(PORT)
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer conn.Close()
}
