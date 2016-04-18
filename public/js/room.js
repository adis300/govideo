// Usage is same as Simple WebRTC
// https://simplewebrtc.com

var muted = false, paused = false, sharingScreen = false;

function mute() {
    muted = !muted;
    console.log(muted);
    if(muted){
        document.getElementById("muteButton").innerHTML = '<img src="../public/img/unmute.png" alt="Unmute" width="16" height="16" />';
        webrtc.mute();
    }else{
        document.getElementById("muteButton").innerHTML = '<img src="../public/img/mute.png" alt="Mute" width="16" height="16" />';
        webrtc.unmute();
    }
}

function pause(){
    paused = !paused;
    console.log(paused);
    if(paused){
        document.getElementById("pauseButton").innerHTML = '<img src="../public/img/play.png" alt="Play" width="16" height="16" />';
        webrtc.pauseVideo();
    }else{
        document.getElementById("pauseButton").innerHTML = '<img src="../public/img/pause.png" alt="Pause" width="16" height="16" />';
        webrtc.resumeVideo();
    }
}

function shareScreen(){
    sharingScreen = !sharingScreen;
    if(sharingScreen){
        webrtc.shareScreen(function (err) {
            if (err) {
                sharingScreen = false;
                console.log("Error sharing screen");
                console.error(err);
                document.getElementById("shareScreenButton").innerHTML = '<img src="../public/img/present.png" alt="Play" width="16" height="16" />';
            } else {
                console.log("Started sharing screen");
                document.getElementById("shareScreenButton").innerHTML = '<img src="../public/img/stop-present.png" alt="Play" width="16" height="16" />';
            }
        });
    }else{
        console.log("Stoped sharing screen");
        webrtc.stopScreenShare();
        document.getElementById("shareScreenButton").innerHTML = '<img src="../public/img/present.png" alt="Play" width="16" height="16" />';

        // Remove screen share previews
        var screenContainer = document.getElementById("localScreenContainer");
        while(screenContainer.firstChild) {
            screenContainer.removeChild(screenContainer.firstChild);
        }
    }
}

var videoCount = 0;
console.log(document.location);
console.log(document.location.host);
var room = document.location.pathname.substring(1);
var secureLink = document.location.protocol === "https:";

var webrtc = new SimpleWebRTC({
    localVideoEl: 'localVideo',
    remoteVideosEl: 'remotes',
    autoRequestMedia: true,
    autoRemoveVideos: true,
    url: document.location.protocol +'//' + document.location.host,
    wsUrl: secureLink? 'wss://' + document.location.host : 'ws://' + document.location.host
}, room);

// a peer video has been added
webrtc.on('videoAdded', function (video, peer) {
    videoCount += 1;
    console.log('video added', peer);
    var remotes = document.getElementById("remoteVideos");
    if (remotes) {
        var container = document.createElement('div');
        container.className = 'remoteVideoContainer';
        container.id = 'container_' + webrtc.getDomId(peer);
        container.appendChild(video);
        // suppress contextmenu
        video.oncontextmenu = function () { return false; };
        container.onclick = function() {
            // clicking the primary video should remove its active status
            if (container.id === videoIndices.primary)
                videoIndices.primary = undefined;
            else
                videoIndices.primary = container.id;
            updateVideoLocations();
        };
        remotes.appendChild(container);
        updateVideoLocations();
    }
});

webrtc.on('videoRemoved', function (video, peer) {
    console.log('video removed ', peer);
    var remotes = document.getElementById("remoteVideos");
    var el = document.getElementById(peer ? 'container_' + webrtc.getDomId(peer) : 'localVideoContainer');
    if (remotes && el && peer) {
        videoCount -= 1;
        remotes.removeChild(el);
        if (videoIndices.id === el.id)
            videoIndices.id = undefined;
        updateVideoLocations();
    }
});

webrtc.on('localScreenAdded', function (video) {
    var screenContainer = document.getElementById("localScreenContainer");
    screenContainer.appendChild(video);
});

webrtc.on('readyToCall', function () {
    console.log("call is ready");
    webrtc.joinRoom(room);
});
