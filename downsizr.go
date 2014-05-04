package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
)

var (
	port             string
	graphite_api_key string

	graphite *HostedGraphite
)

func init() {
	flag.StringVar(&port, "port", "8080", "HTTP bind port")
	flag.StringVar(&graphite_api_key, "graphite_api_key", "", "Hosted Graphite API key")
}

func main() {
	flag.Parse()

	var err error
	graphite, err = NewHostedGraphite(graphite_api_key)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Hosted Graphite adapter connected", graphite.Host, graphite.Port)

	router := mux.NewRouter()
	router.HandleFunc("/resize", downsize).Methods("POST")

	timingHandler := TimingHandler(router)
	loggingHandler := handlers.LoggingHandler(os.Stdout, timingHandler)

	http.Handle("/", loggingHandler)
	log.Println("Downsizr listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type Request struct {
	ImageURL string

	Height uint
	Width  uint
}

type Response struct {
	Resized string
}

func downsize(res http.ResponseWriter, req *http.Request) {
	start := time.Now()

	decoder := json.NewDecoder(req.Body)
	var request Request
	if err := decoder.Decode(&request); err != nil {
		log.Println(err)
		http.Error(res, "malformed request: "+err.Error(), http.StatusBadRequest)
		return
	}
	graphite.Send("parse_request", time.Since(start).Nanoseconds())

	data, content_type, err := download_image(request.ImageURL)
	if err != nil {
		log.Println(err)
		http.Error(res, "error downloading image: "+err.Error(), http.StatusBadRequest)
		return
	}
	graphite.Send("download_image", time.Since(start).Nanoseconds())

	img, err := decode_img(data, content_type)
	graphite.Send("decode_img", time.Since(start).Nanoseconds())

	resized := resize.Resize(request.Width, request.Height, *img, resize.Lanczos3)
	graphite.Send("resize_img", time.Since(start).Nanoseconds())

	data, err = encode_img(resized, content_type)
	if err != nil {
		log.Println(err)
		http.Error(res, "error encoding resized image: "+err.Error(), http.StatusBadRequest)
		return
	}
	graphite.Send("encode_img", time.Since(start).Nanoseconds())

	response := Response{base_64_encode(data, content_type)}

	encoder := json.NewEncoder(res)
	err = encoder.Encode(response)
	if err != nil {
		log.Println(err)
		http.Error(res, "error encoding response: "+err.Error(), http.StatusBadRequest)
		return
	}
	graphite.Send("write_response", time.Since(start).Nanoseconds())
}

func download_image(url string) ([]byte, string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, "", err
	}

	content_type := res.Header.Get("Content-Type")
	return data, content_type, nil
}

func TimingHandler(handler http.Handler) http.Handler {
	timer := func(res http.ResponseWriter, req *http.Request) {
		rec := httptest.NewRecorder()
		start := time.Now()

		handler.ServeHTTP(rec, req)

		for k, v := range rec.Header() {
			res.Header()[k] = v
		}
		res.Header().Set("X-Time-Elapsed", time.Since(start).String())

		res.WriteHeader(rec.Code)
		res.Write(rec.Body.Bytes())
	}

	return http.HandlerFunc(timer)
}
