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

// VersionInfo holds Zabbix version information
type VersionInfo struct {
	Version     string
	Major       int64
	Minor       int64
	Patch       int64
	UseBearer   bool // true for Zabbix 7.2+
	UseUsername bool // true for Zabbix 6.4+
}

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

// API provides access to Zabbix API
type API struct {
	// Auth token, filled by Login() or SetAuth()
	// Warning: Do not set Auth directly, use SetAuth() instead to ensure proper version detection
	Auth        string
	Logger      *log.Logger // request/response logger, nil by default
	url         string
	c           http.Client
	id          int32
	versionInfo *VersionInfo // cached version information
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
	// Ensure version info is available (but skip for APIInfo.version to avoid recursion)
	useBearer := false
	if method != "APIInfo.version" {
		if api.versionInfo == nil {
			_, err = api.Version()
			if err != nil {
				return
			}
		}
		useBearer = api.versionInfo.UseBearer
	}

	id := atomic.AddInt32(&api.id, 1)
	var jsonobj request

	// Zabbix 7.2+ uses Bearer token in header, no auth in JSON body
	// Zabbix < 7.2 uses auth field in JSON body
	if useBearer {
		jsonobj = request{
			Jsonrpc: "2.0",
			Method:  method,
			Params:  params,
			Id:      id,
		}
	} else {
		// Older versions need auth field in JSON body
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

	// Zabbix 7.2+ uses Bearer token in Authorization header
	if useBearer && method != "APIInfo.version" && api.Auth != "" {
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
	// Ensure version info is available
	if api.versionInfo == nil {
		_, err = api.Version()
		if err != nil {
			return
		}
	}

	// Zabbix 6.4+ uses "username", older versions use "user"
	key := "user"
	if api.versionInfo.UseUsername {
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

// Calls "APIInfo.version" API method and caches version information.
func (api *API) Version() (v string, err error) {
	// APIInfo.version doesn't require authentication
	response, err := api.CallWithError("APIInfo.version", Params{})
	if err != nil {
		return
	}

	v = response.Result.(string)

	// Parse version string (e.g., "7.4.0")
	verArr := strings.Split(v, ".")
	if len(verArr) < 2 {
		return v, fmt.Errorf("invalid version format: %s", v)
	}

	major, _ := strconv.ParseInt(verArr[0], 10, 64)
	minor, _ := strconv.ParseInt(verArr[1], 10, 64)
	patch := int64(0)
	if len(verArr) >= 3 {
		patch, _ = strconv.ParseInt(verArr[2], 10, 64)
	}

	// Cache version information
	api.versionInfo = &VersionInfo{
		Version:     v,
		Major:       major,
		Minor:       minor,
		Patch:       patch,
		UseBearer:   major > 7 || (major == 7 && minor >= 2), // Zabbix 7.2+ uses Bearer token
		UseUsername: major > 6 || (major == 6 && minor >= 4), // Zabbix 6.4+ uses "username" instead of "user"
	}

	return v, nil
}

// SetAuth sets the authentication token and determines the Zabbix version
func (api *API) SetAuth(auth string) error {
	api.Auth = auth
	// Get version to determine authentication method
	_, err := api.Version()
	return err
}

// GetVersionInfo returns cached version information
func (api *API) GetVersionInfo() *VersionInfo {
	return api.versionInfo
}
