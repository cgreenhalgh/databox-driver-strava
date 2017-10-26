package main

import (
	"encoding/json"
	//"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	//"strings"
	"html/template"
	"io/ioutil"
	"strconv"
	"sync"
	
	"github.com/gorilla/mux"
	databox "github.com/me-box/lib-go-databox"
)

var dataStoreHref = os.Getenv("DATABOX_STORE_ENDPOINT")

func getStatusEndpoint(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("active\n"))
}

type OauthUris struct {
  AuthUri string
  TokenUri string
}

var oauthUris = OauthUris{ AuthUri: "https://www.strava.com/oauth/authorize?response_type=code&scope=view_private&approval_prompt=force&state=oauth_callback&", TokenUri: "https://www.strava.com/oauth/token"}

type OauthConfig struct {
  ClientID string `json:"client_id"`
  ClientSecret string `json:"client_secret"`
}

type DriverStatus int

const (
	DRIVER_STARTING DriverStatus = iota
	DRIVER_FATAL 
	DRIVER_UNAUTHORIZED 
	DRIVER_OK 
)

type Settings struct {
  Status DriverStatus
  Error string // if there is an error
  ClientID string
  AuthUri string
  Authorized bool
  AthleteID string `json:"athlete_id"`
  Firstname string `json:"firstname"`
  Lastname string `json:"lastname"`
}

type State struct {
  AccessToken string `json:"access_token"`
  AthleteID string `json:"athlete_id"`
  Firstname string `json:"firstname"`
  Lastname string `json:"lastname"`
}

func (state *State) Load() {
	data,err := ioutil.ReadFile("etc/state.json")
	if err != nil {
		log.Print("Unable to read etc/state.json\n")
		return
	}
	err = json.Unmarshal(data, &state)
	if err != nil {
		log.Printf("Unable to unmarshall etc/state.json: %s\n", string(data))
		return
	}
	log.Printf("Read etc/state.json: %s\n", string(data))
}
func (state *State) Save() {
	data,err := json.Marshal(state)
	if err != nil {
		log.Printf("Unable to marshall state\n");
		return
	}
	err = ioutil.WriteFile("etc/state.json", data, 0666);
	if err != nil {
		log.Print("Error writing etc/state.json\n")
		return
	}
	log.Print("Saved etc/state.json\n")
}

var settingsLock = &sync.Mutex{}
var settings = Settings{Status: DRIVER_STARTING, AuthUri: oauthUris.AuthUri, Authorized:false}
var state = State{}

type Athlete struct {
	ID int64 `json:"id"`
	Firstname string `json:"firstname"`
	Lastname string `json:"lastname"`
}
type OauthTokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType string `json:"token_type"`
	Athlete Athlete `json:"athlete"`
}

func handleOauthCode(code string) {
	resp,err := http.PostForm(oauthUris.TokenUri, 
		url.Values{"client_id":{oauthConfig.ClientID}, "client_secret":{oauthConfig.ClientSecret}, "code":{code}})
	if err != nil {
		log.Print("Error getting token\n")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print("Error reading body of oauth token response\n")
		return
	} 
	log.Printf("oauth token resp %s\n", string(body))
	//state.AccessToken = code[0];
	//state.Save()
	var tokenResp = OauthTokenResp{}
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		log.Printf("Error unmarshalling oauth response: %s\n", string(body))
		return
	}
	//if tokenResp.AccessToken != nil {
	log.Printf("Got access token %s\n", tokenResp.AccessToken)
	settingsLock.Lock()
	state.AccessToken = tokenResp.AccessToken
	//if tokenResp.Athlete != nil {
	state.AthleteID = strconv.FormatInt(tokenResp.Athlete.ID, 10)
	state.Firstname = tokenResp.Athlete.Firstname
	state.Lastname = tokenResp.Athlete.Lastname
	//}
	state.Save()
	settings.Authorized = true
	settings.Firstname = state.Firstname
	settings.Lastname = state.Lastname
	settings.AthleteID = state.AthleteID
	//}
	settingsLock.Unlock()
}

func displayUI(w http.ResponseWriter, req *http.Request) {
	// auth callback?
	params := req.URL.Query()
	codes := params["code"]
	if codes != nil && len(codes)>0 {
		log.Printf("Got oauth response code = %s\n", codes[0])
		code := codes[0]
		handleOauthCode(code)
	}

	var templates *template.Template
	templates, err := template.ParseFiles("tmpl/settings.tmpl")
	if err != nil {
		log.Printf("Error parsing template: %s", err.Error())
		w.Write([]byte("error\n"))
		return
	}
	s1 := templates.Lookup("settings.tmpl")
	settingsLock.Lock()
	err = s1.Execute(w,settings)
	settingsLock.Unlock()
	if err != nil {
		log.Printf("Error filling template: %s", err.Error())
		w.Write([]byte("error\n"))
		return
	}
}
/*
func authCallback(w http.ResponseWriter, req *http.Request) {
	url := req.URL.String()
	ix := strings.LastIndex(url,"/")
	log.Printf("url %s -> %s\n", url, string(url[0:ix+1]))
	// proxy defeats redirect?
	//http.Redirect(w, req, string(url[0:ix]), 302)
	w.Write([]byte("<html><head><meta http-equiv=\"refresh\" content=\"0; URL="+string(url[0:ix+1])+"\" /></head></html>"))
}
*/
type data struct {
	Data string `json:"data"`
}

var oauthConfig OauthConfig

func loadSettings() {
	settingsLock.Lock()
	defer settingsLock.Unlock()
	// read config
	data,err := ioutil.ReadFile("etc/oauth.json")
	if err != nil {
		log.Print("Unable to read etc/oauth.json\n")
		settings.Status = DRIVER_FATAL
		settings.Error = "The driver was not build correctly (unable to read etc/oauth.json)"
		return
	}
	err = json.Unmarshal(data, &oauthConfig)
	if err != nil {
		log.Printf("Unable to unmarshall etc/oauth.json: %s\n", string(data))
		settings.Status = DRIVER_FATAL
		settings.Error = "The driver was not build correctly (unable to unmarshall etc/oauth.json)"
		return
	}
	log.Printf("oauth config client ID %s\n", oauthConfig.ClientID)
	settings.ClientID = oauthConfig.ClientID
	
	state.Load()
	if len(state.AccessToken)>0 {
		settings.Authorized = true
	}
	settings.Firstname = state.Firstname
	settings.Lastname = state.Lastname
	settings.AthleteID = state.AthleteID
}

func server(c chan bool) {
	//
	// Handle Https requests
	//
	router := mux.NewRouter()

	router.HandleFunc("/status", getStatusEndpoint).Methods("GET")
	//router.HandleFunc("/ui/auth_callback", authCallback).Methods("GET")
	router.HandleFunc("/ui", displayUI).Methods("GET")

	static := http.StripPrefix("/ui/static", http.FileServer(http.Dir("./www/")))
	router.PathPrefix("/ui/static").Handler(static)

	http.ListenAndServeTLS(":8080", databox.GetHttpsCredentials(), databox.GetHttpsCredentials(), router)
	log.Print("HTTP server exited?!")
	c <- true
}

func main() {

	loadSettings()	

	serverdone := make(chan bool)
	go server(serverdone)
		
	//Wait for my store to become active
	databox.WaitForStoreStatus(dataStoreHref)

	_ = <-serverdone
}
