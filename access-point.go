package arubaos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

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
		return Intf{}, fmt.Errorf(loginWarning)
	}
	req, err := c.genGetReq("/configuration/showcommand")
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
	if err = json.NewDecoder(res.Body).Decode(&mintfs); err != nil {
		return Intf{}, fmt.Errorf("error parsing resp body: %v", err)
	}
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
		return APLldp{}, fmt.Errorf(loginWarning)
	}
	req, err := c.genGetReq("/configuration/showcommand")
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
	if err = json.NewDecoder(res.Body).Decode(&mlldp); err != nil {
		return APLldp{}, fmt.Errorf("error parsing resp body: %v", err)
	}
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

// RebootAp ...
func (c *Client) RebootAp(mac string) error {
	if c.cookie == nil {
		return fmt.Errorf(loginWarning)
	}
	apBoot := map[string]string{"wired-mac": mac}

	j, _ := json.Marshal(apBoot)
	body := strings.NewReader(string(j))
	endpoint := "/configuration/object/apboot"
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, body)
	if err != nil {
		return fmt.Errorf("unabled to create request: %v", err)
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

// LED actions that can be taken on AP LEDs
var LED = map[string][]string{
	"actions": []string{
		"blink", "normal", "fault-disable", "fault-enable",
	},
}

// ApLedActionReq ...
type ApLedActionReq struct {
	MacAddr     string `json:"wired-mac"`
	ApName      string `json:"ap-name"`
	IPAddr      string `json:"ip-addr"`
	All         bool   `json:"all"`
	ApGroup     string `json:"ap-group"`
	LocalGlobal string `json:"local_global"`
	Action      string `json:"action_option"`
}
