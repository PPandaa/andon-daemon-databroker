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
	"gopkg.in/mgo.v2/bson"
)

//MachineRawDataTable ...
func MachineRawDataTable(mode string, groupUnderID ...string) {
	config.DbHealthCheck()
	fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
	fmt.Println("MachineRawDataTable => Mode:", mode, "GroupUnderID:", groupUnderID)
	var groupIDs []string
	httpClient := &http.Client{}

	groupTopoCollection := config.DB.C(config.GroupTopo)
	machineRawDataCollection := config.DB.C(config.MachineRawData)
	machineRawDataHistCollection := config.DB.C(config.MachineRawDataHist)

	if mode == "init" {
		var groupTopoResults []map[string]interface{}
		groupTopoCollection.Find(bson.M{}).All(&groupTopoResults)
		if len(groupTopoResults) != 0 {
			for _, groupTopoResult := range groupTopoResults {
				groupIDs = append(groupIDs, groupTopoResult["GroupID"].(string))
			}
		}
	} else {
		var groupTopoResults map[string]interface{}
		groupTopoCollection.Find(bson.M{"_id": groupUnderID[0]}).One(&groupTopoResults)
		groupIDs = append(groupIDs, groupTopoResults["GroupID"].(string))
	}
	// fmt.Println(groupIDs)

	httpRequestBody, _ := json.Marshal(map[string]interface{}{
		"query":     "query ($groupId: [ID!]!) {   groupsByIds(ids: $groupId) {     id     _id     name     timeZone     machines {       _id       id       name     }   } }",
		"variables": map[string][]string{"groupId": groupIDs},
	})
	request, _ := http.NewRequest("POST", config.IFPURL, bytes.NewBuffer(httpRequestBody))
	request.Header.Set("cookie", config.Token)
	request.Header.Set("Content-Type", "application/json")
	response, _ := httpClient.Do(request)
	m, _ := simplejson.NewFromReader(response.Body)
	// fmt.Println(m)

	if len(m.Get("errors").MustArray()) == 0 {
		groupsLayer := m.Get("data").Get("groupsByIds")
		for indexOfGroups := 0; indexOfGroups < len(groupsLayer.MustArray()); indexOfGroups++ {
			groupID := groupsLayer.GetIndex(indexOfGroups).Get("id").MustString()
			groupName := groupsLayer.GetIndex(indexOfGroups).Get("name").MustString()
			fmt.Println("GroupID:", groupID, " GroupName:", groupName)
			machinesLayer := groupsLayer.GetIndex(indexOfGroups).Get("machines")
			// fmt.Println(machinesLayer)
			for indexOfMachines := 0; indexOfMachines < len(machinesLayer.MustArray()); indexOfMachines++ {
				machineUnderID := machinesLayer.GetIndex(indexOfMachines).Get("_id").MustString()
				machineID := machinesLayer.GetIndex(indexOfMachines).Get("id").MustString()
				machineName := machinesLayer.GetIndex(indexOfMachines).Get("name").MustString()
				fmt.Println("  MachineUnderID:", machineUnderID, "  MachineID:", machineID, " MachineName:", machineName)

				var machineRawDataResult map[string]interface{}
				machineRawDataCollection.Find(bson.M{"MachineID": machineID}).One(&machineRawDataResult)

				machineStatusRequestBody, _ := json.Marshal(map[string]interface{}{
					"query":     "query ($machineId: ID!) {   machine(id: $machineId) {     id     _id     name     parameterByName(name:\"status\"){       _id       id       name       lastValue{         num         ... on TagValue {          mappingCode {           code           message           status{             index             layer1{               index               name             }           }         }         }         time       }     }   } }",
					"variables": map[string]string{"machineId": machineID},
				})
				machineStatusRequest, _ := http.NewRequest("POST", config.IFPURL, bytes.NewBuffer(machineStatusRequestBody))
				machineStatusRequest.Header.Set("cookie", config.Token)
				machineStatusRequest.Header.Set("Content-Type", "application/json")
				machineStatusResponse, _ := httpClient.Do(machineStatusRequest)
				machineLayerWithStatus, _ := simplejson.NewFromReader(machineStatusResponse.Body)
				// fmt.Println(machineLayerWithStatus)

				if len(machineLayerWithStatus.Get("errors").MustArray()) == 0 {
					paraString := "    ParaName: "
					paramaterLayer := machineLayerWithStatus.Get("data").Get("machine").Get("parameterByName")
					// fmt.Println(paramaterLayer)
					paraName := paramaterLayer.Get("name").MustString()
					paraUpdateTime := paramaterLayer.Get("lastValue").Get("time").MustString()
					timestampFS := strings.Replace(paraUpdateTime, "Z", "+00:00", 1)
					timestamp, _ := time.Parse(time.RFC3339, timestampFS)
					statusID := paramaterLayer.Get("_id").MustString()
					statusRawValue := paramaterLayer.Get("lastValue").Get("num").MustInt()
					statusMapValue := paramaterLayer.Get("lastValue").Get("mappingCode").Get("status").Get("index").MustInt()
					statusLay1Value := paramaterLayer.Get("lastValue").Get("mappingCode").Get("status").Get("layer1").Get("index").MustInt()
					paraString += paraName + "  StatusID: " + statusID + "  StatusRawValue: " + strconv.Itoa(statusRawValue) + "  StatusMapValue: " + strconv.Itoa(statusMapValue) + "  StatusLay1Value: " + strconv.Itoa(statusLay1Value) + "  Timestamp: " + timestampFS + " | "

					if len(machineRawDataResult) != 0 {
						machineRawDataCollection.Update(bson.M{"_id": machineRawDataResult["_id"]}, bson.M{"$set": bson.M{"GroupID": groupID, "GroupName": groupName, "MachineID": machineID, "MachineName": machineName, "StatusID": statusID, "StatusRawValue": statusRawValue, "StatusMapValue": statusMapValue, "StatusLay1Value": statusLay1Value, "Timestamp": timestamp}})
					} else {
						machineRawDataCollection.Insert(&map[string]interface{}{"_id": machineUnderID, "GroupID": groupID, "GroupName": groupName, "MachineID": machineID, "MachineName": machineName, "StatusID": statusID, "StatusRawValue": statusRawValue, "StatusMapValue": statusMapValue, "StatusLay1Value": statusLay1Value, "ManualEvent": 0, "Timestamp": timestamp})
					}

					var machineRawDataHistResult map[string]interface{}
					machineRawDataHistCollection.Find(bson.M{"MachineID": machineID}).Sort("-Timestamp").One(&machineRawDataHistResult)
					if len(machineRawDataHistResult) != 0 {
						if machineRawDataHistResult["StatusLay1Value"].(int) != statusLay1Value {
							machineRawDataHistCollection.Insert(&map[string]interface{}{"GroupID": groupID, "GroupName": groupName, "MachineID": machineID, "MachineName": machineName, "StatusRawValue": statusRawValue, "StatusMapValue": statusMapValue, "StatusLay1Value": statusLay1Value, "Timestamp": timestamp})
						}
					} else {
						machineRawDataHistCollection.Insert(&map[string]interface{}{"GroupID": groupID, "GroupName": groupName, "MachineID": machineID, "MachineName": machineName, "StatusRawValue": statusRawValue, "StatusMapValue": statusMapValue, "StatusLay1Value": statusLay1Value, "Timestamp": timestamp})
					}

					fmt.Println(paraString)
				} else {
					if len(machineRawDataResult) != 0 {
						machineRawDataCollection.Update(bson.M{"_id": machineRawDataResult["_id"]}, bson.M{"$set": bson.M{"GroupID": groupID, "GroupName": groupName, "MachineID": machineID, "MachineName": machineName}})
					} else {
						machineRawDataCollection.Insert(&map[string]interface{}{"_id": machineUnderID, "GroupID": groupID, "GroupName": groupName, "MachineID": machineID, "MachineName": machineName, "StatusID": nil, "StatusRawValue": nil, "StatusMapValue": nil, "StatusLay1Value": nil, "ManualEvent": 0, "Timestamp": nil})
					}
				}
			}
		}
	} else {
		fmt.Println(time.Now().In(config.TaipeiTimeZone), "=>  MachineRawDataTable ->  GraphQL Error ->", m.Get("errors").GetIndex(0).Get("message").MustString())
	}
}

