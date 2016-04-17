package dht

/* Implements functionality related to creation and management of underlying Ring structure for Pastry DHT */

import (
	MP "../messagePasser"
	dns "../dnsService"
	Config "../config"
	lns "../localNameService"
	"math/big"
	"fmt"
	"time"
)

const (
	TRAVERSE_CLOCK_WISE   = iota // Traverse in the direction of next node
	TRAVERSE_ANTI_CLOCK_WISE // Traverse in direction of prev node
	NODE_JOIN_TRIGGERRED_LEAF_TABLE_REFRESH
	NODE_FAILURE_TRIGGERED_LEAF_TABLE_REFRESH
	PERIODIC_LEAF_TABLE_REFRESH
	EVENT_TRIGGERED_LEAF_TABLE_REFRESH

	RING_REPAIR_REQUEST_FAILURE_TIMER = 10


)

/* Constructor */
func NewDHTNode(mp *MP.MessagePasser) (*DHTNode) {
	var dhtNode = DHTNode{mp: mp}
	dhtNode.hashTable = make(map[string][]MemberShipInfo)
	/* Use hash of mac address of the super node as the key for partitioning key space */
	dhtNode.NodeKey = lns.GetLocalName()
	dhtNode.NodeName = lns.GetLocalName()
	dhtNode.IpAddress, _ = dns.ExternalIP()
	dhtNode.curNodeNumericKey =  getBigIntFromString(dhtNode.NodeKey)
	fmt.Println("*****  		Node Initial Config 		*****")
	fmt.Println("		key = "+ dhtNode.NodeKey)
	fmt.Println("		name = "+ dhtNode.NodeName)
	fmt.Println("		ipaddr = "+ dhtNode.IpAddress)
	fmt.Println("")
	if (REPLICATION_FACTOR > NEIGHBOURHOOD_DISTANCE){
		/* To reduce overhead while updating replicas, we would like to directly send messages
		 * to replicas instead of traversing around the ring. To achieve this, we assume replication
		 * factor to be <= configured neighbourhood distance */
		panic("Replication factor needs to be <= neighbourhood distance in our implementation")
	}
	return &dhtNode
}

func getFirstNonSelfIpAddr() (string){
	curAddrList := dns.GetAddr(Config.BootstrapDomainName)
	extIP, _ := dns.ExternalIP()

	for _, ipAddr := range curAddrList {
		if ipAddr == extIP {
			continue
		} else{

			return ipAddr
		}
	}
	return ""
}


/* Computes the direction in which we need to make the traversal. First argument is
 * curNodeKey and second argument is the key towards which we need to traversal */
func computeTraversalDirection(curNodeKey *big.Int, newKey *big.Int ) int{
	status := TRAVERSE_CLOCK_WISE

	zero := getBigIntFromString("0")
	one  := getBigIntFromString("1")
	maxKey := getBigIntFromString(MAX_KEY)

	/* compute the distance in clock wise and anti clock wise direction and travel in the direction
	*  which is shortest */
	k := new(big.Int)
	/* k = curNodeKey - newKey */
	k.Sub(curNodeKey,newKey)

	clockWiseDistance :=  new(big.Int)
	antiClockWiseDistance := new(big.Int)

	if (k.Cmp(zero) > 0) {
		antiClockWiseDistance = k
		clockWiseDistance.Sub(maxKey,antiClockWiseDistance)
		clockWiseDistance.Add(clockWiseDistance,one)
	} else {
		clockWiseDistance.Sub(zero,k)
		antiClockWiseDistance.Sub(maxKey,clockWiseDistance)
		antiClockWiseDistance.Add(antiClockWiseDistance,one)
	}

	if (clockWiseDistance.Cmp(antiClockWiseDistance) > 0){
		status = TRAVERSE_ANTI_CLOCK_WISE
	} else {
		status = TRAVERSE_CLOCK_WISE
	}
	return status
}

func getBigIntFromString(key string) *big.Int{
	numericKey := new(big.Int)
	_,status := numericKey.SetString(key, 16)
	if (false == status){
		panic("WARNING: Unable to convert newNodeKey to a valid value")
	}
	return numericKey
}

