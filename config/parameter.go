package config

import (
	"fmt"
	"net/url"
	"time"

	"gopkg.in/mgo.v2"
)

const (
	EnvPath = "local.env"

	MachineRawData     = "iii.dae.MachineRawData"
	StationRawData     = "iii.dae.StationRawData"
	MachineRawDataHist = "iii.dae.MachineRawDataHist"
	Statistic          = "iii.dae.Statistics"
	DailyStatistics    = "iii.dae.DailyStatistics"
	MonthlyStatistics  = "iii.dae.MonthlyStatistics"
	YearlyStatistics   = "iii.dae.YearlyStatistics"
	EventLatest        = "iii.dae.EventLatest"
	EventHist          = "iii.dae.EventHist"
	GroupTopo          = "iii.cfg.GroupTopology"
	TPCList            = "iii.cfg.TPCList"
)

var (
	TaipeiTimeZone, _ = time.LoadLocation("Asia/Taipei")
	UTCTimeZone, _    = time.LoadLocation("UTC")
	IFPStatus         = "Down"
	ServiceName       = "Andon"

	ServerLocation string
	Datacenter     string
	Cluster        string
	Workspace      string
	Namespace      string
	External       string

	SSO_API_URL       *url.URL
	SSO_USERNAME      string
	SSO_PASSWORD      string
	DASHBOARD_API_URL *url.URL

	IFP_DESK_API_URL  *url.URL
	IFP_DESK_USERNAME string
	IFP_DESK_PASSWORD string

	IFPS_ANDON_UI_URL                    *url.URL
	IFP_DESK_DATASOURCE_API_URL          *url.URL
	IFPS_ANDON_DATASOURCE_API_URL        *url.URL
	IFPS_ANDON_DAEMON_DATABROKER_API_URL *url.URL

	MongodbURL        string
	MongodbUsername   string
	MongodbPassword   string
	MongodbDatabase   string
	MongodbAuthSource string

	DashboardToken string
	IFPToken       string

	DB      *mgo.Database
	Session *mgo.Session
)

func DbHealthCheck() {
	err := Session.Ping()
	if err != nil {
		fmt.Println("----------", time.Now().In(TaipeiTimeZone), "----------")
		fmt.Println("MongoDB", err, "->", "URL:", MongodbURL, " Database:", MongodbDatabase)
		Session.Refresh()
		fmt.Println("----------", time.Now().In(TaipeiTimeZone), "----------")
		fmt.Println("MongoDB Reconnect ->", " URL:", MongodbURL, " Database:", MongodbDatabase)
	}
}
