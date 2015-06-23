// This file does not use encoding/json Marshal because we believe string operation is faster;
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
	return []byte(fmt.Sprintf(`{"type":0,"eventName":"new_peer","data":{"cid":"%s"}}`, cid))
}

func encodePeersMessage(peerCids []string, mycid string) []byte {
	if peerCids != nil {
		peers := ""
		for _, peerCid := range peerCids {
			peers += `"` + peerCid + `",`
		}
		if len(peers) > 0 {
			peers = peers[0 : len(peers)-1]
		}
		return []byte(fmt.Sprintf(`{"type":0,"eventName":"peers","data":{"mycid":"%s","peers":[%s]}}`, mycid, peers))
	}

	return []byte(fmt.Sprintf(`{"type":0,"eventName":"peers","data":{"mycid":"%s","peers":[]}}`, mycid))
}

func encodeRemovePeerMessage(cid string) []byte {
	return []byte(fmt.Sprintf(`{"type":0,"eventName":"remove_peer","data":{"cid":"%s"}}`, cid))
}

func encodeAnswer(senderCid string, senderSdp string) []byte {
	return []byte(fmt.Sprintf(`{"type":0,"eventName":"answer","data":{"cid":"%s","sdp":"%s"}}`, senderCid, senderSdp))
}

func encodeOffer(senderCid string, senderSdp string) []byte {
	return []byte(fmt.Sprintf(`{"type":0,"eventName":"offer","data":{"cid":"%s","sdp":"%s"}}`, senderCid, senderSdp))
}
func encodeIceCandidate(senderCid string, label string, candidate string) []byte {
	return []byte(fmt.Sprintf(`{"type":0,"eventName":"offer","data":{"cid":"%s","label":"%s","candidate":"%s"}}`, senderCid, label, candidate))
}
