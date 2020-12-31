package v1

import (
	"encoding/json"
	"fmt"
	"log"

	"databroker/model"

	"github.com/gin-gonic/gin"
)

func Test() {
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
func PostOutbound_waconn(c *gin.Context) {
	sourceId := c.Param("sourceId") //取得URL中参数
	fmt.Println("waconn:", sourceId)

}

func PostOutbound_ifpcfg(c *gin.Context) {
	sourceId := c.Param("sourceId") //取得URL中参数
	fmt.Println("ifpcfg", sourceId)

	var ifpcfg model.Ifpcfg
	if err := c.ShouldBind(&ifpcfg); err != nil {
		fmt.Println("PostOutbound_ifpcfg err:", err)
		// c.String(http.StatusOK, `the body should be formA`)
	}
}

func PostOutbound_wacfg(c *gin.Context) {
	sourceId := c.Param("sourceId") //取得URL中参数
	fmt.Println("wacfg:", sourceId)

}
