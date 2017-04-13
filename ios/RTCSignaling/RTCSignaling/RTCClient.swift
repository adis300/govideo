//
//  RTCClient.swift
//  WebRTC-iOS-Client
//
//  Created by Disi A on 3/30/17.
//  Copyright © 2017 Innovation. All rights reserved.
//

import UIKit
import WebRTC


@objc public class RTCClient: NSObject {
    
    public static let shared = RTCClient()
    
    fileprivate(set) var audioEnabled = true
    
    fileprivate(set) var videoEnabled = true
    
    public var localVideoTrack: RTCVideoTrack?
    
    public var localAudioTrack: RTCAudioTrack?

    public var localMediaStream: RTCMediaStream?
    
    fileprivate var socket: WebSocket?
    
    fileprivate(set) var iceServers : [RTCIceServer] = [] //[String] = []
    
    fileprivate(set) var roomId: String?
    fileprivate(set) var sessionId: String?
    
    fileprivate(set) var sessionReady = false
    
    fileprivate(set) var peers:[String: RTCPeer] = [:]
    
    public var delegate: RTCClientDelegate?
    
    fileprivate var localVideoRenderer: RTCEAGLVideoView?
    
    fileprivate func filterPeers(peerId: String?, type: String?) -> [RTCPeer]{
        return peers.filter{(peerId, peer) in
            return (peer.peerId == peerId) && (type == nil || peer.type == type)
            }.map{(key, val) in val}
    }
    
    private override init() { //No op
    }
    
    // Utility properties
    // fileprivate var peerConnections:[RTCPeerConnection] = []
    
    public func connect(serverUrl:String, roomId: String, delegate: RTCClientDelegate){
        
        reset()

        self.delegate = delegate
        
        self.localMediaStream = RTCFactory.getPeerConnectionFactory().mediaStream(withStreamId: RTCClientConfig.localMediaStreamId)
        
        // Initialize audio track
        localAudioTrack = RTCFactory.getPeerConnectionFactory().audioTrack(withTrackId: RTCClientConfig.localAudioTrackId)
        localMediaStream?.addAudioTrack(localAudioTrack!)
        
        // let audioDevice = AVCaptureDevice.defaultDevice(withDeviceType: .builtInMicrophone, mediaType: AVMediaTypeAudio, position: .unspecified)
                
        if RTCClientConfig.audioOnly{
            delegate.rtcClientDidSetLocalMediaStream(client: self, authorized: true, audioOnly: RTCClientConfig.audioOnly)
        }else{
            
            let authStatus = AVCaptureDevice.authorizationStatus(forMediaType: AVMediaTypeVideo)
            if authStatus == .restricted || authStatus == .denied {
                print("RTCClient:Initialize:Camera authorization denied");
                delegate.rtcClientDidSetLocalMediaStream(client: self, authorized: false, audioOnly: RTCClientConfig.audioOnly)
            }else{
                
                if let _ = AVCaptureDevice.defaultDevice(withDeviceType: .builtInWideAngleCamera, mediaType: AVMediaTypeVideo, position: .back){
                    let videoSource = RTCFactory.getPeerConnectionFactory().avFoundationVideoSource(with: RTCFactory.getMediaConstraints(receiveMedia: nil))
                    localVideoTrack = RTCFactory.getPeerConnectionFactory().videoTrack(with: videoSource, trackId: RTCClientConfig.localVideoTrackId)
                    localMediaStream?.addVideoTrack(localVideoTrack!)
                    
                    delegate.rtcClientDidSetLocalMediaStream(client: self, authorized: true, audioOnly: RTCClientConfig.audioOnly)
                    self.socketConnect(serverUrl: serverUrl, roomId: roomId)
                }else{
                    print("RTCClient:Initialize:Camera unavailable on this device");
                    delegate.rtcClientDidSetLocalMediaStream(client: self, authorized: false, audioOnly: RTCClientConfig.audioOnly)
                }

            }
        }
        
    }
    
