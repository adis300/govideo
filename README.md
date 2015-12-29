# Go video demo for using webrtc

## The rtckit.js library is adapted from
[simplewebrtc.com](https://github.com/andyet/SimpleWebRTC)

with [socket.io](https://github.com/socketio/socket.io/) layer completely removed for golang implementation.

This demo uses websocket instead.

The actual emit feature is rewritten in rtckit.js with websocket only

## Starting the demo
* cd into %Your project folder%/govideo
* run `./govideo` if you are a mac user
* **for Win & Linux** make sure you have golang properly installed and run `go build` before running the app
* Visit `localhost:8000` or `localhost:8080`
* Enter a room name and start webRTC

## Screen sharing
* For Firefox
For firefox open: about:config   search screensharing and add localhost:8080 to trust.

* For chrome
Screen sharing plugin is needed.
