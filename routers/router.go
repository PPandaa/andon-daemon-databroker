package routers

import (
	v1 "databroker/routers/api/v1"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())

	r.Use(gin.Recovery())

	// gin.SetMode(setting.RunMode)

	//----------------->
	apiv1 := r.Group("/")
	{
		apiv1.GET("/", func(c *gin.Context) {
			c.JSON(200, "This is Daemon-Databroker for iFactory/Andon")
		})
		apiv1.POST("/iot-2/evt/waconn/fmt/:sourceId", v1.PostOutbound_waconn)
		apiv1.POST("/iot-2/evt/ifpcfg/fmt/:sourceId", v1.PostOutbound_ifpcfg)
		apiv1.POST("/iot-2/evt/wacfg/fmt/:sourceId", v1.PostOutbound_wacfg)
		apiv1.POST("/iot-2/evt/wadata/fmt/:sourceId", v1.PostOutbound_wadata)

	}

	return r
}
