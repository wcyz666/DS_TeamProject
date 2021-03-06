package streaming

import (
	DHT "../../dht"
	MP "../../messagePasser"
	SNC "../../superNode/superNodeContext"
	SDataType "../"
	"fmt"
	"../../config"
	DNS "../../dnsService"
	"../../utils"
)

/**
The stream handler class
This class takes care of
	 1. The related requests from the normal nodes
	 2. Sync related information in the DHT and with other supernodes
*/
type StreamingHandler struct {
	dht              *DHT.DHTService
	mp               *MP.MessagePasser
	superNodeContext *SNC.SuperNodeContext
	ProgramList      map[string][]byte
}

func NewStreamingHandler(dHashtable *DHT.DHTService, mp *MP.MessagePasser, superNodeContext *SNC.SuperNodeContext) *StreamingHandler {
	sHandler := StreamingHandler{dht: dHashtable, mp: mp, superNodeContext:superNodeContext}
	sHandler.ProgramList = make(map[string][]byte)
	return &sHandler
}

/* Broadcast service */
func (sHandler *StreamingHandler) broadcast(channelName string, data []byte) {
	// Get all the supernodes
	IPs := DNS.GetAddr(config.BootstrapDomainName)
	for _, ip := range (IPs) {
		//fmt.Println("==========Broadcasting to " + ip)
		sHandler.mp.Send(MP.NewMessage(ip, "", channelName, data))
	}
}

/* A node starts to stream messages */
func (sHandler *StreamingHandler) StreamStart(msg *MP.Message) {
	//Notify all the supernodes
	//Including itself (Since one supernode may have multipy children)
	sHandler.broadcast("stream_program_start", msg.Data)

	//Update DHT table
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	sHandler.dht.Create(controlData.RootStreamer,
		DHT.MemberShipInfo{
			StreamerName:controlData.SrcName,
			StreamerIp: controlData.SrcIp,
			StreamProgramName: controlData.Title,
		})

}

/* A node starts to stream messages */
func (sHandler *StreamingHandler) StreamStop(msg *MP.Message) {
	//Notify all the supernodes
	sHandler.broadcast("stream_program_stop", msg.Data)

	//Update DHT table
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	delete(sHandler.ProgramList, controlData.SrcName)
	sHandler.dht.Delete(controlData.SrcName)
}

/* Update broadcasted from other supernodes */
func (sHandler *StreamingHandler) StreamProgramStart(msg *MP.Message) {
	// Store the program in the supernodes
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	sHandler.ProgramList[controlData.SrcName] = msg.Data

	//Notify the children the new program
	fmt.Println("New programs detected! Sending to all children")
	for _, child := range(sHandler.superNodeContext.Nodes){
		sHandler.mp.Send(MP.NewMessage(child.IP, "", "streaming_new_program", msg.Data))
	}

}

/* Update broadcasted from other supernodes */
func (sHandler *StreamingHandler) StreamProgramStop(msg *MP.Message) {
	// Del the program in the supernodes
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	delete(sHandler.ProgramList, controlData.SrcName)

	//Notify the children the new program
	fmt.Println("New programs detected! Sending to all children")
	for _, child := range(sHandler.superNodeContext.Nodes){
		sHandler.mp.Send(MP.NewMessage(child.IP, "", "streaming_stop_program", msg.Data))
	}
}

/* A child node asks to join a certain streaming group */
func (sHandler *StreamingHandler) StreamJoin(msg *MP.Message) {
	// Notify one of the streamers in the DHT
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)

	root := controlData.RootStreamer
	//fmt.Println(root)

	// Find the streaming group with root in the DHT and update it
	streamers, _ := sHandler.dht.Get(root)

	// Changed Apr.12   Use randomly selected streamer as the streaming parent
	if len(streamers) == 0{
		return
	}

	fmt.Print("Current Streamers: ")
	fmt.Println(streamers)
	// TODO: Temp workaround here. For rejoining nodes.
	var streamer DHT.MemberShipInfo
	if controlData.Type == "rejoin"{
		streamer = streamers[0]
	}else {
		streamer = streamers[utils.RandomChoice(0, len(streamers))]
		// Update the dht, append the guy into dht
		sHandler.dht.Append(root, DHT.MemberShipInfo{
			StreamerName:controlData.SrcName,
			StreamerIp:controlData.SrcIp,
			StreamProgramName: controlData.Title,
		})
	}
	sHandler.mp.Send(MP.NewMessage(streamer.StreamerIp, streamer.StreamerName, "streaming_join", msg.Data))
}

func (sHandler *StreamingHandler) NewChildJoin(childIp string, childName string) {
	for _, data := range (sHandler.ProgramList) {
		// Notify as a new program
		sHandler.mp.Send(MP.NewMessage(childIp, childName, "streaming_new_program", data))
	}
}

/* Take care of attached node failure
	If is root streamer: delete the streaming group in DHT
*/
func (sHandler *StreamingHandler) HandleErrorMsg(msg *MP.Message){
	failNode := MP.FailClientInfo{}
	MP.DecodeData(&failNode, msg.Data)

	for _, node := range(sHandler.superNodeContext.Nodes){
		// If fail node is one of the child
		if failNode.IP == node.IP{
			if _, ok := sHandler.ProgramList[failNode.Name]; ok{
				sHandler.dht.Delete(failNode.Name)
				delete(sHandler.ProgramList, failNode.Name)
				data := SDataType.StreamControlMsg{
					SrcName: failNode.Name,
					SrcIp: failNode.IP,
					RootStreamer: failNode.Name,
				}
				sHandler.broadcast("stream_program_stop", MP.EncodeData(data))
			}
		}
	}
}

func (sHandler *StreamingHandler) RemoveFromDht(msg *MP.Message) {
	failNode := SDataType.RemoveFromDht{}
	MP.DecodeData(&failNode, msg.Data)
	// Update dht here
	sHandler.dht.Remove(failNode.RootStreamer, DHT.MemberShipInfo{
		StreamerName: failNode.FailNodeName,
		StreamerIp: failNode.FailNodeIp,
		//StreamProgramName:
	})
}


