package superNodeContext

import (
    "fmt"
    LNS "../../localNameService"
    DNS "../../dnsService"
)

type SuperNodeContext struct {
    LocalName string
    IP        string
    Nodes     map[string]*nodeInfo
}

func (sc *SuperNodeContext) GetAllChildrenName() []string {
    names := make([]string, 0, len(sc.Nodes))
    for name := range(sc.Nodes){
        names = append(names, name)
    }
    return names
}

func (sc *SuperNodeContext) GetNodeCount() int {
    return len(sc.Nodes)
}

func (sc *SuperNodeContext) GetIPByName(nodeName string) string {
    return sc.Nodes[nodeName].IP;
}


func NewSuperNodeContext() (* SuperNodeContext) {
    nodes := make(map[string]*nodeInfo)
    IP, _ := DNS.ExternalIP()
    return &SuperNodeContext{Nodes : nodes, IP: IP, LocalName: LNS.GetLocalName()}
}

func (sc *SuperNodeContext) AddNode(nodeName string, msgIP string)  {
    fmt.Printf("Supernode Context: add node %s\n", nodeName)
    sc.Nodes[nodeName] = &nodeInfo{isLive: true, IP: msgIP}
}

func (sc *SuperNodeContext) RemoveNodes(nodeName string)  {
    fmt.Printf("Supernode Context: remove node %s\n", nodeName)
    delete(sc.Nodes, nodeName)
}

func (sc *SuperNodeContext) SetAlive(nodeName string, nodeIP string) {
    //fmt.Printf("Supernode Context: set alive node %s\n", nodeName)
    _, exists := sc.Nodes[nodeName]
    if (exists == false) {
        sc.AddNode(nodeName, nodeIP)
    }
    sc.Nodes[nodeName].isLive = true
}

func (sc *SuperNodeContext) ResetState() {
    for _, value := range sc.Nodes {
        value.isLive = false
    }
}

func (sc *SuperNodeContext) CheckDead() (hasDead bool, deadNodes []string) {

    hasDead = false
    for name, value := range sc.Nodes {
        if value.isLive == false {
            fmt.Printf("Supernode Context: find dead node %s\n", name)
            hasDead = true
            deadNodes = append(deadNodes, name)
        }
    }

    return hasDead, deadNodes
}

type nodeInfo struct {
    isLive bool
    IP string
}
