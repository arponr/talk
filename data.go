package main

type thread struct {
	Id    int
	Name  string
	Users []string
}

type message struct {
	Username string
	Body     string
}

type data struct {
	Threads  []thread
	Messages []message
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