func isKeyPresentInKeyspaceRange(numericKey *big.Int, prevNodeNumericKey *big.Int, curNodeNumericKey *big.Int) bool {

	zero := getBigIntFromString("0")
	maxKey := getBigIntFromString(MAX_KEY)

	/* If curNodeKey > prevNodeKey, check if new key in (prevNodeKey, curNodeKey]
	 * If not, check if new key is in (prevNodeKey, Maxkey) or [0, curNodeKey]
	*/
	if (curNodeNumericKey.Cmp(prevNodeNumericKey) > 0){
		if ((numericKey.Cmp(prevNodeNumericKey) > 0) &&
		(numericKey.Cmp(curNodeNumericKey) <= 0)){
			return true
		} else {
			return false
		}
	} else {
		if (((numericKey.Cmp(prevNodeNumericKey) > 0) && (numericKey.Cmp(maxKey) <=0)) ||
		((numericKey.Cmp(zero)>=0) && (numericKey.Cmp(curNodeNumericKey) <=0))) {
			return true
		} else {
			return false
		}
	}
}

/* Given a key, function will check whether key is within key space managed by this node
 * KeyspaceRange is from (previous node's key + 1) to current node's key
*/
func (dhtNode *DHTNode) isKeyPresentInMyKeyspaceRange(key string) bool {
	numericKey := getBigIntFromString(key)
	if  ((dhtNode.leafTable.nextNode == nil) &&(dhtNode.leafTable.prevNode == nil)){
		return true
	}

	zero := getBigIntFromString("0")
	maxKey := getBigIntFromString(MAX_KEY)

	/* If curNodeKey > prevNodeKey, check if new key in (prevNodeKey, curNodeKey]
	 * If not, check if new key is in (prevNodeKey, Maxkey) or [0, curNodeKey]
	*/
	if (dhtNode.curNodeNumericKey.Cmp(dhtNode.prevNodeNumericKey) > 0){
		if ((numericKey.Cmp(dhtNode.prevNodeNumericKey) > 0) &&
		    (numericKey.Cmp(dhtNode.curNodeNumericKey) <= 0)){
			return true
		} else {
			return false
		}
	} else {
		if (((numericKey.Cmp(dhtNode.prevNodeNumericKey) > 0) && (numericKey.Cmp(maxKey) <=0)) ||
			((numericKey.Cmp(zero)>=0) && (numericKey.Cmp(dhtNode.curNodeNumericKey) <=0))) {
			return true
		} else {
			return false
		}
	}
}

/*TODO Function responsible for updating leaf table and prefix table based on new information */
func (dhtNode *DHTNode)updateLeafAndPrefixTablesWithNewNode(newNodeIpAddress string, newNodeName string,
																	 newNodeKey string, isPrevNode bool){

	/* Update prev and next node information for now */
	newNodeNumericKey := getBigIntFromString(newNodeKey)

	var node = Node{newNodeIpAddress, newNodeName, newNodeKey}
	if (true == isPrevNode){
		dhtNode.leafTable.prevNode = &node
		dhtNode.prevNodeNumericKey = newNodeNumericKey
	} else{
		dhtNode.leafTable.nextNode = &node
	}

	/*if (dhtNode.leafTable.prevNode != nil ){
		fmt.Println("Previous node is "+ dhtNode.leafTable.prevNode.IpAddress)
	}
	if (dhtNode.leafTable.nextNode != nil){
		fmt.Println("Next node is " + dhtNode.leafTable.nextNode.IpAddress)
	}*/
}

func (dhtNode *DHTNode)getPredecessorFromLeafTable()(*Node)  {
	return dhtNode.leafTable.prevNode
}

/*TODO currently we are implementing a simple find successor scheme */
func (dhtNode *DHTNode) findSuccessor(key string) (*Node){
	if (computeTraversalDirection(dhtNode.curNodeNumericKey,getBigIntFromString(key)) == TRAVERSE_ANTI_CLOCK_WISE){
		return dhtNode.leafTable.prevNode
	} else {
		return dhtNode.leafTable.nextNode
	}
}

