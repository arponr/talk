package main

import (
	"encoding/json"
	"html/template"
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

type encSet map[*json.Encoder]bool

var cache = struct {
	sync.RWMutex
	m map[int]encSet
}{m: make(map[int]encSet)}

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
		cache.m[threadId] = make(encSet)
	}
	return rows.Err()
}

func send(threadId int, c *connection) (err error) {
	src := json.NewDecoder(c)
	for {
		var m message
		if err = src.Decode(&m); err == io.EOF {
			return nil
		} else if err != nil {
			return
		}
		if m.Markdown {
			m.Body = markdown(m.Body, m.Tex)
		} else {
			m.Body = template.HTMLEscapeString(m.Body)
		}
		m.Username = c.user.username
		stmt := "INSERT INTO messages (username, body, tex, thread_id) " +
			"VALUES ($1, $2, $3, $4) RETURNING time"
		err = db.QueryRow(stmt, m.Username, m.Body, m.Tex, threadId).Scan(&m.Time)
		if err != nil {
			return
		}
		cache.RLock()
		conns := cache.m[threadId]
		cache.RUnlock()
		for dst, _ := range conns {
			dst.Encode(&m)
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
	enc := json.NewEncoder(s)
	cache.Lock()
	cache.m[threadId][enc] = true
	cache.Unlock()
	if err := send(threadId, c); err != nil {
		s.Write([]byte("Your connection closed with an error:" + err.Error()))
	}
	cache.Lock()
	delete(cache.m[threadId], enc)
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
		Threads  []*thread
		Messages []*message
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
		Threads []*thread
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
	stmt := "INSERT INTO threads (thread_name) VALUES ($1) RETURNING thread_id"
	if err = db.QueryRow(stmt, threadName).Scan(&threadId); err != nil {
		return err
	}
	ids, err := nameToId(usernames)
	if err != nil {
		return err
	}
	ids = append(ids, c.user.id)
	stmt = "INSERT INTO user_threads (user_id, thread_id) VALUES ($1, $2)"
	for _, id := range ids {
		if _, err = db.Exec(stmt, id, threadId); err != nil {
			return err
		}
	}
	cache.Lock()
	cache.m[threadId] = make(encSet)
	cache.Unlock()
	http.Redirect(w, r, "/thread/"+strconv.Itoa(threadId), http.StatusSeeOther)
	return nil
}
