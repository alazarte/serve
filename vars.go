package main

import (
	"fmt"
)

var (
	defaultPort = ":8080"

	// TODO read the URL from request
	schemeHostnameDefault = fmt.Sprintf("http://127.0.0.1%s", defaultPort)

	filepathError400     = "/400.html"
	filepathError404     = "/404.html"
	filepathError500     = "/500.html"
	filepathErrorUnknown = "unknown.html"
)
