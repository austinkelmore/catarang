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

	// check to see if we want to clean everything, first
	clean := flag.Bool("clean", false, "Cleans everything so you start of fresh for testing")
	qtest := flag.Bool("qtest", false, "Sets up a quick test assuming you have the normal build setup.")
	flag.Parse()
	if *qtest {
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
		os.Remove(catarang.ConfigFileName)
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
