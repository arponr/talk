package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	"code.google.com/p/go.net/websocket"
)

type connection struct {
	io.ReadWriter
	user *user
}

type thread map[*connection]bool

var threads = struct {
	sync.RWMutex
	m map[int]thread
}{m: make(map[int]thread)}

func send(threadId int, src *connection) error {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			md := markdown(buf[0:nr])
			sql := "INSERT INTO messages (author, body, thread) VALUES ($1, $2, $3)"
			if _, ew := db.Exec(sql, src.user.id, md, threadId); ew != nil {
				return ew
			}
			threads.RLock()
			t := threads.m[threadId]
			threads.RUnlock()
			for dst, _ := range t {
				dst.Write(md)
			}
		}
		if er == io.EOF {
			return nil
		}
		if er != nil {
			return er
		}
	}
}

var (
	socketHandler = websocket.Handler(
		func(s *websocket.Conn) {
			r := s.Request()
			u, err := getUser(r)
			if err != nil {
				// serve error
			}
			c := &connection{s, u}
			threadId, err := strconv.Atoi(r.URL.Path[len("/socket/"):])
			if err != nil {
				// bad URL
			}
			threads.Lock()
			threads.m[threadId][c] = true
			threads.Unlock()
			if err := send(threadId, c); err != nil {
				log.Println(err)
			}
			threads.Lock()
			delete(threads.m[threadId], c)
			threads.Unlock()
		})

	threadHandler = userHandler(
		func(w http.ResponseWriter, r *http.Request, u *user) error {
			id, err := strconv.Atoi(r.URL.Path[len("/thread/"):])
			sql := "SELECT body FROM messages WHERE thread = $1"
			rows, err := db.Query(sql, id)
			if err != nil {
				return err
			}
			var msgs []string
			for rows.Next() {
				var msg string
				err = rows.Scan(&msg)
				if err != nil {
					return err
				}
				msgs = append(msgs, msg)
			}
			if err = rows.Err(); err != nil {
				return err
			}
			return render(w, "thread", msgs)
		})

	newThreadHandler = userHandler(
		func(w http.ResponseWriter, r *http.Request, u *user) error {
			switch r.Method {
			case "GET":
				return render(w, "newthread", nil)
			case "POST":
				name := r.FormValue("name")
				var threadId int
				sql := "INSERT INTO threads (thread_name) VALUES ($1) RETURNING thread_id"
				if err := db.QueryRow(sql, name).Scan(&threadId); err != nil {
					return err
				}
				threads.Lock()
				threads.m[threadId] = make(map[*connection]bool)
				threads.Unlock()
			}
			return nil
		})
)
