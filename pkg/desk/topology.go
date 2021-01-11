package desk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bitly/go-simplejson"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// GetTopology initial 時 拿一次 Enabler  Group Topology
func GetTopology(token string, db *mgo.Database) {
	//fmt.Println(token, "=>  Topo Activation")
	grouptopologyCollection := db.C("iii.cfg.GroupTopology")

	httpClient := &http.Client{}

	httpRequestBody, _ := json.Marshal(map[string]interface{}{
		"query": "query groupsWithInboundConnector {   groups {     id     name     parentId     inboundConnector {       id       __typename        }             __typename   } }",
	})
	request, _ := http.NewRequest("POST", "https://ifp-organizer-training-eks011.hz.wise-paas.com.cn/graphql", bytes.NewBuffer(httpRequestBody))
	request.Header.Set("cookie", token)
	request.Header.Set("Content-Type", "application/json")
	response, _ := httpClient.Do(request)
	m, _ := simplejson.NewFromReader(response.Body)

	if len(m.Get("errors").MustArray()) == 0 {
		groupsLayer := m.Get("data").Get("groups")
		for indexOfGroups := 0; indexOfGroups < len(groupsLayer.MustArray()); indexOfGroups++ {
			groupID := groupsLayer.GetIndex(indexOfGroups).Get("id").MustString()
			groupName := groupsLayer.GetIndex(indexOfGroups).Get("name").MustString()
			parentID := groupsLayer.GetIndex(indexOfGroups).Get("parentId").MustString()
			fmt.Println("  GroupID:", groupID, " GroupName:", groupName, "ParentID:", parentID)
			var lastStatusRawValueResult map[string]interface{}
			grouptopologyCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}}).One(&lastStatusRawValueResult)
			if len(lastStatusRawValueResult) == 0 {
				grouptopologyCollection.Insert(&map[string]interface{}{"GroupID": groupID, "GroupName": groupName, "ParentID": parentID})
			}
		}
	} else {
		taipeiTimeZone, _ := time.LoadLocation("Asia/Taipei")
		fmt.Println(time.Now().In(taipeiTimeZone), "=>  GraphQL Error ->", m.Get("errors").GetIndex(0).Get("message").MustString())
	}
}
