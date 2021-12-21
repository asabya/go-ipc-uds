package uds

import (
	"context"
	"errors"
	"io"
	"net"
)

type noLogger struct{}

func (noLogger) Debug(args ...interface{}) {}

func (noLogger) Error(args ...interface{}) {}

// Options is used to create Listener/Dialer.
// SocketPath is the location of the `sock` file.
// Size is the size for byte array
type Options struct {
	Size       uint64
	SocketPath string
	Logger     Logger
}

type Client struct {
	Conn net.Conn
	Size uint64

	logger Logger
}

func (c *Client) Read() ([]byte, error) {
	var b []byte
	if c.Size == 0 {
		// make a temporary bytes var to read from the connection
		tmp := make([]byte, 1024)
		// make 0 length data bytes (since we'll be appending)
		data := make([]byte, 0)
		// keep track of full length read
		length := 0
		for {
			n, err := c.Conn.Read(tmp)
			if err != nil {
				if err != io.EOF {
					c.logger.Error("Read error :", err)
					c.Conn.Close()
				}
				break
			}
			data = append(data, tmp[:n]...)
			length += n
			if n < 1024 {
				break
			}
		}
		return data[:length], nil
	} else {
		b = make([]byte, c.Size)
		n, err := c.Conn.Read(b)
		if err != nil {
			if err == io.EOF {
				return []byte{}, nil
			} else {
				c.logger.Error("Read error :", err)
				c.Conn.Close()
				return nil, err
			}
		}
		return b[:n], nil
	}

}

func (c *Client) Write(d []byte) error {
	_, err := c.Conn.Write(d)
	if err != nil {
		c.Conn.Close()
		return err
	}
	return nil
}

func (c *Client) Close() error {
	return c.Conn.Close()
}

// Listener creates a uds listener
func Listener(ctx context.Context, opts Options) (chan *Client, error) {
	if opts.Logger == nil {
		opts.Logger = &noLogger{}
	}
	in := make(chan *Client)
	if opts.SocketPath == "" {
		return in, errors.New("invalid socket path")
	}
	l, err := net.Listen("unix", opts.SocketPath)
	if err != nil {
		return in, err
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
					opts.Logger.Error("Listener error", err)
					return
				}
				in <- &Client{
					Conn:   conn,
					Size:   opts.Size,
					logger: opts.Logger,
				}
			}
		}
	}()
	return in, nil
}

// Dialer creates a uds dialer
func Dialer(opts Options) (Read func() (string, error), Write func(st string) error, Close func() error, err error) {
	if opts.Logger == nil {
		opts.Logger = &noLogger{}
	}
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
			opts.Logger.Error(err)
			return err
		}
		return nil
	}
	Read = func() (string, error) {
		opts.Logger.Debug("dialer read")
		var b []byte
		if opts.Size == 0 {
			// make a temporary bytes var to read from the connection
			tmp := make([]byte, 1024)
			// make 0 length data bytes (since we'll be appending)
			data := make([]byte, 0)
			// keep track of full length read
			length := 0
			for {
				n, err := conn.Read(tmp)
				if err != nil {
					if err != io.EOF {
						opts.Logger.Error("Read error :", err)
						conn.Close()
					}
					break
				}
				data = append(data, tmp[:n]...)
				length += n
				if n < 1024 {
					break
				}
			}
			return string(data[:length]), nil
		} else {
			b = make([]byte, opts.Size)
			n, err := conn.Read(b)
			if err != nil && err != io.EOF {
				opts.Logger.Error("Read error :", err)
				conn.Close()
				return "", err
			}
			return string(b[:n]), nil
		}
	}
	return
}

func IsIPCListening(socketPath string) bool {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
