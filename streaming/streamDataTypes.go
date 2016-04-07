package streamer

type StreamControlMsg struct {
	Type string
	// The src name and dest name indicates where the requests actually come from
	// The message may be redirected by the supernodes
	// So we need another identifier other than that in the message passer
	SrcName string
	RootStreamer string
	GroupId int
	StreamID int
	Title string
}


