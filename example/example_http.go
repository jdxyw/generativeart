package main

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	address := ":8090"
	log.Println("Server started at", address)

	// Initialize handlers
	http.HandleFunc("/", drawHandler)
	log.Fatal(http.ListenAndServe(address, nil))
}

// drawHandler is writes a piece of generative art
// as a response to an http request
func drawHandler(w http.ResponseWriter, r *http.Request) {

	// Log Requests
	log.Printf("method=%s path=%s ", r.Method, r.RequestURI)

	// Draw the image to bytes
	b, err := drawMyBytes()
	if err != nil {
		log.Fatal(err)
	}

	// Set content headers
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))

	// Write image to response
	if _, err = w.Write(b); err != nil {
		log.Fatal("unable to write image.")
	}
}

func drawMyBytes() ([]byte, error) {

	// Generate a new image
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.SetColorSchema(common.DarkRed)
	c.SetForeground(common.LightPink)
	c.Draw(arts.NewJanus(10, 0.2))

	// Return the image as []byte
	return c.ToBytes()
}
