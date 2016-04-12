package nodeContext

const (
    NodeHello = iota
    ParentElec
    Joined
)

type NodeContext struct {
    LocalName    string
    LocalIp      string
    ParentIP     string
    ParentName   string
    State        int
}

type FailClientInfo struct {
    name     string
    IP		 string
}

func NewNodeContext() *NodeContext {
    return &NodeContext{State: NodeHello}
}

func (nodeContext *NodeContext) SetLocalName(name string) {
    nodeContext.LocalName = name
}

func (nodeContext *NodeContext) SetState(state int) {
    nodeContext.State = state
}
