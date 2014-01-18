package main

import (
	"net/http"

	"code.google.com/p/go.crypto/bcrypt"
)

type user struct {
	id        int
	username  string
	passhash  []byte
	firstName string
	lastName  string
}

func getUser(r *http.Request) (*user, error) {
	sess, _ := store.Get(r, "gotalk")
	id, ok := sess.Values["user"].(int)
	if !ok {
		return &user{id: -1}, nil
	}
	var username, firstName, lastName string
	var passhash []byte
	sql := "SELECT username, passhash, first_name, last_name FROM users WHERE user_id = $1"
	err := db.QueryRow(sql, id).Scan(&username, &passhash, &firstName, &lastName)
	if err != nil {
		return nil, err
	}
	return &user{id, username, passhash, firstName, lastName}, nil
}

type view func(http.ResponseWriter, *http.Request) error
type userView func(http.ResponseWriter, *http.Request, *user) error

func handler(v view) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := v(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func userHandler(v userView) http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		u, err := getUser(r)
		if err != nil {
			return err
		}
		if u.id < 0 {
			return render(w, "login", nil)
		}
		return v(w, r, u)
	})
}

var (
	rootHandler = userHandler(
		func(w http.ResponseWriter, r *http.Request, u *user) error {
			return render(w, "root", nil)
		})

	loginHandler = handler(
		func(w http.ResponseWriter, r *http.Request) error {
			username := r.FormValue("username")
			password := r.FormValue("password")
			var id int
			var passhash []byte
			sql := "SELECT user_id, passhash FROM users WHERE username = $1"
			if err := db.QueryRow(sql, username).Scan(&id, &passhash); err != nil {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return nil
			}
			if err := bcrypt.CompareHashAndPassword(passhash, []byte(password)); err != nil {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return nil
			}
			sess, _ := store.Get(r, "gotalk")
			sess.Values["user"] = id
			if err := sess.Save(r, w); err != nil {
				return err
			}
			http.Redirect(w, r, "/thread/1", http.StatusSeeOther)
			return nil
		})

	registerHandler = handler(
		func(w http.ResponseWriter, r *http.Request) error {
			firstName := r.FormValue("first_name")
			lastName := r.FormValue("last_name")
			username := r.FormValue("username")
			password := r.FormValue("password")
			passhash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			var id int
			sql := "INSERT INTO users (username, passhash, first_name, last_name) " +
				"VALUES ($1, $2, $3, $4) RETURNING user_id"
			if err = db.QueryRow(
				sql, username, passhash, firstName, lastName).Scan(&id); err != nil {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return nil
			}
			sess, _ := store.Get(r, "gotalk")
			sess.Values["user"] = id
			if err = sess.Save(r, w); err != nil {
				return err
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return nil
		})
)
