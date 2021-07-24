package uds

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

const sockPath = "./testing.sock"

func rmSockFile() error {
	_, err := os.Stat(sockPath)
	if !os.IsNotExist(err) {
		err := os.Remove(sockPath)
		if err != nil {
			log.Error(err)
			return err
		}
		return nil
	}
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func TestIsIPCListening(t *testing.T) {
	isListening := IsIPCListening(sockPath)
	if isListening {
		t.Fatal("Listener is not running yet online")
	}
}

func TestListener(t *testing.T) {
	if err := rmSockFile(); err != nil {
		t.Fatal(err)
	}
	opts := Options{
		Size:       0,
		SocketPath: sockPath,
	}
	_, err := Listener(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Listener(context.Background(), opts)
	if err == nil {
		t.Fatal(err)
	}
	isListening := IsIPCListening(sockPath)
	if !isListening {
		t.Fatal("Listener is not running yet online")
	}
}

func TestDialer(t *testing.T) {
	if err := rmSockFile(); err != nil {
		t.Fatal(err)
	}
	opts := Options{
		Size:       0,
		SocketPath: sockPath,
	}
	_, _, _, err := Dialer(opts)
	if err == nil {
		t.Fatal(err)
	}
	_, err = Listener(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, err = Dialer(opts)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDialersWithLimit(t *testing.T) {
	if err := rmSockFile(); err != nil {
		t.Fatal(err)
	}
	opts := Options{
		Size:       100,
		SocketPath: sockPath,
	}
	in, err := Listener(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	go func() {
		for {
			select {
			case conn := <-in:
				count++
				go func(conn *Client) {
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
			case <-time.After(time.Minute * 2):
				fmt.Println("processed ", count)
				t.Fatal("lister could not process all dialers")
			}

		}
	}()
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			r, w, c, err := Dialer(opts)
			if err != nil {
				return
			}
			defer c()
			for j := 0; j < 100000; j++ {
				msg := fmt.Sprintf("%ddialer%d", i, j)
				w(msg)
				res, err := r()
				if err != nil {
					log.Error("Read error ", err)
				}
				if res != msg+"return" {
					t.Fatalf("Got %s instread of %sreturn", res, msg)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestDialersWithoutLimit(t *testing.T) {
	if err := rmSockFile(); err != nil {
		t.Fatal(err)
	}
	opts := Options{
		Size:       0,
		SocketPath: sockPath,
	}
	in, err := Listener(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	go func() {
		for {
			select {
			case conn := <-in:
				count++
				go func(conn *Client) {
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
			case <-time.After(time.Second * 5):
				fmt.Println("processed ", count)
				t.Fatal("lister could not process all dialers")
			}

		}
	}()
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			r, w, c, err := Dialer(opts)
			if err != nil {
				return
			}
			defer c()
			for j := 0; j < 100000; j++ {
				msg := fmt.Sprintf("%ddialer%d", i, j)
				w(msg)
				res, err := r()
				if err != nil {
					log.Error("Read error ", err)
				}
				if res != msg+"return" {
					t.Fatalf("Got %s instread of %sreturn", res, msg)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
