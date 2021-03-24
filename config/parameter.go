package config

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
)

const (
	EnvName = "ifps-dev.env"
)

var (
	IFPURL            string
	MongodbURL        string
	MongodbUsername   string
	MongodbPassword   string
	MongodbDatabase   string
	AdminUsername     string
	AdminPassword     string
	Token             string
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
