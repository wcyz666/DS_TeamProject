package main

import (

	"bufio"
	"fmt"
	"os"
	messagePasser "./messagePasser"
	node "./node"
)


func main() {

	// Start reading from the receive message queue
	mp := messagePasser.NewMessagePasser("bob");
	//go mp.Receive()
	// Start listening
	//go mp.Listen("bob")

	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')     // send to socket
		// Send the message, trim the last \n from input
		go mp.Send(messagePasser.NewMessage("p2plive", "hello", text[:len(text)-1]))
		fmt.Println("Send Message " + text)
	}


	node.Start()

}