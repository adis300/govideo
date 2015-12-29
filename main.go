package main

import (
	"log"
	"net/http"

	"govideo/Godeps/_workspace/src/github.com/bitly/go-simplejson"
	"govideo/Godeps/_workspace/src/github.com/gorilla/mux"
	"govideo/Godeps/_workspace/src/github.com/gorilla/websocket"
)

func getDefaultClientResources() ClientResources {
	return ClientResources{Screen: false, Video: true, Audio: false}
}

var roomView = LoadView("room")
var homeView = LoadView("home")
var roomSecureView = LoadView("room-secure")

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// w.Write(LoadView("home")) //Easier for debug
	w.Write(homeView)
}

func roomHandler(w http.ResponseWriter, r *http.Request) {
	roomPath := mux.Vars(r)["room"]
	log.Println(roomPath)
	if len(roomPath) < 3 { // roomPath is not available or not valid
		w.Write([]byte("Invalid room url"))
	} else {
		// w.Write(LoadView("room")) //Easier for debug
		w.Write(roomView)
	}
}

func roomSecureHandler(w http.ResponseWriter, r *http.Request) {
	roomPath := mux.Vars(r)["room"]
	log.Println(roomPath)
	if len(roomPath) < 3 { // roomPath is not available or not valid
		w.Write([]byte("Invalid room url"))
	} else {
		// w.Write(LoadView("room-secure")) //Easier for debug
		w.Write(roomSecureView)
	}
}

func roomMessageHandler(meeting *Meeting, rawMsg []byte, thisConn *websocket.Conn) {
	log.Println("Event: Message received from client:" + meeting.Room)
	log.Println(string(rawMsg))
	clientMessage, err := simplejson.NewJson(rawMsg)
	if err != nil {
		log.Println("ERROR: Json parsing error: " + meeting.Room)
		return
	}
	if thisClient := meeting.Clients[thisConn]; thisClient != nil {
		switch clientMessage.Get("event").MustString() {
		case "join":
			if rm := clientMessage.Get("data").MustString(); rm == meeting.Room {
				// Send a join message with room description
				if err := thisConn.WriteMessage(websocket.TextMessage, meeting.describeMeeting(thisConn)); err != nil {
					log.Println("ERROR: sending join message with room description")
					meeting.removeClient(thisConn)
				}
			} else {
				log.Println("WARNING: Might be a hack! room to join is different!")
			}
		case "message":
			if details := clientMessage.Get("data"); details != nil {
				if to := details.Get("to").MustString(); len(to) > 0 {
					if otherClientConn := meeting.getConn(to); otherClientConn != nil {
						details.Set("from", thisClient.SessionID)
						clientMessage.Set("data", details)
						newMsg, err := clientMessage.MarshalJSON()
						if err != nil {
							log.Println("Forwarding Message: Marshal json error!")
							return
						}
						if err := otherClientConn.WriteMessage(websocket.TextMessage, newMsg); err != nil {
							log.Println("Forwarding Message socket error!")
							return
						}
					} else {
						log.Println("ERROR: No target connection found!")
					}
				} else {
					log.Println("ERROR: No ~to~ attribute specified in data!")
				}
			} else {
				log.Println("ERROR: No data field in raw message")
			}
		case "shareScreen":
			thisClient.Resources.Screen = true
		case "unshareScreen":
			thisClient.Resources.Screen = false
			meeting.removeFeed(thisConn, "screen")
		case "leave":
			meeting.removeClient(thisConn)
		case "disconnect":
			meeting.removeClient(thisConn)
		case "trace": // Log all the bugs
			log.Println("Trace:")
			log.Println(clientMessage.Get("data").MarshalJSON())
		// case "create":
		// case "join"
		default:
			log.Println("ERROR: Unknown message received")
		}
	}

}

func lockRoom(meeting *Meeting, lockFlag bool) {
	//lockRoom(meeting, clientMessage.Get("flag").MustBool())
	if meeting.Locked != lockFlag {
		meeting.Locked = lockFlag
		// TODO broadcast a lock state change message;
		log.Println("Meeting lock state changed for: " + meeting.Room)
	}
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	roomPath := mux.Vars(r)["room"]
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	defer removeConnectionFromRoom(conn, roomPath) // Gauranteed removal of a connection
	if err == nil {
		if len(roomPath) > 2 {
			meeting := addClientToRoom(roomPath, conn)
			for {
				_, msg, readErr := conn.ReadMessage()
				if readErr == nil {
					roomMessageHandler(meeting, msg, conn)
				} else {
					log.Println("Socket read error(This socket might be dropped):")
					log.Println(readErr)
					return
				}
			}
		}
	} else {
		log.Println("ERROR: Socket connection error: ")
		log.Println(err)
	}
}

/*
func init() {
	allMeetings = make(map[string]*Meeting)
}*/

func main() {
	// handle all requests by serving a file of the same name
	fileServer := http.FileServer(http.Dir("public"))

	router := mux.NewRouter()
	router.HandleFunc("/", homeHandler)

	if SERVE_SECURE {
		router.HandleFunc("/{room}", roomSecureHandler)
	} else {
		router.HandleFunc("/{room}", roomHandler)
	}

	router.PathPrefix("/public").Handler(http.StripPrefix("/public", fileServer))
	router.HandleFunc("/ws/{room}", socketHandler)

	http.Handle("/", router)

	log.Println("APP: Serving on" + PORT)
	log.Println("APP: Go to localhost" + PORT)

	if SERVE_SECURE {
		go func() {
			log.Println("HTTP redirecting on port:" + PORT)
			httpErr := http.ListenAndServe(PORT, secureRedirectHandler(http.StatusFound))
			if httpErr != nil {
				panic("ERROR: " + httpErr.Error())
			}
		}()
		log.Println("APP: Security is on using HTTPS")
		if err := http.ListenAndServeTLS(PORT_SECURE, "ssl/cert.crt", "ssl/server.key", nil); err != nil {
			log.Fatal("ERROR: ListenAndServeTLS:", err)
		}
	} else {
		if err := http.ListenAndServe(PORT, nil); err != nil {
			log.Fatal("ERROR: ListenAndServe:", err)
		}
	}

}
