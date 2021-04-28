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
	err := godotenv.Load(config.EnvPath)
	if err != nil {
		log.Fatalf("Error loading env file")
	}
	config.IFPURL = os.Getenv("IFP_DESK_API_URL")
	config.AdminUsername = os.Getenv("IFP_DESK_USERNAME")
	config.AdminPassword = os.Getenv("IFP_DESK_PASSWORD")
	config.OutboundURL = os.Getenv("IFPS_ANDON_DAEMON_DATABROKER_API_URL")
	ensaasService := os.Getenv("ENSAAS_SERVICES")
	if len(ensaasService) != 0 {
		tempReader := strings.NewReader(ensaasService)
		m, _ := simplejson.NewFromReader(tempReader)
		mongodb := m.Get("mongodb").GetIndex(0).Get("credentials").MustMap()
		config.MongodbURL = mongodb["externalHosts"].(string)
		config.MongodbDatabase = mongodb["database"].(string)
		config.MongodbUsername = mongodb["username"].(string)
		config.MongodbPassword = mongodb["password"].(string)
	} else {
		config.MongodbURL = os.Getenv("MONGODB_URL")
		config.MongodbDatabase = os.Getenv("MONGODB_DATABASE")
		config.MongodbUsername = os.Getenv("MONGODB_USERNAME")
		config.MongodbPassword = os.Getenv("MONGODB_PASSWORD")
	}
	fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
	fmt.Println("IFP -> URL:", config.IFPURL, " Username:", config.AdminUsername)
	fmt.Println("Outbound API -> URL:", config.OutboundURL)

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

func registerOutbound() {
	fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
	fmt.Println("registerOutbound")
	httpClient := &http.Client{}
	content := map[string]interface{}{"name": "ifps-andon", "sourceId": "scada_ifpsandon", "url": config.OutboundURL, "active": true}
	variable := map[string]interface{}{"input": content}
	httpRequestBody, _ := json.Marshal(map[string]interface{}{
		"query":     "mutation ($input: AddOutboundInput!) {     addOutbound(input: $input) {         outbound {             id             name             url             sourceId             allowUnauthorized             active             connected         }     } }",
		"variables": variable,
	})
	request, _ := http.NewRequest("POST", config.IFPURL, bytes.NewBuffer(httpRequestBody))
	request.Header.Set("cookie", config.Token)
	request.Header.Set("Content-Type", "application/json")
	response, _ := httpClient.Do(request)
	m, _ := simplejson.NewFromReader(response.Body)
	if len(m.Get("errors").MustArray()) == 0 {
		fmt.Println("Outbound ifps-andon:", config.OutboundURL)
	} else {
		fmt.Println("Outbound ifps-andon already exist")
	}
}

//TopoStarter ...
func topoStarter() {
	desk.GetTopology(config.DB)
	desk.MachineRawDataTable("init")
	desk.StationRawDataTable("init")
}

var wg sync.WaitGroup

func main() {
	wg.Add(1)

	initGlobalVar()
	go refreshToken()
	time.Sleep(10 * time.Second)
	registerOutbound()
	topoStarter()

	router := routers.InitRouter()
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: router,
	}
	s.ListenAndServe()

	wg.Wait()
}
