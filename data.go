package main

import "net/http"

type thread struct {
	Id    int
	Name  string
	Users []string
	Last  message
}

type message struct {
	Username string
	Body     string
}

type user struct {
	id       int
	username string
	passhash []byte
}

func getUser(r *http.Request) (*user, error) {
	sess, _ := store.Get(r, "gotalk")
	id, ok := sess.Values["user"].(int)
	if !ok {
		return nil, nil
	}
	var username string
	var passhash []byte
	sql := "SELECT username, passhash FROM users WHERE user_id = $1"
	err := db.QueryRow(sql, id).Scan(&username, &passhash)
	if err != nil {
		return nil, err
	}
	return &user{id, username, passhash}, nil
}

func threadUsers(threadId, userId int) ([]string, error) {
	sql := "SELECT u.username FROM users u, user_threads ut " +
		"WHERE ut.thread_id = $1 AND u.user_id = ut.user_id AND u.user_id != $2"
	rows, err := db.Query(sql, threadId, userId)
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
	sql := "SELECT user_id FROM user_threads ut " +
		"WHERE ut.thread_id = $1 AND ut.user_id = $2"
	var g int
	return db.QueryRow(sql, threadId, userId).Scan(&g) == nil
}

func userThreads(userId int) ([]thread, error) {
	sql := "SELECT t.thread_id, t.thread_name FROM threads t, user_threads ut " +
		"WHERE ut.user_id = $1 AND t.thread_id = ut.thread_id"
	rows, err := db.Query(sql, userId)
	if err != nil {
		return nil, err
	}
	var threads []thread
	for rows.Next() {
		var t thread
		err = rows.Scan(&t.Id, &t.Name)
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
		threads = append(threads, t)
	}
	return threads, rows.Err()
}

func threadMessages(threadId int) ([]message, error) {
	sql := "SELECT username, body FROM messages WHERE thread_id = $1"
	rows, err := db.Query(sql, threadId)
	if err != nil {
		return nil, err
	}
	var msgs []message
	for rows.Next() {
		var m message
		err = rows.Scan(&m.Username, &m.Body)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func lastMessage(threadId int) (message, error) {
	sql := "SELECT username, body FROM messages WHERE thread_id = $1 " +
		"ORDER BY message_id DESC LIMIT 1"
	var m message
	err := db.QueryRow(sql, threadId).Scan(&m.Username, &m.Body)
	return m, err
}

func nameToId(names []string) ([]int, error) {
	var ids []int
	sql := "SELECT user_id FROM users WHERE username = $1"
	for _, v := range names {
		var id int
		if err := db.QueryRow(sql, v).Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
