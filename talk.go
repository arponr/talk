package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
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

func socket(s *websocket.Conn) {
	r := s.Request()
	u, err := getUser(r)
	if err != nil || u == nil {
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
}

func readThread(w http.ResponseWriter, r *http.Request, c *context) (err error) {
	threadId, err := strconv.Atoi(r.URL.Path[len("/thread/"):])
	if err != nil {
		serveDNE(w, r)
		return nil
	}
	if !userInThread(threadId, c.user.id) {
		serveDNE(w, r)
		return nil
	}
	var data struct {
		Threads  []thread
		Messages []message
	}
	data.Threads, err = userThreads(c.user.id)
	if err != nil {
		return err
	}
	data.Messages, err = threadMessages(threadId)
	if err != nil {
		return err
	}
	return render(w, "thread", data)
}

func root(w http.ResponseWriter, r *http.Request, c *context) (err error) {
	var data struct {
		Threads []thread
	}
	data.Threads, err = userThreads(c.user.id)
	if err != nil {
		return err
	}
	return render(w, "root", data)
}

func newThread(w http.ResponseWriter, r *http.Request, c *context) (err error) {
	threadName := r.FormValue("name")
	usernames := strings.Split(r.FormValue("users"), " ")
	var threadId int
	sql := "INSERT INTO threads (thread_name) VALUES ($1) RETURNING thread_id"
	if err = db.QueryRow(sql, threadName).Scan(&threadId); err != nil {
		return err
	}
	ids, err := nameToId(usernames)
	if err != nil {
		return err
	}
	ids = append(ids, c.user.id)
	sql = "INSERT INTO user_threads (user_id, thread_id) VALUES ($1, $2)"
	for _, id := range ids {
		if _, err = db.Exec(sql, id, threadId); err != nil {
			return err
		}
	}
	cache.Lock()
	cache.m[threadId] = make(connSet)
	cache.Unlock()
	http.Redirect(w, r, "/thread/"+strconv.Itoa(threadId), http.StatusSeeOther)
	return nil
}
