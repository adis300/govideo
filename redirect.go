package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"
)

func secureRedirectHandler(code int) http.Handler {
	return &secureHandler{code}
}

func httpRedirectHandler(url string, code int) http.Handler {
	return &redirectHandler{url, code}
}

type secureHandler struct {
	code int
}

func (rh *secureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	oldpath := r.URL.Path
	if oldpath == "" { // should not happen, but avoid a crash if it does
		oldpath = "/"
	}
	// Use oldHost value to be the new url
	newUrl := "https://" + strings.Replace(r.Host, portStr, portSecureStr, 1) + oldpath // Replace one instance
	w.Header().Set("Location", newUrl)
	w.WriteHeader(rh.code)
	// RFC2616 recommends that a short note "SHOULD" be included in the
	// response because older user agents may not understand 301/307.
	// Shouldn't send the response for POST or HEAD; that leaves GET.
	if r.Method == "GET" {
		note := "<a href=\"" + template.HTMLEscapeString(newUrl) + "\">" + http.StatusText(rh.code) + "</a>.\n"
		fmt.Fprintln(w, note)
	}
}

type redirectHandler struct {
	url  string
	code int
}

func (rh *redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := url.Parse(rh.url); err == nil {
		oldpath := r.URL.Path
		if oldpath == "" { // should not happen, but avoid a crash if it does
			oldpath = "/"
		}
		rh.url += oldpath
		w.Header().Set("Location", rh.url)
		w.WriteHeader(rh.code)
		// RFC2616 recommends that a short note "SHOULD" be included in the
		// response because older user agents may not understand 301/307.
		// Shouldn't send the response for POST or HEAD; that leaves GET.
		if r.Method == "GET" {
			note := "<a href=\"" + template.HTMLEscapeString(rh.url) + "\">" + http.StatusText(rh.code) + "</a>.\n"
			fmt.Fprintln(w, note)
		}
	}
}
