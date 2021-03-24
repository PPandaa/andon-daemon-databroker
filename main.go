package main

import (
	"bytes"
	"databroker/config"
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
	fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
	fmt.Println("IFP ->", " URL:", config.IFPURL, " Username:", config.AdminUsername)

	newSession, err := mgo.Dial(config.MongodbURL)
	if err != nil {
		fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
		fmt.Println("MongoDB", err, "->", "URL:", config.MongodbURL, " Database:", config.MongodbDatabase)
		for err != nil {
			newSession, err = mgo.Dial(config.MongodbURL)
			time.Sleep(5 * time.Second)
		}
	}
	config.Session = newSession
	config.DB = config.Session.DB(config.MongodbDatabase)
	config.DB.Login(config.MongodbUsername, config.MongodbPassword)
	fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
	fmt.Println("MongoDB Connect ->", " URL:", config.MongodbURL, " Database:", config.MongodbDatabase)
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
				fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
				fmt.Println("retry refreshToken")
				httpRequestBody, _ = json.Marshal(map[string]interface{}{
					"query":     "mutation signIn($input: SignInInput!) {   signIn(input: $input) {     user {       name       __typename     }     __typename   } }",
					"variables": variable,
				})
				request, _ = http.NewRequest("POST", config.IFPURL, bytes.NewBuffer(httpRequestBody))
				request.Header.Set("Content-Type", "application/json")
				response, _ = httpClient.Do(request)
				m, _ = simplejson.NewFromReader(response.Body)
				time.Sleep(6 * time.Minute)
			}
		}
		header := response.Header
		cookie := header["Set-Cookie"]
		var ifpToken, eiToken string
		for _, cookieContent := range cookie {
			tempSplit := strings.Split(cookieContent, ";")
			if strings.HasPrefix(tempSplit[0], "IFPToken") {
				ifpToken = tempSplit[0]
			} else if strings.HasPrefix(tempSplit[0], "EIToken") {
				eiToken = tempSplit[0]
			}
		}
		if eiToken == "" {
			config.Token = ifpToken
		} else {
			config.Token = ifpToken + ";" + eiToken
		}
		fmt.Println("Token:", config.Token)
		time.Sleep(60 * time.Minute)
	}
}

//TopoStarter ...
func TopoStarter() {
	desk.GetTopology(config.DB)
	desk.MachineRawDataTable("init")
}

var wg sync.WaitGroup

func main() {
	wg.Add(1)

	initGlobalVar()
	go refreshToken()
	time.Sleep(10 * time.Second)
	TopoStarter()

	router := routers.InitRouter()
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: router,
	}
	s.ListenAndServe()
	wg.Wait()
}
