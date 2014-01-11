package main

import (
	"fmt"
	"net/http"

	"code.google.com/p/go.crypto/bcrypt"
	"github.com/gorilla/sessions"
)

type Context struct {
	Sess      *sessions.Session
	UserId    int
	Username  string
	Password  string
	FirstName string
	LastName  string
}

type view func(w http.ResponseWriter, r *http.Request) error
type contextView func(w http.ResponseWriter, r *http.Request, c *Context) error

func handler(v view) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := v(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func contextHandler(v contextView) http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		sess, _ := store.Get(r, "gotalk")
		userId, ok := sess.Values["user"].(int)
		if !ok {
			return v(w, r, &Context{Sess: sess, UserId: -1})
		}
		var username, password, firstName, lastName string
		sql := "SELECT username, password, first_name, last_name FROM users WHERE user_id = $1"
		err := db.QueryRow(sql, userId).Scan(&username, &password, &firstName, &lastName)
		if err != nil {
			return err
		}
		return v(w, r, &Context{sess, userId, username, password, firstName, lastName})
	})
}

func userHandler(v contextView) http.HandlerFunc {
	return contextHandler(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		if c.UserId < 0 {
			return render(w, "login", nil)
		}
		return v(w, r, c)
	})
}

func root(w http.ResponseWriter, r *http.Request, c *Context) error {
	return render(w, "root", nil)
}

func login(w http.ResponseWriter, r *http.Request, c *Context) error {
	username := r.FormValue("username")
	password := r.FormValue("password")
	var userId int
	var hash string
	sql := "SELECT user_id, password FROM users WHERE username = $1"
	if err := db.QueryRow(sql, username).Scan(&userId, &hash); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}
	c.Sess.Values["user"] = userId
	if err := c.Sess.Save(r, w); err != nil {
		return err
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func register(w http.ResponseWriter, r *http.Request, c *Context) error {
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	username := r.FormValue("username")
	password := r.FormValue("password")
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	sql := "INSERT INTO users (username, password, first_name, last_name) " +
		"VALUES ($1, $2, $3, $4)"
	if _, err := db.Exec(sql, username, string(hash), firstName, lastName); err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}
	var userId int
	sql = "SELECT user_id FROM users WHERE username = $1"
	if err := db.QueryRow(sql, username).Scan(&userId); err != nil {
		return err
	}
	c.Sess.Values["user"] = userId
	if err := c.Sess.Save(r, w); err != nil {
		return err
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}
