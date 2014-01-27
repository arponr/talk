package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"code.google.com/p/go.net/websocket"
	gc "github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

var db *sql.DB
var store sessions.Store

func dburl() string {
	// return os.Getenv("DATABASE_URL")
	regex := regexp.MustCompile("(?i)^postgres://(?:([^:@]+):([^@]*)@)?([^@/:]+):(\\d+)/(.*)$")
	matches := regex.FindStringSubmatch(os.Getenv("DATABASE_URL"))
	if matches == nil {
		log.Fatalf("DATABASE_URL variable must look like: "+
			"postgres://username:password@hostname:port/dbname (not '%v')",
			os.Getenv("DATABASE_URL"))
	}
	sslmode := os.Getenv("PGSSL")
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		matches[1], matches[2], matches[3], matches[4], matches[5], sslmode)
}

func main() {
	var err error
	db, err = sql.Open("postgres", dburl())
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
