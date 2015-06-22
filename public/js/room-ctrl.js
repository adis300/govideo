var app = angular.module('app',[]);
app.controller('RoomCtrl', function($scope) {
    // Find the path of the room, establish a connection via websocket.
    $scope.connected = false;
    $scope.isEditingName = false;
    var path = document.location.pathname;
    console.log("Path is :" + path);
    var room = path.substring(1);
    $scope.displayName = $scope.path.substring(1);

    SOCKET_STATES = {
        CONNECTING : 0,
        OPEN: 1,
        CLOSING: 2,
        CLOSED: 3
    };
    MESSAGE_TYPES ={
        TYPE_UPDATE_DISPLAY_NAME: 1,
        TYPE_UPDATE_DISPLAY_NAME_RESP: 2,
        LOCK_ROOM: 3,
        LOCK_ROOM_RESP: 4
    };

    var socket = new WebSocket("ws://"+document.location.host + "/ws" + $scope.path);
    socket.onopen = function(e){
        socket.send(JSON.stringify(
                {
                    "message": "hahaha",
                    "handle": "a31"
                }
            ));
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
                case 2:
                    $scope.$apply(function(){ $scope.displayName = msg.content; });
                    break;
                default:
                    alert("Unknown message from server received");
            }
        }

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

    $scope.updateName = function(newName){
        var rawMsg = JSON.stringify({
            type: MESSAGE_TYPES.TYPE_UPDATE_DISPLAY_NAME,
            content:newName
        });
        sendMessage(rawMsg);
    };

    $scope.lockRoom = function(lockFlag){
        var rawMsg = JSON.stringify({
            type: MESSAGE_TYPES.LOCK_ROOM,
            content: lockFlag? "lc": "ulc"
        });
        sendMessage(rawMsg);
    };

});
