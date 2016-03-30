package main

import (
	"bufio"
	"fmt"
	"os"
	mp "./messagePasser"
	dht "./dht"
	"strconv"
)


func checkForDataPresence(hashTable dht.DHT, key string, expectedLength int){
	memberList, status := hashTable.Get(key)
	fmt.Println("status is "+ strconv.FormatBool(status))
	if (false == status){
		panic("Entry should have been present")
	}

	fmt.Println("len of member list is "+ strconv.Itoa(len(memberList)))

	for i,v := range memberList{
		fmt.Println("Member " + strconv.Itoa(i) + " is "+ v.SuperNodeIp);
	}

	if (len(memberList) != expectedLength){
		panic ("Incorrect length for retreived list ")
	}

}

/**
This is a file to test the message passer
 */

func testDHT(hashTable dht.DHT ,key string){
	/* Initialize DHT */
	hashTable.Initialize()

	/* Insert data into hash table */
	hashTable.Append(key,dht.MemberShipInfo{"1.2.3.4"})

	/* Get data and returned list contains 1 values */
	checkForDataPresence(hashTable,key,1)

	/* Remove key from hash table */
	hashTable.Delete(key)

	/* check that no entry exists for the key */
	_, status := hashTable.Get(key)
	fmt.Println("After deletion status is "+ strconv.FormatBool(status))
	if (false != status){
		panic("Entry should not have been present")
	}

	/* Insert data into hash table */
	hashTable.Append(key,dht.MemberShipInfo{"1.2.3.5"})

	/* Append data into hash table */
	hashTable.Append(key,dht.MemberShipInfo{"1.2.3.6"})

	/* Get data and returned list contains 2 values */
	checkForDataPresence(hashTable,key,2)

	/* Remove data in the hash table*/
	hashTable.Remove(key,dht.MemberShipInfo{"1.2.3.5"})

	/* Get data and returned list contains 1 value */
	checkForDataPresence(hashTable,key,1)

	/* Remove key */
	hashTable.Delete(key)

	/* check that no entry exists for the key */
	_, status = hashTable.Get(key)
	fmt.Println("After deletion status is "+ strconv.FormatBool(status))
	if (false != status){
		panic("Entry should not have been present")
	}
}

func main() {
	// Start reading from the receive message queue
	go mp.Receive()
	// Start listening
	go mp.Listen("bob")

	hashTable := dht.DHT{}
	testDHT(hashTable,"new_key")
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')     // send to socket
		// Send the message, trim the last \n from input
		go mp.Send(mp.NewMessage("p2plive", text[:len(text)-1]))
		fmt.Println("Send Message " + text)
	}

}