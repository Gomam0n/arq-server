package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
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

	clientData := make([]byte, 17)
	var sendData []byte
	udpAddr, err := getUDPAddr(0)

	fmt.Println(udpAddr.Port, udpAddr.IP, udpAddr.Network())

	listenConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return
	}
	defer listenConn.Close()
	portUse := listenConn.LocalAddr().String()[strings.Index(listenConn.LocalAddr().String(), ":")+1:]
	var ch1 = make(chan []byte, 2)

	content, err := ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	go ListenClient(listenConn, clientData, ch1)

	sendData = []byte(portUse)
	conn.Write(sendData)
	fmt.Println(conn.RemoteAddr())
	fmt.Println("use port ", portUse, " to communicate with ", conn.RemoteAddr())

	sequenceNumber := 0
	i := 0
	//timeout := true
	fmt.Println(sequenceNumber, i)
	timer := time.NewTimer(time.Millisecond * 300)
	for {
		// timer may be not active, and fired
		if !timer.Stop() {
			select {
			case <-timer.C: //try to drain from the channel
			default:
			}
		}
		timer.Reset(time.Millisecond * 300)
		select {
		case clientData := <-ch1:
			fmt.Println("reveive ", string(clientData), " from ", conn.RemoteAddr())
			if clientData[0] == 'A' && clientData[1] == 'C' && clientData[2] == 'K' && clientData[3] == byte(sequenceNumber+'0') {
				if i > len(content) {
					break
				} else {
					sendData = nil
					sequenceNumber = 1 - sequenceNumber
					sendData = append(sendData, byte(sequenceNumber+'0'))
					if i+16 > len(content) {
						sendData = append(sendData, content[i:]...)
					} else {
						sendData = append(sendData, content[i:i+16]...)
					}
					i += 16
					conn.Write(sendData)
					fmt.Println("send next packet to ", conn.RemoteAddr())
					continue
				}
			}
		case <-timer.C:
			fmt.Println(time.Now(), ":timer expired")
			conn.Write(sendData)
			continue
		}
	}
	/*
		for{
			if clientData[0] == 'A' &&clientData[1] == 'C' &&clientData[2] == 'K'&& clientData[3] == byte(sequenceNumber+'0') {
				clientData[0],clientData[1],clientData[2],clientData[3] = '0','0','0','0'
				if i>len(content){
					break
				}else{
					sendData = nil
					sequenceNumber = ^sequenceNumber
					sendData= append(sendData,byte(sequenceNumber+'0') )
					sendData = append(sendData,content[i:i+16]...)
					i+=16
					conn.Write(sendData)
					timer.Stop()
					timer.Reset(time.Millisecond*300)
				}
			}else if timeout == true{
				conn.Write(sendData)
				timer.Reset(time.Millisecond*300)
			}

		}
	*/

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
	// error?
	conn, err = net.Dial("udp", clientAddr)
	if err != nil {
		fmt.Println("The connection with " + addr.IP.String() + " has error: " + err.Error())
	}
	return
}

func ListenClient(conn *net.UDPConn, clientData []byte, ch1 chan []byte) {
	fmt.Println(conn.LocalAddr())
	for {
		len, _, err := conn.ReadFromUDP(clientData)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(string(clientData[:len]))
		ch1 <- clientData[:len]
		clientData = make([]byte, 17)
	}
}
func getUDPAddr(port int) (udpAddr *net.UDPAddr, err error) {
	serverAddress := IP + ":" + strconv.Itoa(port)
	udpAddr, err = net.ResolveUDPAddr("udp", serverAddress)
	return
}
func count(cnt *int) {
	timer := time.NewTimer(time.Millisecond)
	for {
		<-timer.C
		timer.Reset(time.Millisecond)
		*cnt++
		fmt.Println(*cnt)
	}
}
