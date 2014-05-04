package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
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

	Height uint
	Width  uint
}

type Response struct {
	Resized string
}

func downsize(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var request Request
	if err := decoder.Decode(&request); err != nil {
		log.Println(err)
		http.Error(res, "malformed request: "+err.Error(), http.StatusBadRequest)
		return
	}

	img, content_type, err := download_image(request.ImageURL)
	if err != nil {
		log.Println(err)
		http.Error(res, "error downloading image: "+err.Error(), http.StatusBadRequest)
		return
	}

	resized := resize.Resize(request.Width, request.Height, *img, resize.Lanczos3)
	data, err := encode_img(resized, content_type)
	if err != nil {
		log.Println(err)
		http.Error(res, "error encoding resized image: "+err.Error(), http.StatusBadRequest)
		return
	}

	response := Response{base_64_encode(data, content_type)}
	encoder := json.NewEncoder(res)
	err = encoder.Encode(response)
	if err != nil {
		log.Println(err)
		http.Error(res, "error encoding response: "+err.Error(), http.StatusBadRequest)
		return
	}
}

func download_image(url string) (*image.Image, string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()

	content_type := res.Header.Get("Content-Type")
	img, err := decode_img(res.Body, content_type)

	return img, content_type, err
}

func decode_img(data io.ReadCloser, content_type string) (*image.Image, error) {
	var (
		img image.Image
		err error
	)

	switch content_type {
	case "image/jpeg":
		img, err = jpeg.Decode(data)
	case "image/gif":
		img, err = gif.Decode(data)
	case "image/png":
		img, err = png.Decode(data)

	default:
		err = errors.New("Unknown content type " + content_type)
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &img, nil
}

func encode_img(img image.Image, content_type string) ([]byte, error) {
	var (
		buffer bytes.Buffer
		err    error
	)

	switch content_type {
	case "image/jpeg":
		err = jpeg.Encode(&buffer, img, nil)
	case "image/gif":
		err = gif.Encode(&buffer, img, nil)
	case "image/png":
		err = png.Encode(&buffer, img)

	default:
		err = errors.New("Unknown content type " + content_type)
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return buffer.Bytes(), nil
}

func base_64_encode(data []byte, content_type string) string {
	return fmt.Sprintf("data:%s;base64,%s",
		content_type, base64.StdEncoding.EncodeToString(data))
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
