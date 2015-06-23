var app = angular.module('app',[]);
app.controller('RoomCtrl', function($scope) {
    // Find the path of the room, establish a connection via websocket.
    // adapter.js could be included
    var PeerConnection = (window.PeerConnection || window.webkitPeerConnection00 || window.webkitRTCPeerConnection || window.mozRTCPeerConnection);
    var URL = (window.URL || window.webkitURL || window.msURL || window.oURL);
    var getUserMedia = (navigator.getUserMedia || navigator.webkitGetUserMedia || navigator.mozGetUserMedia || navigator.msGetUserMedia);
    var nativeRTCIceCandidate = (window.mozRTCIceCandidate || window.RTCIceCandidate);
    var nativeRTCSessionDescription = (window.mozRTCSessionDescription || window.RTCSessionDescription); // order is very important: "RTCSessionDescription" defined in Nighly but useless
    var moz = !!navigator.mozGetUserMedia;
    var iceServer = {
        "iceServers": [{
            "url": "stun:stun.l.google.com:19302"
        }]
    };
    var packetSize = 1024;

    var path = document.location.pathname;
    var room = path.substring(1);

    $scope.localMediaStream = null;
    //$scope.fileData = {};
    $scope.peerConns = {}; //peerCid as key
    $scope.peerCids = [];
    $scope.numStreams = 0;
    $scope.initializedStreams = 0;
    $scope.peerVideos = {};
    // All datachannels, peerCid as key, created by PeerConnection.createChannel
    // $scope.dataChannels = {};
    // Similar to dataChannels, will be used later.
    //$scope.fileChannels = {};
    //$scope.receiveFiles = {};

    // App state variables
    $scope.connected = false;
    $scope.displayName = "";//$scope.path.substring(1);
    $scope.isEditingName = true;
    $scope.mycid = "";
    var socket = null;
    $scope.locked = false;

    var pc;
    SOCKET_STATES = {
        CONNECTING : 0,
        OPEN: 1,
        CLOSING: 2,
        CLOSED: 3
    };
    MESSAGE_TYPES ={
        RTC: 0,
        UPDATE_DISPLAY_NAME: 1,
        UPDATE_DISPLAY_NAME_RESP: 2,
        LOCK_ROOM: 3,
        LOCK_ROOM_RESP: 4
    };

    var handleRtcMessage = function(msg){
        console.log("Handling a rtc message");
        var data = msg.data;
        switch (msg.eventName) {
            case "peers":
                $scope.connected = true;
                $scope.mycid = data.mycid;
                $scope.createStream();
                break;
            case "new_peer":
                $scope.peerCids.push(data.cid);
                pc = $scope.createPeerConn(data.cid);
                pc.addStream($scope.localMediaStream);
                break;
            case "ice_candidate":
                var candidate = new nativeRTCIceCandidate(data);
                $scope.peerConns[data.cid].addIceCandidate(candidate);
                break;
            case "remove_peer":
                $scope.closePeerConnection($scope.peerConns[data.cid]);
                delete $scope.peerConns[data.cid];
                delete $scope.peerVideos[data.cid];
                /*delete $scope.dataChannels[data.cid];
                for (sendId in that.fileChannels[data.socketId]) {
                    that.emit("send_file_error", new Error("Connection has been closed"), data.socketId, sendId, that.fileChannels[data.socketId][sendId].file);
                }
                delete that.fileChannels[data.socketId];*/
                break;
            case "offer":
                pc = $scope.peerConns[data.cid];
                pc.setRemoteDescription(new nativeRTCSessionDescription(data.sdp));
                pc.createAnswer(function(sessionDesc) {
                    pc.setLocalDescription(sessionDesc);
                    $scope.sendMessage(
                        JSON.stringify({
                            "eventName": "answer",
                            "data": {
                                "cid": data.cid,
                                "sdp": sessionDesc
                            }
                        }));
                        }, function(error) {
                            console.log(error);
                    });
                break;
            case "answer":
                $scope.peerConns[data.cid].setRemoteDescription(
                    new nativeRTCSessionDescription(data.sdp));
                break;
            default:

        }
        console.log($scope.connected);
    };
    $scope.connect = function(){
        socket = new WebSocket("ws://"+document.location.host + "/ws" + path);
        socket.onopen = function(e){
            console.log("Waiting for server session info");
            //$scope.$apply(function(){$scope.connected = true; });
            //TODO: that.emit("socket_opened", socket);
        };
        socket.onclose = function(e){
            $scope.$apply(function(){
                if($scope.localMediaStream)$scope.localMediaStream.close();
                var pcs = $scope.peerConns;
                for (i = pcs.length; i--;) {
                    $scope.closePeerConnection(pcs[i]);
                }
                $scope.peerConns = [];
                $scope.dataChannels = {};
                //$scope.fileChannels = {};
                $scope.peerCids = [];
                //$scope.fileData = {};
                $scope.connected = false;
                });
            //TODO: that.emit('socket_closed', socket);
        };
        socket.onmessage = function(e){
            console.log("Raw message received:");
            var msg = JSON.parse(e.data);
            console.log(msg);
            if(msg){
                switch (msg.type) {
                    case MESSAGE_TYPES.RTC:
                        handleRtcMessage(msg);
                        break;
                    case MESSAGE_TYPES.UPDATE_DISPLAY_NAME_RESP:
                        $scope.$apply(function(){ $scope.displayName = msg.name; });
                        break;
                    case MESSAGE_TYPES.LOCK_ROOM_RESP:
                        $scope.$apply(function(){ $scope.locked = msg.flag; });
                        break;
                    default:
                        alert("Unknown message from server received");
                }
            }
        };
        socket.onerror = function(err){
            console.log("Socket error:");
            console.log(err);
        };
    };

    //~~~~~~~~~~~~~~~~~~~~~~~ Stream and signaling ~~~~~~~~~~~~~~~~~
    $scope.sendOffers = function() {
        var pcCreateOfferCbGen = function(pc, cid) {
                return function(sessionDesc) {
                    pc.setLocalDescription(sessionDesc);
                    $scope.sendMessage(
                        JSON.stringify({
                            eventName: "offer",
                            data: {
                                sdp: sessionD,
                                cid: cid
                            }
                        })
                    );
                };
            };
        var pcCreateOfferErrorCb = function(error) {
            console.log(error);
        };
        $scope.peerCids.forEach(function(peerCid){
            pc = peerConns[peerCid];
            pc.createOffer(pcCreateOfferCbGen(pc, peerCid), pcCreateOfferErrorCb);
        });
    };
    $scope.createStream = function(){
        if (getUserMedia) {
            var options = {video: true, audio: true};
            $scope.numStreams++;
            getUserMedia.call(navigator, options, function(stream) {
                $scope.localMediaStream = stream;
                $scope.initializedStreams++;
                //TODO: play the created stream
                if ($scope.initializedStreams === $scope.numStreams) {
                    $scope.createPeerConns();
                    $scope.addStreams();
                    //$scope.addDataChannels();
                    $scope.sendOffers();
                }
            },function(error) {
                that.emit("stream_create_error", error);
            });
        } else {
            that.emit("stream_create_error", new Error('WebRTC is not yet supported in this browser.'));
        }
    };

    $scope.connect();

    var sendMessage = function(msg){
        console.log("Raw message to send:");
        console.log(msg);
        if(socket){
            if(socket.readyState == SOCKET_STATES.OPEN){
                socket.send(msg);
            }else alert('Connection is not open');
        }else alert('Connection is closed');
    };

    $scope.updateName = function(){
        $scope.isEditingName = false;
        if($scope.displayName.length > 0){
            var rawMsg = JSON.stringify({
                type: MESSAGE_TYPES.TYPE_UPDATE_DISPLAY_NAME,
                name:$scope.displayName
            });
            sendMessage(rawMsg);
        }
    };

    $scope.lockRoom = function(lockFlag){
        var rawMsg = JSON.stringify({
            type: MESSAGE_TYPES.LOCK_ROOM,
            flag: lockFlag
        });
        sendMessage(rawMsg);
    };
    $scope.shouldShowNameEdit = function(){
        if($scope.isEditingName) return true;
        else{
            if($scope.displayName.length === 0) return true;
            else return false;
        }
    };
    $scope.editNameKeyboardEvent = function(e){
        if(e.keyCode == 13 || e.keyCode == 27){
            $scope.updateName();
            console.log("Enter or escape key pressed");
        }else{
            console.log("Regular editing");
        }
    };

    // ---------- Point to point connection --------------------
    $scope.createPeerConns = function() {
        $scope.peerCids.forEach(function(peerCid){
            $scope.createPeerConn(peerCid);
        });
    };

    $scope.createPeerConn = function(cid) {
        var peerConn = new PeerConnection(iceServer);
        $scope.peerConns[cid] = peerConn;
        peerConn.onicecandidate = function(evt) {
            if (evt.candidate){
                var rawMsg = JSON.stringify({
                    "eventName": "ice_candidate",
                    data:{
                        "label": evt.candidate.sdpMLineIndex,
                        "candidate": evt.candidate.candidate,
                        "cid": cid
                    }
                });
                sendMessage(rawMsg);
                //TODO: that.emit("pc_get_ice_candidate", evt.candidate, cid, peerConn);
            }
        };
        peerConn.onopen = function() {
            that.emit("pc_opened", socketId, pc);
        };
        peerConn.onaddstream = function(evt) {
            $scope.attachPeerStream(evt.stream, cid);
            //that.emit('pc_add_stream', evt.stream, socketId, pc);
        };
        peerConn.ondatachannel = function(evt) {
            $scope.addDataChannel(cid, evt.channel);
            //that.emit('pc_add_data_channel', evt.channel, socketId, pc);
        };
        return peerConn;
    };

    // ~~~~~~~~~~~~~~~~~~~~~Data channel section --------------
    $scope.addDataChannels = function() {
        console.log("Should add a data channel!");
        /* this section is all for file transfer
        $scope.peerConns.forEach(function(peerConn){
            $scope.createDataChannel(peerConn);
        });
        */
    };
    // ~~~~~~~~~~~~~~~~~~~~~View methods --------------

    // Add local stream to all peers
    $scope.addStreams = function() {
        $scope.peerCids.forEach(function(peerCid){
            peerConns[peerCid].addStream($scope.localMediaStream);
        });
    };

    $scope.attachPeerStream = function(stream, cid){
        console.log("Attaching a peer stream");
        $scope.peerVideos[cid] = {
            src: webkitURL.createObjectURL(stream)
        };
    };

    $scope.localStreamCreated = function(stream){

        /*TODO:
        document.getElementById('me').src = URL.createObjectURL(stream);
        document.getElementById('me').play();
        */
    };
    $scope.peerConnAddStream = function(stream, cid){
        /*TODO:
        var newVideo = document.createElement("video"),
        id = "other-" + socketId;
        newVideo.setAttribute("class", "other");
        newVideo.setAttribute("autoplay", "autoplay");
        newVideo.setAttribute("id", id);
        videos.appendChild(newVideo);
        */
    };

});
