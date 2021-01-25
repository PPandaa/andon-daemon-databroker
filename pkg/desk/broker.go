package desk

import (
	"bytes"
	"databroker/config"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//TransmitData ...
func TransmitData(nowTime time.Time, db *mgo.Database) {
	machineRawDataCollection := db.C("iii.dae.MachineRawData")
	groupTopoCollection := db.C("iii.cfg.GroupTopology")
	var groupIDs []string
	//  ------------------------ GroupData
	//    ---------------------- GraphQl
	httpClient := &http.Client{}
	// httpRequestBody, _ := json.Marshal(map[string]string{
	// 	"query": "query groups {	groups {		id		name	}}",
	// })
	// request, _ := http.NewRequest("POST", "https://ifp-organizer-training-eks011.hz.wise-paas.com.cn/graphql", bytes.NewBuffer(httpRequestBody))
	// request.Header.Set("cookie", eiToken)
	// request.Header.Set("Content-Type", "application/json")
	// response, err := httpClient.Do(request)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// m, _ := simplejson.NewFromReader(response.Body)
	// for indexOfGroups := 0; indexOfGroups < len(m.Get("data").Get("groups").MustArray()); indexOfGroups++ {
	// 	groupIDs = append(groupIDs, m.Get("data").Get("groups").GetIndex(indexOfGroups).Get("id").MustString())
	// }
	var groupTopo []map[string]interface{}
	groupTopoCollection.Find(bson.M{}).All(&groupTopo)
	for _, groupV := range groupTopo {
		groupIDs = append(groupIDs, groupV["GroupID"].(string))
	}
	// groupIDs = []string{"R3JvdXA.X-0tJMYGAgAG-fkZ"} //,"R3JvdXA.X-LiKkwnAwAG1mqq"}
	// ------------------------------------------------------ MachineData
	// fmt.Println("-- GraphQL API Start", time.Now().In(taipeiTimeZone))
	httpRequestBody, _ := json.Marshal(map[string]interface{}{
		// "query":     "query bigbang($groupId: [ID!]!) {   groupsByIds(ids: $groupId) {     id     name     timeZone     machines {       id       name       parameters(first: 300) {         nodes{           id           name           lastValue{             num             mappingCode{               code               message               status{                 index                 layer1{                   index                   name                 }               }             }             time           }         }       }     }   } }",
		"query":     "query bigbang($groupId: [ID!]!) {   groupsByIds(ids: $groupId) {     id     name     timeZone     machines {       id       name       parameters(first: 10) {         nodes{           id           name           lastValue{             num             mappingCode{               code               message               status{                 index                 layer1{                   index                   name                 }               }             }             time           }         }       }parameterByName(name: \"status\") {           id           name           lastValue{             num             mappingCode{               code               message               status{                 index                 layer1{                   index                   name                 }               }             }             time           }                }     }   } }",
		"variables": map[string][]string{"groupId": groupIDs},
	})
	request, _ := http.NewRequest("POST", "https://ifp-organizer-tienkang-eks002.sa.wise-paas.com/graphql", bytes.NewBuffer(httpRequestBody))
	request.Header.Set("cookie", config.Token)
	request.Header.Set("Content-Type", "application/json")
	response, _ := httpClient.Do(request)
	// fmt.Println("-- GraphQL API End", time.Now().In(taipeiTimeZone))

	m, _ := simplejson.NewFromReader(response.Body)
	// fmt.Println(m)
	if len(m.Get("errors").MustArray()) == 0 {
		groupsLayer := m.Get("data").Get("groupsByIds")
		for indexOfGroups := 0; indexOfGroups < len(groupsLayer.MustArray()); indexOfGroups++ {
			groupID := groupsLayer.GetIndex(indexOfGroups).Get("id").MustString()
			groupName := groupsLayer.GetIndex(indexOfGroups).Get("name").MustString()
			// fmt.Println("GroupID:", groupID, " GroupName:", groupName)
			machinesLayer := groupsLayer.GetIndex(indexOfGroups).Get("machines")
			// fmt.Println(machinesLayer)
			for indexOfMachines := 0; indexOfMachines < len(machinesLayer.MustArray()); indexOfMachines++ {
				machineID := machinesLayer.GetIndex(indexOfMachines).Get("id").MustString()
				machineName := machinesLayer.GetIndex(indexOfMachines).Get("name").MustString()
				// fmt.Println("  MachineID:", machineID, " MachineName:", machineName)
				// paramaterLayer := machinesLayer.GetIndex(indexOfMachines).Get("parameters").Get("nodes")
				paramaterLayer := machinesLayer.GetIndex(indexOfMachines).Get("parameterByName")
				paraString := "    ParaName: "
				// fmt.Println(paramaterLayer)
				paraName := paramaterLayer.Get("name").MustString()
				paraUpdateTime := paramaterLayer.Get("lastValue").Get("time").MustString()
				timestampFS := strings.Replace(paraUpdateTime, "Z", "+00:00", 1)
				timestamp, _ := time.Parse(time.RFC3339, timestampFS)
				if paramaterLayer.Get("name").MustString() == "status" {
					var lastStatusRawValueResult map[string]interface{}
					machineRawDataCollection.Pipe([]bson.M{{"$match": bson.M{"MachineID": machineID}}, {"$sort": bson.M{"ts": -1}}}).One(&lastStatusRawValueResult)
					// fmt.Println(lastStatusRawValueResult)
					statusRawValue := paramaterLayer.Get("lastValue").Get("num").MustInt()
					statusMapValue := paramaterLayer.Get("lastValue").Get("mappingCode").Get("status").Get("index").MustInt()
					statusLay1Value := paramaterLayer.Get("lastValue").Get("mappingCode").Get("status").Get("layer1").Get("index").MustInt()
					paraString += paraName + "  StatusRawValue: " + strconv.Itoa(statusRawValue) + "  StatusMapValue: " + strconv.Itoa(statusMapValue) + "  StatusLay1Value: " + strconv.Itoa(statusLay1Value) + "  Timestamp: " + timestampFS + " | "
					//  DB Insert
					if len(lastStatusRawValueResult) == 0 {
						machineRawDataCollection.Insert(&map[string]interface{}{"Timestamp": timestamp, "GroupID": groupID, "GroupName": groupName, "MachineID": machineID, "MachineName": machineName, "StatusRawValue": statusRawValue, "StatusMapValue": statusMapValue, "StatusLay1Value": statusLay1Value, "ManualEvent": 0})
					} else {
						if lastStatusRawValueResult["Timestamp"] != timestamp {
							machineRawDataCollection.Update(bson.M{"_id": lastStatusRawValueResult["_id"]}, bson.M{"$set": bson.M{"Timestamp": timestamp, "GroupID": groupID, "GroupName": groupName, "MachineID": machineID, "MachineName": machineName, "StatusRawValue": statusRawValue, "StatusMapValue": statusMapValue, "StatusLay1Value": statusLay1Value}})
						}
					}
				} else {
					// paraValue := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("num").MustInt()
					// paraString += paraName + "  ParaValue: " + strconv.Itoa(paraValue) + "  Timestamp: " + timestampFS + " | "
				}
				// for indexOfParamater := 0; indexOfParamater < len(paramaterLayer.MustArray()); indexOfParamater++ {
				// 	paraName := paramaterLayer.GetIndex(indexOfParamater).Get("name").MustString()
				// 	paraUpdateTime := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("time").MustString()
				// 	timestampFS := strings.Replace(paraUpdateTime, "Z", "+00:00", 1)
				// 	timestamp, _ := time.Parse(time.RFC3339, timestampFS)
				// 	// if paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("mappingCode").Get("code").MustString() != "" {
				// 	if paramaterLayer.GetIndex(indexOfParamater).Get("name").MustString() == "status" {
				// 		var lastStatusRawValueResult map[string]interface{}
				// 		machineRawDataCollection.Pipe([]bson.M{{"$match": bson.M{"MachineID": machineID}}, {"$sort": bson.M{"ts": -1}}}).One(&lastStatusRawValueResult)
				// 		// fmt.Println(lastStatusRawValueResult)
				// 		statusRawValue := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("num").MustInt()
				// 		statusMapValue := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("mappingCode").Get("status").Get("index").MustInt()
				// 		statusLay1Value := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("mappingCode").Get("status").Get("layer1").Get("index").MustInt()
				// 		paraString += paraName + "  StatusRawValue: " + strconv.Itoa(statusRawValue) + "  StatusMapValue: " + strconv.Itoa(statusMapValue) + "  StatusLay1Value: " + strconv.Itoa(statusLay1Value) + "  Timestamp: " + timestampFS + " | "
				// 		//  DB Insert
				// 		if len(lastStatusRawValueResult) == 0 {
				// 			machineRawDataCollection.Insert(&map[string]interface{}{"Timestamp": timestamp, "GroupID": groupID, "GroupName": groupName, "MachineID": machineID, "MachineName": machineName, "StatusRawValue": statusRawValue, "StatusMapValue": statusMapValue, "StatusLay1Value": statusLay1Value})
				// 		} else {
				// 			if lastStatusRawValueResult["Timestamp"] != timestamp {
				// 				machineRawDataCollection.Update(bson.M{"_id": lastStatusRawValueResult["_id"]}, bson.M{"$set": bson.M{"Timestamp": timestamp, "GroupID": groupID, "GroupName": groupName, "MachineID": machineID, "MachineName": machineName, "StatusRawValue": statusRawValue, "StatusMapValue": statusMapValue, "StatusLay1Value": statusLay1Value}})
				// 			}
				// 		}
				// 	} else {
				// 		// paraValue := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("num").MustInt()
				// 		// paraString += paraName + "  ParaValue: " + strconv.Itoa(paraValue) + "  Timestamp: " + timestampFS + " | "
				// 	}
				// }
				// fmt.Println(paraString)
			}
		}
	} else {
		taipeiTimeZone, _ := time.LoadLocation("Asia/Taipei")
		fmt.Println(time.Now().In(taipeiTimeZone), "=>  GraphQL Error ->", m.Get("errors").GetIndex(0).Get("message").MustString())
	}
	// }
}
