package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
)

var IP = "127.0.0.1"
var serverPort = 8888
var filename = "shakespeare.txt"

func main() {

	udpAddr, err := getUDPAddr(serverPort)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	for {

		data := make([]byte, 16)
		len, addr, err := conn.ReadFromUDP(data)
		if err != nil {
			fmt.Println(err)
			continue
		}
		go transport(addr, string(data[:len]))
	}
}
func transport(addr *net.UDPAddr, data string) {
	fmt.Println("Connection from " + addr.String())

	conn, err := DialClient(addr, data)
	if err != nil {
		fmt.Println(err)
		return
	}
	clientData := make([]byte, 16)
	fmt.Println(clientData)

	udpAddr, err := getUDPAddr(0)
	listenConn, err := net.ListenUDP("udp", udpAddr)
	if err == nil {
		return
	}
	defer listenConn.Close()

	ack := "Ready to sent"
	conn.Write([]byte(ack))

	// the first re-transmission logic

	content, err := ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < len(content); i += 16 {
		if i+16 < len(content) {
			conn.Write(content[i : i+16])
		} else {
			conn.Write(content[i:])
		}
	}
	return
}

func ReadFile(name string) (content []byte, err error) {
	file, err := os.OpenFile(name, os.O_RDONLY, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	content, err = ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	return
}

func DialClient(addr *net.UDPAddr, data string) (conn net.Conn, err error) {
	port, err := strconv.ParseUint(data, 10, 16)
	if err != nil {
		fmt.Println("Parse unsigned int error: " + err.Error())
		return
	}
	if port < 1024 {
		fmt.Println(strconv.Itoa(addr.Port) + ": the port should be 1024~65535")
		return
	}
	clientAddr := addr.IP.String() + ":" + strconv.Itoa(int(port))
	conn, err = net.Dial("udp", clientAddr)
	if err != nil {
		fmt.Println("The connection with " + addr.IP.String() + " has error: " + err.Error())
	}
	return
}

func ListenClient(addr *net.UDPAddr, clientData *[]byte) {

}
func getUDPAddr(port int) (udpAddr *net.UDPAddr, err error) {
	serverAddress := IP + ":" + strconv.Itoa(port)
	udpAddr, err = net.ResolveUDPAddr("udp", serverAddress)
	return
}
