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
	serverAddress := IP + ":" + strconv.Itoa(serverPort)
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddress)
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
		/*
			fmt.Println(rAddr.IP)
			fmt.Println(rAddr.Port)
			fmt.Println(conn.RemoteAddr())
			fmt.Println(rAddr.Network())
			if err != nil {
				fmt.Println(err)
				continue
			}

			strData := string(data[:len])
			fmt.Println("Received:", strData)

			upper := strings.ToUpper(strData)
			_, err = conn.WriteToUDP([]byte(upper), rAddr)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("Send:", upper)
		*/
	}
}
func transport(addr *net.UDPAddr, data string) {
	fmt.Println(addr)
	port, err := strconv.ParseUint(data, 10, 16)
	if err != nil {
		fmt.Println("Parse unsigned int error: " + err.Error())
		return
	}
	if port < 1024 {
		fmt.Println(strconv.Itoa(addr.Port) + ": the port should be 1024~65535")
	}

	clientAddr := addr.IP.String() + ":" + strconv.Itoa(int(port))
	conn, err := net.Dial("udp", clientAddr)
	if err != nil {
		fmt.Println("The connection with " + addr.IP.String() + " has error: " + err.Error())
		return
	}
	fmt.Println(conn)

	file, err := os.OpenFile(filename, os.O_RDONLY, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
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
