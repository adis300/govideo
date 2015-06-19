package main

// Meeting represents information about dogs.
type Meeting struct {
	CreatorName string
	Room        string
	Locked      bool
	Users       []User
}

// User is a struct that has connections;
type User struct {
	Name    string
	Handle  string
	IsOwner bool
	//Conn    *websocket.Conn
}