func (dhtNode *DHTNode) CreateOrJoinRing()int{
	ipAddr := getFirstNonSelfIpAddr()
	if ("" == ipAddr){
		/* No entries exist or your are the only one. This means you are like
		 * Apocalypse, the first mutant. Create a DHT*/
		fmt.Println("[DHT]	Creating New DHT")
		dhtNode.leafTable.nextNode = nil
		dhtNode.leafTable.prevNode = nil
		dhtNode.curReplicaCount = 1
		dhtNode.StartPeriodicLeafTableRefresh()
		return NEW_DHT_CREATED
	} else {
		/* Send a message to one of the super nodes requesting to provide successor node's information
		 * based on key provided
		 */
		fmt.Println("[DHT]	Attempting to Join Existing DHT. Sending Request to " + ipAddr)
		ip,name := dhtNode.mp.GetNodeIpAndName()
		dhtNode.mp.Send(MP.NewMessage(ipAddr, "", "join_dht_req", MP.EncodeData(JoinRequest{dhtNode.NodeKey,ip,name})))
		/* Ip address of the peer in DNS might be allocated to another EC2 instance that no longer runs super node.
		 * Wait for join response from the peer for 2 seconds afterwhich move to new node */
		timer1 := time.NewTimer(time.Second * 2)
		go func(){
			<-timer1.C
			if (dhtNode.DhtState == DHT_WAIT_FOR_JOIN_RESPONSE){
				fmt.Println("No response from peer")
				msg := MP.NewMessage(dhtNode.IpAddress, "self", "join_dht_conn_failed",
						MP.EncodeData(MP.FailClientInfo{"",ipAddr,"No Response received from peer"}))
				dhtNode.mp.Messages["join_dht_conn_failed"] <- &msg
			}
		}()
		dhtNode.DhtState = DHT_WAIT_FOR_JOIN_RESPONSE
		return JOINING_EXISTING_DHT
	}
}

func (dhtNode *DHTNode) sendJoinReq(node *Node){
	ip,name := dhtNode.mp.GetNodeIpAndName()
	dhtNode.mp.Send(MP.NewMessage(node.IpAddress, "", "join_dht_req", MP.EncodeData(JoinRequest{dhtNode.NodeKey,ip,name})))
}

func (dhtNode *DHTNode) AmITheOnlyNodeInDHT()(bool){
	if ((nil == dhtNode.leafTable.prevNode) &&
	    (nil == dhtNode.leafTable.nextNode)){
		return true
	}
	return false
}

func logNodeList(nodeList []Node){
	for _,node := range nodeList {
		fmt.Println("		IP: "+ node.IpAddress + " Key: "+ node.Key)
	}
}

func (dhtNode *DHTNode) GetNextNodeToForwardInRing(key string) (*Node){
	return dhtNode.findSuccessor(key)
}

