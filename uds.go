package uds

import (
	"context"
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
				go func(conn net.Conn) {
					dontWait := make(chan struct{})
					go func(c net.Conn) {
						for {
							select {
							case <-dontWait:
								c.Close()
								return
							case in := <-external:
								_, err = c.Write([]byte(in))
								if err != nil {
									log.Error("Write error", err)
									c.Close()
									return
								}
							}
						}
					}(conn)
					for {
						b := make([]byte, opts.Size)
						n, err := conn.Read(b)
						if err != nil {
							if err == io.EOF {
								dontWait <- struct{}{}
								return
							} else {
								log.Error("Read error :", err)
								conn.Close()
								return
							}
						}
						if n > 0 {
							out <- string(b[:n])
						}
					}
				}(conn)
			}
		}
	}()
	return out, external, nil
}

// Dialer creates a uds dialer
func Dialer(opts Options) (Read func() (string, error), Write func(st string) error, Close func() error, err error) {
	Close = func() error {
		return nil
	}
	Write = func(st string) error {
		return nil
	}
	Read = func() (string, error) {
		return "", nil
	}
	if opts.SocketPath == "" {
		return Read, Write, Close, errors.New("invalid socket path")
	}
	if opts.Size == 0 {
		opts.Size = DefaultSize
	}
	conn, err := net.Dial("unix", opts.SocketPath)
	if err != nil {
		return Read, Write, Close, err
	}
	Close = func() error {
		return conn.Close()
	}
	Write = func(st string) error {
		_, err := conn.Write([]byte(st))
		if err != nil {
			log.Error(err)
			return err
		}
		return nil
	}
	Read = func() (string, error) {
		b := make([]byte, opts.Size)
		n, err := conn.Read(b)
		if err != nil && err != io.EOF {
			log.Error("Read error :", err)
			conn.Close()
			return "", err
		}
		return string(b[:n]), nil
	}
	return
}

func IsIPCListing(socketPath string) bool {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