    private func socketConnect(serverUrl: String, roomId: String){
        
        guard delegate != nil else {
            assertionFailure("RTCClient:Not initialized, please call rtcClient.initialize(delegate)")
            return
        }
        guard  localMediaStream != nil && (localAudioTrack != nil || localVideoTrack != nil) else {
            assertionFailure("RTCClient:Local media not ready")
            return
        }
                
        self.roomId = roomId
        socket = WebSocket(url: URL(string: serverUrl + roomId)!)
        if !RTCClientConfig.validateSsl{
            socket?.disableSSLCertValidation = true
        }
        socket?.delegate = self
        self.socket?.connect()
        
    }
    
    public func setAudio(on: Bool){
        localAudioTrack?.isEnabled = on
    }
    
    public func setVideo(on: Bool){
        localVideoTrack?.isEnabled = on && !RTCClientConfig.audioOnly
    }
    
    public func setAudioOutput(useSpeaker: Bool){
        do{
            if useSpeaker {
                try AVAudioSession.sharedInstance().overrideOutputAudioPort(.speaker)
            }else{
                try AVAudioSession.sharedInstance().overrideOutputAudioPort(.none)
            }
            
        }catch{
            print(error.localizedDescription)
        }
        
    }
    
    public func setLocalVideoContainer(view: UIView){
        if !RTCClientConfig.audioOnly{
            let videoRenderer = RTCEAGLVideoView(frame: CGRect(x: 0, y: 0, width: view.frame.width, height: view.frame.height))
            view.addSubview(videoRenderer)
            localVideoTrack?.add(videoRenderer)
        }
    }
    
    fileprivate func handleServerMessage(msg: JSON){
        
        print(msg)

        if let event = msg["event"].string{
            switch event {
            case "connect":
                let data = msg["data"]
                if let turnServers = data["turnservers"].array{
                    for turnServer in turnServers{
                        iceServers.append(RTCIceServer(urlStrings: [turnServer["url"].string!]))
                    }
                }
                
                if let stunServers = data["stunservers"].array{
                    for stunServer in stunServers{
                        iceServers.append(RTCIceServer(urlStrings: [stunServer["url"].string!]))
                    }
                }
                
                if let sid = data["sessionid"].string{
                    sessionId = sid
                    sessionReady = true
                    joinRoom()
                }else{
                    assertionFailure("Invalid sessionId received")
                }
                
                // emit sessionReady event
            case "_join": //On join implementation
                let data = msg["data"]
                if let err = data["err"].string{
                    assertionFailure("Server room description error: \(err)")
                    return
                }
                let roomDescription = data["roomDescription"]
                onJoin(roomDescription: roomDescription)
            
            case "message":
                let data = msg["data"]
                onMessage(data)
            case "remove":
                let peerId = msg["data"]["id"].string!
                removePeer(peerId: peerId)
            default:
                assertionFailure("ERROR: Unknown server envent")
            }
        }else{
            assertionFailure("ERROR: Event marker missing")
        }
    }
    
    private func joinRoom(){
        guard roomId != nil else {
            assertionFailure("roomId unavailable when join")
            return
        }
        
        sendMessage(event: "join", data: roomId!)
    }
    
    public func disconect(){
        reset()
        socket?.disconnect()
    }
    
    private func onJoin(roomDescription: JSON) {
        print(roomDescription.count)
        for (id, clientResource) in roomDescription {
            for (type, typeEnabled) in clientResource{
                if typeEnabled.boolValue {
                    let receiveMedia = [
                        "mandatory": [
                            "OfferToReceiveAudio": type != "screen" && RTCClientConfig.offerToReceiveAudio,
                            "OfferToReceiveVideo": RTCClientConfig.offerToReceiveVideo && type != "audio"
                        ]
                    ]
                    let peer = RTCPeer(id: id, options: ["type": type, "enableDataChannels": RTCClientConfig.DEFAULT_ENABLE_DATA_CHANNELS && type != "screen", "receiveMedia": receiveMedia], parent:self)

                    peer.start()
                    peer.sendOffer()
                    peers[id] = peer
                }
            }
        }
    }
    
