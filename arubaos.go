package arubaos

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client struct used for the Connection
// To an Aruba MM and/or Controller
type Client struct {
	BaseURL  string
	Username string
	Password string

	http     *http.Client
	cookie   *http.Cookie
	uidAruba string
}

// AP the properties that exist on AccessPoints
type AP struct {
}

// APDatabase the response from a show ap database long cmd on a MM/WLC
type APDatabase struct {
	AP []struct {
		Type     string `json:"AP Type"`
		Group    string `json:"Group"`
		IPAddrs  string `json:"IP Address"`
		Name     string `json:"Name"`
		Serial   string `json:"Serial #"`
		Status   string `json:"Status"`
		SwitchIP string `json:"Switch IP"`
		MACAddr  string `json:"Wired MAC Address"`
	} `json:"AP Database"`
}

// RenameAPReq is the object definition needed for MM/WLC
// To Rename an AP from Default
type RenameAPReq struct {
	MacAddr   string `json:"wired-mac"`
	ApName    string `json:"ap-name"`
	SerialNum string `json:"serial-num"`
	NewName   string `json:"new-name"`
}

// APAssoc show user-table
type APAssoc struct {
	Users []struct {
		APName        string      `json:"AP name"`
		AgeDHM        string      `json:"Age(d:h:m)"`
		Auth          interface{} `json:"Auth"`
		EssidBssidPhy string      `json:"Essid/Bssid/Phy"`
		ForwardMode   string      `json:"Forward mode"`
		HostName      interface{} `json:"Host Name"`
		IP            string      `json:"IP"`
		MAC           string      `json:"MAC"`
		Name          interface{} `json:"Name"`
		Profile       string      `json:"Profile"`
		Roaming       string      `json:"Roaming"`
		Role          string      `json:"Role"`
		Type          string      `json:"Type"`
		UserType      string      `json:"User Type"`
		VPNLink       interface{} `json:"VPN link"`
	} `json:"Users"`
}

// ArubaAuthResp in login/logout methods
type ArubaAuthResp struct {
	GlobalRes struct {
		Status    string `json:"status"`
		StatusStr string `json:"status_str"`
		UIDAruba  string `json:"UIDARUBA"`
	} `json:"_global_result"`
}

// New creates a new reference to the Client struct
func New(host, user, pass string, ignoreSSL bool) *Client {
	return &Client{
		BaseURL:  fmt.Sprintf("https://%s:4343/v1", host),
		Username: user,
		Password: pass,
		http: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: ignoreSSL,
				},
			},
			Timeout: 8 * time.Second,
		},
	}
}

func (c *Client) login() error {
	data := url.Values{}
	data.Set("username", c.Username)
	data.Set("password", c.Password)
	creds := strings.NewReader(data.Encode())
	req, err := http.NewRequest("POST", c.BaseURL+"/api/login", creds)
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
	req, err := http.NewRequest("GET", c.BaseURL+"/api/logout", nil)
	if err != nil {
		return ArubaAuthResp{}, fmt.Errorf("failed to create a new request: %v", err)
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
		return authObj, nil
	}
	return authObj, nil
}

// GetAPDatabase retrieves AccessPoints associated with a WLC
// show ap database long
func (c *Client) GetAPDatabase() (APDatabase, error) {
	err := c.login()
	if err != nil {
		return APDatabase{}, fmt.Errorf("%v", err)
	}
	endpoint := "/configuration/showcommand"
	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return APDatabase{}, fmt.Errorf("unabled to create a new request: %v", err)
	}
	req.AddCookie(c.cookie)
	q := req.URL.Query()
	q.Add("command", "show ap database long")
	q.Add("UIDARUBA", c.uidAruba)
	req.URL.RawQuery = q.Encode()
	res, err := c.http.Do(req)
	if err != nil {
		return APDatabase{}, fmt.Errorf("%v", err)
	}
	defer res.Body.Close()
	var apDatabase APDatabase
	json.NewDecoder(res.Body).Decode(&apDatabase)
	return apDatabase, nil
}

// CreateVAP
// Rename AP
// Get User (show user-table mac <mac-addr>)
