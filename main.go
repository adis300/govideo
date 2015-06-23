package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
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
	Cid     string
	Name    string
	IsOwner bool
}

func (meeting *Meeting) getPeerCids() []string {
	peers := []string{}
	for _, peer := range meeting.Users {
		peers = append(peers, peer.Cid)
	}
	return peers
}

func (meeting *Meeting) getConn(cid string) *websocket.Conn {
	for conn, user := range meeting.Users {
		if user.Cid == cid {
			return conn
		}
	}
	return nil
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
	clientMessage, err := simplejson.NewJson(msg)
	if err != nil {
		log.Println("Json parsing error")
		return
	}
	switch clientMessage.Get("type").MustInt() {
	case RTC:
		handleRtc(room, clientMessage, conn)
	case UPDATE_DISPLAY_NAME:
		updateUserName(room, clientMessage.Get("name").MustString(), conn)
	case LOCK_ROOM:
		lockRoom(room, clientMessage.Get("flag").MustBool())
	default:
		log.Println("Unknown message received")
	}
}
func handleRtc(room string, clientMessage *simplejson.Json, conn *websocket.Conn) {
	rtcData := clientMessage.Get("data")
	rm := rtcData.Get("rm").MustString()
	if rm == room {
		if meeting := getMeeting(room); meeting != nil {
			targetCid := rtcData.Get("cid").MustString()
			sender := meeting.Users[conn]
			if sender != nil && len(targetCid) > 0 {
				if targetConn := meeting.getConn(targetCid); targetConn != nil {
					switch clientMessage.Get("eventName").MustString() {
					case "answer":
						log.Println("Answer received:")
						if senderSdp := rtcData.Get("sdp"); senderSdp != nil {
							if err := targetConn.WriteMessage(websocket.TextMessage, encodeOfferAnswer("answer", sender.Cid, senderSdp)); err != nil {
								deleteUser(targetConn, meeting)
							}
						}
					case "offer":
						log.Println("Offer received:")
						log.Println(rtcData.Get("sdp").MustString())
						if senderSdp := rtcData.Get("sdp"); senderSdp != nil {
							log.Println("Sdp information is available")
							if err := targetConn.WriteMessage(websocket.TextMessage, encodeOfferAnswer("offer", sender.Cid, senderSdp)); err != nil {
								deleteUser(targetConn, meeting)
								log.Println("Offer failed to send")
							}
							log.Println("Offer sent if not failed")
						}
					case "ice_candidate":
						log.Println("Ice candidate received:")
						label := rtcData.Get("label").MustInt()
						candidate := rtcData.Get("candidate").MustString()
						if len(candidate) > 0 {
							if err := targetConn.WriteMessage(websocket.TextMessage, encodeIceCandidate(sender.Cid, strconv.Itoa(label), candidate)); err != nil {
								deleteUser(targetConn, meeting)
							}
						}
					default:
						log.Println("Unknown rtc message")
					}
				}
			}
		}
	} else {
		log.Println("Invalid room from socket. May be a hack!")
	}

}

func updateUserName(room string, newName string, conn *websocket.Conn) {
	log.Println("Broadcast new name to all other connections")
}

func lockRoom(room string, lockFlag bool) {
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
					log.Println("Socket read error(This socket might be dropped):")
					log.Println(readErr)
					return
				}
			}
		}
	} else {
		log.Println("Socket connection error: ")
		log.Println(err)
	}
}

func getMeeting(room string) *Meeting {
	return allMeetings[room]
}

func createUser(name string, isOwner bool) *User {
	cid := uuid.NewRandom().String()
	return &User{Cid: cid, Name: name, IsOwner: isOwner}
}

func addUserToRoom(room string, conn *websocket.Conn, name string, isOwner bool) *Meeting {
	meeting := getMeeting(room)
	user := createUser(name, isOwner)
	if meeting == nil {
		meeting = createMeeting(room, conn, user)
		if er := conn.WriteMessage(websocket.TextMessage, encodePeersMessage(nil, user.Cid)); er != nil {
			conn.Close()
		}
	} else {
		log.Println("Meeting found: " + room)
		if er := conn.WriteMessage(websocket.TextMessage, encodePeersMessage(meeting.getPeerCids(), user.Cid)); er != nil {
			conn.Close()
		}
		for userConnection := range meeting.Users {
			if err := userConnection.WriteMessage(websocket.TextMessage, encodeNewPeerMessage(user.Cid)); err != nil {
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
			if err := userConnection.WriteMessage(websocket.TextMessage, encodeRemovePeerMessage(user.Cid)); err != nil {
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
