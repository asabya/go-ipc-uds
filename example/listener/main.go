package main

import (
	"context"
	"log"

	uds "github.com/asabya/go-ipc-uds"
)

var SockPath = "/tmp/uds.sock"

func main() {
	opts := uds.Options{
		Size:       512,
		SocketPath: SockPath,
	}
	out, ext, err := uds.Listener(context.Background(), opts)
	if err != nil {
		log.Fatal(err)
	}
	for {
		data := <-out
		log.Println("Got data : ", data)
		d := "My data"
		ext <- d
		log.Println("Sent data : ", d)
	}
}
