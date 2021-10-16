package main

import (
	"github.com/gorilla/mux"
	"log"
	"mime"
	"net/http"
)

//
//func home (w http.ResponseWriter, r *http.Request) {
//	_, _ = w.Write([]byte("Test!!!"))
//}

func main() {
	log.Println("Starting...")

	router := mux.NewRouter()

	_ = mime.AddExtensionType(".js", "application/javascript; charset=utf8")

	spa := &spaHandler{
		staticPath: "./client/dist",
		indexPath:  "./client/dist/index.html",
	}

	router.PathPrefix("/").Handler(spa)

	err := http.ListenAndServe(":8888", router)

	log.Fatal(err)
}