func (dhtNode *DHTNode) HandleJoinReq(msg *MP.Message) {
	var joinReq JoinRequest
	MP.DecodeData(&joinReq,msg.Data)
	var joinRes JoinResponse
	fmt.Println("[DHT] HandleJoinReq")
        
	if (true == dhtNode.AmITheOnlyNodeInDHT()){
		/* Me apocolyse, got my first disciple. Join request received for a DHT ring of one node */
		/* Send a Join Response indicating that new Node's predecessor and successor is myself */
		node := Node{dhtNode.IpAddress, dhtNode.NodeName, dhtNode.NodeKey}
		joinRes.Predecessor = node
		fmt.Println("[DHT] Adding my first disciple (i.e.) second node in DHT.")
	} else {
		/* Forward the message if key is not managed by you */
		if (false == dhtNode.isKeyPresentInMyKeyspaceRange(joinReq.Key)){
			/* Find successor node and send it in the response */
			successor := dhtNode.findSuccessor(joinReq.Key)
			if (nil == successor){
				joinRes.Status = FAILURE
				/* Send failure message to Join Request originator */
				dhtNode.mp.Send(MP.NewMessage(joinReq.OriginIpAddress,
					joinReq.OriginName, "join_dht_res", MP.EncodeData(joinRes)))
			} else {
				/* Forward the message towards successor */
				dhtNode.mp.Send(MP.NewMessage(successor.IpAddress, "", "join_dht_req", MP.EncodeData(joinReq)))
			}
			return
		}
		joinRes.Predecessor = *(dhtNode.getPredecessorFromLeafTable())
	}


	if (true == dhtNode.IsRingUpdateInProgress){
		fmt.Println("[DHT] Join In Progress. Retry later")
		joinRes.Status = JOIN_IN_PROGRESS_RETRY_LATER
		dhtNode.mp.Send(MP.NewMessage(joinReq.OriginIpAddress,
			joinReq.OriginName, "join_dht_res", MP.EncodeData(joinRes)))
		return
	}

	/* Since we are transferring a portion of our hashtable to new node and the process is still in progress
	 * set this flag */
	dhtNode.IsRingUpdateInProgress = true
	/* Retrieve entries which are less than new node's key and create a map out of it.*/
	nodeKey := getBigIntFromString(joinReq.Key)
	var entryKey *big.Int
	joinRes.HashTable = make(map[string][]MemberShipInfo)

	for k,v := range dhtNode.hashTable {
		entryKey = getBigIntFromString(k)
		/* If entry key is within new node's key space, transfer the data to new node */
		if (false == isKeyPresentInKeyspaceRange(entryKey, nodeKey, dhtNode.curNodeNumericKey)){
			joinRes.HashTable[k] = v
		}
	}

	/* Send the map in the response to Join Request originator */
	joinRes.Status = SUCCESS
	joinRes.Successor = Node{dhtNode.IpAddress, dhtNode.NodeName, dhtNode.NodeKey}
        
	fmt.Println("[DHT] Sent Predecessor is " + joinRes.Predecessor.IpAddress)
	fmt.Println("[DHT] Sent Predecessor name is " + joinRes.Predecessor.Name)
	fmt.Println("[DHT] Sending Successful Join Response to " + joinReq.OriginIpAddress)
	dhtNode.mp.Send(MP.NewMessage(joinReq.OriginIpAddress, "" , "join_dht_res", MP.EncodeData(joinRes)))
}

func (dhtNode *DHTNode) HandleJoinRes(msg *MP.Message) (int,*Node) {
	var joinRes JoinResponse
	MP.DecodeData(&joinRes,msg.Data)
	var node *Node = nil

	if joinRes.Status == FAILURE {
		panic ("Join procedure for DHT failed")
	} else if (joinRes.Status == JOIN_IN_PROGRESS_RETRY_LATER) {
		node = &(joinRes.Successor)
	} else {
		/* SUCCESS case */
		fmt.Println("[DHT] Join Response with Success received")
		/* 1. Add received map to local DHT table */
		dhtNode.hashTable = joinRes.HashTable

		/* Update prev and next nodes */
		dhtNode.updateLeafAndPrefixTablesWithNewNode(joinRes.Successor.IpAddress, joinRes.Successor.Name,
		                                     joinRes.Successor.Key,false)
		dhtNode.updateLeafAndPrefixTablesWithNewNode(joinRes.Predecessor.IpAddress, joinRes.Predecessor.Name,
			joinRes.Predecessor.Key,true)

		fmt.Println("Sending Join complete message to "+ joinRes.Successor.IpAddress + " with key " + joinRes.Successor.Name)
		/* 2. Send Join complete to successor */
		dhtNode.mp.Send(MP.NewMessage(joinRes.Successor.IpAddress, joinRes.Successor.Name, "join_dht_complete",
			                  MP.EncodeData(JoinComplete{SUCCESS, dhtNode.NodeKey})))

		fmt.Println("Sending Join notify message to "+ joinRes.Predecessor.IpAddress + " with key " + joinRes.Predecessor.Name)
		/* 3. Send join notification to predecessor */
		dhtNode.mp.Send(MP.NewMessage(joinRes.Predecessor.IpAddress, joinRes.Predecessor.Name, "join_dht_notify",
			                              MP.EncodeData(JoinNotify{dhtNode.NodeKey})))

		dhtNode.StartPeriodicLeafTableRefresh()
	}
	return joinRes.Status,node
}

