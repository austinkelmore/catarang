package main

import (
	"fmt"
	"net/http"
	"text/template"
)

func webHandler(w http.ResponseWriter, r *http.Request) {
	root, err := template.ParseFiles("root.html")
	if err != nil {
		fmt.Println("Can't open root.html file, does it exist in the root directory?")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	root.Execute(w, root)
}

func main() {
	http.HandleFunc("/", webHandler)
	http.ListenAndServe(":8080", nil)
}
