package main

import (
	"database/sql"
	"net/http"
	"sort"
	"time"
)

type thread struct {
	Id    int
	Name  string
	Users []string
	Time  time.Time
	Last  *message
}

type message struct {
	Username string
	RawBody  string
	FmtBody  string
	Markdown bool
	Tex      bool
	Time     time.Time
}

type user struct {
	id       int
	username string
	passhash []byte
}

type byTime []*thread

func (t byTime) Len() int           { return len(t) }
func (t byTime) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byTime) Less(i, j int) bool { return t[i].Time.After(t[j].Time) }

func getUser(r *http.Request) (*user, error) {
	sess, _ := store.Get(r, "gotalk")
	id, ok := sess.Values["user"].(int)
	if !ok {
		return nil, nil
	}
	var username string
	var passhash []byte
	q := "SELECT username, passhash FROM users WHERE user_id = $1"
	err := db.QueryRow(q, id).Scan(&username, &passhash)
	if err != nil {
		return nil, err
	}
	return &user{id, username, passhash}, nil
}

func threadUsers(threadId, userId int) (users []string, err error) {
	q := "SELECT u.username FROM users u, user_threads ut" +
		" WHERE ut.thread_id = $1 AND u.user_id = ut.user_id AND u.user_id != $2"
	rows, err := db.Query(q, threadId, userId)
	if err != nil {
		return
	}
	for rows.Next() {
		var u string
		err = rows.Scan(&u)
		if err != nil {
			return
		}
		users = append(users, u)
	}
	err = rows.Err()
	return
}

func userThreads(userId, threadId int) (threads []*thread, cur *thread, err error) {
	q := "SELECT t.thread_id, t.thread_name, t.time" +
		" FROM threads t, user_threads ut" +
		" WHERE ut.user_id = $1 AND t.thread_id = ut.thread_id"
	rows, err := db.Query(q, userId)
	if err != nil {
		return
	}
	for rows.Next() {
		var t thread
		err = rows.Scan(&t.Id, &t.Name, &t.Time)
		if err != nil {
			return
		}
		if t.Id == threadId {
			cur = &t
		}
		t.Users, err = threadUsers(t.Id, userId)
		if err != nil {
			return
		}
		t.Last, err = lastMessage(t.Id)
		if err != nil {
			return
		}
		if t.Last != nil {
			t.Time = t.Last.Time
		}
		threads = append(threads, &t)
	}
	if err = rows.Err(); err != nil {
		return
	}
	sort.Sort(byTime(threads))
	return
}

func threadMessages(threadId int) (msgs []*message, err error) {
	q := "SELECT username, raw_body, fmt_body, markdown, tex, time" +
		" FROM messages WHERE thread_id = $1"
	rows, err := db.Query(q, threadId)
	if err != nil {
		return
	}
	for rows.Next() {
		var m message
		err = rows.Scan(&m.Username, &m.RawBody, &m.FmtBody,
			&m.Markdown, &m.Tex, &m.Time)
		if err != nil {
			return
		}
		msgs = append(msgs, &m)
	}
	err = rows.Err()
	return
}

func lastMessage(threadId int) (*message, error) {
	q := "SELECT username, raw_body, fmt_body, markdown, tex, time" +
		" FROM messages WHERE thread_id = $1" +
		" ORDER BY message_id DESC LIMIT 1"
	var m message
	err := db.QueryRow(q, threadId).Scan(&m.Username, &m.RawBody, &m.FmtBody,
		&m.Markdown, &m.Tex, &m.Time)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &m, err
}

func usernameToId(usernames []string) (ids []int, err error) {
	q := "SELECT user_id FROM users WHERE username = $1"
	for _, u := range usernames {
		var id int
		if err = db.QueryRow(q, u).Scan(&id); err != nil {
			return
		}
		ids = append(ids, id)
	}
	return
}

func insertMessage(threadId int, m *message) error {
	q := "INSERT INTO messages (thread_id, username, raw_body, fmt_body," +
		" markdown, tex, time) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	m.Time = time.Now().UTC()
	_, err := db.Exec(q, threadId, m.Username, m.RawBody, m.FmtBody,
		m.Markdown, m.Tex, m.Time)
	return err
}

func insertThread(name string, users []int) (threadId int, err error) {
	q := "INSERT INTO threads (thread_name, time) VALUES ($1, $2)" +
		" RETURNING thread_id"
	if err = db.QueryRow(q, name, time.Now().UTC()).Scan(&threadId); err != nil {
		return
	}
	for _, u := range users {
		if err = insertUserThread(u, threadId); err != nil {
			return
		}
	}
	return
}

func insertUserThread(userId, threadId int) error {
	q := "INSERT INTO user_threads (user_id, thread_id) VALUES ($1, $2)"
	_, err := db.Exec(q, userId, threadId)
	return err
}

func userInThread(threadId, userId int) bool {
	q := "SELECT user_id FROM user_threads ut" +
		" WHERE ut.thread_id = $1 AND ut.user_id = $2"
	var g int
	return db.QueryRow(q, threadId, userId).Scan(&g) == nil
}
