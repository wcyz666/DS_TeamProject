package dht

/* Structure of messages exchanged between DHT nodes */

type JoinRequest struct {
	Key 			string
	OriginIpAddress string
	OriginName      string
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

/* NewEntry structures */
type CreateNewEntryRequest struct {
	Key string
	Data MemberShipInfo
}

type CreateNewEntryResponse struct {
	Status int
}

type BroadcastMessage struct {
	/* Currently used for debugging. List of Maps {NodeIpAddress: Node Key }*/
	TraversedNodesList [] Node
	OriginIpAddress string
	OriginName      string
}