package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	port string
)

func init() {
	flag.StringVar(&port, "port", "8080", "HTTP bind port")
}

func main() {
	flag.Parse()

	router := mux.NewRouter()
	router.HandleFunc("/resize", downsize).Methods("POST")

	timingHandler := TimingHandler(router)
	loggingHandler := handlers.LoggingHandler(os.Stdout, timingHandler)

	http.Handle("/", loggingHandler)
	log.Println("Downsizr listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type Request struct {
	SourceURL string
	Sizes     []string
}

type Response struct {
}

func downsize(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var request Request
	if err := decoder.Decode(&request); err != nil {
		log.Println(err)
		http.Error(res, "malformed request", http.StatusBadRequest)
		return
	}

	fmt.Fprintln(res, "hello, world")
}

func TimingHandler(handler http.Handler) http.Handler {
	timer := func(res http.ResponseWriter, req *http.Request) {
		rec := httptest.NewRecorder()
		now := time.Now()

		handler.ServeHTTP(rec, req)

		for k, v := range rec.Header() {
			res.Header()[k] = v
		}
		res.Header().Set("X-Time-Elapsed", time.Since(now).String())

		res.WriteHeader(rec.Code)
		res.Write(rec.Body.Bytes())
	}

	return http.HandlerFunc(timer)
}