    private func onMessage(_ data: JSON){
        
        if let type = data["type"].string{
            
            let sessionPeers = filterPeers(peerId: data["from"].string, type: data["roomType"].string)
            var peer: RTCPeer?

            if type == "offer"{
                

                for p in sessionPeers {
                    if p.sid  == data["sid"].string!{
                        peer = p
                        break
                    }
                }
                
                if peer == nil{
                    let isScreen = data["roomType"].string! == "screen"
                    
                    var enableDataChannels = false, sharemyscreen = false, broadcaster: String? = nil
                    
                    if !isScreen && RTCClientConfig.DEFAULT_ENABLE_DATA_CHANNELS{
                        enableDataChannels = true
                    }
                    if isScreen {
                        if data["broadcaster"] == JSON.null {
                            sharemyscreen = true
                            broadcaster = sessionId!
                        }
                    }
                    
                    peer = RTCPeer(id: data["from"].string!, options: [
                        "sid": data["sid"].string,
                        "type": data["roomType"].string,
                        "enableDataChannels": enableDataChannels,
                        "sharemyscreen": sharemyscreen,
                        "broadcaster": broadcaster
                        ], parent: self)

                    peer?.start()
                    peers["id"] = peer!
                }
                peer?.handleMessage(data)
                
            }else{
                for p in sessionPeers {
                    let sid = data["sid"].string
                    if sid == nil || sid == p.sid{
                        p.handleMessage(data);
                    }
                }
            }
        }else{
            assertionFailure("type unavailable, bad offer")
        }
    
    }
}

extension RTCClient {
    
    fileprivate func sendMessage(event: String, data: Any){
        
        let msg = ["event": event, "data": data]
        sendMessage(msg)
    }
    
    private func sendMessage(_ msg: [String: Any]){
        
        if let sock = socket{
            do {
                let jsonData = try JSONSerialization.data(withJSONObject: msg) //, options: .prettyPrinted
                sock.write(data: jsonData)
            } catch {
                print("RTCClient:sendMessage: error serializing msg")
                print(error.localizedDescription)
            }
        }else{
            assertionFailure("Trying to send socket message before session is initialized")
        }
    }
    
    fileprivate func removePeer(peerId: String){
        if let peer = peers[peerId]{
            if let stream = localMediaStream {
                peer.peerConnection.remove(stream)
            }
            // Remove renderers
            if let remoteVideoRenderer = peer.remoteVideoRenderer{
                remoteVideoRenderer.removeFromSuperview()
                peer.remoteVideoTrack?.remove(remoteVideoRenderer)
            }
            peer.peerConnection.close()
            peers.removeValue(forKey: peerId)
        }
    }
    
    fileprivate func reset(){
        
        setAudioOutput(useSpeaker: RTCClientConfig.DEFAULT_USE_SPEAKER)

        for (_, peer) in peers{
            if let stream = localMediaStream {
                peer.peerConnection.remove(stream)
                peer.peerConnection.close()
            }
        }
        if let localRenderer = self.localVideoRenderer {
            localRenderer.removeFromSuperview()
            localVideoTrack?.remove(localRenderer)
            localVideoRenderer = nil
        }
        peers = [:]
        self.iceServers = []
        self.sessionId = nil
        self.sessionReady = false
        self.localVideoTrack = nil
        self.localAudioTrack = nil
        self.localMediaStream = nil
        self.socket = nil
        self.roomId = nil
        self.audioEnabled = RTCClientConfig.offerToReceiveAudio
        self.videoEnabled = RTCClientConfig.offerToReceiveVideo
    }
    
}


extension RTCClient: WebSocketDelegate{
    public func websocketDidConnect(socket: WebSocket){
        print("Did connect")
    }
    public func websocketDidDisconnect(socket: WebSocket, error: NSError?){
        print("Did disconnect")
    }
    public func websocketDidReceiveMessage(socket: WebSocket, text: String){
        if let data = text.data(using: .utf8){
            let json = JSON(data: data)
            handleServerMessage(msg: json)
            
        }else{
            assertionFailure("ERROR: JSON decoding failure with utf8")
        }
    }
    
    public func websocketDidReceiveData(socket: WebSocket, data: Data){
        print("Received some data \(data.count)")
    }
    
}

extension RTCPeer: RTCPeerConnectionDelegate{
    
    /** Called when the SignalingState changed. */
    public func peerConnection(_ peerConnection: RTCPeerConnection, didChange stateChanged: RTCSignalingState){
        print("DEBUG:RTCPeer:RTCPeerConnectionDelegate: peer connection state changed to \(stateChanged.rawValue)")
    }
    
