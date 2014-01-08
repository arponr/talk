package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

const addr = "localhost:8080"

var partner = make(chan io.ReadWriteCloser)

func send(t io.Writer, f io.Reader, errch chan<- error) {
	_, err := io.Copy(t, f)
	errch <- err
}

func talk(u, v io.ReadWriteCloser) {
	fmt.Fprintln(u, "Found one!")
	fmt.Fprintln(v, "Found one!")
	errch := make(chan error, 1)
	go send(u, v, errch)
	go send(v, u, errch)
	if err := <-errch; err != nil {
		log.Println(err)
	}
	u.Close()
	v.Close()
}

func match(u io.ReadWriteCloser) {
	fmt.Fprint(u, "Waiting for a partner...")
	select {
	case partner <- u:
	case v := <-partner:
		talk(u, v)
	}
}

func main() {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		u, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go match(u)
	}
}
