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

//UpdateMachineStatus ...
func UpdateMachineStatus(StatusID string) {
	fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
	fmt.Println("UpdateMachineRaw  =>  StatusID:", StatusID)
	var machineIDs []string
	httpClient := &http.Client{}
	session, _ := mgo.Dial(config.MongodbURL)
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
	machineRawDataCollection := db.C("iii.dae.MachineRawData")
	// startTime := time.Now().In(config.TaipeiTimeZone)

	var machineRawDataResult map[string]interface{}
	machineRawDataCollection.Find(bson.M{"StatusID": StatusID}).One(&machineRawDataResult)
	if len(machineRawDataResult) != 0 {
		fmt.Println("StatusID:", StatusID, "-> MachineID:", machineRawDataResult["MachineID"])
		machineIDs = append(machineIDs, machineRawDataResult["MachineID"].(string))

		httpRequestBody, _ := json.Marshal(map[string]interface{}{
			"query":     "query bigbang($machineID: [ID!]!) {   machinesByIds(ids:$machineID){     _id     id     name     parameterByName(name:\"status\"){       id       name       lastValue{         num         mappingCode{           code           message           status{             index             layer1{               index               name             }           }         }         time       }     }   } }",
			"variables": map[string][]string{"machineID": machineIDs},
		})
		request, _ := http.NewRequest("POST", config.IFPURL, bytes.NewBuffer(httpRequestBody))
		request.Header.Set("cookie", config.Token)
		request.Header.Set("Content-Type", "application/json")
		response, _ := httpClient.Do(request)
		m, _ := simplejson.NewFromReader(response.Body)
		// fmt.Println(m)

		machinesLayer := m.Get("data").Get("machinesByIds")
		if len(m.Get("errors").MustArray()) == 0 {
			for indexOfMachines := 0; indexOfMachines < len(machinesLayer.MustArray()); indexOfMachines++ {
				machineUnderID := machinesLayer.GetIndex(indexOfMachines).Get("_id").MustString()
				machineID := machinesLayer.GetIndex(indexOfMachines).Get("id").MustString()
				machineName := machinesLayer.GetIndex(indexOfMachines).Get("name").MustString()
				if machineRawDataResult["_id"].(string) == machineUnderID {
					fmt.Println("  MachineID:", machineID, " MachineName:", machineName)
					paraString := "    ParaName: "
					paramaterLayer := machinesLayer.GetIndex(indexOfMachines).Get("parameterByName")
					// fmt.Println(paramaterLayer)
					paraName := paramaterLayer.Get("name").MustString()
					paraUpdateTime := paramaterLayer.Get("lastValue").Get("time").MustString()
					timestampFS := strings.Replace(paraUpdateTime, "Z", "+00:00", 1)
					timestamp, _ := time.Parse(time.RFC3339, timestampFS)
					if paramaterLayer.Get("name").MustString() == "status" {
						paramaterID := paramaterLayer.Get("_id").MustString()
						statusRawValue := paramaterLayer.Get("lastValue").Get("num").MustInt()
						statusMapValue := paramaterLayer.Get("lastValue").Get("mappingCode").Get("status").Get("index").MustInt()
						statusLay1Value := paramaterLayer.Get("lastValue").Get("mappingCode").Get("status").Get("layer1").Get("index").MustInt()
						paraString += paraName + "  StatusRawValue: " + strconv.Itoa(statusRawValue) + "  StatusMapValue: " + strconv.Itoa(statusMapValue) + "  StatusLay1Value: " + strconv.Itoa(statusLay1Value) + "  Timestamp: " + timestampFS + " | "
						machineRawDataCollection.Update(bson.M{"_id": machineRawDataResult["_id"]}, bson.M{"$set": bson.M{"Timestamp": timestamp, "MachineID": machineID, "MachineName": machineName, "StatusRawValue": statusRawValue, "StatusMapValue": statusMapValue, "StatusLay1Value": statusLay1Value, "ParamaterID": paramaterID}})
					}
					fmt.Println(paraString)
				} else {
					// paraValue := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("num").MustInt()
					// paraString += paraName + "  ParaValue: " + strconv.Itoa(paraValue) + "  Timestamp: " + timestampFS + " | "
				}
			}

			// paramaterLayer := machinesLayer.GetIndex(indexOfMachines).Get("parameters").Get("nodes")
			// for indexOfParamater := 0; indexOfParamater < len(paramaterLayer.MustArray()); indexOfParamater++ {
			// 	paraName := paramaterLayer.GetIndex(indexOfParamater).Get("name").MustString()
			// 	paraUpdateTime := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("time").MustString()
			// 	timestampFS := strings.Replace(paraUpdateTime, "Z", "+00:00", 1)
			// 	timestamp, _ := time.Parse(time.RFC3339, timestampFS)
			// 	if paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("mappingCode").Get("code").MustString() != "" {
			// 	if paramaterLayer.GetIndex(indexOfParamater).Get("name").MustString() == "status" {
			// 		var lastStatusRawValueResult map[string]interface{}
			// 		machineRawDataCollection.Pipe([]bson.M{{"$match": bson.M{"MachineID": machineID}}, {"$sort": bson.M{"ts": -1}}}).One(&lastStatusRawValueResult)
			// 		fmt.Println(lastStatusRawValueResult)
			// 		statusRawValue := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("num").MustInt()
			// 		statusMapValue := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("mappingCode").Get("status").Get("index").MustInt()
			// 		statusLay1Value := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("mappingCode").Get("status").Get("layer1").Get("index").MustInt()
			// 		paraString += paraName + "  StatusRawValue: " + strconv.Itoa(statusRawValue) + "  StatusMapValue: " + strconv.Itoa(statusMapValue) + "  StatusLay1Value: " + strconv.Itoa(statusLay1Value) + "  Timestamp: " + timestampFS + " | "
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
			// endtime := time.Now().In(config.TaipeiTimeZone)
			// exectime := endtime.Sub(startTime)
			// fmt.Printf("%s =>  UpdateMachineRaw ->  %.1f Sec\n", time.Now().In(config.TaipeiTimeZone), exectime.Seconds())
		} else {
			taipeiTimeZone, _ := time.LoadLocation("Asia/Taipei")
			fmt.Println(time.Now().In(taipeiTimeZone), "=>  UpdateMachineRaw ->  GraphQL Error ->", m.Get("errors").GetIndex(0).Get("message").MustString())
		}
	}
	session.Close()
}
