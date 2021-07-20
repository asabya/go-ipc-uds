package main

import (
	"fmt"

	uds "github.com/asabya/go-ipc-uds"
)

func main() {
	opts := uds.Options{
		Size:       512,
		SocketPath: "/tmp/uds.sock",
	}
	r, w, c, err := uds.Dialer(opts)
	if err != nil {
		return
	}
	defer c()

	w("asd 1")
	fmt.Println(r())

	w("qwe 1")
	fmt.Println(r())
}
