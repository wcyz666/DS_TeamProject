package dht

/* Defines common data types used in DHT package */

import (
	MP "../messagePasser"
	"math/big"
)

/* Planning to use MD5 to generate the Hash. Hence 128 bit */
const HASH_KEY_SIZE = 128
const MAX_KEY = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"

type Node struct {
	IpAddress string
}

type LeafTable struct {
	/* currently leaf table is of size 2 and stores only previous and next nodes */
	prevNode *Node
	nextNode *Node
}

/* Initial DHT version contains details of next and previous nodes */
type DHTNode struct {
	/* We interpret Hash value as a hexadecimal stream so each digit is 4 bit long */
	prefixForwardingTable [HASH_KEY_SIZE/4][HASH_KEY_SIZE/4]*Node
	leafTable LeafTable
	/* TODO Can we use concurrent maps as described in https://github.com/streamrail/concurrent-map */
	hashTable             	map[string][]MemberShipInfo
	mp                    	*MP.MessagePasser
	nodeKey               	string
	prevNodeNumericKey    	*big.Int
	curNodeNumericKey      	*big.Int
	/* When a super node is already involved in a ring update (i.e.) transferring portion
	 * of its hash table as part of new node joining, flag is set. This is to avoid
	 * multiple join operations happening at same super node at the same time which
	 * may result in incorrect splitting of hash table among the super nodes in the ring */
	isRingUpdateInProgress 	bool
}

type DHTService struct {
	DhtNode *DHTNode
}


/* TODO: Need to revisit data structures. Temporarily adding superNodeIp*/
type MemberShipInfo struct {
	SuperNodeIp string
}

/* Looks like Go does not support enums. So have to define all status related constants here (even though it is bit ugly)*/
const (
	/* Generic Status*/
	SUCCESS = iota
	FAILURE

	/*DHT Data management related status */
	KEY_NOT_PRESENT
	SUCCESS_ENTRY_OVERWRITTEN

	/*Ring management related status */
	// If successor node is already involved in another join procedure
	JOIN_IN_PROGRESS_RETRY_LATER
)