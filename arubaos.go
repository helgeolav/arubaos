package arubaos

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

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
			Timeout: 8 * time.Second,
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
		c.cookie = nil
		c.uidAruba = ""
		return authObj, nil
	}
	return authObj, nil
}

// AFilter URI Params for Get Reqs
type AFilter struct {
	Count int
}

// MMApDB the response when retrieving APs from a Mobility Master
type MMApDB struct {
	AP []MMAp `json:"AP Database"`
}

// MMAp the properties that exist on APs from the Mobility Master
type MMAp struct {
	MacAddr string `json:"apmac"`
	Name    string `json:"apname"`
	Group   string `json:"apgroup"`
	Model   string `json:"model"`
	Serial  string `json:"serialno"`
	IPAddr  string `json:"ipaddress"`
	Status  string `json:"status"`
	WLCIp   string `json:"switchip"`
}

// GetMMApDB the Mobility Master has a unique API Call
// to retrieve APs from its Database
func (c *Client) GetMMApDB(f AFilter) ([]MMAp, error) {
	if c.cookie == nil {
		return nil, fmt.Errorf("you must first login to perform this action")
	}
	endpoint := "/configuration/object/apdatabase"
	req, _ := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	// Custom QueryString for Request
	qs := map[string]string{"config_path": "/md"}
	if f.Count != 0 {
		qs["count"] = strconv.Itoa(f.Count)
	}
	// Add Common Values to the REQ
	c.updateReq(req, qs)
	res, err := c.http.Do(req)
	if err != nil {
		return []MMAp{}, fmt.Errorf("%v", err)
	}
	defer res.Body.Close()
	var apDb MMApDB
	json.NewDecoder(res.Body).Decode(&apDb)
	return apDb.AP, nil
}

// APDatabase the response from a show ap database long cmd on a MM/WLC
type APDatabase struct {
	AP []AP `json:"AP Database"`
}

// AP the properties that exist on AccessPoints
type AP struct {
	MacAddr      string `json:"Wired MAC Address"`
	Name         string `json:"Name"`
	Group        string `json:"Group"`
	Model        string `json:"AP Type"`
	Serial       string `json:"Serial #"`
	IPAddr       string `json:"IP Address"`
	Status       string `json:"Status"`
	PrimaryWlc   string `json:"Switch IP"`
	SecondaryWlc string `json:"Standby IP"`
}

// GetApDB retrieves AccessPoints associated with a WLC
// show ap database long
func (c *Client) GetApDB() ([]AP, error) {
	if c.cookie == nil {
		return nil, fmt.Errorf("you must first login to perform this action")
	}
	endpoint := "/configuration/showcommand"
	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return []AP{}, fmt.Errorf("unabled to create a new request: %v", err)
	}
	qs := map[string]string{"command": "show ap database long"}
	c.updateReq(req, qs)
	res, err := c.http.Do(req)
	if err != nil {
		return []AP{}, fmt.Errorf("%v", err)
	}
	defer res.Body.Close()
	var apDatabase APDatabase
	json.NewDecoder(res.Body).Decode(&apDatabase)
	return apDatabase.AP, nil
}

func (c *Client) updateReq(req *http.Request, qs map[string]string) {
	req.Header.Add("Content-Type", "application/json")
	req.AddCookie(c.cookie)
	q := req.URL.Query()
	for key, val := range qs {
		q.Add(key, val)
	}
	q.Add("UIDARUBA", c.uidAruba)
	req.URL.RawQuery = q.Encode()
}

// Intf the Aruba AP Interface Information
type Intf struct {
	Duplex    string `json:"Duplex"`
	MAC       string `json:"MAC"`
	Oper      string `json:"Oper"`
	Port      string `json:"Port"`
	RXBytes   string `json:"RX-Bytes"`
	RXPackets string `json:"RX-Packets"`
	Speed     string `json:"Speed"`
	TXBytes   string `json:"TX-Bytes"`
	TXPackets string `json:"TX-Packets"`
}

// GetApPortStatus retrieves Interface statistics of an AP
// This Command Must be run from a Controller *NOT MM
func (c *Client) GetApPortStatus(mac string) (Intf, error) {
	if c.cookie == nil {
		return Intf{}, fmt.Errorf("you must first login to perform this action")
	}
	endpoint := "/configuration/showcommand"
	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return Intf{}, fmt.Errorf("%v", err)
	}
	cmd := fmt.Sprintf("show ap port status wired-mac %s", mac)
	qs := map[string]string{"command": cmd}
	c.updateReq(req, qs)
	res, err := c.http.Do(req)
	if err != nil {
		return Intf{}, fmt.Errorf("%v", err)
	}
	defer res.Body.Close()
	var intf Intf

	// This Block of Code map[string][]Slice
	// Is Used Because the Property/Field of the returned
	// JSON Object is Dynamic/Non-Deterministic, so IT needs
	// To Be PARSED and Stripped OFF
	var mintfs map[string][]Intf
	json.NewDecoder(res.Body).Decode(&mintfs)
	for k, intfs := range mintfs {
		// Ignore these Fields
		if k == "_meta" || k == "_data" {
			continue
		}
		for _, ints := range intfs {
			if ints.Oper == "up" {
				intf = ints
				break
			}
		}
	}
	return intf, nil
}

// APLldp the properties of a Neighbor Connected to the AP
type APLldp struct {
	APName         string `json:"AP"`
	RemoteHostname string `json:"Chassis Name/ID"`
	RemoteIP       string `json:"Mgmt. Address"`
	RemoteIntfDesc string `json:"Port Desc"`
	RemoteIntf     string `json:"Port ID"`
}

// GetApLLDPInfo gets LLDP Info of Device Connecting to the AP
// This Command MUST be run from the Controller *NOT MM
func (c *Client) GetApLLDPInfo(apName string) (APLldp, error) {
	if c.cookie == nil {
		return APLldp{}, fmt.Errorf("you must first login to perform this action")
	}
	endpoint := "/configuration/showcommand"
	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return APLldp{}, fmt.Errorf("%v", err)
	}
	cmd := fmt.Sprintf("show ap lldp neighbors ap-name %s", apName)
	qs := map[string]string{"command": cmd}
	c.updateReq(req, qs)
	res, err := c.http.Do(req)
	if err != nil {
		return APLldp{}, fmt.Errorf("%v", err)
	}
	defer res.Body.Close()
	var lldp APLldp

	// This Block of Code map[string][]Slice
	// Is Used Because the Property/Field of the returned
	// JSON Object is Dynamic/Non-Deterministic, so IT needs
	// To Be PARSED and Stripped OFF
	var mlldp map[string][]APLldp
	json.NewDecoder(res.Body).Decode(&mlldp)
	for k, lldps := range mlldp {
		// Ignore These Fields
		if k == "_data" || k == "_meta" {
			continue
		}
		for _, l := range lldps {
			lldp = l
			break
		}
	}
	return lldp, nil
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

// Get User (show user-table mac <mac-addr>)
