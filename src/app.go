package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	//"strings"
	"text/template"

	"github.com/gorilla/mux"
	databox "github.com/me-box/lib-go-databox"
)

var dataStoreHref = os.Getenv("DATABOX_STORE_ENDPOINT")

func getStatusEndpoint(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("active\n"))
}

type Settings struct {
}

var settings = Settings{}

func displayUI(w http.ResponseWriter, req *http.Request) {
	var templates *template.Template
	templates, err := template.ParseFiles("tmpl/settings.tmpl")
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("error\n"))
		return
	}
	s1 := templates.Lookup("settings.tmpl")
	err = s1.Execute(w,settings)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("error\n"))
		return
	}
}

type data struct {
	Data string `json:"data"`
}

func main() {

	//Wait for my store to become active
	databox.WaitForStoreStatus(dataStoreHref)

	//
	// Handle Https requests
	//
	router := mux.NewRouter()

	router.HandleFunc("/status", getStatusEndpoint).Methods("GET")
	router.HandleFunc("/ui", displayUI).Methods("GET")

	static := http.StripPrefix("/ui/static", http.FileServer(http.Dir("./www/")))
	router.PathPrefix("/ui/static").Handler(static)

	log.Fatal(http.ListenAndServeTLS(":8080", databox.GetHttpsCredentials(), databox.GetHttpsCredentials(), router))
}
