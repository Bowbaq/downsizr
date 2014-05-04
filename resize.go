package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"

	"github.com/nfnt/resize"
)

func resize_img(req *Request, img *image.Image) image.Image {
	var algorithm resize.InterpolationFunction
	switch req.Algorithm {
	case "NearestNeighbor":
		algorithm = resize.NearestNeighbor
	case "Bilinear":
		algorithm = resize.Bilinear
	case "Bicubic":
		algorithm = resize.Bicubic
	case "MitchellNetravali":
		algorithm = resize.MitchellNetravali
	case "Lanczos2":
		algorithm = resize.Lanczos2
	case "Lanczos3":
		algorithm = resize.Lanczos3
	default:
		algorithm = resize.Lanczos3
	}

	return resize.Resize(req.Width, req.Height, *img, algorithm)
}

func decode_img(data []byte, content_type string) (*image.Image, error) {
	var (
		img image.Image
		err error
	)

	buffer := bytes.NewBuffer(data)

	switch content_type {
	case "image/jpeg":
		img, err = jpeg.Decode(buffer)
	case "image/gif":
		img, err = gif.Decode(buffer)
	case "image/png":
		img, err = png.Decode(buffer)

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
