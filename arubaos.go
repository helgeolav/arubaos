package arubaos

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const loginWarning string = "you must first login to perform this action"

// Client struct used for the Connection
// To an Aruba MM and/or Controller
type Client struct {
	BaseURL  string
	Username string
	Password string
	IP       string

	http     *http.Client
	cookie   *http.Cookie
	uidAruba string
}

// New creates a reference to the Client struct
func New(host, user, pass string, ignoreSSL bool) *Client {
	return &Client{
		BaseURL:  fmt.Sprintf("https://%s:4343/v1", host),
		Username: user,
		Password: pass,
		IP:       host,
		http: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: ignoreSSL,
				},
			},
			Timeout: 60 * time.Second,
		},
	}
}

// ArubaAuthResp in login/logout methods
type ArubaAuthResp struct {
	GlobalRes struct {
		Status    string `json:"status"`
		StatusStr string `json:"status_str"`
		UIDAruba  string `json:"UIDARUBA"`
	} `json:"_global_result"`
}

// Login establishes a session with an Aruba Device
func (c *Client) Login() error {
	data := url.Values{}
	data.Set("username", c.Username)
	data.Set("password", c.Password)
	body := strings.NewReader(data.Encode())
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/login", body)
	if err != nil {
		return fmt.Errorf("failed to create a new request: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}
	defer res.Body.Close()

	var authObj ArubaAuthResp
	json.NewDecoder(res.Body).Decode(&authObj)
	if authObj.GlobalRes.Status != "0" {
		return fmt.Errorf(authObj.GlobalRes.StatusStr)
	}
	// if we've logged in successfully we'll be able to
	// grab the AUTH Token AND AuthCookie from the Resp
	c.uidAruba = authObj.GlobalRes.UIDAruba
	c.cookie = res.Cookies()[0]
	return nil
}

// Logout of the Controller
func (c *Client) Logout() (ArubaAuthResp, error) {
	req, err := c.genGetReq("/api/logout")
	if err != nil {
		return ArubaAuthResp{}, err
	}
	req.AddCookie(c.cookie)
	res, err := c.http.Do(req)
	if err != nil {
		return ArubaAuthResp{}, fmt.Errorf("failed to logout: %v", err)
	}
	defer res.Body.Close()

	var authObj ArubaAuthResp
	json.NewDecoder(res.Body).Decode(&authObj)
	if authObj.GlobalRes.StatusStr == "You've been logged out successfully" {
		c.cookie = nil
		c.uidAruba = ""
		return authObj, nil
	}
	return authObj, nil
}

// genGetReq returns a new http.Request object for a GET with the BaseURL as prefix to url
func (c *Client) genGetReq(url string) (*http.Request, error) {
	return http.NewRequest(http.MethodGet, c.BaseURL+url, nil)
}

// AFilter URI Params for Get Reqs
type AFilter struct {
	Count   int
	CfgPath string
}

// updateReq enhances a http.Request object with query values needed to query ArubaOS
func (c *Client) updateReq(req *http.Request, qs map[string]string) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.AddCookie(c.cookie)
	q := req.URL.Query()
	for key, val := range qs {
		q.Add(key, val)
	}
	q.Add("UIDARUBA", c.uidAruba)
	req.URL.RawQuery = q.Encode()
}

// WirelessClient ...
type WirelessClient struct {
	ApName     string `json:"AP name"`
	Auth       string `json:"Auth"`
	BSSID      string `json:"Bssid"`
	Controller string `json:"Current switch"`
	SSID       string `json:"Essid"`
	MacAddr    string `json:"MAC"`
	IPAddr     string `json:"IP"`
	DeviceType string `json:"Type"`
}

// GetClients ...
func (c *Client) GetClients() ([]WirelessClient, error) {
	var clients []WirelessClient
	if c.cookie == nil {
		return clients, errors.New("missing cookie")
	}
	req, err := c.genGetReq("/configuration/showcommand")
	if err != nil {
		return clients, err
	}
	cmd := "show global-user-table list"
	qs := map[string]string{"command": cmd}
	c.updateReq(req, qs)
	res, err := c.http.Do(req)
	if err != nil {
		return clients, err
	}
	defer res.Body.Close()
	type ClientResp map[string][]WirelessClient
	var clientResp ClientResp
	json.NewDecoder(res.Body).Decode(&clientResp)
	clients = clientResp["Global Users"]
	return clients, nil
}

// ControllerLicense ...
type ControllerLicense struct {
	Expires     string    `json:"Expires(Grace period expiry)"`
	Installed   time.Time `json:"Installed"`
	Key         string    `json:"Key"`
	ServiceType string    `json:"Service Type"`
}
