package main

import (
	"github.com/Sab94/uds"
	"log"
)

var SockPath = "/tmp/uds.sock"

func main() {
	opts := uds.ListenerOptions{
		Size:       512,
		SocketPath: SockPath,
	}
	out := uds.Listener(opts)
	for {
		data := <-out
		if data.Error != nil {
			log.Println("Got Error :", data.Error.Error())
			return
		}
		log.Println("Got data : ", data.Data)
	}
}
