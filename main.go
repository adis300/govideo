package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Meeting represents information about dogs.
type Meeting struct {
	CreatorName string
	Room        string
	Locked      bool
	Users       map[*websocket.Conn]*User
}

// User is a struct that has connections;
type User struct {
	Name    string
	IsOwner bool
}

// ClientMessage is a client message structure record
type ClientMessage struct {
	Type    int
	Content string
}

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
		w.Write(LoadView("room"))
	}
}

func roomMessageHandler(room string, msg []byte, conn *websocket.Conn) {
	log.Println("Message received from client:" + room)
	var clientMessage ClientMessage
	log.Println(string(msg))
	json.Unmarshal(msg, &clientMessage)
	switch clientMessage.Type {
	case UPDATE_DISPLAY_NAME:
		updateDisplayName(room, clientMessage.Content, conn)
	case LOCK_ROOM:
		lockRoom(room, clientMessage.Content)
	default:
		log.Println("Unknown message received")
	}
}

func updateDisplayName(room string, newName string, conn *websocket.Conn) {
}
func lockRoom(room string, lockCmd string) {
	// getMeeting from here.
	lockFlag := lockCmd == "lc"
	if lockFlag {
		log.Println("Broadcast a lock message")
	} else {
		log.Println("Broadcast an unlock message")
	}
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	roomPath := mux.Vars(r)["room"]
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	defer removeConnection(conn, roomPath) // Gauranteed removal of a connection
	if err == nil {
		if len(roomPath) > 2 {
			isOwner := false
			displayName := "Test name"
			_ = addUserToRoom(roomPath, conn, displayName, isOwner)
			for {
				_, msg, readErr := conn.ReadMessage()
				if readErr == nil {
					roomMessageHandler(roomPath, msg, conn)
				} else {
					log.Println(readErr)
					return
				}
			}
		}
	} else {
		log.Println(err)
	}
}

func getMeeting(room string) *Meeting {
	return allMeetings[room]
}

func createUser(name string, isOwner bool) *User {
	return &User{Name: name, IsOwner: isOwner}
}

func addUserToRoom(room string, conn *websocket.Conn, name string, isOwner bool) *Meeting {
	meeting := getMeeting(room)
	user := createUser(name, isOwner)

	if meeting == nil {
		meeting = createMeeting(room, conn, user)
	} else {
		log.Println("Meeting found: " + room)

		for userConnection := range meeting.Users {
			if err := userConnection.WriteMessage(websocket.TextMessage, []byte("New person has arrived")); err != nil {
				deleteUser(userConnection, meeting)
			}
		}
		meeting.Users[conn] = user
		log.Println("User added")
	}
	return meeting
}

func deleteUser(conn *websocket.Conn, meeting *Meeting) {
	defer conn.Close()
	delete(meeting.Users, conn)
	if len(meeting.Users) > 0 {
		for userConnection, user := range meeting.Users {
			if err := userConnection.WriteMessage(websocket.TextMessage, []byte("Some one left: "+user.Name)); err != nil {
				deleteUser(userConnection, meeting)
			}
		}
	} else {
		delete(allMeetings, meeting.Room)
	}
}

func removeConnection(conn *websocket.Conn, room string) {
	meeting := getMeeting(room)
	if meeting == nil {
		conn.Close()
	} else {
		deleteUser(conn, meeting)
	}
}

func createMeeting(room string, conn *websocket.Conn, user *User) *Meeting {
	users := make(map[*websocket.Conn]*User)
	users[conn] = user
	var meeting *Meeting
	if user.IsOwner {
		meeting = &Meeting{CreatorName: user.Name, Room: room, Locked: false, Users: users}
	} else {
		meeting = &Meeting{CreatorName: "", Room: room, Locked: false, Users: users}
	}
	allMeetings[room] = meeting
	log.Println("Meeting created: " + room)
	return meeting
}

func init() {
	allMeetings = make(map[string]*Meeting)
}

func main() {
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
