package main

import "fmt"

// This file does not use encoding/json Marshal because the messages are too simple
// Encoding with string formatter for simple json messages (lighter weight & faster)

func getConnectMessage(sessionid string) []byte {
	// Currently no turnservers
	return []byte(fmt.Sprintf(`{"event":"connect","data":{"sessionid":"%s","stunservers":[{"url":"%s"}],"turnservers":[]}}`, sessionid, STUN))
}

func getRemoveFeedMessage(sessionid string, tp string) []byte {

	return []byte(fmt.Sprintf(`{"event":"remove","data":{"id":"%s","type":"%s"}}`, sessionid, tp))
}

func getRemoveClientMessage(sessionid string) []byte {
	return []byte(fmt.Sprintf(`{"event":"remove","data":{"id":"%s"}}`, sessionid))
}
