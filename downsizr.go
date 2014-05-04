package main

import (
	"encoding/json"
	"flag"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
  "errors"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
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
	ImageURL string

	Height int
	Width  int
}

type Response struct {
}

func downsize(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var request Request
	if err := decoder.Decode(&request); err != nil {
		log.Println(err)
		http.Error(res, "malformed request: " + err.Error(), http.StatusBadRequest)
		return
	}

	img, err := download_image(request.ImageURL)
	if err != nil {
		log.Println(err)
		http.Error(res, err.Error(), http.StatusBadRequest)
    return
	}

	resize.Resize(1000, 0, *img, resize.Lanczos3)
}

func download_image(url string) (*image.Image, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	img_encoding := res.Header.Get("Content-Type")

	var img image.Image
	switch img_encoding {
	case "image/jpeg":
		img, err = jpeg.Decode(res.Body)

	default:
    err = errors.New("Unknown encoding " + img_encoding)
	}

	if err != nil {
    log.Println(err)
		return nil, err
	}

	return &img, nil
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
