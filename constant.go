package main

// setup mongodb
const DB_NAME string = "rtc_meeting"
const PORT string = ":9727"

//var DB_USERS *mgo.Collection
//var DB_SESSIONS *mgo.Collection

// used for error reporting
var ERRSOURCE string

// The location where the HTML Templates are kept.
const TEMPLATE_PATH string = "view"

// USER_COOKIE is the name of the cookie used for
// user sessions.
//const USER_COOKIE string = "application-login"

// This Email RE will only fit 99% of all emails...
const EMAIL_RGX string = `(?i)[A-Z0-9._%+-]+@(?:[A-Z0-9-]+\.)+[A-Z]{2,6}`

const INVALID_ROOM int = 9

const TYPE_UPDATE_DISPLAY_NAME = 1
const TYPE_UPDATE_DISPLAY_NAME_RESP = 2
