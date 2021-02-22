package v1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"databroker/config"
	"databroker/db"
	"databroker/model"
	"databroker/pkg/desk"

	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"

	// . "github.com/logrusorgru/aurora"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//由於目前對方發送過來的post body內容無法預測規則性，因此收到後先全部存db

func PostOutbound_wadata(c *gin.Context) {
	// sourceId := c.Param("sourceId") //取得URL中参数
	// fmt.Println(BrightBlue("------------------wadata-------------------"), sourceId)

	body, _ := ioutil.ReadAll(c.Request.Body)
	// fmt.Println(BrightBlue(string(body)))

	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		glog.Error(err)
	}

	// 1/26 取消存入db
	// err := db.Insert(db.Wadata, v)
	// if err == nil {
	// 	glog.Info("---wadata inserted---")
	// }

	//------------

	// 1/25新增
	var wadata model.Wadata
	if err := json.Unmarshal(body, &wadata); err != nil {
		glog.Error(err)
	}
	// fmt.Printf("%+v", wadata)
	wadata.Service("status")
}

func PostOutbound_waconn(c *gin.Context) {
	// sourceId := c.Param("sourceId") //取得URL中参数
	// fmt.Println(BrightBlue("------------------waconn-------------------"), sourceId)

	body, _ := ioutil.ReadAll(c.Request.Body)
	// fmt.Println(BrightBlue(string(body)))

	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		glog.Error(err)
	}

	// err := db.Insert(db.Waconn, v)
	// if err == nil {
	// 	glog.Info("---waconn inserted---")
	// }
}

func PostOutbound_ifpcfg(c *gin.Context) {
	// sourceId := c.Param("sourceId") //取得URL中参数
	// fmt.Println(BrightBlue("------------------ifpcfg-------------------"), sourceId)

	body, _ := ioutil.ReadAll(c.Request.Body)
	// fmt.Println(BrightBlue(string(body)))

	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		glog.Error(err)
	}

	err := db.Insert(db.Ifpcfg, v)
	if err == nil {
		glog.Info("---ifpcg inserted---")
	}

	simpleJsonBody, _ := simplejson.NewJson(body)

	// remove
	if simpleJsonBody.Get("group").Get("removed") != nil {
		for i := 0; i < len(simpleJsonBody.Get("group").Get("removed").MustArray()); i++ {
			removeid := simpleJsonBody.Get("group").Get("removed").GetIndex(i).Get("id").MustString()
			msg := bson.M{"_id": removeid}
			db.Delete(db.Topo, msg)
		}
	}
	if simpleJsonBody.Get("machine").Get("removed") != nil {
		for i := 0; i < len(simpleJsonBody.Get("machine").Get("removed").MustArray()); i++ {
			removeid := simpleJsonBody.Get("machine").Get("removed").GetIndex(i).Get("id").MustString()
			msg := bson.M{"_id": removeid}
			db.Delete(db.MachineRawData, msg)
		}
	}
	if simpleJsonBody.Get("parameter").Get("removed") != nil {
		for i := 0; i < len(simpleJsonBody.Get("parameter").Get("removed").MustArray()); i++ {
			StatusID := simpleJsonBody.Get("parameter").Get("removed").GetIndex(i).Get("id").MustString()
			query := bson.M{"StatusID": StatusID}
			value := bson.M{"$set": bson.M{
				"StatusRawValue":  nil,
				"StatusLay1Value": nil,
				"StatusMapValue":  nil,
			}}
			db.Upadte(db.MachineRawData, query, value)
		}
	}

	// insert or update
	if simpleJsonBody.Get("group").Get("items") != nil {
		if len(simpleJsonBody.Get("machine").Get("items").MustArray()) == 0 {
			if len(simpleJsonBody.Get("parameter").Get("items").MustArray()) == 0 {
				fmt.Println("insert&upadte=>  Topo Activation")
				session, _ := mgo.Dial(config.MongodbURL)
				db := session.DB(config.MongodbDatabase)
				db.Login(config.MongodbUsername, config.MongodbPassword)
				desk.GetTopology(db)
			}
		} else {
			// 如果 machine items > 0
			// 拿到 simpleJsonBody group _id (可能有多個)
			for i := 0; i < len(simpleJsonBody.Get("group").Get("items").MustArray()); i++ {
				groupid := simpleJsonBody.Get("group").Get("items").GetIndex(i).Get("id").MustString()
				//  將 groupid 傳給 peter borker func 去打GQL 拿該GROUP 底下的所有資料
				desk.MachineRawDataTable("group", groupid)
			}

		}
	}

	// method1: use gjson to get the field you want
	// gjson.GetBytes(req, "factoryId")

	// method2: convert json to struct, 但目前不確定完整格式可能會產生 Bug EOF
	// var ifpcfg model.Ifpcfg
	// if err := c.ShouldBind(&ifpcfg); err != nil {
	// 	fmt.Println("PostOutbound_ifpcfg err:", err)
	// 	// c.String(http.StatusOK, `the body should be formA`)
	// }
	// fmt.Println(BrightCyan(ifpcfg))
}

func PostOutbound_wacfg(c *gin.Context) {
	// sourceId := c.Param("sourceId") //取得URL中参数
	// fmt.Println(BrightBlue("------------------wacfg-------------------"), sourceId)

	body, _ := ioutil.ReadAll(c.Request.Body)
	// fmt.Println(BrightBlue(string(body)))

	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		glog.Error(err)
	}

	// err := db.Insert(db.Wacfg, v)
	// if err == nil {
	// 	glog.Info("---wacfg inserted---")
	// }
}

func JustForTest() {
	b := GetOutboundSample()
	var ifpcfg model.Ifpcfg
	err := json.Unmarshal(b, &ifpcfg)
	if err != nil {
		log.Println(err)
	}
	// fmt.Printf("%+v", ifpcfg)

	groupItems := ifpcfg.Group.Items
	// machineItems := ifpcfg.Machine
	// parameterItems := ifpcfg.Parameter

	//method1
	//unmarshal裡面寫匿名struct, 在把mapp到struct的值給去 &point
	//method2
	//tag加上omitempty

	for _, i := range groupItems {
		fmt.Printf("%+v", i)
	}
}
