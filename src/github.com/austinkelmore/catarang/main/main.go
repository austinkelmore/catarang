package main

import (
	"log"
	"net/http"

	"github.com/austinkelmore/catarang/catarang"
	"github.com/austinkelmore/catarang/web"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Running Catarang!")
	catarang.ReadInConfig()

	r := web.CreateRoutes()
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
