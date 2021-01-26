package config

import "time"

const (
	EnvName = "tienkang.env"
)

var (
	IFPURL          string
	MongodbURL      string
	MongodbUsername string
	MongodbPassword string
	MongodbDatabase string
	AdminUsername   string
	AdminPassword   string
	Token           string
	TaipeiTimeZone  *time.Location
	UTCTimeZone     *time.Location
)
