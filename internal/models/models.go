package models

import "sync"

var URLs = make(map[string]string) // key - full URL, value - short URL
var Mu sync.Mutex
