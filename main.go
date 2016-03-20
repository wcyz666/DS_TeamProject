package main

import (
	"bufio"
	"fmt"
	"os"
	mp "./messagePasser"
)

/**
This is a file to test the message passer
 */

func main() {
	// Start reading from the receive message queue
	go mp.Receive()
	// Start listening
	go mp.Listen("bob")

	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')     // send to socket
		// Send the message, trim the last \n from input
		go mp.Send(mp.NewMessage("alice", text[:len(text)-1]))
		fmt.Println("Send Message " + text)
	}
}