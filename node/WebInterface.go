package node

import (
	"github.com/hoisie/web"
	"net/http"
	Streamer "../streaming/streamer"
	Json "encoding/json"
	"fmt"
	NodeContext "../node/nodeContext"
)

var context *NodeContext.NodeContext

func apiHello()string{
	return "Hello World!"
}

func apiStart(title string){
	streamer.Stop()
	streamer.Start(title)
}

func apiStop(){
	streamer.Stop()
}

func apiJoin(programId string){
	streamer.Stop()
	streamer.Join(programId)
}

func apiStream(data string){
	streamer.Stream(data)
}

func apiReceive() string{
	return streamer.Receive()
}

func apiGetPrograms(ctx *web.Context, val string) string{
	programList := streamer.ProgramList
	delete(programList, context.LocalName)
	json, _ := Json.Marshal(programList)
	fmt.Println(ctx.Params["callback"] + "(" + string(json) + ")")
	return ctx.Params["callback"] + "(" + string(json) + ")"

}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!")
}

func webInterface(streamer *Streamer.Streamer, nodeContext *NodeContext.NodeContext) {
	context = nodeContext


	web.Get("/", apiHello)
	web.Get("/start/(.*)", apiStart)
	web.Get("/stop/", apiStop)
	web.Get("/join/(.*)", apiJoin)
	web.Get("/stream/(.*)", apiStream)
	web.Get("/receive/", apiReceive)
	web.Get("/allPrograms/(.*)", apiGetPrograms)
	
	web.Get("/(.*)",  http.FileServer(http.Dir(".")))
	web.Run("0.0.0.0:9999")

}