package routers

import (
	"databroker/config"
	v1 "databroker/routers/api/v1"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(gin.Logger())

	r.Use(gin.Recovery())
	apiv1 := r.Group("/")
	{
		apiv1.GET("/", func(c *gin.Context) {
			if config.IFPStatus == "Up" {
				c.JSON(200, "This is Daemon-Databroker for iFactory/Andon")
			} else if config.IFPStatus == "Down" {
				c.JSON(500, "IFP Desk is not available")
			}
		})
		apiv1.POST("/iot-2/evt/waconn/fmt/:sourceId", v1.PostOutbound_waconn)
		apiv1.POST("/iot-2/evt/ifpcfg/fmt/:sourceId", v1.PostOutbound_ifpcfg)
		apiv1.POST("/iot-2/evt/wacfg/fmt/:sourceId", v1.PostOutbound_wacfg)
		apiv1.POST("/iot-2/evt/wadata/fmt/:sourceId", v1.PostOutbound_wadata)
	}

	return r
}