    /** Called when media is received on a new stream from remote peer. */
    public func peerConnection(_ peerConnection: RTCPeerConnection, didAdd stream: RTCMediaStream){
        
        print("RTCPeer:RTCPeerConnectionDelegate: peer connection did add stream")
        if !RTCClientConfig.audioOnly{
            remoteVideoTrack = stream.videoTracks.first
        }
        DispatchQueue.main.async {
            self.parent.delegate?.rtcClientDidAddRemoteMediaStream(peer: self, stream: stream, audioOnly: RTCClientConfig.audioOnly)
        }
    }
    
    /** Called when a remote peer closes a stream. */
    public func peerConnection(_ peerConnection: RTCPeerConnection, didRemove stream: RTCMediaStream){
        print("DEBUG:RTCPeer:RTCPeerConnectionDelegate: peer connection did remove stream")
    }
    
    /** Called when negotiation is needed, for example ICE has restarted. */
    public func peerConnectionShouldNegotiate(_ peerConnection: RTCPeerConnection){
        print("DEBUG:RTCPeer:RTCPeerConnectionDelegate: peer connection should negotiate, no op")
    }
    
    /** Called any time the IceConnectionState changes. */
    public func peerConnection(_ peerConnection: RTCPeerConnection, didChange newState: RTCIceConnectionState){
        if newState == .failed {
            sendPeerMessage(messageType: "connectivityError", payload: nil)
        }
        print("RTCPeer:RTCPeerConnectionDelegate: peer connection ice state did change")
    }
    
    
    /** Called any time the IceGatheringState changes. */
    public func peerConnection(_ peerConnection: RTCPeerConnection, didChange newState: RTCIceGatheringState){
        print("DEBUG:RTCPeer:RTCPeerConnectionDelegate: peer connection ice gathering state did change")
    }
    
    /** New ice candidate has been found. */
    public func peerConnection(_ peerConnection: RTCPeerConnection, didGenerate candidate: RTCIceCandidate){
        print(candidate.sdp)
        let candidate:[String: Any] = ["candidate":candidate.sdp,"sdpMid":candidate.sdpMid!,"sdpMLineIndex":candidate.sdpMLineIndex]
        sendPeerMessage(messageType: "candidate", payload: ["candidate":candidate])
        print("RTCPeer:RTCPeerConnectionDelegate: peer connection did generate new ICE candidate")
    }
    
    /** Called when a group of local Ice candidates have been removed. */
    public func peerConnection(_ peerConnection: RTCPeerConnection, didRemove candidates: [RTCIceCandidate]){
        print("DEBUG:RTCPeer:RTCPeerConnectionDelegate: peer connection did remove candidates")
    }
    
    /** New data channel has been opened. */
    public func peerConnection(_ peerConnection: RTCPeerConnection, didOpen dataChannel: RTCDataChannel){
        print("DEBUG:RTCPeer:RTCPeerConnectionDelegate: peer connection did open data channel")
    }
    
    func sedPeerMessage(messageType: String, payload: String){
        // send via signalling channel
        let data: [String: Any]
        if let broadcaster = self.broadcaster{
            data = [
                "to": self.peerId,
                "sid": self.sid,
                "broadcaster": broadcaster,
                "roomType": self.type,
                "type": messageType,
                "payload": payload,
                "prefix": browserPrefix
            ]
        }else{
            data = [
                "to": self.peerId,
                "sid": self.sid,
                "roomType": self.type,
                "type": messageType,
                "payload": payload,
                "prefix": browserPrefix
            ]
        }
        parent.sendMessage(event: "message", data: data)
    }
}

@objc public class RTCPeer: NSObject {
    
    private(set) public var peerConnection: RTCPeerConnection!
    private(set) public var peerId: String
    private(set) public var type = RTCClientConfig.DEFAULT_PEER_TYPE   //default peer type to video
    private(set) public var oneway = false
    private(set) public var sharemyscreen = false
    private(set) var browserPrefix = "webkit"
    private(set) public var enableDataChannels = RTCClientConfig.DEFAULT_ENABLE_DATA_CHANNELS
    private(set) public var sid: String = String(Date().timeIntervalSince1970)
    private(set) public var broadcaster: String?
    //var stream
    //var channels
    //var receiveMedia
    private(set) var receiveMedia:[String: Any] = RTCClientConfig.defaultReceiveMedia
    public let parent: RTCClient
    fileprivate(set) public var remoteVideoTrack: RTCVideoTrack?
    
