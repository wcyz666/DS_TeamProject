package node

import (
	"github.com/hoisie/web"
	Streamer "../streaming/streamer"
	Json "encoding/json"
	"fmt"
)

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
	json, _ := Json.Marshal(streamer.ProgramList)
	fmt.Println(ctx.Params["callback"] + "(" + string(json) + ")")
	return ctx.Params["callback"] + "(" + string(json) + ")"

}

func webInterface(streamer *Streamer.Streamer) {
	web.Get("/", apiHello)
	web.Get("/start/(.*)", apiStart)
	web.Get("/stop/", apiStop)
	web.Get("/join/(.*)", apiJoin)
	web.Get("/stream/(.*)", apiStream)
	web.Get("/receive/", apiReceive)
	web.Get("/allPrograms/(.*)", apiGetPrograms)
	web.Run("0.0.0.0:9999")
}