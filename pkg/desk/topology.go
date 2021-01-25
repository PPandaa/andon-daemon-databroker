package desk

import (
	"bytes"
	"databroker/config"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bitly/go-simplejson"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// GetTopology initial 時 拿一次 Enabler  Group Topology
func GetTopology(db *mgo.Database) {
	grouptopologyCollection := db.C("iii.cfg.GroupTopology")

	httpClient := &http.Client{}

	httpRequestBody, _ := json.Marshal(map[string]interface{}{
		"query": "query groupsWithInboundConnector {   groups {     _id     id     name     parentId     timeZone     inboundConnector {       id       __typename        }             __typename   } }",
	})
	request, _ := http.NewRequest("POST", "https://ifp-organizer-tienkang-eks002.sa.wise-paas.com/graphql", bytes.NewBuffer(httpRequestBody))
	request.Header.Set("cookie", config.Token)
	request.Header.Set("Content-Type", "application/json")
	response, _ := httpClient.Do(request)
	m, _ := simplejson.NewFromReader(response.Body)

	if len(m.Get("errors").MustArray()) == 0 {
		groupsLayer := m.Get("data").Get("groups")
		for indexOfGroups := 0; indexOfGroups < len(groupsLayer.MustArray()); indexOfGroups++ {
			ID := groupsLayer.GetIndex(indexOfGroups).Get("_id").MustString()
			groupID := groupsLayer.GetIndex(indexOfGroups).Get("id").MustString()
			groupName := groupsLayer.GetIndex(indexOfGroups).Get("name").MustString()
			parentID := groupsLayer.GetIndex(indexOfGroups).Get("parentId").MustString()
			timeZone := groupsLayer.GetIndex(indexOfGroups).Get("timeZone").MustString()
			//fmt.Println("  GroupID:", groupID, " GroupName:", groupName, "ParentID:", parentID)
			var lastStatusRawValueResult map[string]interface{}
			grouptopologyCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}}).One(&lastStatusRawValueResult)
			if len(lastStatusRawValueResult) == 0 {
				grouptopologyCollection.Insert(&map[string]interface{}{"_id": ID, "GroupID": groupID, "GroupName": groupName, "ParentID": parentID, "TimeZone": timeZone})
			} else {
				grouptopologyCollection.Update(bson.M{"_id": lastStatusRawValueResult["_id"]}, bson.M{"$set": bson.M{"GroupID": groupID, "GroupName": groupName, "ParentID": parentID, "TimeZone": timeZone}})
			}
		}
	} else {
		taipeiTimeZone, _ := time.LoadLocation("Asia/Taipei")
		fmt.Println(time.Now().In(taipeiTimeZone), "=>  GraphQL Error ->", m.Get("errors").GetIndex(0).Get("message").MustString())
	}
}
