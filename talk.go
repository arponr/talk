package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"code.google.com/p/go.net/websocket"
)

func main() {
	http.HandleFunc("/", rootHandler)
	http.Handle("/socket", websocket.Handler(socketHandler))
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatal(err)
	}
}

var rootTemplate = template.Must(template.ParseFiles("root.html"))

func rootHandler(w http.ResponseWriter, r *http.Request) {
	rootTemplate.Execute(w, nil)
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
	errc := make(chan error, 1)
	go send(u, v, errc)
	go send(v, u, errc)
	if err := <-errc; err != nil {
		log.Println(err)
	}
	u.Close()
	v.Close()
}

// modified io.Copy
func send(dst io.Writer, src io.ReadWriter, errc chan<- error) {
	var err error
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			md := markdown(buf[0:nr])

			nw, ew := src.Write(md)
			if ew != nil {
				err = ew
				break
			}
			if nw != len(md) {
				err = io.ErrShortWrite
				break
			}

			nw, ew = dst.Write(md)
			if ew != nil {
				err = ew
				break
			}
			if nw != len(md) {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	errc <- err
}
