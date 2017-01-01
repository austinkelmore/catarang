package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

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

	clean := flag.Bool("clean", false, "Cleans everything so you start of fresh for testing")
	dev := flag.Bool("dev", false, "Sets up a dev environment to point to the web assets, but run in the bin dir assuming you have the normal build setup.")
	flag.Parse()

	if *dev {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal("Couldn't get the directory of executable")
		}

		if err := os.Chdir(dir); err != nil {
			log.Fatal("Couldn't set working directory to", dir)
		}

		log.Println("Set working directory to", dir)
		web.SetWebDir("../web")
	}
	if *clean {
		catarang.Clean()
		os.RemoveAll("jobs/")
		os.RemoveAll("results/")

		log.Println("Cleaned catarang to start from scratch")
	}

	catarang.ReadInConfig()

	r := web.CreateRoutes()
	http.Handle("/", r)

	// greeter.ServerStart(os.Args)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
