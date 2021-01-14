package v1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"databroker/db"
	"databroker/model"
	"databroker/pkg/desk"

	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	. "github.com/logrusorgru/aurora"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//由於目前對方發送過來的post body內容無法預測規則性，因此收到後先全部存db

func PostOutbound_wadata(c *gin.Context) {
	sourceId := c.Param("sourceId") //取得URL中参数
	fmt.Println(BrightBlue("------------------wadata-------------------"), sourceId)

	body, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println(BrightBlue(string(body)))

	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		glog.Error(err)
	}

	err := db.Insert(db.Wadata, v)
	if err == nil {
		glog.Info("---wadata inserted---")
	}
}

func PostOutbound_waconn(c *gin.Context) {
	sourceId := c.Param("sourceId") //取得URL中参数
	fmt.Println(BrightBlue("------------------waconn-------------------"), sourceId)

	body, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println(BrightBlue(string(body)))

	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		glog.Error(err)
	}

	err := db.Insert(db.Waconn, v)
	if err == nil {
		glog.Info("---waconn inserted---")
	}
}

func PostOutbound_ifpcfg(c *gin.Context) {
	sourceId := c.Param("sourceId") //取得URL中参数
	fmt.Println(BrightBlue("------------------ifpcfg-------------------"), sourceId)

	body, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println(BrightBlue(string(body)))

	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		glog.Error(err)
	}

	err := db.Insert(db.Ifpcfg, v)
	if err == nil {
		glog.Info("---ifpcg inserted---")
	}

	simpleJsonBody, _ := simplejson.NewJson(body)
	if simpleJsonBody.Get("group").Get("removed") != nil {
		for i := 0; i < len(simpleJsonBody.Get("group").Get("removed").MustArray()); i++ {
			removeid := simpleJsonBody.Get("group").Get("removed").GetIndex(i).Get("id").MustString()
			msg := bson.M{"_id": removeid}
			db.Delete(db.Topo, msg)
		}
	}
	if simpleJsonBody.Get("group").Get("items") != nil {
		fmt.Println("insert&upadte=>  Topo Activation")
		session, _ := mgo.Dial(os.Getenv("MONGODB_URL"))
		db := session.DB(os.Getenv("MONGODB_DATABASE"))
		db.Login(os.Getenv("MONGODB_USERNAME"), os.Getenv("MONGODB_PASSWORD"))
		desk.GetTopology(db)
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
	sourceId := c.Param("sourceId") //取得URL中参数
	fmt.Println(BrightBlue("------------------wacfg-------------------"), sourceId)

	body, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println(BrightBlue(string(body)))

	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		glog.Error(err)
	}

	err := db.Insert(db.Wacfg, v)
	if err == nil {
		glog.Info("---wacfg inserted---")
	}
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