//StationRawDataTable
func StationRawDataTable(mode string, groupUnderID ...string) {
	config.DbHealthCheck()
	fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
	fmt.Println("StationRawDataTable => Mode:", mode, "GroupUnderID:", groupUnderID)
	var groupIDs []string
	httpClient := &http.Client{}

	groupTopoCollection := config.DB.C(config.GroupTopo)
	stationRawDataCollection := config.DB.C(config.StationRawData)

	if mode == "init" {
		var groupTopoResults []map[string]interface{}
		groupTopoCollection.Find(bson.M{}).All(&groupTopoResults)
		if len(groupTopoResults) != 0 {
			for _, groupTopoResult := range groupTopoResults {
				groupIDs = append(groupIDs, groupTopoResult["GroupID"].(string))
			}
		}
	} else {
		var groupTopoResults map[string]interface{}
		groupTopoCollection.Find(bson.M{"_id": groupUnderID[0]}).One(&groupTopoResults)
		groupIDs = append(groupIDs, groupTopoResults["GroupID"].(string))
	}
	// fmt.Println(groupIDs)

	httpRequestBody, _ := json.Marshal(map[string]interface{}{
		"query":     "query ($groupId: [ID!]!) {   groupsByIds(ids: $groupId) {     id     _id     name     timeZone     machines(isStation: true) {       _id       id       name     }   } }",
		"variables": map[string][]string{"groupId": groupIDs},
	})
	request, _ := http.NewRequest("POST", config.IFPURL, bytes.NewBuffer(httpRequestBody))
	request.Header.Set("cookie", config.Token)
	request.Header.Set("Content-Type", "application/json")
	response, _ := httpClient.Do(request)
	m, _ := simplejson.NewFromReader(response.Body)
	// fmt.Println(m)

	if len(m.Get("errors").MustArray()) == 0 {
		groupsLayer := m.Get("data").Get("groupsByIds")
		for indexOfGroups := 0; indexOfGroups < len(groupsLayer.MustArray()); indexOfGroups++ {
			groupID := groupsLayer.GetIndex(indexOfGroups).Get("id").MustString()
			groupName := groupsLayer.GetIndex(indexOfGroups).Get("name").MustString()
			fmt.Println("GroupID:", groupID, " GroupName:", groupName)
			machinesLayer := groupsLayer.GetIndex(indexOfGroups).Get("machines")
			// fmt.Println(machinesLayer)
			for indexOfMachines := 0; indexOfMachines < len(machinesLayer.MustArray()); indexOfMachines++ {
				stationUnderID := machinesLayer.GetIndex(indexOfMachines).Get("_id").MustString()
				stationID := machinesLayer.GetIndex(indexOfMachines).Get("id").MustString()
				stationName := machinesLayer.GetIndex(indexOfMachines).Get("name").MustString()
				fmt.Println("  StationUnderID:", stationUnderID, "  StationID:", stationID, " StationName:", stationName)

				var stationRawDataResult map[string]interface{}
				stationRawDataCollection.Find(bson.M{"StationID": stationID}).One(&stationRawDataResult)
				if len(stationRawDataResult) != 0 {
					stationRawDataCollection.Update(bson.M{"_id": stationRawDataResult["_id"]}, bson.M{"$set": bson.M{"GroupID": groupID, "GroupName": groupName, "StationID": stationID, "StationName": stationName}})
				} else {
					stationRawDataCollection.Insert(&map[string]interface{}{"_id": stationUnderID, "GroupID": groupID, "GroupName": groupName, "StationID": stationID, "StationName": stationName, "ManualEvent": 0})
				}
			}
		}
	} else {
		fmt.Println(time.Now().In(config.TaipeiTimeZone), "=>  StationRawDataTable ->  GraphQL Error ->", m.Get("errors").GetIndex(0).Get("message").MustString())
	}
}

