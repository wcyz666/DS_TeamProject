package nodeContext

type NodeContext struct {
    LocalName string
    ParentIP string
}

func (nodeContext *NodeContext) SetLocalName(name string) {
    nodeContext.LocalName = name
}