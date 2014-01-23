package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"code.google.com/p/go.net/websocket"
	gc "github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

var db *sql.DB
var store sessions.Store

func main() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	if err = initCache(); err != nil {
		log.Fatal(err)
	}
	store = sessions.NewCookieStore(
		securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))

	http.HandleFunc("/", handler(root, true))
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/login", handler(login, false))
	http.HandleFunc("/register", handler(register, false))
	http.Handle("/thread/", handler(readThread, true))
	http.Handle("/newthread", handler(newThread, true))
	http.Handle("/socket/", websocket.Handler(socket))
	http.Handle("/static/", http.FileServer(http.Dir(".")))

	err = http.ListenAndServe(":"+os.Getenv("PORT"), gc.ClearHandler(http.DefaultServeMux))
	if err != nil {
		log.Fatal(err)
	}
}
