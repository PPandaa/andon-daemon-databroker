package desk

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"databroker/config"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
)

func RefreshToken() {
	for {
		fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
		if len(config.Datacenter) == 0 {
			fmt.Println("refreshToken")
			httpClient := &http.Client{}
			content := map[string]string{"username": config.AdminUsername, "password": config.AdminPassword}
			variable := map[string]interface{}{"input": content}
			httpRequestBody, _ := json.Marshal(map[string]interface{}{
				"query":     "mutation signIn($input: SignInInput!) {   signIn(input: $input) {     user {       name       __typename     }     __typename   } }",
				"variables": variable,
			})
			request, _ := http.NewRequest("POST", config.IFPURL, bytes.NewBuffer(httpRequestBody))
			request.Header.Set("Content-Type", "application/json")
			response, _ := httpClient.Do(request)
			m, _ := simplejson.NewFromReader(response.Body)
			for {
				if len(m.Get("errors").MustArray()) == 0 {
					break
				} else {
					fmt.Println("----------", time.Now().In(config.TaipeiTimeZone), "----------")
					fmt.Println("retry refreshToken")
					httpRequestBody, _ = json.Marshal(map[string]interface{}{
						"query":     "mutation signIn($input: SignInInput!) {   signIn(input: $input) {     user {       name       __typename     }     __typename   } }",
						"variables": variable,
					})
					request, _ = http.NewRequest("POST", config.IFPURL, bytes.NewBuffer(httpRequestBody))
					request.Header.Set("Content-Type", "application/json")
					response, _ = httpClient.Do(request)
					m, _ = simplejson.NewFromReader(response.Body)
					time.Sleep(6 * time.Minute)
				}
			}
			header := response.Header
			cookie := header["Set-Cookie"]
			var ifpToken, eiToken string
			for _, cookieContent := range cookie {
				tempSplit := strings.Split(cookieContent, ";")
				if strings.HasPrefix(tempSplit[0], "IFPToken") {
					ifpToken = tempSplit[0]
				} else if strings.HasPrefix(tempSplit[0], "EIToken") {
					eiToken = tempSplit[0]
				}
			}
			if eiToken == "" {
				config.Token = ifpToken
			} else {
				config.Token = ifpToken + ";" + eiToken
			}
			fmt.Println("Token:", config.Token)
			time.Sleep(60 * time.Minute)
		} else {
			fmt.Println("refreshClientSecret")
			timestamp := time.Now()
			options := &newSRPTokenOptions{Timestamp: &timestamp}
			result := newSrpToken("OEE", options)
			httpClient := &http.Client{}
			request, _ := http.NewRequest("GET", config.SSOURL+"/clients/OEE", nil)
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("X-Auth-SRPToken", result)
			q := request.URL.Query()
			if config.Namespace == "ifpsdev" || config.Namespace == "ifpsdemo" {
				q.Add("cluster", "eks011")
				q.Add("workspace", "53e8c8bd-b724-4c87-a905-5bbc5c30a36c")
				q.Add("namespace", "training")
			} else {
				q.Add("cluster", config.Cluster)
				q.Add("workspace", config.Workspace)
				q.Add("namespace", config.Namespace)
			}
			q.Add("serviceName", "OEE")
			request.URL.RawQuery = q.Encode()
			response, _ := httpClient.Do(request)
			m, _ := simplejson.NewFromReader(response.Body)
			config.Token = m.Get("clientSecret").MustString()
			fmt.Println("Token:", config.Token)
			time.Sleep(60 * time.Minute)
		}
	}
}

type ecb struct {
	b         cipher.Block
	blockSize int
}

type ecbEncrypter ecb

func (x *ecbEncrypter) BlockSize() int { return x.blockSize }

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

func newECBEncrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypter)(newECB(b))
}

type newSRPTokenOptions struct {
	Timestamp *time.Time
}

// PKCS7Padding adds padding to data
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func newSrpToken(serviceName string, opts ...*newSRPTokenOptions) string {
	now := time.Now()
	timestamp := &now

	for _, opt := range opts {
		if opt.Timestamp != nil {
			timestamp = opt.Timestamp
		}
	}

	key := "ssoisno12345678987654321"
	src := fmt.Sprintf("%v-%v", timestamp.Unix(), serviceName)

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}

	blockSize := block.BlockSize()
	data := PKCS7Padding([]byte(src), blockSize)

	encryptData := make([]byte, len(data))

	ecb := newECBEncrypter(block)
	ecb.CryptBlocks(encryptData, data)

	token := base64.URLEncoding.EncodeToString(encryptData)
	token = strings.ReplaceAll(token, "=", "")
	return token
}
