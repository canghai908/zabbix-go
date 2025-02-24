package zabbix

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Params map[string]interface{}

var isZbx64 bool

type request struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Auth    string      `json:"auth,omitempty"`
	Id      int32       `json:"id"`
}

type Response struct {
	Jsonrpc string      `json:"jsonrpc"`
	Error   *Error      `json:"error"`
	Result  interface{} `json:"result"`
	Id      int32       `json:"id"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d (%s): %s", e.Code, e.Message, e.Data)
}

type ExpectedOneResult int

func (e *ExpectedOneResult) Error() string {
	return fmt.Sprintf("Expected exactly one result, got %d.", *e)
}

type ExpectedMore struct {
	Expected int
	Got      int
}

func (e *ExpectedMore) Error() string {
	return fmt.Sprintf("Expected %d, got %d.", e.Expected, e.Got)
}

// 为了向后兼容，保留Auth字段但添加警告注释
type API struct {
	// Auth token, filled by Login() or SetAuth()
	// Warning: Do not set Auth directly, use SetAuth() instead to ensure proper version detection
	Auth   string
	Logger *log.Logger // request/response logger, nil by default
	url    string
	c      http.Client
	id     int32
}

// Creates new API access object.
// Typical URL is http://host/api_jsonrpc.php or http://host/zabbix/api_jsonrpc.php.
// It also may contain HTTP basic auth username and password like
// http://username:password@host/api_jsonrpc.php.
func NewAPI(url string) (api *API) {
	return &API{url: url, c: http.Client{}}
}

// Allows one to use specific http.Client, for example with InsecureSkipVerify transport.
func (api *API) SetClient(c *http.Client) {
	api.c = *c
}

func (api *API) printf(format string, v ...interface{}) {
	if api.Logger != nil {
		api.Logger.Printf(format, v...)
	}
}

func (api *API) callBytes(method string, params interface{}) (b []byte, err error) {
	id := atomic.AddInt32(&api.id, 1)
	var jsonobj request

	// 7.2及以上版本完全不需要auth字段
	if isZbx64 {
		jsonobj = request{
			Jsonrpc: "2.0",
			Method:  method,
			Params:  params,
			Id:      id,
		}
	} else {
		// 6.4以下版本需要auth字段
		auth := ""
		if method != "APIInfo.version" {
			auth = api.Auth
		}
		jsonobj = request{
			Jsonrpc: "2.0",
			Method:  method,
			Params:  params,
			Auth:    auth,
			Id:      id,
		}
	}

	b, err = json.Marshal(jsonobj)
	if err != nil {
		return
	}
	api.printf("Request (POST): %s", b)

	// make the http client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 5 * time.Second, Transport: tr}
	req, err := http.NewRequest("POST", api.url, bytes.NewReader(b))
	if err != nil {
		return
	}
	req.ContentLength = int64(len(b))
	req.Header.Add("Content-Type", "application/json-rpc")
	req.Header.Add("User-Agent", "github.com/AlekSi/zabbix")

	// 6.4及以上版本使用Bearer认证
	if isZbx64 && method != "APIInfo.version" {
		req.Header.Add("Authorization", "Bearer "+api.Auth)
	}

	res, err := client.Do(req)
	if err != nil {
		api.printf("Error   : %s", err)
		return
	}
	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		api.printf("Error   : %s", err)
		return
	}
	api.printf("Response (%d): %s", res.StatusCode, b)
	return
}

// Calls specified API method. Uses api.Auth if not empty.
// err is something network or marshaling related. Caller should inspect response.Error to get API error.
func (api *API) Call(method string, params interface{}) (response Response, err error) {
	b, err := api.callBytes(method, params)
	if err == nil {
		err = json.Unmarshal(b, &response)
	}
	return
}

// Uses Call() and then sets err to response.Error if former is nil and latter is not.
func (api *API) CallWithError(method string, params interface{}) (response Response, err error) {
	response, err = api.Call(method, params)
	if err == nil && response.Error != nil {
		err = response.Error
	}
	return
}

// Calls "user.login" API method and fills api.Auth field.
// This method modifies API structure and should not be called concurrently with other methods.
func (api *API) Login(user, password string) (auth string, err error) {
	// 先获取版本并设置isZbx64
	_, err = api.Version()
	if err != nil {
		return
	}

	key := "user"
	if isZbx64 {
		key = "username"
	}

	params := map[string]string{
		key:        user,
		"password": password,
	}

	response, err := api.CallWithError("user.login", params)
	if err != nil {
		return
	}
	auth = response.Result.(string)
	api.Auth = auth
	return
}

// Calls "APIInfo.version" API method.
func (api *API) Version() (v string, err error) {
	// APIInfo.version 方法不需要认证
	response, err := api.CallWithError("APIInfo.version", Params{})
	if err != nil {
		return
	}

	v = response.Result.(string)

	// 判断版本并设置认证方式
	verArr := strings.Split(v, ".")
	ZbxMasterVer, _ := strconv.ParseInt(verArr[0], 10, 64)
	ZbxMiddleVer, _ := strconv.ParseInt(verArr[1], 10, 64)

	isZbx64 = ZbxMasterVer > 6 || (ZbxMasterVer == 6 && ZbxMiddleVer >= 4)
	return
}

// SetAuth sets the authentication token and determines the Zabbix version
func (api *API) SetAuth(auth string) error {
	api.Auth = auth
	// 获取版本并设置isZbx64
	_, err := api.Version()
	return err
}
