package main

import (
	"context"
	"log"

	uds "github.com/asabya/go-ipc-uds"
)

var SockPath = "/tmp/uds.sock"

func main() {
	opts := uds.Options{
		Size:       0,
		SocketPath: SockPath,
	}
	in, err := uds.Listener(context.Background(), opts)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn := <-in
		go func(conn *uds.Client) {
			for {
				d, err := conn.Read()
				if err != nil {
					return
				}
				r := string(d) + "return"
				err = conn.Write([]byte(r))
				if err != nil {
					return
				}
			}
		}(conn)
	}
}
