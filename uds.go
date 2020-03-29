package uds

import (
	"errors"
	"net"
)

const (
	DEFAULT_SIZE = 512
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

// Listener creates a uds listener
func Listener(opts ListenerOptions) chan Envelop {
	out := make(chan Envelop)
	r := Envelop{
		Data: "",
		Error: nil,
	}
	if opts.SocketPath == "" {
		r.Error = errors.New("invalid socket path")
		out <- r
		return out
	}
	if opts.Size == 0 {
		opts.Size = DEFAULT_SIZE
	}
	l, err := net.Listen("unix", opts.SocketPath)
	if err != nil {
		r.Error = err
		out <- r
		return out
	}
	go func() {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				r.Error = err
				out <- r
			}
			for {
				b := make([]byte, opts.Size)
				_, err := conn.Read(b)
				if err != nil {
					conn.Close()
					break
				}
				r.Data = string(b)
				out <- r
			}
		}
	}()
	return out
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