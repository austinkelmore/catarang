package main

import (
	"log"
	"net/http"

	"github.com/austinkelmore/catarang/catarang"
	"github.com/austinkelmore/catarang/web"
)

func main() {
	log.Println("Running Catarang!")
	// if len(os.Args) >= 2 && os.Args[1] != "" {
	// 	greeter.ClientPluginStart()
	// 	return
	// }

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	catarang.ReadInConfig()

	catarang.AddJob("ping_test", "../ping_test")

	r := web.CreateRoutes()
	http.Handle("/", r)

	// greeter.ServerStart(os.Args)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
