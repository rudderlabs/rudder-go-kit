package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
)

const serverURL = "http://localhost:7777/hello-world"

func main() {
	logger := log.New(os.Stderr, "plainClient", log.Ldate|log.Ltime|log.Llongfile)

	var body []byte

	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "GET", serverURL, nil)

	logger.Println("Sending request...")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	body, err = io.ReadAll(res.Body)
	_ = res.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	logger.Printf("Response Received: %s", body)
}
