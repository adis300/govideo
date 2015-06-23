// This file does not use encoding/json Marshal because we believe string operation is faster;
package main

import (
	"fmt"
	"log"

	"github.com/bitly/go-simplejson"
)

/*
type OfferAnswerData struct {
	Cid string      `json:"cid"`
	Sdp interface{} `json:"sdp"`
}

type Offer struct {
	Type      int             `json:"type"`
	EventName string          `json:"eventName"`
	Offer     OfferAnswerData `json:"offer"`
}

type Answer struct {
	Type      int             `json:"type"`
	EventName string          `json:"eventName"`
	Offer     OfferAnswerData `json:"answer"`
}

func encodeNewPeerMessage(cid string) []byte {
	s := newPeerJson{EventName: "new_peer", Cid: cid}
    j, _ := json.Marshal(s)z
	return
}*/

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

func encodeOfferAnswer(offerOrAnswer string, senderCid string, senderSdp *simplejson.Json) []byte {

	answerData := simplejson.New()
	answerData.Set("cid", senderCid)
	answerData.Set("sdp", senderSdp)
	answerJSON := simplejson.New()
	answerJSON.Set("type", RTC)
	answerJSON.Set("eventName", offerOrAnswer)
	answerJSON.Set("data", answerData)
	json, err := answerJSON.MarshalJSON()
	if err != nil {
		log.Println("Answer json encoding error")
		return []byte("")
	}
	log.Println(string(json))
	return json
}

func encodeIceCandidate(senderCid string, label string, candidate string) []byte {
	return []byte(fmt.Sprintf(`{"type":0,"eventName":"ice_candidate","data":{"cid":"%s","label":%s,"candidate":"%s"}}`, senderCid, label, candidate))
}
