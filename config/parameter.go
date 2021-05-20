package config

import (
	"fmt"
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
	IFPURL            string
	MongodbURL        string
	MongodbUsername   string
	MongodbPassword   string
	MongodbDatabase   string
	AdminUsername     string
	AdminPassword     string
	OutboundURL       string
	Token             string
	Datacenter        string
	Cluster           string
	Workspace         string
	Namespace         string
	SSOURL            string
	IFPStatus         = "Down"
	TaipeiTimeZone, _ = time.LoadLocation("Asia/Taipei")
	UTCTimeZone, _    = time.LoadLocation("UTC")

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