    fileprivate var remoteVideoRenderer: RTCEAGLVideoView?
    
    private func getReceiveMedia () -> [String: Any]{
        
        var receiveMedia:[String: Any] = [:]
        var mandatory:[String: Bool] = [:]
        
        if type == "screen"{
            mandatory["OfferToReceiveAudio"] = false
        }else{
            mandatory["OfferToReceiveAudio"] = RTCClientConfig.offerToReceiveAudio
        }
        
        mandatory["OfferToReceiveVideo"] = RTCClientConfig.offerToReceiveVideo
        
        receiveMedia["mandatory"] = mandatory
        
        return receiveMedia
    }
    

    public init(id: String, options: [String: Any?], parent: RTCClient) {
        
        self.parent = parent
        peerId = id
        
        if let val  = options["type"] as? String{
            type = val
        }
        if let val  = options["sid"] as? String{
            sid = val
        }
        if let val  = options["browserPrefix"] as? String{
            browserPrefix = val
        }
        if let val  = options["enableDataChannels"] as? Bool{
            enableDataChannels = val
        }
        if let val  = options["sharemyscreen"] as? Bool{
            sharemyscreen = val
        }
        if let val  = options["oneway"] as? Bool{
            oneway = val
        }
        
        if let val = options["receiveMedia"] as? [String: Any]{
            receiveMedia = val
        }else{
            receiveMedia = RTCClientConfig.defaultReceiveMedia
        }
    }
    
    public func setRemoteVideoContainer(view: UIView){
        if !RTCClientConfig.audioOnly{
            remoteVideoRenderer = RTCEAGLVideoView(frame: CGRect(x: 0, y: 0, width: view.frame.width, height: view.frame.height))
            remoteVideoTrack?.add(remoteVideoRenderer!)
            DispatchQueue.main.async {
                view.addSubview(self.remoteVideoRenderer!)
            }
        }
    }
    
    fileprivate func start() {
        let mediaConstraints = RTCFactory.getMediaConstraints(receiveMedia: receiveMedia)
        peerConnection = RTCFactory.getPeerConnectionFactory().peerConnection(with: RTCClientConfig.getRTCConfiguration(iceServers: parent.iceServers), constraints: mediaConstraints, delegate: self)
        
        // handle screensharing/broadcast mode
        if type == "screen" {
            if sharemyscreen { //&& has a localScreen
                print("adding local screen stream to peer connection");
                // peerConnection.add(<#T##stream: RTCMediaStream##RTCMediaStream#>) // Add a local screen
                // broadcaster = opt["broadcaster"] as! String
                assertionFailure("TODO: Screensharing is not yet supported")
            }
        } else {
            if let localStream = parent.localMediaStream{
                peerConnection.add(localStream)
            }else{
                assertionFailure("Failed to add local stream, not ready.")
            }
        }
        
        if (self.enableDataChannels) {
            let dataChannel = peerConnection.dataChannel(forLabel: "simplewebrtc", configuration: RTCClientConfig.dataChannelConfiguration)
            // TODO: Make use of dataChannel
        }
    }
    
    fileprivate func sendOffer(){

        peerConnection.offer(for: RTCFactory.getMediaConstraints(receiveMedia: receiveMedia)) { (sessionDescription, error) in
            if let err = error{
                print(err.localizedDescription)
                return
            }
            if let sessionDesc = sessionDescription{
                self.peerConnection.setLocalDescription(sessionDesc, completionHandler: { (localSdpError) in
                    if let e = localSdpError{
                        assertionFailure(e.localizedDescription)
                    }
                })
                self.sendPeerMessage(messageType: "offer", payload: ["type":"offer", "sdp": sessionDesc.sdp])
            }else{
                assertionFailure("Failed to get session description")
            }
        }
    }
    fileprivate func sendAnswer(){
        
        peerConnection.answer(for: RTCFactory.getMediaConstraints(receiveMedia: receiveMedia)) { (sessionDescription, error) in
            if let err = error{
                assertionFailure(err.localizedDescription)
                return
            }
            if let sessionDesc = sessionDescription{
                self.peerConnection.setLocalDescription(sessionDesc, completionHandler: { (localSdpError) in
                    if let e = localSdpError{
                        assertionFailure(e.localizedDescription)
                    }
                })
                self.sendPeerMessage(messageType: "answer", payload: ["type":"answer", "sdp": sessionDesc.sdp])
            }else{
                assertionFailure("Failed to get session description")
            }
        }
    }
    
