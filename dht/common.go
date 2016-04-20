package dht

/* Defines common data types used in DHT package */

import (
	MP "../messagePasser"
	"math/big"
)

const (
	DHT_WAIT_FOR_JOIN_RESPONSE = iota
	DHT_JOIN_IN_PROGRESS
	DHT_JOINED
	DHT_JOIN_FAILED_MAX_ATTEMPTS
	DHT_JOIN_FAILED
	DHT_RING_REPAIR_IN_PROGRESS
)

/* Planning to use MD5 to generate the Hash. Hence 128 bit */
const HASH_KEY_SIZE = 128
const MAX_KEY = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"
const NEIGHBOURHOOD_DISTANCE = 2 // Neighbourhood distance on each direction
const PERIODIC_LEAF_TABLE_REFRESH_DURATION = 120
const REPLICATION_FACTOR = 2 // 1 Primary and 1 Backup
const REPLICATION_UPDATE_RESPONSE_TIMER_EXPIRY = 6

type Node struct {
	IpAddress string
	Name      string
	Key       string
}

type LeafTable struct {
	/* prev and current nodes maintained separately for easy access*/
	prevNode     *Node
	nextNode     *Node
	/* List of previous nodes in the neighbourhood */
	PrevNodeList [] Node
	/* List of next nodes in the neighbourhood */
	NextNodeList [] Node
}

/* Initial DHT version contains details of next and previous nodes */
type DHTNode struct {
	/* We interpret Hash value as a hexadecimal stream so each digit is 4 bit long */
	prefixForwardingTable  [HASH_KEY_SIZE/4][HASH_KEY_SIZE/4]*Node
	leafTable              LeafTable
	/* TODO Can we use concurrent maps as described in https://github.com/streamrail/concurrent-map */
	hashTable              map[string][]MemberShipInfo
	mp                     *MP.MessagePasser
	NodeKey                string
	NodeName               string
	IpAddress              string
	prevNodeNumericKey     *big.Int
	curNodeNumericKey      *big.Int
	/* When a super node is already involved in a ring update (i.e.) transferring portion
	 * of its hash table as part of new node joining, flag is set. This is to avoid
	 * multiple join operations happening at same super node at the same time which
	 * may result in incorrect splitting of hash table among the super nodes in the ring */
	IsRingUpdateInProgress bool
	DhtState               int
	curReplicaCount        int // This includes primary as well in the replica count
}

type DHTService struct {
	DhtNode *DHTNode
}


/* TODO: Need to revisit data structures. Temporarily adding superNodeIp*/
type MemberShipInfo struct {
	StreamerIp string
	StreamerName string
}

/* Looks like Go does not support enums. So have to define all status related constants here (even though it is bit ugly)*/
const (
	/* Generic Status*/
	SUCCESS = iota
	FAILURE
	INVALID_INPUT_PARAMS

	/*DHT Data management related status */
	KEY_NOT_PRESENT
	SUCCESS_ENTRY_OVERWRITTEN

	/*DHT Replication Management related status*/
	SUCCESS_REDUCED_REPLICATION  // If operation was successful only in a subset of replicas

	/*Ring management related status */
	// If successor node is already involved in another join procedure
	JOIN_IN_PROGRESS_RETRY_LATER
)