package mylib

import (
	"bufio"
	"net"
	"fmt"
	dns "../dnsService"
)
const (
	localPort = "6666"
)

/*
This is a simple message passing object.
Refer to the sample here: https://gist.github.com/drewolson/3950226
 */


var Incoming = make(chan Message)
var connections *Connections

type Client struct {
	name string
	incoming chan Message
	outgoing chan Message // Act as a sending message queue
	reader   *bufio.Reader
	writer   *bufio.Writer
}


//The go routine here, keep waiting messages from connected client Blocking.
func (client *Client) Read() {
	for {
		line, err := client.reader.ReadBytes('\xfe')
		if err != nil {
			fmt.Println("Client " + client.name + " disconneted!")
			return
		}
		msg := new(Message)
		msg.Deserialize(line)
		Incoming <- *msg  // Since there is only one socket, directly put all the received
				  // msgs into the global receiving channel (message queue)
	}
}

//go routine: Keep writing
func (client *Client) Write() {
	for {
		msg := <-client.outgoing
		seri, _ := msg.Serialize()
		client.writer.Write(seri)
		client.writer.Flush()
	}
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

// Constructor
func NewClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	client := &Client{
		incoming: make(chan Message),
		outgoing: make(chan Message),
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
		NewClient(conn)
		//clientName, _ := bufio.NewReader(conn).ReadString('\n')
		//fmt.Println("Client : " + clientName + " connected!")
		//client.name = clientName
		//connect.clients[clientName] = client

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

/*
A blocking version of receive function
Can be changed in future
 */
func Receive() {
	for {
		data := <-Incoming
		fmt.Println("Received data: ")
		fmt.Println(data)
	}
}

/*
Send a message
 */
func Send(msg Message) {
	msg.SetSrc(connections.localname)
	dest := msg.GetDest()
	if client, ok := connections.clients[dest]; ok {
		// Already contains the dest peer
		client.outgoing <- msg
	}else{
		// Not contain
		// Try connecting to the peer
		conn, _ := net.Dial("tcp", dns.GetAddr(dest) + ":" + localPort)
		client := NewClient(conn)
		client.name = dest
		connections.clients[dest] = client
		client.outgoing <- msg
	}
	
}
