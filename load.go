package main

import "io/ioutil"

// LoadFile loads a file data in to a byte array
func LoadFile(filepath string) []byte {
	data, _ := ioutil.ReadFile(filepath)
	return data
}

// LoadView is a wrapper around LoadFile to load templates
func LoadView(templateName string) []byte {
	filepath := "view/" + templateName + ".html"
	return LoadFile(filepath)
}
