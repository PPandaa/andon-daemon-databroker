package main

import (
	"bytes"
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

const (
	statusCodeParaName = "GYR"
)

//TransmitData ...
func TransmitData(eiToken string, nowTime time.Time, taipeiTimeZone *time.Location, mongodbURL string, mongodbUsername string, mongodbPassword string, mongodbDatabase string) {
	session, _ := mgo.Dial(mongodbURL)
	db := session.DB(mongodbDatabase)
	db.Login(mongodbUsername, mongodbPassword)
	machineRawDataCollection := db.C("iii.dae.MachineRawData")
	var groupIDs []string
	//  ------------------------ GroupData
	//    ---------------------- GraphQl
	httpClient := &http.Client{}
	httpRequestBody, _ := json.Marshal(map[string]string{
		"query": "query groups {	groups {		id		name	}}",
	})
	request, _ := http.NewRequest("POST", "https://ifp-organizer-training-eks011.hz.wise-paas.com.cn/graphql", bytes.NewBuffer(httpRequestBody))
	request.Header.Set("cookie", eiToken)
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		fmt.Println(err)
	} else {
		// m, _ := simplejson.NewFromReader(response.Body)
		// for indexOfGroups := 0; indexOfGroups < len(m.Get("data").Get("groups").MustArray()); indexOfGroups++ {
		// 	groupIDs = append(groupIDs, m.Get("data").Get("groups").GetIndex(indexOfGroups).Get("id").MustString())
		// }
		// fmt.Println(groupIDs)
		groupIDs = []string{"R3JvdXA.X-LiKkwnAwAG1mqq"} //,"R3JvdXA.X9MpbyaqrwAG2CZV"} //, "R3JvdXA.X9MfFCaqrwAG2CZP"}
		// ------------------------------------------------------ MachineData
		httpRequestBody, _ = json.Marshal(map[string]interface{}{
			"query":     "query bigbang($groupId: [ID!]!) {   groupsByIds(ids: $groupId) {     id     name     timeZone     machines {       id       name       parameters(first: 2) {         nodes{           id           name           lastValue{             num             mappingCode{               code               message               status{                 index                 layer1{                   index                   name                 }               }             }             time           }         }       }     }   } }",
			"variables": map[string][]string{"groupId": groupIDs},
		})
		request, _ = http.NewRequest("POST", "https://ifp-organizer-training-eks011.hz.wise-paas.com.cn/graphql", bytes.NewBuffer(httpRequestBody))

		request.Header.Set("cookie", eiToken)
		request.Header.Set("Content-Type", "application/json")
		response, err = httpClient.Do(request)
		if err != nil {
			fmt.Println(err)
		} else {
			m, _ := simplejson.NewFromReader(response.Body)
			// fmt.Println(m)
			groupsLayer := m.Get("data").Get("groupsByIds")
			for indexOfGroups := 0; indexOfGroups < len(groupsLayer.MustArray()); indexOfGroups++ {
				groupID := groupsLayer.GetIndex(indexOfGroups).Get("id").MustString()
				groupName := groupsLayer.GetIndex(indexOfGroups).Get("name").MustString()
				fmt.Println("GroupID:", groupID, " GroupName:", groupName)
				machinesLayer := groupsLayer.GetIndex(indexOfGroups).Get("machines")
				// fmt.Println(machinesLayer)
				for indexOfMachines := 0; indexOfMachines < len(machinesLayer.MustArray()); indexOfMachines++ {
					machineID := machinesLayer.GetIndex(indexOfMachines).Get("id").MustString()
					machineName := machinesLayer.GetIndex(indexOfMachines).Get("name").MustString()
					fmt.Println("  MachineID:", machineID, " MachineName:", machineName)
					paramaterLayer := machinesLayer.GetIndex(indexOfMachines).Get("parameters").Get("nodes")
					paraString := "    ParaName: "
					// fmt.Println(paramaterLayer)
					for indexOfParamater := 0; indexOfParamater < len(paramaterLayer.MustArray()); indexOfParamater++ {
						paraName := paramaterLayer.GetIndex(indexOfParamater).Get("name").MustString()
						paraUpdateTime := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("time").MustString()
						timestampFS := strings.Replace(paraUpdateTime, "Z", "+00:00", 1)
						timestamp, _ := time.Parse(time.RFC3339, timestampFS)
						if paraName == statusCodeParaName {
							var lastStatusRawValueResult map[string]interface{}
							machineRawDataCollection.Pipe([]bson.M{{"$match": bson.M{"MachineID": machineID}}, {"$sort": bson.M{"ts": -1}}}).One(&lastStatusRawValueResult)
							// fmt.Println(lastStatusRawValueResult)
							statusRawValue := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("num").MustInt()
							statusMapValue := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("mappingCode").Get("status").Get("index").MustInt()
							statusLay1Value := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("mappingCode").Get("status").Get("layer1").Get("index").MustInt()
							paraString += paraName + "  StatusRawValue: " + strconv.Itoa(statusRawValue) + "  StatusMapValue: " + strconv.Itoa(statusMapValue) + "  StatusLay1Value: " + strconv.Itoa(statusLay1Value) + "  Timestamp: " + timestampFS + " | "
							//  DB Insert
							if len(lastStatusRawValueResult) == 0 {
								machineRawDataCollection.Insert(&map[string]interface{}{"Timestamp": timestamp, "GroupID": groupID, "GroupName": groupName, "MachineID": machineID, "MachineName": machineName, "StatusRawValue": statusRawValue, "StatusMapValue": statusMapValue, "StatusLay1Value": statusLay1Value})
							} else {
								if lastStatusRawValueResult["Timestamp"] != timestamp {
									machineRawDataCollection.Update(bson.M{"_id": lastStatusRawValueResult["_id"]}, bson.M{"$set": bson.M{"Timestamp": timestamp, "GroupID": groupID, "GroupName": groupName, "MachineID": machineID, "MachineName": machineName, "StatusRawValue": statusRawValue, "StatusMapValue": statusMapValue, "StatusLay1Value": statusLay1Value}})
								}
							}
						} else {
							// paraValue := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("num").MustInt()
							// paraString += paraName + "  ParaValue: " + strconv.Itoa(paraValue) + "  Timestamp: " + timestampFS + " | "
						}
					}
					fmt.Println(paraString)
				}
			}
		}
	}
}
