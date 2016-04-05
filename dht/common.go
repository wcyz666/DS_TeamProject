package dht

/* Defines common data types used in DHT package */

import (
	MP "../messagePasser"
)

/* Planning to use MD5 to generate the Hash. Hence 128 bit */
const HASH_KEY_SIZE = 128

type Node struct {
	nodeKey string /*TODO we may need it for optimizing the node to be stored in this slot */
	IpAddress string
	port int
}

/* Initial DHT version contains details of next and previous nodes */
type DHT struct {
	/* We interpret Hash value as a hexadecimal stream so each digit is 4 bit long */
	prefixForwardingTable [HASH_KEY_SIZE/4][HASH_KEY_SIZE/4]Node
	/* TODO Can we use concurrent maps as described in https://github.com/streamrail/concurrent-map */
	hashTable             map[string][]MemberShipInfo
	mp                    *MP.MessagePasser
	nodeKey               string
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
	SUCCESSOR_REDIRECTION
)