func (dhtNode *DHTNode) HandleJoinComplete(msg *MP.Message) {
	var joinComplete JoinComplete
	MP.DecodeData(&joinComplete,msg.Data)
	fmt.Println("[DHT] Join Complete received")

	/* Update routing information to include this new node */
	dhtNode.updateLeafAndPrefixTablesWithNewNode(msg.Src, msg.SrcName, joinComplete.Key,true)

	/* Delete entries transferred to new node from the farthest replica */
	/* In a ring, since it is a circular list, if one of the nodes achieves the desired replication factor, everyone
	 * else achieves the replication factor */

	/*TODO Update current replication counts accordingly*/
	/*Replication count < desired replication factor, don't delete the contents*/
	if (dhtNode.curReplicaCount == REPLICATION_FACTOR){
		/* Send delete replica request */
		nodeToForward := dhtNode.leafTable.NextNodeList[REPLICATION_FACTOR-1]
		dhtNode.mp.Send(MP.NewMessage(nodeToForward.IpAddress, nodeToForward.Name , "dht_delete_replica_req",
			MP.EncodeData(DeleteReplicaRequest{joinComplete.Key,dhtNode.NodeKey})))
		dhtNode.ReplicationState = REPLICA_DELETION_IN_PROGRESS

	} else {
		fmt.Println("No need to delete existing replicas since we are yet to reach the desired replicatio factor")
	}

	dhtNode.IsRingUpdateInProgress = false
}


func (dhtNode *DHTNode) HandleJoinNotify(msg *MP.Message) {
	var joinNotify JoinNotify
	MP.DecodeData(&joinNotify,msg.Data)
	fmt.Println("[DHT] Join Notify received")

	/* Update routing information to include this new node */
	dhtNode.updateLeafAndPrefixTablesWithNewNode(msg.Src, msg.SrcName, joinNotify.Key,false)
}



/*Failure Handling relation functions */

func (dhtNode *DHTNode)CommunicationFailureHandler(msg *MP.Message){
	var failClientInfo MP.FailClientInfo
	MP.DecodeData(&failClientInfo,msg.Data)

	switch dhtNode.DhtState  {
	case DHT_JOIN_IN_PROGRESS:
		dhtNode.mp.Messages["join_dht_conn_failed"] <- msg
	case DHT_JOINED:
		dhtNode.NodeFailureDetected(failClientInfo.IP, 0)
	case DHT_RING_REPAIR_IN_PROGRESS:
		if (len(dhtNode.leafTable.PrevNodeList)>1){
			if ((dhtNode.leafTable.PrevNodeList[1].IpAddress) == failClientInfo.IP){
				dhtNode.mp.Messages["dht_ring_repair_req_conn_failed"]	<- msg
			}
		}
	}
}

func (dhtNode *DHTNode) RemoveFailedSuperNode(IpAddress string){
	/* Remove failed node from DNS */
	dns.ClearAddrRecords(Config.BootstrapDomainName, IpAddress)
}

