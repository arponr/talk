package main

import (
	"fmt"
	"net/http"

	"code.google.com/p/go.crypto/bcrypt"
	"github.com/gorilla/sessions"
)

type Context struct {
	Session   *sessions.Session
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
		s, _ := store.Get(r, "gotalk")
		userId, ok := s.Values["user"].(int)
		if !ok {
			return v(w, r, &Context{Session: s, UserId: -1})
		}
		var username, password, firstName, lastName string
		err := db.QueryRow("SELECT username, password, first_name, last_name FROM users WHERE user_id = $1", userId).Scan(&username, &password, &firstName, &lastName)
		if err != nil {
			return err
		}
		return v(w, r, &Context{s, userId, username, password, firstName, lastName})
	})
}

func userHandler(v contextView) http.HandlerFunc {
	return contextHandler(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		if c.UserId < 0 {
			return loginTemplate.Execute(w, nil)
		}
		return v(w, r, c)
	})
}

func root(w http.ResponseWriter, r *http.Request, c *Context) error {
	return rootTemplate.Execute(w, nil)
}

func login(w http.ResponseWriter, r *http.Request, c *Context) error {
	username := r.FormValue("username")
	password := r.FormValue("password")
	var userId int
	var hash string
	err := db.QueryRow("SELECT user_id, password FROM users WHERE username = $1", username).Scan(&userId, &hash)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}
	c.Session.Values["user"] = userId
	if err = c.Session.Save(r, w); err != nil {
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
	fmt.Println(1)
	_, err = db.Exec("INSERT INTO users (username, password, first_name, last_name) VALUES ($1, $2, $3, $4)", username, string(hash), firstName, lastName)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}
	fmt.Println(2)
	var userId int
	err = db.QueryRow("SELECT user_id FROM users WHERE username = $1", username).Scan(&userId)
	if err != nil {
		return err
	}
	fmt.Println(3)
	c.Session.Values["user"] = userId
	if err = c.Session.Save(r, w); err != nil {
		return err
	}
	fmt.Println(4)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}
