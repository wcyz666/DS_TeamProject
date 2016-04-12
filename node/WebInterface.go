package node

import (
	"github.com/hoisie/web"
	Streamer "../streaming/streamer"
)


func webInterface(streamer *Streamer.Streamer) {
	web.Get("/start/(.*)", streamer.Start)
	web.Get("/stop/", streamer.Stop)
	web.Get("/join/(.*)", streamer.Join)
	web.Get("/stream/(.*)", streamer.Stream)
	web.Run("0.0.0.0:9999")
}

