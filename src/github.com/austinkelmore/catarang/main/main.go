package main

import (
	"log"
	"net/http"
	"os"

	"github.com/austinkelmore/catarang/catarang"
	"github.com/austinkelmore/catarang/greeter"
	"github.com/austinkelmore/catarang/web"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] != "" {
		greeter.ClientPluginStart()
		return
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Running Catarang!")
	catarang.ReadInConfig()

	r := web.CreateRoutes()
	http.Handle("/", r)

	greeter.ServerStart(os.Args)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
