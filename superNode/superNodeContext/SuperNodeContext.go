package superNodeContext

import (
    "fmt"
    LNS "../../localNameService"
)

type SuperNodeContext struct {
    LocalName string
    nodes map[string]*nodeInfo
}

func NewSuperNodeContext() (* SuperNodeContext) {
    nodes := make(map[string]*nodeInfo)
    return &SuperNodeContext{nodes : nodes, LocalName: LNS.GetLocalName()}
}

func (sc *SuperNodeContext) AddNode(nodeName string)  {
    fmt.Printf("Supernode Context: add node %s\n", nodeName)
    sc.nodes[nodeName] = &nodeInfo{isLive: true}
}

func (sc *SuperNodeContext) RemoveNodes(nodeNames []string)  {
    for _, nodeName := range nodeNames {
        fmt.Printf("Supernode Context: remove node %s\n", nodeName)
        delete(sc.nodes, nodeName)
    }
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
}
