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
	Successor        Node
}

type JoinComplete struct {
	Status int
	Key    string
}

type JoinNotify struct {
	Key string
}


type DataOperationRequest struct {
	Key 		string			/* Key entry */
	Data 		MemberShipInfo		/* Data entry, used for Update operations only */
	Remove 		bool			/* indicate removal in Update operation */
	Add 		bool			/* indicate append in Update operation */
	OriginIpAddress	string			/* Source IP address, used for response message */
	OriginName	string			/* Dest IP address, used for response message */
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

type DataOperationResponse struct {
	Status	int				/* status of the requested data operation */
	Data []MemberShipInfo			/* data response from GetData operation */
}

type NeighbourhoodDiscoveryMessage struct {
	ResidualHopCount   int
	TraversalDirection int
	NodeList           [] Node
	OriginIpAddress    string
	OriginName         string
}

type RingRepairRequest struct {
	Key    string
}

type RingRepairResponse struct {
	Status int
	Key    string
}