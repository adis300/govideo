package main

import "fmt"

/*
import "encoding/json"

type newPeerJson struct {
	EventName string `json:"eventName"`
	Cid       string `json:"cid"`
}

func encodeNewPeerMessage(cid string) []byte {
	s := newPeerJson{EventName: "new_peer", Cid: cid}
    j, _ := json.Marshal(s)z
	return
}

*/

func encodeNewPeerMessage(cid string) []byte {
	return []byte(fmt.Sprintf(`{"type":0,"eventName":"new_peer","cid":"%s"}`, cid))
}

func encodePeersMessage(peerCids []string, mycid string) []byte {
	peers := ""
	for _, peerCid := range peerCids {
		peers += `"` + peerCid + `",`
	}
	if len(peers) > 0 {
		peers = peers[1 : len(peers)-1]
	}
	return []byte(fmt.Sprintf(`{"type":0,"eventName":"peers","mycid":"%s","peers":[%s]}`, mycid, peers))
}
