package main

import (
	"fmt"
	"github.com/asabya/go-ipc-uds"
)

func main() {
	in, out, err := uds.Dialer("/tmp/uds.sock")
	if err != nil {
		return
	}
	in <- "My message"
	fmt.Println(<-out)
}
