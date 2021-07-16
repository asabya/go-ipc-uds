package main

import (
	"github.com/asabya/uds"
)

func main() {
	in, err := uds.Dialer("/tmp/uds.sock")
	if err != nil {
		return
	}
	in <- "My message"
}
