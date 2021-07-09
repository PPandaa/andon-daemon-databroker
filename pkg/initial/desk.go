package initial

import (
	"bytes"
	"databroker/config"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitly/go-simplejson"
)

func RegisterCommandCenter(regName string) {
	isAppRegisterCommandCenter := false
	httpClient := &http.Client{}
	httpRequestBody, _ := json.Marshal(map[string]interface{}{
		"query": "query { iApps { name } }",
	})
	request, _ := http.NewRequest("POST", config.IFP_DESK_API_URL.String(), bytes.NewBuffer(httpRequestBody))
	request.Header.Set("Content-Type", "application/json")
	response, _ := httpClient.Do(request)
	m, _ := simplejson.NewFromReader(response.Body)
	for _, appName := range m.Get("data").Get("iApps").MustArray() {
		if appName.(map[string]interface{})["name"] == regName {
			isAppRegisterCommandCenter = true
		}
	}
	if isAppRegisterCommandCenter {
		fmt.Println("      iApp", regName, "is already exist")
	} else {
		var content map[string]interface{}
		if config.ServerLocation == "Cloud" {
			content = map[string]interface{}{"name": regName, "link": config.DASHBOARD_API_URL.String() + "/frame/" + config.ServiceName + "?orgId=1&language=en-US&theme=gray&refresh=5s", "iconUrl": config.IFPS_ANDON_UI_URL.String() + "/_nuxt/img/Andon@1x.ff8bda8.svg", "display": "Show"}
		} else {
			content = map[string]interface{}{"name": regName, "link": "http://127.0.0.1:8080/frame/" + config.ServiceName + "?orgId=1&language=en-US&theme=gray&refresh=5s", "iconUrl": "http://127.0.0.1:5000/_nuxt/img/Andon@1x.ff8bda8.svg", "display": "Show"}
		}
		variable := map[string]interface{}{"iAppInput": content}
		httpRequestBody, _ = json.Marshal(map[string]interface{}{
			"query":     "mutation ($iAppInput: AddIAppInput!) { addIApp(input: $iAppInput) { iApp { name } } }",
			"variables": variable,
		})
		request, _ = http.NewRequest("POST", config.IFP_DESK_API_URL.String(), bytes.NewBuffer(httpRequestBody))
		if config.ServerLocation == "Cloud" {
			request.Header.Set("X-Ifp-App-Secret", config.IFPToken)
		} else {
			request.Header.Set("cookie", config.IFPToken)
		}
		request.Header.Set("Content-Type", "application/json")
		response, _ = httpClient.Do(request)
		m, _ = simplejson.NewFromReader(response.Body)
		if len(m.Get("errors").MustArray()) == 0 {
			fmt.Println("      Register iApp", regName, "Success")
		} else {
			fmt.Println("      Register iApp", regName, "Fail")
		}
	}
}

func RegisterOutbound(regName string) {
	httpClient := &http.Client{}
	content := map[string]interface{}{"name": regName, "sourceId": "scada_ifpsandon", "url": config.IFPS_ANDON_DAEMON_DATABROKER_API_URL.String(), "active": true}
	variable := map[string]interface{}{"input": content}
	httpRequestBody, _ := json.Marshal(map[string]interface{}{
		"query":     "mutation ($input: AddOutboundInput!) {     addOutbound(input: $input) {         outbound {             id             name             url             sourceId             allowUnauthorized             active             connected         }     } }",
		"variables": variable,
	})
	request, _ := http.NewRequest("POST", config.IFP_DESK_API_URL.String(), bytes.NewBuffer(httpRequestBody))
	if config.ServerLocation == "Cloud" {
		request.Header.Set("X-Ifp-App-Secret", config.IFPToken)
	} else {
		request.Header.Set("cookie", config.IFPToken)
	}
	request.Header.Set("Content-Type", "application/json")
	response, _ := httpClient.Do(request)
	if response.StatusCode == 200 {
		config.IFPStatus = "Up"
	} else {
		config.IFPStatus = "Down"
	}
	m, _ := simplejson.NewFromReader(response.Body)
	if len(m.Get("errors").MustArray()) == 0 {
		fmt.Println("      Register Outbound " + regName + " Success")
	} else {
		fmt.Println("      Outbound " + regName + " is already exist")
	}
}
