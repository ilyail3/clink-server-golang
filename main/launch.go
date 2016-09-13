package main

import (
	"github.com/op/go-logging"
	"os"

	"clink.com/server"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main(){
	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)

	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)

	srv, err := server.NewServer("/tmp/clink.sqlite3", logging.MustGetLogger("server"))
	defer srv.Close()

	if(err != nil){
		panic("Error opening db:" + err.Error())
	}

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/connect", srv.Connect).
		Methods("POST").
		Headers("Content-Type", "application/json")

	log.Fatal(http.ListenAndServe(":8080", router))
}
