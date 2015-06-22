var app = angular.module('app',[]);
app.controller('RoomCtrl', function($scope) {
    // Find the path of the room, establish a connection via websocket.
    $scope.connected = false;
    var path = document.location.pathname;
    console.log("Path is :" + path);
    var room = path.substring(1);
    $scope.displayName = "";//$scope.path.substring(1);
    $scope.isEditingName = true;
    $scope.mycid = "";
    var socket;
    $scope.locked = false;
    $scope.connect();

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
        switch (msg.eventName) {
            case "peers":
                $scope.connected = true;
                $scope.mycid = msg.mycid;
                break;
            case "new_peer":
                break;
            default:

        }
    };
    $scope.connect = function(){
        socket = new WebSocket("ws://"+document.location.host + "/ws" + path);
        socket.onopen = function(e){
            $scope.$apply(function(){$scope.connected = true; });
        };
        socket.onclose = function(e){
            $scope.$apply(function(){ $scope.connected = false; });
        };
        socket.onmessage = function(e){
            console.log("Raw message received:");
            var msg = e.data;
            console.log(e.data);
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
});
