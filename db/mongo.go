package db

import (
	"databroker/config"
	"fmt"

	"github.com/golang/glog"
	. "github.com/logrusorgru/aurora"
	"gopkg.in/mgo.v2"
)

var (
	//public mongo connection
	MongoDB *mongoDB
)

func StartMongo() {
	MongoDB = NewMongo()
	fmt.Println(BrightRed("StartMongo..."))
}

type mongoDB struct {
	Db *mgo.Database
}

func NewMongo() *mongoDB {
	return &mongoDB{
		Db: CreateConnection(),
	}
}

func CreateConnection() *mgo.Database {
	glog.Infoln("create mongodb connection...")
	session, err := mgo.Dial(config.MongodbURL)

	if err != nil {
		panic(err)
	}

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	db := session.DB(config.MongodbDatabase)
	err = db.Login(config.MongodbUsername, config.MongodbPassword)
	if err != nil {
		panic(err)
	}
	return db
}

func (mongodb *mongoDB) UseCollection(collection string) *mgo.Collection {
	return mongodb.Db.C(collection)
}

// Use Go original mongo library (more compplecated)-------------------------------------------------
/*
type mongoDB struct {
	mongoClient *mongo.Client
}

func NewMongo() *mongoDB {
	return &mongoDB{
		mongoClient: createConnection(),
	}
}

func createConnection() *mongo.Client {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://10.0.1.19:27017,10.0.1.20:27017,10.0.1.21:27017")
	// client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	return client
}
*/

func Insert(collection string, v interface{}) error {
	c := MongoDB.UseCollection(collection)
	err := c.Insert(v)
	if err != nil {
		glog.Error(err)
	}
	return err
}

func Delete(collection string, v interface{}) error {
	c := MongoDB.UseCollection(collection)
	err := c.Remove(v)
	if err != nil {
		glog.Error(err)
	}
	return err
}