func (dhtNode *DHTNode) NodeFailureDetected(IpAddress string, depth int){
	fmt.Println("Node failure detected for node " + IpAddress)

	if (dhtNode.leafTable.prevNode == nil){
		return
	}

	//fmt.Println("Prev Node is "+ dhtNode.leafTable.prevNode.IpAddress)

	//fmt.Println("NodeFailureDetected : prev Node list is ")
	//logNodeList(dhtNode.leafTable.PrevNodeList)
	//fmt.Println("NodeFailureDetected: next Node list is ")
	//logNodeList(dhtNode.leafTable.NextNodeList)

	/* Previous Node failure detected. Ip Address parameter is the
	 * Ip Address of the node that failed */

	/* Trigger recovery if I am the successor of the node. Otherwise
	 * wait for successor to trigger recovery */
	if (dhtNode.leafTable.prevNode.IpAddress == IpAddress){
		/*Now previous node's key space becomes mine.*/
		prevNodeList := dhtNode.leafTable.PrevNodeList
		if (len(prevNodeList) > 1){
			//fmt.Println("prev Node list length > 1")

			/* Recover the ring from failure */
			newPrevNode := dhtNode.leafTable.PrevNodeList[1]

			/* Send a ring repair request along with my node information */
			dhtNode.mp.Send(MP.NewMessage(newPrevNode.IpAddress, newPrevNode.Name, "dht_ring_repair_req",
								MP.EncodeData(RingRepairRequest{dhtNode.NodeKey})))
			dhtNode.IsRingUpdateInProgress = true
			for {
				select {
				case ring_repair_res := <- dhtNode.mp.Messages["dht_ring_repair_res"]:
					//fmt.Println("Removing failed node with IP "+ IpAddress +" from DNS ")
					/* Remove failed node from DNS */
					dns.ClearAddrRecords(Config.BootstrapDomainName, IpAddress)
					dhtNode.DhtState = DHT_JOINED
					dhtNode.HandleRingRepairResponse(ring_repair_res)
					return
				case  _ = <- dhtNode.mp.Messages["dht_ring_repair_req_conn_failed"]:
					fmt.Println("Ring Repair request failed. Probably this node has failed too. Move to its previous node")
					dhtNode.leafTable.PrevNodeList = dhtNode.leafTable.PrevNodeList[1:]
					dhtNode.NodeFailureDetected(dhtNode.leafTable.prevNode.IpAddress, depth+1)
					//fmt.Println("Removing failed node with IP "+ IpAddress +" from DNS ")
					/* Remove failed node from DNS */
					dns.ClearAddrRecords(Config.BootstrapDomainName, IpAddress)
					return
				}
			}
		} else {
			fmt.Println("prev Node list length == 1")
			if (dhtNode.leafTable.prevNode.IpAddress == dhtNode.leafTable.nextNode.IpAddress){
				dhtNode.leafTable.nextNode = nil
				dhtNode.leafTable.NextNodeList = nil
			}
			dhtNode.leafTable.prevNode = nil
			dhtNode.leafTable.PrevNodeList = nil
			dhtNode.DhtState = DHT_JOINED
			//fmt.Println("Removing failed node with IP "+ IpAddress +" from DNS ")
			/* Remove failed node from DNS */
			dns.ClearAddrRecords(Config.BootstrapDomainName, IpAddress)
			/* Since I am the only one in the network, cannot have a replica */
			dhtNode.curReplicaCount = 1
		}
	}
}
func (dhtNode *DHTNode) HandleRingRepairRequest(msg *MP.Message){
	var ringRepairReq RingRepairRequest
	MP.DecodeData(&ringRepairReq,msg.Data)
	fmt.Println("[DHT] Ring Repair Request received")

	/* Update routing information to include this new node */
	dhtNode.updateLeafAndPrefixTablesWithNewNode(msg.Src, msg.SrcName, ringRepairReq.Key,false)
	dhtNode.mp.Send(MP.NewMessage(msg.Src, msg.SrcName, "dht_ring_repair_res",
		MP.EncodeData(RingRepairResponse{SUCCESS, dhtNode.NodeKey})))
}

func (dhtNode *DHTNode) HandleRingRepairResponse(msg *MP.Message){
	var ringRepairRes RingRepairResponse
	MP.DecodeData(&ringRepairRes,msg.Data)
	fmt.Println("[DHT] Ring Repair Response received")

	/* Update routing information to include this new node */
	dhtNode.updateLeafAndPrefixTablesWithNewNode(msg.Src, msg.SrcName, ringRepairRes.Key,true)
	dhtNode.IsRingUpdateInProgress = false

	dhtNode.RefreshLeafTable(NODE_FAILURE_TRIGGERED_LEAF_TABLE_REFRESH)
	//fmt.Println("HandleRingRepairResponse: prev Node list is ")
	//logNodeList(dhtNode.leafTable.PrevNodeList)
	//fmt.Println("HandleRingRepairResponse: next Node list is ")
	//logNodeList(dhtNode.leafTable.NextNodeList)
}

