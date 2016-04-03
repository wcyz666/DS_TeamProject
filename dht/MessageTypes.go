package dht

const (
	SUCCESS = iota
	FAILURE
	NEW_SUCCESSOR
)

type JoinRequest struct {
	key                  string
}

type JoinResponse struct {
	status                int
	hashTable             map[string][]MemberShipInfo
	predecessor           Node
	newSuccessorNode      Node // Used as re-direction mechanism when key is no longer managed by this node
}

type JoinComplete struct {
	status                int
}

type JoinNotify struct {
	key                  string
}

type SuccessorInfoReq struct {
	key                  string
}

type SuccessorInfoRes struct {
	status               int
	node                 Node
}