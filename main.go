package main

import (

	//"bufio"
	//"fmt"
	//"os"
	//messagePasser "./messagePasser"
	"./node"
	superNode "./superNode"
	//"fmt"
	dns "./dnsService"
	"flag"
	//"os"
	"fmt"
)

const (
	bootstrap_dns = "DS.supernodes.com"
)

func start() {
	me := flag.String("class", "node", "the identity of the current node")

	flag.Parse()

	if *me == "node" {
		helloIP := dns.GetAddr(bootstrap_dns)[0]
		node.Start()
		node.NodeJoin(helloIP)
	} else {
		superNode.Start()
	}
}

func main() {

	// Start reading from the receive message queue
	//mp := messagePasser.NewMessagePasser("cheng");
	//go mp.Receive()
	// Start listening
	//go mp.Listen("bob")
	/*
		reader := bufio.NewReader(os.Stdin)
		for {
			text, _ := reader.ReadString('\n')     // send to socket
			// Send the message, trim the last \n from input
			go mp.Send(messagePasser.NewMessage("p2plive", "hello", text[:len(text)-1]))
			fmt.Println("Send Message " + text)
		}

	*/

	start()
}