/* Leaf table refresh and neighbourhood discovery */

func (dhtNode *DHTNode) PerformPeriodicLeafTableRefresh(){
	ticker := time.NewTicker(time.Second * PERIODIC_LEAF_TABLE_REFRESH_DURATION)
	go func() {
		for _ = range ticker.C {
			if (dhtNode.AmITheOnlyNodeInDHT()){
				continue
			}

			//fmt.Println("Triggering Periodic Neighbourhood discovery")
			var neighbourhoodDiscovery = NeighbourhoodDiscoveryMessage{OriginIpAddress: dhtNode.IpAddress, OriginName:
			dhtNode.NodeName, ResidualHopCount: NEIGHBOURHOOD_DISTANCE, OriginKey: dhtNode.NodeKey, Event: PERIODIC_LEAF_TABLE_REFRESH}

			neighbourhoodDiscovery.TraversalDirection = TRAVERSE_ANTI_CLOCK_WISE
			dhtNode.mp.Send(MP.NewMessage(dhtNode.leafTable.prevNode.IpAddress,dhtNode.leafTable.prevNode.Name,
				"dht_neighbourhood_discovery",MP.EncodeData(neighbourhoodDiscovery)))

			neighbourhoodDiscovery.TraversalDirection = TRAVERSE_CLOCK_WISE
			dhtNode.mp.Send(MP.NewMessage(dhtNode.leafTable.nextNode.IpAddress,dhtNode.leafTable.nextNode.Name,
				"dht_neighbourhood_discovery",MP.EncodeData(neighbourhoodDiscovery)))
		}
	}()
}

/* Apart from periodically refreshing the table, there might be other events where we want to immediately
*  refresh the table instead of waiting for the timer to exprire. Invoke this method during those cases */
func (dhtNode *DHTNode) RefreshLeafTable(event int){
	//fmt.Println("Refresh Leaf Table for event " + strconv.Itoa(event))
	if (false == dhtNode.AmITheOnlyNodeInDHT()) {
		//fmt.Println("Refresh leaf table: More than 1 node in the DHT")
		//fmt.Println("Triggering Periodic Neighbourhood discovery")
		var neighbourhoodDiscovery = NeighbourhoodDiscoveryMessage{OriginIpAddress: dhtNode.IpAddress, OriginName:
		dhtNode.NodeName, ResidualHopCount: NEIGHBOURHOOD_DISTANCE, OriginKey: dhtNode.NodeKey, Event:event}

		neighbourhoodDiscovery.TraversalDirection = TRAVERSE_ANTI_CLOCK_WISE
		dhtNode.mp.Send(MP.NewMessage(dhtNode.leafTable.prevNode.IpAddress, dhtNode.leafTable.prevNode.Name,
			"dht_neighbourhood_discovery", MP.EncodeData(neighbourhoodDiscovery)))

		neighbourhoodDiscovery.TraversalDirection = TRAVERSE_CLOCK_WISE
		dhtNode.mp.Send(MP.NewMessage(dhtNode.leafTable.nextNode.IpAddress, dhtNode.leafTable.nextNode.Name,
			"dht_neighbourhood_discovery", MP.EncodeData(neighbourhoodDiscovery)))
	}
}

func (dhtNode *DHTNode) StartPeriodicLeafTableRefresh (){
	/* Schedule a trigger to query about neighbourhood details after 3 seconds */
	timer1 := time.NewTimer(time.Second * 3)
	go func(){
		<-timer1.C
		fmt.Println("Initiating periodic leaf table refresh procedure")
		dhtNode.RefreshLeafTable(NODE_JOIN_TRIGGERRED_LEAF_TABLE_REFRESH)
		dhtNode.PerformPeriodicLeafTableRefresh()
	}()
}

