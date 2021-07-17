package uds

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"

	logging "github.com/ipfs/go-log/v2"
)

const (
	DefaultSize = 512
)

var log = logging.Logger("uds")

// Options is used to create Listener/Dialer.
// SocketPath is the location of the `sock` file.
// Size is the size for byte array
type Options struct {
	Size       uint64
	SocketPath string
}

// Listener creates a uds listener
func Listener(ctx context.Context, opts Options) (chan string, chan string, error) {
	out := make(chan string)
	external := make(chan string)
	if opts.SocketPath == "" {
		return out, external, errors.New("invalid socket path")
	}
	if opts.Size == 0 {
		opts.Size = DefaultSize
	}
	l, err := net.Listen("unix", opts.SocketPath)
	if err != nil {
		return out, external, err
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
					log.Error("Listener error", err)
					return
				}
				go func(c net.Conn) {
					in := <-external
					b, _ := json.Marshal(in)
					_, err = c.Write(b)
					if err != nil {
						log.Error("Write error", err)
						c.Close()
						return
					}
				}(conn)
				b := make([]byte, opts.Size)
				n, err := conn.Read(b)
				if err != nil && err != io.EOF {
					log.Error("Read error :", err)
					conn.Close()
					break
				}
				out <- string(b[:n])
			}
		}
	}()
	return out, external, nil
}

// Dialer creates a uds dialer
func Dialer(opts Options) (chan string, chan string, error) {
	in := make(chan string)
	out := make(chan string)
	if opts.SocketPath == "" {
		return in, out, errors.New("invalid socket path")
	}
	if opts.Size == 0 {
		opts.Size = DefaultSize
	}
	conn, err := net.Dial("unix", opts.SocketPath)
	if err != nil {
		return in, out, err
	}
	go func() {
		defer conn.Close()
		_, err := conn.Write([]byte(<-in))
		if err != nil {
			log.Error(err)
			return
		}
		b := make([]byte, opts.Size)
		n, err := conn.Read(b)
		if err != nil && err != io.EOF {
			log.Error("Read error :", err)
			conn.Close()
			return
		}
		out <- string(b[:n])
	}()
	return in, out, nil
}
