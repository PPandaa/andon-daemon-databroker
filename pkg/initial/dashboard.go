package initial

import (
	"bytes"
	"databroker/config"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bitly/go-simplejson"
)

var (
	serviceFolderID int
)

func importPluginRequest(filePath string) {
	httpClient := &http.Client{}
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	httpRequestBody := &bytes.Buffer{}
	httpRequestBodyWriter := multipart.NewWriter(httpRequestBody)
	part, err := httpRequestBodyWriter.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		fmt.Println(err)
	}
	io.Copy(part, file)
	httpRequestBodyWriter.Close()
	request, _ := http.NewRequest("POST", config.DASHBOARD_API_URL.String()+"/file/up", httpRequestBody)
	request.Header.Set("Content-Type", httpRequestBodyWriter.FormDataContentType())
	request.Header.Add("Authorization", config.DashboardToken)
	response, _ := httpClient.Do(request)
	if response.StatusCode != 200 {
		fmt.Println("      ", response.Status, "| Plugin Import Fail")
	}
}

func FromLocalImportPlugins() {
	files, _ := ioutil.ReadDir("./plugins")
	for _, file := range files {
		fmt.Println("    -", file.Name())
		importPluginRequest("./plugins/" + file.Name())
	}
}

func CreateDatasource(dsName string, dsType string, dsURL string) {
	httpClient := &http.Client{}
	httpRequestBody, _ := json.Marshal(map[string]interface{}{
		"name":      dsName,
		"type":      dsType,
		"url":       dsURL,
		"access":    "proxy",
		"basicAuth": false,
		"isDefault": false,
		"readOnly":  false,
	})
	request, _ := http.NewRequest("POST", config.DASHBOARD_API_URL.String()+"/datasources", bytes.NewBuffer(httpRequestBody))
	request.Header.Set("accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", config.DashboardToken)
	response, _ := httpClient.Do(request)
	if response.StatusCode != 200 {
		if response.StatusCode == 409 {
			fmt.Println("      " + dsName + " Is Already Exist")
		} else {
			fmt.Println("      " + dsName + " Create Fail")
		}
	} else {
		fmt.Println("      " + dsName + " Create Success")
	}
}

func CreateDashboardFolder(folderName string) {
	httpClient := &http.Client{}
	httpRequestBody, _ := json.Marshal(map[string]string{
		"title": folderName,
	})
	request, _ := http.NewRequest("POST", config.DASHBOARD_API_URL.String()+"/folders", bytes.NewBuffer(httpRequestBody))
	request.Header.Set("accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", config.DashboardToken)
	response, _ := httpClient.Do(request)
	m, _ := simplejson.NewFromReader(response.Body)
	if response.StatusCode != 200 {
		if response.StatusCode == 400 {
			getFoldersRequest, _ := http.NewRequest("GET", config.DASHBOARD_API_URL.String()+"/search?query="+folderName+"&type=dash-folder", nil)
			getFoldersResponse, _ := httpClient.Do(getFoldersRequest)
			folders, _ := simplejson.NewFromReader(getFoldersResponse.Body)
			for i := 0; i < len(folders.MustArray()); i++ {
				if folders.GetIndex(i).Get("title").MustString() == folderName {
					serviceFolderID = folders.GetIndex(i).Get("id").MustInt()
				}
			}
			fmt.Println("     ", folderName, "Folder Is Already Exist ->", folderName, "Folder ID:", serviceFolderID)
		} else {
			fmt.Println("     ", response.Status, "| Folder Create Fail")
		}
	} else {
		serviceFolderID = m.Get("id").MustInt()
		fmt.Println("     ", folderName, "Folder Create Success ->", folderName, "Folder ID:", serviceFolderID)
	}
}

func importDashboardRequest(filePath string) {
	httpClient := &http.Client{}
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	byteValue, _ := ioutil.ReadAll(file)
	var dashboardRequestJSON map[string]interface{}
	json.Unmarshal([]byte(byteValue), &dashboardRequestJSON)
	dashboardRequestJSON["folderId"] = serviceFolderID
	httpRequestBody, _ := json.Marshal(dashboardRequestJSON)
	request, _ := http.NewRequest("POST", config.DASHBOARD_API_URL.String()+"/dashboards/import", bytes.NewBuffer(httpRequestBody))
	request.Header.Set("accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", config.DashboardToken)
	response, _ := httpClient.Do(request)
	if response.StatusCode != 200 {
		if response.StatusCode == 500 {
			fmt.Println("       Dashboard Is Already Exist")
		} else {
			fmt.Println("      ", response.Status, "| Import Dashboard Fail")
		}
	}
}

func ImportDashboard() {
	files, _ := ioutil.ReadDir("./requestDashboardTemplates")
	for _, file := range files {
		fmt.Println("    -", file.Name())
		importDashboardRequest("./requestDashboardTemplates/" + file.Name())
	}
}

func importSRPRequest(filePath string) {
	httpClient := &http.Client{}
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	byteValue, _ := ioutil.ReadAll(file)
	var srpRequestJSON []map[string]interface{}
	json.Unmarshal([]byte(byteValue), &srpRequestJSON)
	for _, json := range srpRequestJSON {
		configJs := json["configJs"].(string)
		json["srpName"] = config.ServiceName
		json["logoImg"] = "http://127.0.0.1:5000/_nuxt/img/Andon@1x.ff8bda8.svg"
		json["configJs"] = strings.ReplaceAll(configJs, "dashboardHostname", config.DASHBOARD_API_URL.Hostname())
		json["configJs"] = strings.ReplaceAll(configJs, "serviceName", config.ServiceName)
		json["configJs"] = strings.ReplaceAll(configJs, "serviceFolderID", strconv.Itoa(serviceFolderID))
	}
	httpRequestBody, _ := json.Marshal(srpRequestJSON)
	request, _ := http.NewRequest("POST", config.DASHBOARD_API_URL.String()+"/frame/imports", bytes.NewBuffer(httpRequestBody))
	request.Header.Set("accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", config.DashboardToken)
	response, _ := httpClient.Do(request)
	if response.StatusCode != 200 {
		fmt.Println("     ", response.Status, "| Import SRP Fail")
	}
}

func CreateSRP(srpName string) {
	httpClient := &http.Client{}
	request, _ := http.NewRequest("GET", config.DASHBOARD_API_URL.String()+"/frame/search/?name="+srpName, nil)
	request.Header.Set("accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, _ := httpClient.Do(request)
	m, _ := simplejson.NewFromReader(response.Body)
	if response.StatusCode != 200 {
		fmt.Println("     ", response.Status, "| SRP Create Fail")
	} else {
		if len(m.MustArray()) == 0 {
			files, _ := ioutil.ReadDir("./requestSRPTemplates")
			for _, file := range files {
				fmt.Println("    -", file.Name())
				importSRPRequest("./requestSRPTemplates/" + file.Name())
			}
		} else {
			fmt.Println("     ", srpName, "SRP Is Already Exist")
		}
	}
}