func (dhtNode *DHTNode) HandleNeighbourhoodDiscovery(msg *MP.Message){
	var discoveryMsg NeighbourhoodDiscoveryMessage
	MP.DecodeData(&discoveryMsg,msg.Data)
	//fmt.Println("Received Neigbhiurhood request message from "+ msg.Src + " with direction "+
	//strconv.Itoa(discoveryMsg.TraversalDirection) + "with hop "+ strconv.Itoa(discoveryMsg.ResidualHopCount))

	if (discoveryMsg.OriginIpAddress == dhtNode.IpAddress){
		/* Check if hop count = 0 . If so, populate it into the corresponding leaf table list.
		   Otherwise append your IP address and append it to the list.*/
		if (discoveryMsg.ResidualHopCount != 0){
			node := Node{dhtNode.IpAddress, dhtNode.NodeName, dhtNode.NodeKey}
			discoveryMsg.NodeList = append(discoveryMsg.NodeList, node)
		}

		if (discoveryMsg.TraversalDirection == TRAVERSE_ANTI_CLOCK_WISE){
			dhtNode.leafTable.PrevNodeList = discoveryMsg.NodeList

		} else {
			dhtNode.leafTable.NextNodeList = discoveryMsg.NodeList
			/* Next node is up-to-date. Trigger Replica synchronization procedure*/
//			dhtNode.PerformReplicaSync()
		}

		//fmt.Println("[DHT] Lead Table contents")
		//fmt.Println("[DHT]	Previous Node List")
		//logNodeList(dhtNode.leafTable.PrevNodeList)
		//fmt.Println("[DHT]	Next Node List")
		//logNodeList(dhtNode.leafTable.NextNodeList)

	} else{
		node := Node{dhtNode.IpAddress, dhtNode.NodeName, dhtNode.NodeKey}
		discoveryMsg.NodeList = append(discoveryMsg.NodeList, node)

		discoveryMsg.ResidualHopCount--
		if (discoveryMsg.ResidualHopCount == 0){
			//fmt.Println("Forwarded message to origin: "+ discoveryMsg.OriginIpAddress)
			//logNodeList(discoveryMsg.NodeList)
			dhtNode.mp.Send(MP.NewMessage(discoveryMsg.OriginIpAddress, discoveryMsg.OriginName,
				"dht_neighbourhood_discovery", MP.EncodeData(discoveryMsg)))
		} else {
			var nodeToForward *Node
			if (discoveryMsg.TraversalDirection == TRAVERSE_ANTI_CLOCK_WISE){
				nodeToForward = dhtNode.leafTable.prevNode
			} else {
				nodeToForward = dhtNode.leafTable.nextNode
			}
			//fmt.Println("Forwarded message to "+ nodeToForward.IpAddress)
			//logNodeList(discoveryMsg.NodeList)
			dhtNode.mp.Send(MP.NewMessage(nodeToForward.IpAddress, nodeToForward.Name,
				"dht_neighbourhood_discovery", MP.EncodeData(discoveryMsg)))
		}

		/* If leaf table refresh is due to node joining or leaving, trigger a refresh so
		 * as to be up-to-date with your surroundings */
		if (discoveryMsg.Event == NODE_JOIN_TRIGGERRED_LEAF_TABLE_REFRESH){
			/* TODO We can probably deduce information from the messages being exchanged instead of triggering one more
			 * neighbourhood discovery procedure. Deferring it as it is low priority work item*/
			dhtNode.RefreshLeafTable(EVENT_TRIGGERED_LEAF_TABLE_REFRESH)
			if (dhtNode.curReplicaCount < REPLICATION_FACTOR){

			}
			//dhtNode.ScheduleReplicaCreation()
		} else if (discoveryMsg.Event == NODE_FAILURE_TRIGGERED_LEAF_TABLE_REFRESH) {
			dhtNode.RefreshLeafTable(EVENT_TRIGGERED_LEAF_TABLE_REFRESH)
		}
	}
}