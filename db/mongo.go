package db

import (
	"databroker/config"

	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
)

func UseCollection(collection string) *mgo.Collection {
	return config.DB.C(collection)
}

func Insert(collection string, v interface{}) error {
	config.DbHealthCheck()
	c := UseCollection(collection)
	err := c.Insert(v)
	if err != nil {
		glog.Error(err)
	}
	return err
}

func Upadte(collection string, query, v interface{}) error {
	config.DbHealthCheck()
	c := UseCollection(collection)
	err := c.Update(query, v)
	if err != nil {
		glog.Error(err)
	}
	return err
}

func Delete(collection string, v interface{}) error {
	config.DbHealthCheck()
	c := UseCollection(collection)
	err := c.Remove(v)
	if err != nil {
		glog.Error(err)
	}
	return err
}
