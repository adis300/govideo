package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var allMeetings map[string]*Meeting

func entranceHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(LoadView("home"))
}
func meetingHandler(w http.ResponseWriter, r *http.Request) {
	roomPath := mux.Vars(r)["room"]
	log.Println(roomPath)
	if len(roomPath) < 3 { // roomPath is not available or not valid
		w.Write([]byte("Invalid room url"))
	} else {
		owner := r.URL.Query().Get("on") // this is a hashed value of owner to be checked
		ownerName := r.URL.Query().Get("oname")
		participantName := r.URL.Query().Get("pname")
		User := ValidateUser(roomPath, owner, ownerName, participantName)
		fmt.Println(User)
		w.Write(LoadView("room"))
	}
}
func socketHandler(w http.ResponseWriter, r *http.Request) {
	roomPath := mux.Vars(r)["room"]
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	defer conn.Close()
	if err == nil {
		if len(roomPath) > 2 {
			for {
				_, msg, readErr := conn.ReadMessage()
				if readErr == nil {
					log.Println("Message received from client:" + roomPath)
					log.Println(string(msg))
				} else {
					log.Println(readErr)
					return
				}
				/*
					log.Println(string(msg))
					if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
						return
					}
				*/
			}
		}
	} else {
		log.Println(err)
	}
}

func socketSendMessage

// searches the books for the book with `id` and returns the book and it's index, or -1 for 404
func getMeeting(room string) (*Meeting, bool) {
	meeting := allMeetings[room]

	if meeting != nil {
		return meeting, true
	}
	return nil, false
}

func main() {
	//allConnections = make(map[*websocket.Conn]bool)
	allMeetings = make(map[string]*Meeting)

	// handle all requests by serving a file of the same name
	fileServer := http.FileServer(http.Dir("public"))

	router := mux.NewRouter()
	router.HandleFunc("/", entranceHandler)
	router.HandleFunc("/{room}", meetingHandler)
	router.PathPrefix("/public").Handler(http.StripPrefix("/public", fileServer))
	router.HandleFunc("/ws/{room}", socketHandler)
	http.Handle("/", router)

	log.Println("serving")
	if err := http.ListenAndServe(PORT, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
