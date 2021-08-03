package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"
)

// IP local IP address
var IP = "0.0.0.0"

// port used to listening requests
var serverPort = 8888
var filename = "shakespeare.txt"

func main() {

	// listen to the port
	udpAddr, err := getUDPAddr(serverPort)
	checkError(err)
	conn, err := net.ListenUDP("udp", udpAddr)
	checkError(err)

	defer conn.Close()

	for {
		data := make([]byte, 16)
		len, addr, err := conn.ReadFromUDP(data)
		if err != nil {
			fmt.Println(err)
			continue
		}
		// transport file to client
		go transport(addr, string(data[:len]))
	}
}

// The main transportation function
// addr contains the address of client, data contains the port of client
func transport(addr *net.UDPAddr, data string) {
	fmt.Println("Connection from " + addr.String())

	conn, err := DialClient(addr, data)

	if err != nil {
		fmt.Println(err)
		return
	}

	// receive ACK from client
	clientData := make([]byte, 17)

	var sendData []byte
	// randomly choose a port number to receive ACK from client
	udpAddr, err := getUDPAddr(0)

	listenConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listenConn.Close()

	// a channel, if receives ACK from the client, the data will
	// be pushed into the channel
	var ch1 = make(chan []byte, 2)

	// read the file
	content, err := ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	// use a goroutine to listen to ACK from client
	go ListenClient(listenConn, clientData, ch1)

	// tell client the port number
	portUse := getPortFromConn(listenConn.LocalAddr().String())
	sendData = []byte(portUse)
	conn.Write(sendData)

	fmt.Println("use port ", portUse, " to communicate with ", conn.RemoteAddr())

	sequenceNumber := 0
	// byte number to be transmitted to client
	i := 0

	// count the timeout rate
	timeout, newpac := 0, 0

	// setup the timer
	timer := time.NewTimer(time.Millisecond * 300)
	for {
		// timer may be not active, and fired
		if !timer.Stop() {
			select {
			case <-timer.C: //try to drain from the channel
			default:
			}
		}
		// reset the timer
		timer.Reset(time.Millisecond * 300)
		select {
		// if there is data from client
		case clientData := <-ch1:
			fmt.Println("reveive ", string(clientData), " from ", conn.RemoteAddr())
			// check sequence number
			if clientData[0] == 'A' && clientData[1] == 'C' && clientData[2] == 'K' && clientData[3] == byte(sequenceNumber+'0') {
				if i > len(content) {
					fmt.Println("send all data to conn.RemoteAddr()")
					fmt.Println("Packet re-transmit: ", timeout, ", total packet: ", timeout+newpac)
					return
				} else {
					// transmit the next packet
					sendData = nil
					sequenceNumber = 1 - sequenceNumber
					sendData = append(sendData, byte(sequenceNumber+'0'))
					if i+16 > len(content) {
						sendData = append(sendData, content[i:]...)
					} else {
						sendData = append(sendData, content[i:i+16]...)
					}
					i += 16
					newpac++
					conn.Write(sendData)
					continue
				}
			}
		// if timeout
		case <-timer.C:
			timeout++
			conn.Write(sendData)
			continue
		}
	}

	return
}

// ReadFile read file accounding to the filename
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

// DialClient dial the client, IP address from addr, port from data
func DialClient(addr *net.UDPAddr, data string) (conn net.Conn, err error) {
	// the port number should be 1024-65535
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

func ListenClient(conn *net.UDPConn, clientData []byte, ch1 chan []byte) {
	//fmt.Println(conn.LocalAddr())
	for {
		len, _, err := conn.ReadFromUDP(clientData)
		if err != nil {
			break
		}
		ch1 <- clientData[:len]
		clientData = make([]byte, 17)
	}
}
func getUDPAddr(port int) (udpAddr *net.UDPAddr, err error) {
	serverAddress := IP + ":" + strconv.Itoa(port)
	udpAddr, err = net.ResolveUDPAddr("udp", serverAddress)
	return
}

// parse the connection string to get port number
func getPortFromConn(conn string) (port string) {
	for i := len(conn) - 1; i >= 0; i-- {
		if conn[i] == ':' {
			port = conn[i+1:]
			break
		}
	}
	return
}
func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
