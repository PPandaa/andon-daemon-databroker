package main

import (
	"bytes"
	"databroker/pkg/desk"
	"databroker/routers"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/joho/godotenv"
	"gopkg.in/mgo.v2"
)

var taipeiTimeZone, utcTimeZone *time.Location
var eiToken, mongodbURL, mongodbUsername, mongodbPassword, mongodbDatabase, adminUsername, adminPassword, ssoURL string

func initGlobalVar() {
	taipeiTimeZone, _ = time.LoadLocation("Asia/Taipei")
	utcTimeZone, _ = time.LoadLocation("UTC")
	err := godotenv.Load("dev.env")
	if err != nil {
		log.Fatalf("Error loading env file")
	}
	adminUsername = os.Getenv("ADMIN_USERNAME")
	adminPassword = os.Getenv("ADMIN_PASSWORD")
	ssoURL = os.Getenv("SSO_URL")
	mongodbURL = os.Getenv("MONGODB_URL")
	mongodbUsername = os.Getenv("MONGODB_USERNAME")
	mongodbPassword = os.Getenv("MONGODB_PASSWORD")
	mongodbDatabase = os.Getenv("MONGODB_DATABASE")
	fmt.Println("MongoDB ->", " URL:", mongodbURL, " Database:", mongodbDatabase)
}

func refreshEIToken() {
	for {
		log.Println("RefreshEIToken")
		httpClient := &http.Client{}
		//----------get EIToken => var(eiToken)----------
		httpRequestBody, _ := json.Marshal(map[string]string{
			"username": adminUsername,
			"password": adminPassword,
		})
		request, _ := http.NewRequest("POST", ssoURL+"/v4.0/auth/native", bytes.NewBuffer(httpRequestBody))
		request.Header.Set("accept", "application/json")
		request.Header.Set("Content-Type", "application/json")
		response, err := httpClient.Do(request)
		for err != nil {
			fmt.Println("Refresh EIToken Fail , Retry!!")
			response, err = httpClient.Do(request)
		}
		m, _ := simplejson.NewFromReader(response.Body)
		eiToken = "EIToken=" + m.Get("accessToken").MustString()
		time.Sleep(60 * time.Minute)
	}
}

//BrokerStarter ...
func BrokerStarter() {
	// var nowTime time.Time
	log.Printf("Broker Activation")
	session, _ := mgo.Dial(mongodbURL)
	db := session.DB(mongodbDatabase)
	db.Login(mongodbUsername, mongodbPassword)
	for {
		// nowTime = time.Now().In(taipeiTimeZone)
		// if nowTime.Minute()%2 == 0 && nowTime.Second() == 0 {
		// 	databroker.TransmitData(eiToken, nowTime, taipeiTimeZone, mongodbURL, mongodbUsername, mongodbPassword, mongodbDatabase)
		// }
		// --------- broker.go
		// fmt.Println("-- TransmitData Start", time.Now().In(taipeiTimeZone))
		desk.TransmitData(eiToken, db)
		// fmt.Println("-- TransmitData End", time.Now().In(taipeiTimeZone))
		time.Sleep(1 * time.Second)
	}
}

var wg sync.WaitGroup

func main() {
	log.Printf("Main Activation")

	initGlobalVar()
	go refreshEIToken()
	go BrokerStarter()

	//------------------------->
	// v1.Test()

	router := routers.InitRouter()

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: router,
		// ReadTimeout:    ReadTimeout,
		// WriteTimeout:   WriteTimeout,
		// MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}
