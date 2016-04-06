package mylib

import (
	dns "../dnsService"
	"bufio"
	"fmt"
	"net"
)

const (
	localPort = "6666"
)

/*
This is a simple message passing object.
Refer to the sample here: https://gist.github.com/drewolson/3950226
*/

type MessagePasser struct {
	Incoming    chan *Message
	connections *Connections
	Messages    map[string]chan *Message
}

type Client struct {
	name     string
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
	clients   map[string]*Client
	joins     chan net.Conn
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

		_, exists := mp.connections.clients[msg.SrcName]
		if exists == false {
			// This is first message received from this src
			// Store this connection
			client.name = msg.SrcName
			mp.connections.clients[msg.SrcName] = client
		}
		//Also save the src into the connection map
		_, exists = mp.connections.clients[msg.Src]
		if exists == false {
			mp.connections.clients[msg.Src] = client
		}

		mp.Incoming <- msg // Since there is only one socket, directly put all the received
		// msgs into the global receiving channel (message queue)
	}
}

//go routine: Keep writing
func (client *Client) Write(mp *MessagePasser) {
	for {
		msg := <-client.outgoing
		seri, _ := msg.Serialize()

		_, err := client.writer.Write(seri)
		if err != nil {
			errorMsg := NewMessage("self", mp.connections.localname, "conn_error", EncodeData(err.Error()))
			mp.Messages["error"] <- &errorMsg
		}
		client.writer.Flush()
	}
}

func (client *Client) Listen(mp *MessagePasser) {
	go client.Read(mp)
	go client.Write(mp)
}

// Constructor
func NewClient(connection net.Conn, mp *MessagePasser) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	client := &Client{
		incoming: make(chan *Message),
		outgoing: make(chan *Message),
		reader:   reader,
		writer:   writer,
	}

	client.Listen(mp)
	return client
}

// Constructor
func newConnections(localname string, mp *MessagePasser) *Connections {
	connections := &Connections{
		localname: localname,
		clients:   make(map[string]*Client),
		joins:     make(chan net.Conn),
	}

	go connections.Listen(mp)
	return connections
}

// Listening for new connected clients
func (connect *Connections) Listen(mp *MessagePasser) {
	for {
		conn := <-connect.joins
		NewClient(conn, mp)

		//client.name = clientName
		//connect.clients[clientName] = client

	}
}

func NewMessagePasser(localname string) *MessagePasser {
	mp := &MessagePasser{}
	mp.Incoming = make(chan *Message)
	mp.Messages = make(map[string]chan *Message)
	mp.connections = newConnections(localname, mp)

	go mp.receiveMapping()
	go mp.listen()

	return mp
}

// The global listening go routine
func (mp *MessagePasser) listen() {
	fmt.Println("Listening on " + localPort)
	listener, _ := net.Listen("tcp", ":"+localPort)

	for {
		conn, _ := listener.Accept()
		fmt.Println("New clients joined!")
		mp.connections.joins <- conn
	}
}

/*
   Create new mapping & channel to the messagePasser
 */
func (mp *MessagePasser) AddMapping(kind string) {
	fmt.Print("Initialized the channel: ")
	fmt.Println(kind)
	mp.Messages[kind] = make(chan *Message, 100)
}

func (mp *MessagePasser) AddMappings(kinds []string) {
	for _, kind := range kinds {
		mp.AddMapping(kind)
	}
}

/*
Organize the received messages into different channels in the map [kind][channel *Message]
Store in the Message map and To be used by the upper layer handlers
*/
func (mp *MessagePasser) receiveMapping() {
	for {
		msg := <-mp.Incoming

		_, exists := mp.Messages[msg.Kind]
		if exists == false {
			mp.AddMapping(msg.Kind)
		}
		mp.Messages[msg.Kind] <- msg
	}
}

/*
Send a message
*/
func (mp *MessagePasser) Send(msg Message)  {
	msg.SrcName = mp.connections.localname
	msg.Src, _ = dns.ExternalIP()

	dest := msg.DestName

	if _, ok := mp.connections.clients[dest]; ok == false {
		dest = msg.Dest
	}
	if client, ok := mp.connections.clients[dest]; ok {
		// Already contains the dest peer
		client.outgoing <- &msg
	} else {
		// Try connecting to the peer

		addr := dest
		conn, err := net.Dial("tcp", addr + ":" + localPort)
		if (err != nil) {
			errMsg := NewMessage("self", mp.connections.localname, "error", EncodeData(err.Error()))
			mp.Messages["error"] <- &errMsg
			return
		}
		client := NewClient(conn, mp)
		client.name = dest
		mp.connections.clients[dest] = client
		client.outgoing <- &msg
	}

}

func (mp *MessagePasser) GetNodeIpAndName()(string,string){
	nodeName := mp.connections.localname
	nodeIP, _ := dns.ExternalIP()
	return nodeIP,nodeName
}

