package node

import (
    dns "../dnsService"
    MP "../messagePasser/"
    "fmt"
)

const (
    bootstrap_dns = "DS.supernodes.com"
)

var mp *MP.MessagePasser
var localName string
var parentIP string

func NodeJoin(IP string) {
    helloMsg := MP.NewMessage(IP, "join", "hello, my name is Bay Max, you personal healthcare companion")
    mp.Send(helloMsg)
    echoMsg := <- mp.Messages["ack"]
    fmt.Println(echoMsg)
}

func setLocalName(name string) {
    localName = name
}

func NodeStart()  {
    setLocalName("bob")

    mp = MP.NewMessagePasser(localName);
    helloIP := dns.GetAddr(bootstrap_dns)[0]

    NodeJoin(helloIP)
}
