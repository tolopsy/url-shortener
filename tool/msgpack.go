package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/tolopsy/url-shortener/shortener"
	"github.com/vmihailenco/msgpack"
)

func httpPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	return fmt.Sprintf(":%s", port)
}

func main() {
	urlToShorten := flag.String("url", "https://github.com/tolopsy", "The URL you want to shorten")
	flag.Parse()

	address := fmt.Sprintf("http://localhost%s", httpPort())
	redirect := shortener.Redirect{}
	redirect.URL = *urlToShorten

	body, err := msgpack.Marshal(&redirect)
	if err != nil {
		log.Fatalln(err)
	}

	response, err := http.Post(address, "application/x-msgpack", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalln(err)
	}
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}

	msgpack.Unmarshal(body, &redirect)
	log.Printf("%v\n", redirect)
}
