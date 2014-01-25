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
	Body     string
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
	stmt := "SELECT username, passhash FROM users WHERE user_id = $1"
	err := db.QueryRow(stmt, id).Scan(&username, &passhash)
	if err != nil {
		return nil, err
	}
	return &user{id, username, passhash}, nil
}

func threadUsers(threadId, userId int) ([]string, error) {
	stmt := "SELECT u.username FROM users u, user_threads ut " +
		"WHERE ut.thread_id = $1 AND u.user_id = ut.user_id AND u.user_id != $2"
	rows, err := db.Query(stmt, threadId, userId)
	if err != nil {
		return nil, err
	}
	var users []string
	for rows.Next() {
		var u string
		err = rows.Scan(&u)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func userInThread(threadId, userId int) bool {
	stmt := "SELECT user_id FROM user_threads ut " +
		"WHERE ut.thread_id = $1 AND ut.user_id = $2"
	var g int
	return db.QueryRow(stmt, threadId, userId).Scan(&g) == nil
}

func userThreads(userId int) ([]*thread, error) {
	stmt := "SELECT t.thread_id, t.thread_name, t.time FROM threads t, user_threads ut " +
		"WHERE ut.user_id = $1 AND t.thread_id = ut.thread_id"
	rows, err := db.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	var threads []*thread
	for rows.Next() {
		var t thread
		err = rows.Scan(&t.Id, &t.Name, &t.Time)
		if err != nil {
			return nil, err
		}
		t.Users, err = threadUsers(t.Id, userId)
		if err != nil {
			return nil, err
		}
		t.Last, err = lastMessage(t.Id)
		if err != nil {
			return nil, err
		}
		if t.Last != nil {
			t.Time = t.Last.Time
		}
		threads = append(threads, &t)
	}
	sort.Sort(byTime(threads))
	return threads, rows.Err()
}

func threadMessages(threadId int) ([]*message, error) {
	stmt := "SELECT username, body, tex, time FROM messages WHERE thread_id = $1"
	rows, err := db.Query(stmt, threadId)
	if err != nil {
		return nil, err
	}
	var msgs []*message
	for rows.Next() {
		var m message
		err = rows.Scan(&m.Username, &m.Body, &m.Tex, &m.Time)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, &m)
	}
	return msgs, rows.Err()
}

func lastMessage(threadId int) (*message, error) {
	stmt := "SELECT username, body, tex, time FROM messages WHERE thread_id = $1 " +
		"ORDER BY message_id DESC LIMIT 1"
	var m message
	err := db.QueryRow(stmt, threadId).Scan(&m.Username, &m.Body, &m.Tex, &m.Time)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &m, err
}

func usernameToId(usernames []string) ([]int, error) {
	var ids []int
	stmt := "SELECT user_id FROM users WHERE username = $1"
	for _, u := range usernames {
		var id int
		if err := db.QueryRow(stmt, u).Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func insertMessage(threadId int, m *message) error {
	stmt := "INSERT INTO messages (username, body, tex, time, thread_id) " +
		"VALUES ($1, $2, $3, $4, $5)"
	m.Time = time.Now().UTC()
	_, err := db.Exec(stmt, m.Username, m.Body, m.Tex, m.Time, threadId)
	return err
}

func insertThread(name string, users []int) (threadId int, err error) {
	stmt := "INSERT INTO threads (thread_name, time) VALUES ($1, $2) RETURNING thread_id"
	if err = db.QueryRow(stmt, name, time.Now().UTC()).Scan(&threadId); err != nil {
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
	stmt := "INSERT INTO user_threads (user_id, thread_id) VALUES ($1, $2)"
	_, err := db.Exec(stmt, userId, threadId)
	return err
}
