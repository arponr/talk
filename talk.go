package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"code.google.com/p/go.net/websocket"
)

const addr = "localhost:4000"

func main() {
	http.HandleFunc("/", rootHandler)
	http.Handle("/socket", websocket.Handler(socketHandler))
	http.Handle("/static/", http.FileServer(http.Dir("./")))
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

var rootTemplate = template.Must(template.ParseFiles("root.html"))

func rootHandler(w http.ResponseWriter, r *http.Request) {
	rootTemplate.Execute(w, addr)
}

type socket struct {
	io.ReadWriter
	done chan bool
}

func (s socket) Close() error {
	s.done <- true
	return nil
}

func socketHandler(ws *websocket.Conn) {
	s := socket{ws, make(chan bool)}
	go match(s)
	<-s.done
}

var partner = make(chan io.ReadWriteCloser)

func match(u io.ReadWriteCloser) {
	fmt.Fprint(u, "Waiting for a partner...")
	select {
	case partner <- u:
	case v := <-partner:
		talk(u, v)
	}
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

func send(t io.Writer, f io.Reader, errch chan<- error) {
	_, err := io.Copy(t, f)
	errch <- err
}
