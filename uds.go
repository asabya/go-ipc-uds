package uds

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
)

const (
	DefaultSize = 512
)

// ListenerOptions is used to create Listener.
// SocketPath is the location of the `sock` file.
// Size is the size for byte array
type ListenerOptions struct {
	Size uint64
	SocketPath string
}

// Envelop is used to read data on Listener
type Envelop struct {
	Data string
	Error error
}

type OutMessage struct {
	Topic string
	Data string
}

// Listener creates a uds listener
func Listener(ctx context.Context, opts ListenerOptions) (chan Envelop, chan OutMessage) {
	out := make(chan Envelop)
	external := make(chan OutMessage)
	r := Envelop{
		Data: "",
		Error: nil,
	}
	if opts.SocketPath == "" {
		r.Error = errors.New("invalid socket path")
		out <- r
		return out, external
	}
	if opts.Size == 0 {
		opts.Size = DefaultSize
	}
	l, err := net.Listen("unix", opts.SocketPath)
	if err != nil {
		r.Error = err
		out <- r
		return out, external
	}
	go func() {
		defer l.Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := l.Accept()
				if err != nil {
					r.Error = err
					out <- r
				}
				go func(c net.Conn) {
					log.Println("looking for in")
					in := <-external
					b, _ := json.Marshal(in)
					_, err = c.Write(b)
					if err != nil {
						log.Println("Write error", err)
						c.Close()
						return
					}
				}(conn)
				b := make([]byte, opts.Size)
				n, err := conn.Read(b)
				if err != nil && err != io.EOF {
					log.Println("Read error :" , err)
					conn.Close()
					break
				}
				r.Data = string(b[:n])
				out <- r
			}
		}
	}()
	return out, external
}

// Dialer creates a uds dialer
func Dialer(sockPath string) (chan string ,error) {
	in := make(chan string)
	c, err := net.Dial("unix", sockPath)
	if err != nil {
		return in, err
	}
	go func() {
		defer c.Close()
		for {
			_, err := c.Write([]byte(<-in))
			if err != nil {
				return
			}
		}
	}()
	return in, nil
}