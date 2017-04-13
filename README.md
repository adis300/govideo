# Go video demo for using webrtc

* The rtckit.js library is adapted from
[simplewebrtc.com](https://github.com/andyet/SimpleWebRTC)

* **completely removed**[socket.io](https://github.com/socketio/socket.io/) layer for Go implementation.

* The actual emit feature from socket.io is rewritten in with websocket only

## Starting the demo
* cd into %Your project folder%/govideo
* run `./govideo -port=8080 -portSecure=8443 -secure` if you are a mac user
* **for Win & Linux** make sure you have golang properly installed and run `go build` before running the app
* Visit `localhost:8000` or `localhost:8080`
* Enter a room name and start webRTC

## Screen sharing
* For Firefox
For firefox open: about:config   search screensharing and add localhost:8080 to trust.

* For chrome
Screen sharing plugin is needed.

* For creating iOS Client don't forget to add the following permissions
```
Privacy - Camera Usage Description    : Your reason
Privacy - Microphone Usage Description    : You reason
Also disable bitcode of your project

```
