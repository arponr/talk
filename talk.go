package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"

	"code.google.com/p/go.net/websocket"
)

type connection struct {
	io.ReadWriter
	user *user
}

type connSet map[*connection]bool

var cache = struct {
	sync.RWMutex
	m map[int]connSet
}{m: make(map[int]connSet)}

func initCache() error {
	rows, err := db.Query("SELECT thread_id FROM threads")
	if err != nil {
		return err
	}
	for rows.Next() {
		var threadId int
		err = rows.Scan(&threadId)
		if err != nil {
			return err
		}
		cache.m[threadId] = make(connSet)
	}
	return rows.Err()
}

func send(threadId int, src *connection) error {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			md := markdown(buf[0:nr])
			msg := message{src.user.username, md}
			sql := "INSERT INTO messages (username, body, thread_id) VALUES ($1, $2, $3)"
			if _, ew := db.Exec(sql, msg.Username, msg.Body, threadId); ew != nil {
				return ew
			}
			cache.RLock()
			t := cache.m[threadId]
			cache.RUnlock()
			for dst, _ := range t {
				enc := json.NewEncoder(dst)
				enc.Encode(&msg)
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
			if err != nil || u.id < 0 {
				s.Write([]byte("Your login is invalid"))
				return
			}
			c := &connection{s, u}
			threadId, err := strconv.Atoi(r.URL.Path[len("/socket/"):])
			if err != nil {
				s.Write([]byte("You're trying to write to an invalid thread url"))
			}
			cache.Lock()
			cache.m[threadId][c] = true
			cache.Unlock()
			if err := send(threadId, c); err != nil {
				s.Write([]byte("Your connection closed with an error:" + err.Error()))
			}
			cache.Lock()
			delete(cache.m[threadId], c)
			cache.Unlock()
		})

	threadHandler = userHandler(
		func(w http.ResponseWriter, r *http.Request, u *user) error {
			threadId, err := strconv.Atoi(r.URL.Path[len("/thread/"):])
			if err != nil {
				return err
			}
			var data struct {
				Threads  []thread
				Messages []message
			}
			data.Threads, err = userThreads(u.id)
			if err != nil {
				return err
			}
			data.Messages, err = threadMessages(threadId)
			if err != nil {
				return err
			}
			return render(w, "thread", data)
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
				cache.Lock()
				cache.m[threadId] = make(connSet)
				cache.Unlock()
				http.Redirect(w, r, "/thread/"+strconv.Itoa(threadId), http.StatusSeeOther)
			}
			return nil
		})
)
