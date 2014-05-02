package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	port string
)

func init() {
	flag.StringVar(&port, "port", "8080", "HTTP bind port")
}

func main() {
	flag.Parse()

	http.HandleFunc("/", hello)
	log.Println("Downsizr listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, world")
}
