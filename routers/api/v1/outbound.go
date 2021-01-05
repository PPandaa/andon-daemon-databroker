package v1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"databroker/db"
	"databroker/model"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	. "github.com/logrusorgru/aurora"
)

//由於目前對方發送過來的post body內容無法預測規則性，因此收到後先全部存db

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
