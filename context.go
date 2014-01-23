package main

import (
	"net/http"

	"code.google.com/p/go.crypto/bcrypt"
	"github.com/gorilla/sessions"
)

type context struct {
	sess *sessions.Session
	user *user
}

type view func(http.ResponseWriter, *http.Request, *context) error

func handler(v view, requireLogin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			if err != nil {
				serveError(w, err)
			}
		}()
		var c context
		c.sess, _ = store.Get(r, "gotalk")
		if requireLogin {
			c.user, err = getUser(r)
			if err != nil {
				return
			}
			if c.user == nil {
				flashes := c.sess.Flashes()
				if err = c.sess.Save(r, w); err != nil {
					return
				}
				err = render(w, "login", flashes)
				return
			}
		}
		err = v(w, r, &c)
	}
}

func login(w http.ResponseWriter, r *http.Request, c *context) (err error) {
	defer func() {
		err = c.sess.Save(r, w)
		if err == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()
	username := r.FormValue("username")
	password := r.FormValue("password")
	var id int
	var passhash []byte
	sql := "SELECT user_id, passhash FROM users WHERE username = $1"
	if err = db.QueryRow(sql, username).Scan(&id, &passhash); err != nil {
		c.sess.AddFlash("invalid username")
		return nil
	}
	if err = bcrypt.CompareHashAndPassword(passhash, []byte(password)); err != nil {
		c.sess.AddFlash("username and password do not match")
		return nil
	}
	c.sess.Values["user"] = id
	return
}

func register(w http.ResponseWriter, r *http.Request, c *context) (err error) {
	defer func() {
		err = c.sess.Save(r, w)
		if err == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()
	username := r.FormValue("username")
	password := r.FormValue("password")
	passwordAgain := r.FormValue("password_again")
	if password != passwordAgain {
		c.sess.AddFlash("passwords do not match")
		return nil
	}
	if len(password) < 6 {
		c.sess.AddFlash("password must be at least 6 characters")
		return nil
	}
	passhash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	var id int
	sql := "INSERT INTO users (username, passhash) VALUES ($1, $2) RETURNING user_id"
	if err = db.QueryRow(sql, username, passhash).Scan(&id); err != nil {
		c.sess.AddFlash("username is taken already")
		return nil
	}
	c.sess.Values["user"] = id
	return
}
