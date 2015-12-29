package main

// Meeting represents information about dogs.
import (
	"govideo/Godeps/_workspace/src/github.com/bitly/go-simplejson"
	"govideo/Godeps/_workspace/src/github.com/gorilla/websocket"
)

// Client is a struct that has connections;
type Client struct {
	SessionID string
	Resources ClientResources
}

// ClientResources is stream resources that sits in client
type ClientResources struct {
	Screen bool `json:"screen"`
	Video  bool `json:"video"`
	Audio  bool `json:"audio"`
}

// Meeting object describes a meeting with its state and members
type Meeting struct {
	Room    string
	Locked  bool
	Clients map[*websocket.Conn]*Client
}

func (meeting *Meeting) getNumberOfClients() int {
	return len(meeting.Clients)
}

func (meeting *Meeting) getConn(sessionid string) *websocket.Conn {
	for conn, client := range meeting.Clients {
		if client.SessionID == sessionid {
			return conn
		}
	}
	return nil
}

func (meeting *Meeting) describeMeeting(thisConn *websocket.Conn) []byte {
	clients := simplejson.New()
	for conn, client := range meeting.Clients {
		if conn != thisConn {
			clients.Set(client.SessionID, client.Resources)
		}
	}
	data := simplejson.New()
	data.Set("roomDescription", clients)
	joinMsg := simplejson.New()
	joinMsg.Set("event", "_join")
	joinMsg.Set("data", data)
	msg, err := joinMsg.MarshalJSON()
	if err != nil {
		return []byte(`{"event":"_join","data":{"err":"Json marshal error"}}`)
	}
	return msg
}

func (meeting *Meeting) removeFeed(thisConn *websocket.Conn, tp string) {
	thisClient := meeting.Clients[thisConn]
	if thisClient != nil {
		if len(tp) == 0 {
			meeting.removeClient(thisConn)
		} else {
			removeMsg := getRemoveFeedMessage(thisClient.SessionID, tp)
			for conn := range meeting.Clients {
				if conn != thisConn {
					if err := conn.WriteMessage(websocket.TextMessage, removeMsg); err != nil {
						meeting.removeClient(conn)
					}
				}
			}
		}
	}
}

func (meeting *Meeting) removeClient(thisConn *websocket.Conn) {
	defer thisConn.Close()
	thisClient := meeting.Clients[thisConn]
	if thisClient != nil {
		delete(meeting.Clients, thisConn)
		if len(meeting.Clients) > 0 {
			// TODO: Notify other users
			for clientConnection := range meeting.Clients {
				if err := clientConnection.WriteMessage(websocket.TextMessage, getRemoveClientMessage(thisClient.SessionID)); err != nil {
					meeting.removeClient(clientConnection)
				}
			}
		} else {
			delete(allMeetings, meeting.Room)
		}
	}
}
