package main

import (
	"bufio"
	"net"
	"fmt"
	"os"
	mylib "./mylib"
)
const (
	localPort = "6666"
)

var Incoming = make(chan string)
var dnsMap = make(map[string]string)
var connections *Connections

type Client struct {
	name string
	incoming chan string
	outgoing chan string
	reader   *bufio.Reader
	writer   *bufio.Writer
}

func (client *Client) Read() {
	for {
		line, err := client.reader.ReadString('\n')
		fmt.Println("read " + line)
		if err != nil {
			fmt.Println("Client " + client.name + " disconneted!")
			return
		}
		Incoming <- line
	}
}

func (client *Client) Write() {
	for {
		data := <-client.outgoing
		fmt.Println("out " + data)
		client.writer.WriteString(data)
		client.writer.Flush()
	}
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

func NewClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	client := &Client{
		incoming: make(chan string),
		outgoing: make(chan string),
		reader: reader,
		writer: writer,
	}

	client.Listen()

	return client
}

/**
	Define the interfaces for all the connections alive in the current system
 */
type Connections struct {
	localname string
	clients map[string]*Client
	joins chan net.Conn
}

// Constructor
func NewConnections(localname string) *Connections {
	connections := &Connections{
		localname : localname,
		clients: make(map[string]*Client),
		joins: make(chan net.Conn),
	}

	go connections.Listen()
	return connections
}

// Listening for new connected clients
func (connect *Connections) Listen(){
	for {
		conn := <-connect.joins
		client := NewClient(conn)
		clientName, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Println("Client : " + clientName + " connected!")
		client.name = clientName
		connect.clients[clientName] = client

	}
}

// The global listening go routine
func Listen(localname string) {
	connections = NewConnections(localname)
	fmt.Println("Listening on " + localPort)
	listener, _ := net.Listen("tcp", ":" + localPort)

	for {
		conn, _ := listener.Accept()
		fmt.Println("New clients joined!")
		connections.joins <- conn
	}
}

func Receive() {
	for {
		data := <-Incoming
		fmt.Println("Received data: " + data)
	}
}

func Send(msg string, dest string) {
	if client, ok := connections.clients[dest]; ok {
		// Already contains the dest client
		client.outgoing <- msg
	}else{
		// Not contain
		conn, _ := net.Dial("tcp", dnsMap[dest] + ":" + "6666")
		client := NewClient(conn)
		client.name = dest
		connections.clients[dest] = client
		client.writer.WriteString(dest)
		client.writer.Flush()
		client.outgoing <- msg
	}


}

func main() {
	dnsMap["alice"] = "127.0.0.1"
	dnsMap["bob"] = "127.0.0.1"

	go Receive()
	go Listen("alice")

	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')     // send to socket
		go Send(text, "bob")
		fmt.Println("Send Message " + text)
	}





}