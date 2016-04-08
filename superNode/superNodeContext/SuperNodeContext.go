package superNodeContext

import (
    "fmt"
    LNS "../../localNameService"
)

type SuperNodeContext struct {
    LocalName string
    nodes map[string]*nodeInfo
}

func (sc *SuperNodeContext) GetNodeCount() int {
    return len(sc.nodes)
}

func (sc *SuperNodeContext) GetIPByName(nodeName string) string {
    return sc.nodes[nodeName].IP;
}


func NewSuperNodeContext() (* SuperNodeContext) {
    nodes := make(map[string]*nodeInfo)
    return &SuperNodeContext{nodes : nodes, LocalName: LNS.GetLocalName()}
}

func (sc *SuperNodeContext) AddNode(nodeName string, msgIP string)  {
    fmt.Printf("Supernode Context: add node %s\n", nodeName)
    sc.nodes[nodeName] = &nodeInfo{isLive: true, IP: msgIP}
}

func (sc *SuperNodeContext) RemoveNodes(nodeName string)  {
    fmt.Printf("Supernode Context: remove node %s\n", nodeName)
    delete(sc.nodes, nodeName)
}

func (sc *SuperNodeContext) SetAlive(nodeName string) {
    fmt.Printf("Supernode Context: set alive node %s\n", nodeName)
    sc.nodes[nodeName].isLive = true
}

func (sc *SuperNodeContext) ResetState() {
    for _, value := range sc.nodes {
        value.isLive = false
    }
}

func (sc *SuperNodeContext) CheckDead() (hasDead bool, deadNodes []string) {

    hasDead = false
    for name, value := range sc.nodes {
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
