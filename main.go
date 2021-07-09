package main

import (
	"databroker/config"
	"databroker/pkg/auth"
	"databroker/pkg/desk"
	"databroker/pkg/initial"
	"databroker/routers"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

	sso_api_url := os.Getenv("SSO_API_URL")
	if len(sso_api_url) != 0 {
		config.ServerLocation = "Cloud"

		config.Datacenter = os.Getenv("datacenter")
		config.Workspace = os.Getenv("workspace")
		config.Cluster = os.Getenv("cluster")
		config.Namespace = os.Getenv("namespace")
		config.External = os.Getenv("external")

		if config.Namespace == "ifpsdev" || config.Namespace == "ifpsdemo" {
			config.SSO_API_URL, _ = url.Parse("https://api-sso-ensaas.hz.wise-paas.com.cn/v4.0")
		} else {
			config.SSO_API_URL, _ = url.Parse(os.Getenv("SSO_API_URL"))
		}
		config.SSO_USERNAME = os.Getenv("SSO_USERNAME")
		config.SSO_PASSWORD = os.Getenv("SSO_PASSWORD")

		ensaasService := os.Getenv("ENSAAS_SERVICES")
		tempReader := strings.NewReader(ensaasService)
		m, _ := simplejson.NewFromReader(tempReader)
		mongodb := m.Get("mongodb").GetIndex(0).Get("credentials").MustMap()
		config.MongodbURL = mongodb["externalHosts"].(string)
		config.MongodbDatabase = mongodb["database"].(string)
		config.MongodbUsername = mongodb["username"].(string)
		config.MongodbPassword = mongodb["password"].(string)
	} else {
		config.ServerLocation = "On-Premise"

		config.IFP_DESK_USERNAME = os.Getenv("IFP_DESK_USERNAME")
		config.IFP_DESK_PASSWORD = os.Getenv("IFP_DESK_PASSWORD")

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

	ifps_desk_api_url := os.Getenv("IFP_DESK_API_URL")
	if len(ifps_desk_api_url) != 0 {
		config.IFP_DESK_API_URL, _ = url.Parse(ifps_desk_api_url)
	} else {
		config.IFP_DESK_API_URL, _ = url.Parse("https://ifp-organizer-" + config.Namespace + "-" + config.Cluster + "." + config.External + "/graphql")
	}

	ifps_andon_daemon_databroker_api_url := os.Getenv("IFPS_ANDON_DAEMON_DATABROKER_API_URL")
	if len(ifps_andon_daemon_databroker_api_url) != 0 {
		config.IFPS_ANDON_DAEMON_DATABROKER_API_URL, _ = url.Parse(ifps_andon_daemon_databroker_api_url)
	} else {
		config.IFPS_ANDON_DAEMON_DATABROKER_API_URL, _ = url.Parse("https://ifps-andon-daemon-databroker-" + config.Namespace + "-" + config.Cluster + "." + config.External)
	}

	config.DASHBOARD_API_URL, _ = url.Parse(os.Getenv("DASHBOARD_API_URL"))

	ifps_andon_datasource_api_url := os.Getenv("IFPS_ANDON_DATASOURCE_API_URL")
	if len(ifps_andon_datasource_api_url) != 0 {
		config.IFPS_ANDON_DATASOURCE_API_URL, _ = url.Parse(ifps_andon_datasource_api_url)
	} else {
		config.IFPS_ANDON_DATASOURCE_API_URL, _ = url.Parse("https://ifps-andon-datasource-" + config.Namespace + "-" + config.Cluster + "." + config.External)
	}

	ifps_desk_datasource_api_url := os.Getenv("IFP_DESK_DATASOURCE_API_URL")
	if len(ifps_desk_datasource_api_url) != 0 {
		config.IFP_DESK_DATASOURCE_API_URL, _ = url.Parse(ifps_desk_datasource_api_url)
	} else {
		config.IFP_DESK_DATASOURCE_API_URL, _ = url.Parse("https://ifp-data-hub-api-" + config.Namespace + "-" + config.Cluster + "." + config.External)
	}

	ifps_andon_ui_url := os.Getenv("IFPS_ANDON_UI_URL")
	if len(ifps_desk_datasource_api_url) != 0 {
		config.IFPS_ANDON_UI_URL, _ = url.Parse(ifps_andon_ui_url)
	} else {
		config.IFPS_ANDON_UI_URL, _ = url.Parse("https://ifps-andon-" + config.Namespace + "-" + config.Cluster + "." + config.External)
	}

	fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
	fmt.Println("Service Name:", config.ServiceName)
	fmt.Println("Server Location:", config.ServerLocation)
	fmt.Println("SSO API -> URL:", config.SSO_API_URL)
	fmt.Println("DASHBOARD API -> URL:", config.DASHBOARD_API_URL)
	fmt.Println("IFP DESK API -> URL:", config.IFP_DESK_API_URL)
	fmt.Println("IFP DESK DATASOURCE API -> URL:", config.IFP_DESK_DATASOURCE_API_URL)
	fmt.Println("IFPS ANDON UI -> URL:", config.IFPS_ANDON_UI_URL)
	fmt.Println("IFPS ANDON DATASOURCE API -> URL:", config.IFPS_ANDON_DATASOURCE_API_URL)
	fmt.Println("IFPS ANDON DAEMON DATABROKER API -> URL:", config.IFPS_ANDON_DAEMON_DATABROKER_API_URL)

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
	if config.ServerLocation == "Cloud" {
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

//TopoStarter ...
func topoStarter() {
	desk.GetTopology(config.DB)
	desk.MachineRawDataTable("init")
	desk.StationRawDataTable("init")
}

func dashboardPrincipal() {
	fmt.Println()
	fmt.Println("Dashboard")

	fmt.Println("[1/5] From Local Import Plugins")
	initial.FromLocalImportPlugins()

	fmt.Println("[2/5] Create Datasource")
	initial.CreateDatasource("iFactory Datasource", "advantech-ifp-datasource", config.IFP_DESK_DATASOURCE_API_URL.String())
	initial.CreateDatasource(config.ServiceName+" Datasource", "ifps-andon-datasource", config.IFPS_ANDON_DATASOURCE_API_URL.String())

	fmt.Println("[3/5] Create", config.ServiceName, "Folder")
	initial.CreateDashboardFolder(config.ServiceName)

	fmt.Println("[4/5] Import Dashboard")
	initial.ImportDashboard()

	fmt.Println("[5/5] Create", config.ServiceName, "SRP")
	initial.CreateSRP(config.ServiceName)
}

func deskPrincipal() {
	fmt.Println()
	fmt.Println("Desk")

	fmt.Println("[1/1] Register " + config.ServiceName + " In Desk Outbound")
	initial.RegisterOutbound(config.ServiceName)

	fmt.Println("[1/1] Register", "iFactory/"+config.ServiceName, "In Desk CommandCenter")
	initial.RegisterCommandCenter("iFactory/" + config.ServiceName)
}

// var wg sync.WaitGroup

func main() {
	// wg.Add(2)
	initGlobalVar()

	if config.ServerLocation == "Cloud" {
		go auth.CloudSSOToken()
		go auth.CloudIFPToken()
	} else {
		auth.OnPremiseDashboardToken()
		go auth.OnPremiseIFPToken()
	}

	time.Sleep(10 * time.Second)

	fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
	fmt.Println(config.ServiceName, "Init")
	deskPrincipal()
	// dashboardPrincipal()
	fmt.Println()

	go topoStarter()

	router := routers.InitRouter()
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: router,
	}
	s.ListenAndServe()
	// wg.Wait()
}
