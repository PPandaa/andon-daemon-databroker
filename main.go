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

	"github.com/bitly/go-simplejson"
	"github.com/joho/godotenv"
	"gopkg.in/mgo.v2"
)

func initGlobalVar() {
	config.TaipeiTimeZone, _ = time.LoadLocation("Asia/Taipei")
	config.UTCTimeZone, _ = time.LoadLocation("UTC")
	err := godotenv.Load(config.EnvName)
	if err != nil {
		log.Fatalf("Error loading env file")
	}
	config.IFPURL = os.Getenv("IFP_URL")
	config.AdminUsername = os.Getenv("ADMIN_USERNAME")
	config.AdminPassword = os.Getenv("ADMIN_PASSWORD")
	config.MongodbURL = os.Getenv("MONGODB_URL")
	config.MongodbUsername = os.Getenv("MONGODB_USERNAME")
	config.MongodbPassword = os.Getenv("MONGODB_PASSWORD")
	config.MongodbDatabase = os.Getenv("MONGODB_DATABASE")
	fmt.Println("IFP ->", " URL:", config.IFPURL, " Username:", config.AdminUsername)
	fmt.Println("MongoDB ->", " URL:", config.MongodbURL, " Database:", config.MongodbDatabase)
}

func refreshToken() {
	for {
		fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
		fmt.Println("refreshToken")
		httpClient := &http.Client{}
		content := map[string]string{"username": config.AdminUsername, "password": config.AdminPassword}
		variable := map[string]interface{}{"input": content}
		httpRequestBody, _ := json.Marshal(map[string]interface{}{
			"query":     "mutation signIn($input: SignInInput!) {   signIn(input: $input) {     user {       name       __typename     }     __typename   } }",
			"variables": variable,
		})
		request, _ := http.NewRequest("POST", config.IFPURL, bytes.NewBuffer(httpRequestBody))
		request.Header.Set("Content-Type", "application/json")
		response, _ := httpClient.Do(request)
		m, _ := simplejson.NewFromReader(response.Body)
		for {
			if len(m.Get("errors").MustArray()) == 0 {
				break
			} else {
				httpRequestBody, _ = json.Marshal(map[string]interface{}{
					"query":     "mutation signIn($input: SignInInput!) {   signIn(input: $input) {     user {       name       __typename     }     __typename   } }",
					"variables": variable,
				})
				request, _ = http.NewRequest("POST", config.IFPURL, bytes.NewBuffer(httpRequestBody))
				request.Header.Set("Content-Type", "application/json")
				response, _ = httpClient.Do(request)
				m, _ = simplejson.NewFromReader(response.Body)
			}
		}
		// fmt.Println("-- GraphQL API End", time.Now().In(taipeiTimeZone))
		header := response.Header
		// fmt.Println(header)
		// m, _ := simplejson.NewFromReader(response.Header)
		cookie := header["Set-Cookie"]
		tempSplit := strings.Split(cookie[0], ";")
		ifpToken := tempSplit[0]
		tempSplit = strings.Split(cookie[1], ";")
		eiToken := tempSplit[0]
		config.Token = ifpToken + ";" + eiToken
		fmt.Println("Token:", config.Token)
		time.Sleep(60 * time.Minute)
	}
}

//TopoStarter ...
func TopoStarter() {
	session, _ := mgo.Dial(config.MongodbURL)
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
	time.Sleep(5 * time.Second)
	desk.GetTopology(db)
	session.Close()
	desk.MachineRawDataTable("init")
}

var wg sync.WaitGroup

func main() {
	wg.Add(3)
	initGlobalVar()
	go refreshToken()
	time.Sleep(10 * time.Second)
	TopoStarter()

	//------------------------->
	// v1.Test()
	db.StartMongo()

	// test mongo------------------->
	// query := bson.M{"StatusID": 123}
	// query := bson.M{"MachineName": "test0115"}
	// value := bson.M{"$set": bson.M{
	// 	"StatusRawValue":  nil,
	// 	"StatusLay1Value": nil,
	// 	"StatusMapValue":  nil,
	// }}
	// db.Upadte(db.MachineRawData, query, value)

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
