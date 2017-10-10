package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	//"strings"
	"text/template"
	"io/ioutil"

	"github.com/gorilla/mux"
	databox "github.com/me-box/lib-go-databox"
)

var dataStoreHref = os.Getenv("DATABOX_STORE_ENDPOINT")

func getStatusEndpoint(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("active\n"))
}

type OauthConfig struct {
  ClientID string `json:"client_id"`
  ClientSecret string `json:"client_secret"`
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

var oauthConfig OauthConfig

func main() {

	// read config
	data,err := ioutil.ReadFile("etc/oauth.json")
	if err != nil {
		fmt.Print("Unable to read etc/oauth.json")
		panic(err)
	}
	err = json.Unmarshal(data, &oauthConfig)
	if err != nil {
		fmt.Printf("Unable to unmarshall etc/oauth.json: %s\n", string(data))
		panic(err)
	}
	fmt.Printf("oauth config %s,%s from %s\n", oauthConfig.ClientID, oauthConfig.ClientSecret, string(data))

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
