package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func echo(w http.ResponseWriter, r *http.Request) {

	intervalParam, ok := r.URL.Query()["interval"]

	if !ok || len(intervalParam[0]) < 1 {
		log.Println("Url Param 'interval' is missing")
		return
	}

	intervalDuration, err := time.ParseDuration(intervalParam[0])
	if err != nil {
		log.Println("Url Param 'interval' is not a duration")
		return
	}

	backlogParam, ok := r.URL.Query()["backlog"]

	if !ok || len(backlogParam[0]) < 1 {
		log.Println("Url Param 'backlog' is missing")
		return
	}

	backlogDuration, err := time.ParseDuration(backlogParam[0])
	if err != nil {
		log.Println("Url Param 'backlog' is not a duration")
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	aggregateByInterval(intervalDuration, backlogDuration, c)
}

func startServer(address string) error {
	http.HandleFunc("/ws", echo)
	return http.ListenAndServe(address, nil)
}
