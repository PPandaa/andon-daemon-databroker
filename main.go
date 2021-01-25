package main

import (
	"bytes"
	"databroker/config"
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

var taipeiTimeZone, utcTimeZone *time.Location

func initGlobalVar() {
	taipeiTimeZone, _ = time.LoadLocation("Asia/Taipei")
	utcTimeZone, _ = time.LoadLocation("UTC")
	err := godotenv.Load(config.EnvName)
	if err != nil {
		log.Fatalf("Error loading env file")
	}
	config.AdminUsername = os.Getenv("ADMIN_USERNAME")
	config.AdminPassword = os.Getenv("ADMIN_PASSWORD")
	config.MongodbURL = os.Getenv("MONGODB_URL")
	config.MongodbUsername = os.Getenv("MONGODB_USERNAME")
	config.MongodbPassword = os.Getenv("MONGODB_PASSWORD")
	config.MongodbDatabase = os.Getenv("MONGODB_DATABASE")
	fmt.Println("MongoDB ->", " URL:", config.MongodbURL, " Database:", config.MongodbDatabase)
}

func refreshToken() {
	for {
		httpClient := &http.Client{}
		content := map[string]string{"username": config.AdminUsername, "password": config.AdminPassword}
		variable := map[string]interface{}{"input": content}
		httpRequestBody, _ := json.Marshal(map[string]interface{}{
			"query":     "mutation signIn($input: SignInInput!) {   signIn(input: $input) {     user {       name       __typename     }     __typename   } }",
			"variables": variable,
		})
		// request, _ := http.NewRequest("POST", "https://ifp-organizer-training-eks011.hz.wise-paas.com.cn/graphql", bytes.NewBuffer(httpRequestBody))
		request, _ := http.NewRequest("POST", "https://ifp-organizer-tienkang-eks002.sa.wise-paas.com/graphql", bytes.NewBuffer(httpRequestBody))
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
		fmt.Println(time.Now().In(taipeiTimeZone), "=>  Refresh Token ->", config.Token)
		time.Sleep(60 * time.Minute)
	}
}

//BrokerStarter ...
func BrokerStarter() {
	fmt.Println(time.Now().In(taipeiTimeZone), "=>  Broker Activation")
	session, _ := mgo.Dial(config.MongodbURL)
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
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
	session, _ := mgo.Dial(config.MongodbURL)
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
	time.Sleep(5 * time.Second)
	desk.GetTopology(db)
}

var wg sync.WaitGroup

func main() {
	wg.Add(3)
	initGlobalVar()
	go refreshToken()
	go BrokerStarter()
	TopoStarter()

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
