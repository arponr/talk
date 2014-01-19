package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

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

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.Handle("/newthread", newThreadHandler)
	http.Handle("/thread/", threadHandler)
	http.Handle("/socket/", socketHandler)
	http.Handle("/static/", http.FileServer(http.Dir(".")))

	err = http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatal(err)
	}
}
