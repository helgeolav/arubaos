package arubaos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// WdbCpSec whitelist properties
type WdbCpSec struct {
	// optional use only for Modify
	// true
	State bool `json:"state,omitempty"`
	// optional use only for Modify
	// approved-ready-for-cert
	// certified-factory-cert
	Act string `json:"act,omitempty"`
	// optional use only for Modify
	RevokeTxt string `json:"revoke-text,omitempty"`
	// optional
	Description string `json:"description,omitempty"`
	// optional use only for Modify
	// true
	Cert bool `json:"cert-type,omitempty"`
	// optional use only for Modify
	// factory-cert
	// switch-cert
	CertType string `json:"certtype,omitempty"`
	// optional use only for Modify
	// true
	Mode bool `json:"mode,omitempty"`
	// optional use only for Modify
	// enable|disable
	ModeAct string `json:"modeact,omitempty"`
	// Do not use for DEL
	ApName string `json:"ap_name,omitempty"`
	// Do not use for DEL
	ApGroup string `json:"ap_group,omitempty"`
	// Wired-Mac-Address ab:cd:ef:01:23:45
	Name string `json:"name"`
}

// CpSecAdd add APs to Whitelist
func (c *Client) CpSecAdd(aps []WdbCpSec) error {
	if c.cookie == nil {
		return fmt.Errorf(loginWarning)
	}
	type addWhitelist struct {
		CpSecAdd WdbCpSec `json:"wdb_cpsec_add_mac"`
	}
	var apList []addWhitelist
	for _, ap := range aps {
		apList = append(apList, addWhitelist{CpSecAdd: ap})
	}
	type apAddWl struct {
		ApConfList []addWhitelist `json:"_list"`
	}
	apWhitelist := apAddWl{ApConfList: apList}
	j, _ := json.Marshal(apWhitelist)
	fmt.Println(string(j))
	body := strings.NewReader(string(j))

	endpoint := "/configuration/object"
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, body)

	c.updateReq(req, map[string]string{})
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	defer res.Body.Close()
	result, err := ioutil.ReadAll(res.Body)
	fmt.Println(string(result))
	return nil
}

// CpSecModify update APs in Whitelist
func (c *Client) CpSecModify(aps []WdbCpSec) error {
	if c.cookie == nil {
		return fmt.Errorf(loginWarning)
	}
	type modWl struct {
		CpSecMod WdbCpSec `json:"wdb_cpsec_modify_mac"`
	}
	var modAp []modWl
	for _, ap := range aps {
		ap.Act = "approved-ready-for-cert"
		ap.Cert = true
		ap.CertType = "factory-cert"
		ap.Mode = true
		ap.ModeAct = "enable"
		modAp = append(modAp, modWl{CpSecMod: ap})
	}
	type apModWl struct {
		ApConfList []modWl `json:"_list"`
	}
	apModWhitelist := apModWl{ApConfList: modAp}
	j, _ := json.Marshal(apModWhitelist)
	body := strings.NewReader(string(j))
	endpoint := "/configuration/object"
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	c.updateReq(req, map[string]string{})
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer res.Body.Close()
	result, err := ioutil.ReadAll(res.Body)
	fmt.Println(string(result))
	return nil
}

/*
{
	"_list":[
		{ "wdb_cpsec_del_mac":{"name":"bc:9f:e4:c3:0f:90"} }
	]
}
*/

// CpSecDel remove APs from Whitelist
func (c *Client) CpSecDel(aps []WdbCpSec) error {
	if c.cookie == nil {
		return fmt.Errorf(loginWarning)
	}
	// DelWhitelist ...
	type delWhitelist struct {
		CpSecDel WdbCpSec `json:"wdb_cpsec_del_mac"`
	}
	var apList []delWhitelist
	for _, ap := range aps {
		apList = append(apList, delWhitelist{CpSecDel: ap})
	}
	type apDelWl struct {
		ApConfList []delWhitelist `json:"_list"`
	}
	apDel := apDelWl{ApConfList: apList}
	j, _ := json.Marshal(apDel)
	body := strings.NewReader(string(j))

	endpoint := "/configuration/object"
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, body)

	c.updateReq(req, map[string]string{})
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	defer res.Body.Close()
	result, err := ioutil.ReadAll(res.Body)
	fmt.Println(string(result))
	return nil
}

// ClrGapAp deletes APs from LMS(Controller) Database
func (c *Client) ClrGapAp() {}
