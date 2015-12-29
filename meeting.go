package main

import (
	"log"

	"govideo/Godeps/_workspace/src/github.com/gorilla/websocket"
	"govideo/Godeps/_workspace/src/github.com/pborman/uuid"
)

var allMeetings = make(map[string]*Meeting)

func getMeeting(room string) *Meeting {
	return allMeetings[room]
}

func createClient() *Client {
	return &Client{SessionID: uuid.NewRandom().String(), Resources: getDefaultClientResources()}
}

func addClientToRoom(room string, conn *websocket.Conn) *Meeting {
	meeting := getMeeting(room)
	client := createClient()

	// Send connect event with information
	if err := conn.WriteMessage(websocket.TextMessage, getConnectMessage(client.SessionID)); err != nil {
		log.Println("ERROR: sending connect message with turn and stun information")
		conn.Close()
	}

	if meeting == nil {
		meeting = createMeeting(room, conn, client)
		log.Println("EVENT: Meeting not found, created new: " + room)

	} else {
		log.Println("EVENT: Meeting found: " + room)
		meeting.Clients[conn] = client
		log.Println("EVENT: Client added")
	}
	return meeting
}

func removeConnectionFromRoom(conn *websocket.Conn, room string) {
	meeting := getMeeting(room)
	if meeting == nil {
		conn.Close()
	} else {
		meeting.removeClient(conn)
	}
}

func createMeeting(room string, conn *websocket.Conn, client *Client) *Meeting {
	clients := make(map[*websocket.Conn]*Client)
	clients[conn] = client
	var meeting *Meeting
	meeting = &Meeting{Room: room, Locked: false, Clients: clients}
	allMeetings[room] = meeting
	log.Println("EVENT: Meeting created: " + room)
	return meeting
}
