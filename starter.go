package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/joho/godotenv"
)

var nowTime time.Time
var taipeiTimeZone, utcTimeZone *time.Location
var eiToken, mongodbURL, mongodbUsername, mongodbPassword, mongodbDatabase, adminUsername, adminPassword, ssoURL string

func initGlobalVar() {
	taipeiTimeZone, _ = time.LoadLocation("Asia/Taipei")
	utcTimeZone, _ = time.LoadLocation("UTC")
	err := godotenv.Load(".env")
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
	// fmt.Println(eiToken)
}

//Starter ...
func Starter() {
	initGlobalVar()
	refreshEIToken()
	for {
		nowTime = time.Now().In(taipeiTimeZone)
		if nowTime.Minute() == 30 {
			refreshEIToken()
		}
		// if nowTime.Minute()%2 == 0 && nowTime.Second() == 0 {
		// 	DataBroker.TransmitData(eiToken, nowTime, taipeiTimeZone, mongodbURL, mongodbUsername, mongodbPassword, mongodbDatabase)
		// }
		TransmitData(eiToken, nowTime, taipeiTimeZone, mongodbURL, mongodbUsername, mongodbPassword, mongodbDatabase)
		time.Sleep(5 * time.Second)
	}
}
