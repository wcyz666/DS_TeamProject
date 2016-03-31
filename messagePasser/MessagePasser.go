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



type Client struct {
	name string
	incoming chan *Message
	outgoing chan *Message // Act as a sending message queue
	reader   *bufio.Reader
	writer   *bufio.Writer
}

/**
	Define the interfaces for all the connections alive in the current system
 */
type Connections struct {
	localname string
	clients map[string]*Client
	joins chan net.Conn
}

//The go routine here, keep waiting messages from connected client Blocking.
func (client *Client) Read(mp *MessagePasser) {
	for {
		line, err := client.reader.ReadBytes('\xfe')
		if err != nil {
			fmt.Println("Client " + client.name + " disconneted!")
			return
		}
		msg := new(Message)
		msg.Deserialize(line)

		_, exists := mp.connections.clients[msg.GetSrc()]
		if(exists != nil){
			// This is first message received from this src
			// Store this connection
			client.name = msg.GetSrc()
			mp.connections.clients[msg.GetSrc()] = client
		}

		mp.incoming <- msg  // Since there is only one socket, directly put all the received
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

func (client *Client) Listen(mp *MessagePasser) {
	go client.Read(mp)
	go client.Write()
}

// Constructor
func NewClient(connection net.Conn, mp *MessagePasser) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	client := &Client{
		incoming: make(chan *Message),
		outgoing: make(chan *Message),
		reader: reader,
		writer: writer,
	}

	client.Listen(mp)
	return client
}


// Constructor
func newConnections(localname string, mp *MessagePasser) *Connections {
	connections := &Connections{
		localname : localname,
		clients: make(map[string]*Client),
		joins: make(chan net.Conn),
	}

	go connections.Listen(mp)
	return connections
}

// Listening for new connected clients
func (connect *Connections) Listen(mp *MessagePasser){
	for {
		conn := <-connect.joins
		NewClient(conn, mp)
		//clientName, _ := bufio.NewReader(conn).ReadString('\n')
		//fmt.Println("Client : " + clientName + " connected!")
		//client.name = clientName
		//connect.clients[clientName] = client

	}
}


type MessagePasser struct{
	incoming chan *Message
	connections *Connections
	Messages map[string]chan *Message

}

func NewMessagePasser(localname string) *MessagePasser {
	mp := &MessagePasser{}
	mp.incoming = make(chan *Message)
	mp.receiveMapping()
	mp.listen(localname)
	return mp
}

// The global listening go routine
func (mp *MessagePasser)listen(localname string) {
	mp.connections = newConnections(localname, mp)
	fmt.Println("Listening on " + localPort)
	listener, _ := net.Listen("tcp", ":" + localPort)

	for {
		conn, _ := listener.Accept()
		fmt.Println("New clients joined!")
		mp.connections.joins <- conn
	}
}

/*
Organize the received messages into different channels in the map [kind][channel *Message]
Store in the Message map and To be used by the upper layer handlers
 */
func (mp *MessagePasser) receiveMapping() {
	for {
		msg := <-mp.incoming

		_, exists := mp.Messages[msg.kind]
		if (exists == nil){
			mp.Messages[msg.kind] = make(chan *Message)
		}
		mp.Messages[msg.kind] <- msg
	}
}

/*
Send a message
 */
func (mp *MessagePasser) Send(msg Message) {
	msg.SetSrc(mp.connections.localname)
	dest := msg.GetDest()
	if client, ok := mp.connections.clients[dest]; ok {
		// Already contains the dest peer
		client.outgoing <- msg
	}else{
		// Not contain
		/* TODO: Remove code. Temporary code to demonstrate adding supernodes to DNS */
		// dns.RegisterSuperNode(dest)

		// Try connecting to the peer
		addr := dns.GetAddr(dest)
		fmt.Println("Selecting first entry in the list. Address is "+ addr[0])
		conn, _ := net.Dial("tcp", addr[0] + ":" + localPort)
		client := NewClient(conn, mp)
		client.name = dest
		mp.connections.clients[dest] = client
		client.outgoing <- msg
	}
	
}
