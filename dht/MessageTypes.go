package dht

/* Structure of messages exchanged between DHT nodes */

type JoinRequest struct {
	Key string
}

type JoinResponse struct {
	Status           int
	HashTable        map[string][]MemberShipInfo
	Predecessor      Node
	NewSuccessorNode Node // Used as re-direction mechanism when key is no longer managed by this node
}

type JoinComplete struct {
	Status int
	Key    string
}

type JoinNotify struct {
	Key string
}