    // MARK: Peer connection active methods
    
    fileprivate func sendPeerMessage(messageType: String, payload:[String : Any]?){
        var message: [String: Any]
        
        if let broadcaster = self.broadcaster{
            message = ["type": messageType, "to":self.peerId, "sid":self.sid, "broadcaster": broadcaster, "roomType": self.type, "prefix": self.browserPrefix]
        }else{
            message = ["type": messageType, "to":self.peerId, "sid":self.sid, "roomType": self.type, "prefix": self.browserPrefix]
        }
        
        if let msgPayload = payload{
            message["payload"] = msgPayload
        }
        
        parent.sendMessage(event: "message", data: message)
    }
    
    fileprivate func handleMessage(_ message: JSON){

        print("RTCPeer:handleMessage:\(message)")
        let payload = message["payload"]

        if let prefix = message["prefix"].string{
            self.browserPrefix = prefix
        }
        
        let type = message["type"].string!  //Type must by available, protected from top level onMessage
        
        switch type {
        case "offer":
            let remoteSdp = RTCSessionDescription(type: .offer, sdp: payload["sdp"].string!)
            peerConnection.setRemoteDescription(remoteSdp, completionHandler: { (error) in
                if let err = error{
                    assertionFailure("RTCPeer:onAnswer: failed to set remote description: \(err.localizedDescription)")
                }
            })
            sendAnswer()
        case "answer":
            let remoteSdp = RTCSessionDescription(type: .answer, sdp: payload["sdp"].string!)
            peerConnection.setRemoteDescription(remoteSdp, completionHandler: { (error) in
                if let err = error{
                    print("RTCPeer:onAnswer: failed to set remote description: \(err.localizedDescription)")
                }
            })
        case "candidate":
            
            if let _ = parent.peers[message["from"].string!]{
                let candidateJson = payload["candidate"]
                let candidateSdp = candidateJson["candidate"].string!
                let candidate = RTCIceCandidate(sdp: candidateSdp, sdpMLineIndex: candidateJson["sdpMLineIndex"].int32!, sdpMid: candidateJson["sdpMid"].string!)
                
                peerConnection.add(candidate)
            }else{
                assertionFailure("Peer not found for answer")
            }
        case "endOfCandidates":
            print("End of candidate received.")
            /* TODO: Find out about mLine and transceivers.
            var mLines = this.pc.pc.peerconnection.transceivers || [];
            mLines.forEach(function(mLine) {
                if (mLine.iceTransport) mLine.iceTransport.addRemoteCandidate({});
            });
            */
        case "mute":
            if payload["name"].string == "audio"{
                parent.delegate?.rtcRemotePeerDidChangeAudioState(peer: self, on: false)
            }else if payload["name"].string == "video"{
                parent.delegate?.rtcRemotePeerDidChangeVideoState(peer: self, on: false)
            }else{
                assertionFailure("ERROR:Unknown media type to mute")
            }
        case "unmute":
            if payload["name"].string == "audio"{
                parent.delegate?.rtcRemotePeerDidChangeAudioState(peer: self, on: true)
            }else if payload["name"].string == "video"{
                parent.delegate?.rtcRemotePeerDidChangeVideoState(peer: self, on: true)
            }else{
                assertionFailure("ERROR:Unknown media type to unmute")
            }
        case "connectivityError":
            parent.delegate?.rtcRemotePeerFailedToGenerateIceCandidate(peer: self)

        default:
            assertionFailure("RTCPeer:handleMessage: unknown message type")
        }
    }
}

