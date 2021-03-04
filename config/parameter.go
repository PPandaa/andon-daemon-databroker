package config

import (
	"time"

	"gopkg.in/mgo.v2"
)

const (
	EnvName = "local.env"
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
