package node

import (
	"github.com/hoisie/web"
	"net/http"
	Streamer "../streaming/streamer"
	Json "encoding/json"
	"fmt"
	NodeContext "../node/nodeContext"
	"mime"
	"strconv"
	"time"
)

type ServerData struct{
	isStreamer bool
}

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

func apiReceive(ctx *web.Context) string{
	data := map[string]string{"msg":streamer.Receive()}
	json, _ := Json.Marshal(data)
	fmt.Println(ctx.Params["callback"] + "(" + string(json) + ")")
	return ctx.Params["callback"] + "(" + string(json) + ")"
}

func apiGetPrograms(ctx *web.Context, val string) string{
	programList := streamer.ProgramList
	delete(programList, context.LocalName)
	json, _ := Json.Marshal(programList)
	fmt.Println(ctx.Params["callback"] + "(" + string(json) + ")")
	return ctx.Params["callback"] + "(" + string(json) + ")"
}

func apiGetTitle(ctx *web.Context, val string) string{
	title := streamer.ProgramList[streamer.CurrentProgram]
	json, _ := Json.Marshal(map[string]string{"title": title})
	fmt.Println(ctx.Params["callback"] + "(" + string(json) + ")")
	return ctx.Params["callback"] + "(" + string(json) + ")"
}

func apiGetLoad(ctx *web.Context, val string) string{
	loadList := StartLoadTrack().SuperNodeUsages
	json, _ := Json.Marshal(loadList)
	fmt.Println(ctx.Params["callback"] + "(" + string(json) + ")")
	return ctx.Params["callback"] + "(" + string(json) + ")"
}

func apiFakeStream(ctx *web.Context, num string) {
	n, _ := strconv.ParseInt(num, 10, 64)
	for i := 0; i < int(n); i++ {
		streamer.Stream("Gossip " + strconv.Itoa(i))
		time.Sleep(1e8)
	}
}

func apiGetLocalName() string{
	return streamer.CurrentProgram
}

func apiIsStreamer(ctx *web.Context, val string) string {
	data := map[string]bool{"isStreamer":false}
	if streamer.CurrentProgram == nodeContext.LocalName {
		data["isStreamer"] = true
	}
	json, _ := Json.Marshal(data)
	fmt.Println(ctx.Params["callback"] + "(" + string(json) + ")")
	return ctx.Params["callback"] + "(" + string(json) + ")"
}

func webInterface(streamer *Streamer.Streamer, nodeContext *NodeContext.NodeContext) {
	context = nodeContext
	mime.AddExtensionType(".css", "text/css")

	web.Get("/start/(.*)", apiStart)
	web.Get("/stop/", apiStop)
	web.Get("/join/(.*)", apiJoin)
	web.Get("/stream/(.*)", apiStream)
	web.Get("/receive/", apiReceive)
	web.Get("/allPrograms/(.*)", apiGetPrograms)
	web.Get("/load/(.*)", apiGetLoad)
	web.Get("/fakeStream/([0-9]+)", apiFakeStream)
	web.Get("/getLocalName/", apiGetLocalName)
	web.Get("/getTitle", apiGetTitle)
	web.Get("/isStreamer/(.*)", apiIsStreamer)

	web.Get("/(.*)",  http.FileServer(http.Dir(".")))
	web.Run("0.0.0.0:9999")

}