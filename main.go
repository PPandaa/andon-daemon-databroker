package main

import (
	"bytes"
	"databroker/config"
	"databroker/pkg/desk"
	"databroker/routers"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
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

	config.Datacenter = os.Getenv("datacenter")
	config.Workspace = os.Getenv("workspace")
	config.Cluster = os.Getenv("cluster")
	config.Namespace = os.Getenv("namespace")
	if config.Namespace == "ifpsdev" || config.Namespace == "ifpsdemo" {
		config.SSOURL = "https://api-sso-ensaas.hz.wise-paas.com.cn/v4.0"
	} else {
		config.SSOURL = os.Getenv("SSO_API_URL")
	}
	external := os.Getenv("external")

	ifps_desk_api_url := os.Getenv("IFP_DESK_API_URL")
	if len(ifps_desk_api_url) != 0 {
		config.IFPURL = ifps_desk_api_url
	} else {
		config.IFPURL = "https://ifp-organizer-" + config.Namespace + "-" + config.Cluster + "." + external + "/graphql"
	}

	config.AdminUsername = os.Getenv("IFP_DESK_USERNAME")
	config.AdminPassword = os.Getenv("IFP_DESK_PASSWORD")

	ifps_andon_daemon_databroker_api_url := os.Getenv("IFPS_ANDON_DAEMON_DATABROKER_API_URL")
	if len(ifps_andon_daemon_databroker_api_url) != 0 {
		config.OutboundURL = ifps_andon_daemon_databroker_api_url
	} else {
		config.OutboundURL = "https://ifps-andon-daemon-databroker-" + config.Namespace + "-" + config.Cluster + "." + external
	}

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
		config.MongodbAuthSource = os.Getenv("MONGODB_AUTH_SOURCE")
		mongodbPasswordFile := os.Getenv("MONGODB_PASSWORD_FILE")
		if len(mongodbPasswordFile) != 0 {
			mongodbPassword, err := ioutil.ReadFile(mongodbPasswordFile)
			if err != nil {
				fmt.Println("MongoDB Password File", err, "->", "FilePath:", mongodbPasswordFile)
			} else {
				config.MongodbPassword = string(mongodbPassword)
			}
		}
	}

	fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
	fmt.Println("IFP -> URL:", config.IFPURL)
	fmt.Println("SSO -> URL:", config.SSOURL)
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
	if len(ensaasService) != 0 {
		config.DB = config.Session.DB(config.MongodbDatabase)
		config.DB.Login(config.MongodbUsername, config.MongodbPassword)
	} else {
		config.DB = config.Session.DB(config.MongodbAuthSource)
		config.DB.Login(config.MongodbUsername, config.MongodbPassword)
		config.DB = config.Session.DB(config.MongodbDatabase)
	}

	fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
	fmt.Println("MongoDB Connect ->", " URL:", config.MongodbURL, " Database:", config.MongodbDatabase)
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
	if len(config.Datacenter) == 0 {
		request.Header.Set("cookie", config.Token)
	} else {
		request.Header.Set("X-Ifp-App-Secret", config.Token)
	}
	request.Header.Set("Content-Type", "application/json")
	response, _ := httpClient.Do(request)
	if response.StatusCode == 200 {
		config.IFPStatus = "Up"
	} else {
		config.IFPStatus = "Down"
	}
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

// var wg sync.WaitGroup

func main() {
	// wg.Add(2)

	initGlobalVar()
	go desk.RefreshToken()
	time.Sleep(10 * time.Second)
	registerOutbound()
	go topoStarter()

	router := routers.InitRouter()
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: router,
	}
	s.ListenAndServe()

	// wg.Wait()
}
