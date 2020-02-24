package arubaos

import (
	"encoding/json"
	"fmt"
	"strconv"
)

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
		return nil, fmt.Errorf(loginWarning)
	}
	req, err := c.genGetReq("/configuration/object/apdatabase")
	if err != nil {
		return nil, err
	}
	if f.CfgPath == "" {
		f.CfgPath = "/md"
	}
	// Custom QueryString for Request
	qs := map[string]string{"config_path": f.CfgPath}
	if f.Count != 0 {
		qs["count"] = strconv.Itoa(f.Count)
	}
	// Add Common Values to the REQ
	c.updateReq(req, qs)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	defer res.Body.Close()
	var apDb MMApDB
	if err = json.NewDecoder(res.Body).Decode(&apDb); err != nil {
		return nil, fmt.Errorf("error parsing resp body: %v", err)
	}
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
		return nil, fmt.Errorf(loginWarning)
	}
	req, err := c.genGetReq("/configuration/showcommand")
	if err != nil {
		return nil, err
	}
	qs := map[string]string{"command": "show ap database long"}
	c.updateReq(req, qs)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	defer res.Body.Close()
	var apDatabase APDatabase
	if err = json.NewDecoder(res.Body).Decode(&apDatabase); err != nil {
		return nil, fmt.Errorf("error parsing resp body: %v", err)
	}
	return apDatabase.AP, nil
}
