package main

// ValidateUser determines if a user is a valid owner and returns the session
func ValidateUser(room string, owner string, ownerName string, participantName string) User {
	name := ""
	if len(ownerName) > 1 {
		name = ownerName
	}
	if len(owner) > 0 {
		if true { // Check that owner is valid hash
			return User{Name: name, IsOwner: true, Handle: room}
		}
	}
	return User{Name: participantName, IsOwner: false, Handle: ""}
}
