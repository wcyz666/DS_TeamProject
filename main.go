package main

import (

	"./node"
	SuperNode "./superNode"
	//"fmt"
	"flag"
        dns "./dnsService"
        config "./config"
	"fmt"
)

const (
	bootstrap_dns = "DS.supernodes.com"
)

func start(me *string) {
	if *me == "node" {
		fmt.Println("Start as node!")
		node.Start()
	} else {
		fmt.Println("Start as supernode!")
		SuperNode.Start()
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

	 clearDNS:= flag.Bool("clearDNS",false,"set if you want to clear DNS A records")
	 me := flag.String("class", "node", "the identity of the current node")

	 flag.Parse()
	 if (*clearDNS){
	     dns.ClearAddrRecords(config.BootstrapDomainName)
	 } else {
	     start(me)
	 }
}
