package main

import (
	"bytes"
	"databroker/db"
	"databroker/pkg/desk"
	"databroker/routers"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/mgo.v2"
)

const (
	envName = "dev.env"
)

var taipeiTimeZone, utcTimeZone *time.Location
var mongodbURL, mongodbUsername, mongodbPassword, mongodbDatabase, adminUsername, adminPassword, ssoURL string

func initGlobalVar() {
	taipeiTimeZone, _ = time.LoadLocation("Asia/Taipei")
	utcTimeZone, _ = time.LoadLocation("UTC")
	err := godotenv.Load(envName)
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

func refreshToken() {
	for {
		httpClient := &http.Client{}
		content := map[string]string{"username": adminUsername, "password": adminPassword}
		variable := map[string]interface{}{"input": content}
		httpRequestBody, _ := json.Marshal(map[string]interface{}{
			"query":     "mutation signIn($input: SignInInput!) {   signIn(input: $input) {     user {       name       __typename     }     __typename   } }",
			"variables": variable,
		})
		request, _ := http.NewRequest("POST", "https://ifp-organizer-training-eks011.hz.wise-paas.com.cn/graphql", bytes.NewBuffer(httpRequestBody))
		request.Header.Set("Content-Type", "application/json")
		response, _ := httpClient.Do(request)
		// fmt.Println("-- GraphQL API End", time.Now().In(taipeiTimeZone))
		header := response.Header
		// fmt.Println(header)
		// m, _ := simplejson.NewFromReader(response.Header)
		cookie := header["Set-Cookie"]
		tempSplit := strings.Split(cookie[0], ";")
		ifpToken := tempSplit[0]
		tempSplit = strings.Split(cookie[1], ";")
		eiToken := tempSplit[0]
		desk.Token = ifpToken + ";" + eiToken
		fmt.Println(time.Now().In(taipeiTimeZone), "=>  Refresh Token ->", desk.Token)
		time.Sleep(60 * time.Minute)
	}
}

//BrokerStarter ...
func BrokerStarter() {
	fmt.Println(time.Now().In(taipeiTimeZone), "=>  Broker Activation")
	session, _ := mgo.Dial(mongodbURL)
	db := session.DB(mongodbDatabase)
	db.Login(mongodbUsername, mongodbPassword)
	for {
		var nowTime time.Time
		nowTime = time.Now().In(taipeiTimeZone)
		// if nowTime.Minute()%2 == 0 && nowTime.Second() == 0 {
		// 	databroker.TransmitData(eiToken, nowTime, taipeiTimeZone, mongodbURL, mongodbUsername, mongodbPassword, mongodbDatabase)
		// }
		// --------- broker.go
		// fmt.Println("-- TransmitData Start", time.Now().In(taipeiTimeZone))
		desk.TransmitData(nowTime, db)
		if nowTime.Minute() == 0 && nowTime.Second() <= 10 {
			transmitDataEndtime := time.Now().In(taipeiTimeZone)
			transmitDataExectime := transmitDataEndtime.Sub(nowTime)
			fmt.Printf("%s =>  TransmitDataExecTime ->  %.1f Sec\n", nowTime, transmitDataExectime.Seconds())
		}
		// fmt.Println("-- TransmitData End", time.Now().In(taipeiTimeZone))
		time.Sleep(1 * time.Second)
	}
}

//TopoStarter ...
func TopoStarter() {
	fmt.Println(time.Now().In(taipeiTimeZone), "=>  Topo Activation")
	session, _ := mgo.Dial(mongodbURL)
	db := session.DB(mongodbDatabase)
	db.Login(mongodbUsername, mongodbPassword)
	for {
		time.Sleep(10 * time.Second)
		desk.GetTopology(desk.Token, db)
	}
}

var wg sync.WaitGroup

func main() {
	wg.Add(3)
	initGlobalVar()
	go refreshToken()
	go BrokerStarter()
	go TopoStarter()

	//------------------------->
	// v1.Test()

	db.StartMongo()
	router := routers.InitRouter()

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: router,
		// ReadTimeout:    ReadTimeout,
		// WriteTimeout:   WriteTimeout,
		// MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
	wg.Wait()
}
