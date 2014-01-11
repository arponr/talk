package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"code.google.com/p/go.net/websocket"
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

	store = sessions.NewCookieStore(
		securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))

	http.HandleFunc("/", userHandler(root))
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/login", contextHandler(login))
	http.HandleFunc("/register", contextHandler(register))
	http.Handle("/socket", websocket.Handler(socketHandler))
	http.Handle("/static/", http.FileServer(http.Dir(".")))

	err = http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatal(err)
	}
}
