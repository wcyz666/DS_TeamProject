package dht

import "cmd/api/testdata/src/pkg/p1"

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


type DataOperationRequest struct {
	Key 		string			/* Key entry */
	Data 		MemberShipInfo		/* Data entry, used for Update operations only */
	Remove 		bool			/* indicate removal in Update operation */
	Add 		bool			/* indicate append in Update operation */
	OriginIpAddress	string			/* Source IP address, used for response message */
	OriginName	string			/* Dest IP address, used for response message */
}

type DataOperationResponse struct {
	Status	int				/* status of the requested data operation */
	Data []MemberShipInfo			/* data response from GetData operation */
}