//UpdateMachineStatus ...
func UpdateMachineStatus(StatusID string) {
	config.DbHealthCheck()
	fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
	fmt.Println("UpdateMachineRaw  =>  StatusID:", StatusID)
	var machineIDs []string
	httpClient := &http.Client{}

	machineRawDataCollection := config.DB.C(config.MachineRawData)
	machineRawDataHistCollection := config.DB.C(config.MachineRawDataHist)
	// startTime := time.Now().In(config.TaipeiTimeZone)

	var machineRawDataResult map[string]interface{}
	machineRawDataCollection.Find(bson.M{"StatusID": StatusID}).One(&machineRawDataResult)
	if len(machineRawDataResult) != 0 {
		fmt.Println("StatusID:", StatusID, "-> MachineID:", machineRawDataResult["MachineID"])
		machineIDs = append(machineIDs, machineRawDataResult["MachineID"].(string))

		httpRequestBody, _ := json.Marshal(map[string]interface{}{
			"query":     "query bigbang($machineId: [ID!]!) {   machinesByIds(ids: $machineId) {     id     _id     name     parameterByName(name:\"status\"){       _id       id       name       lastValue{         num         ... on TagValue {          mappingCode {           code           message           status{             index             layer1{               index               name             }           }         }         }         time       }     }   } }",
			"variables": map[string][]string{"machineId": machineIDs},
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
						statusRawValue := paramaterLayer.Get("lastValue").Get("num").MustInt()
						statusMapValue := paramaterLayer.Get("lastValue").Get("mappingCode").Get("status").Get("index").MustInt()
						statusLay1Value := paramaterLayer.Get("lastValue").Get("mappingCode").Get("status").Get("layer1").Get("index").MustInt()
						paraString += paraName + "  StatusRawValue: " + strconv.Itoa(statusRawValue) + "  StatusMapValue: " + strconv.Itoa(statusMapValue) + "  StatusLay1Value: " + strconv.Itoa(statusLay1Value) + "  Timestamp: " + timestampFS + " | "
						machineRawDataCollection.Update(bson.M{"_id": machineRawDataResult["_id"]}, bson.M{"$set": bson.M{"Timestamp": timestamp, "StatusRawValue": statusRawValue, "StatusMapValue": statusMapValue, "StatusLay1Value": statusLay1Value}})
						machineRawDataHistCollection.Insert(&map[string]interface{}{"GroupID": machineRawDataResult["GroupID"], "GroupName": machineRawDataResult["GroupName"], "MachineID": machineRawDataResult["MachineID"], "MachineName": machineRawDataResult["MachineName"], "StatusRawValue": statusRawValue, "StatusMapValue": statusMapValue, "StatusLay1Value": statusLay1Value, "Timestamp": timestamp})
					}
					fmt.Println(paraString)
				} else {
					// paraValue := paramaterLayer.GetIndex(indexOfParamater).Get("lastValue").Get("num").MustInt()
					// paraString += paraName + "  ParaValue: " + strconv.Itoa(paraValue) + "  Timestamp: " + timestampFS + " | "
				}
			}
			// endtime := time.Now().In(config.TaipeiTimeZone)
			// exectime := endtime.Sub(startTime)
			// fmt.Printf("%s =>  UpdateMachineRaw ->  %.1f Sec\n", time.Now().In(config.TaipeiTimeZone), exectime.Seconds())
		} else {
			fmt.Println(time.Now().In(config.TaipeiTimeZone), "=>  UpdateMachineRaw ->  GraphQL Error ->", m.Get("errors").GetIndex(0).Get("message").MustString())
		}
	}
}
