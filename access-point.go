package arubaos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
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
	var mintfs map[string][]interface{}
	if err = json.NewDecoder(res.Body).Decode(&mintfs); err != nil {
		return Intf{}, fmt.Errorf("error parsing resp body: %v", err)
	}
	for k, intfs := range mintfs {
		// Ignore these Fields
		if k == "_meta" || k == "_data" {
			continue
		}
		for _, m := range intfs {
			v := reflect.ValueOf(m)
			if v.Kind() == reflect.Map {
				for _, key := range v.MapKeys() {
					k := key.String()
					l := v.MapIndex(key)
					val := l.Interface().(string)
					switch k {
					case "Duplex":
						intf.Duplex = val
					case "MAC":
						intf.MAC = val
					case "Speed":
						intf.Speed = val
					case "Oper":
						if val != "up" {
							intf = Intf{}
							continue
						}
						intf.Oper = val
					case "Port":
						intf.Port = val
					case "RX-Bytes":
						intf.RXBytes = val
					case "RX-Packets":
						intf.RXPackets = val
					case "TX-Bytes":
						intf.TXBytes = val
					case "TX-Packets":
						intf.TXPackets = val
					}
				}
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
	// b, _ := ioutil.ReadAll(res.Body)
	// fmt.Println(string(b))
	var lldp APLldp

	// This Block of Code map[string][]Slice
	// Is Used Because the Property/Field of the returned
	// JSON Object is Dynamic/Non-Deterministic, so IT needs
	// To Be PARSED and Stripped OFF
	var mlldp map[string][]interface{}
	if err = json.NewDecoder(res.Body).Decode(&mlldp); err != nil {
		return APLldp{}, fmt.Errorf("error parsing resp body: %v", err)
	}
	for k, lldps := range mlldp {
		// Ignore These Fields
		if k == "_data" || k == "_meta" {
			continue
		}
		for _, m := range lldps {
			v := reflect.ValueOf(m)
			if v.Kind() == reflect.Map {
				for _, key := range v.MapKeys() {
					k := key.String()
					l := v.MapIndex(key)
					val := l.Interface().(string)
					switch k {
					case "AP":
						lldp.APName = val
					case "Chassis Name/ID":
						lldp.RemoteHostname = val
					case "Mgmt. Address":
						lldp.RemoteIP = val
					case "Port ID":
						lldp.RemoteIntf = val
					}
				}
				break
			}
		}
	}
	return lldp, nil
}

// RebootAp ...
func (c *Client) RebootAp(ap AP) (string, error) {
	if c.cookie == nil {
		return "", fmt.Errorf(loginWarning)
	}
	var apBoot map[string]string
	switch {
	case ap.Name != "":
		apBoot = map[string]string{"ap-name": ap.Name}
	case ap.MacAddr != "":
		apBoot = map[string]string{"wired-mac": ap.MacAddr}
	}

	j, _ := json.Marshal(apBoot)
	body := strings.NewReader(string(j))
	endpoint := "/configuration/object/apboot"
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, body)
	if err != nil {
		return "", fmt.Errorf("unabled to create request: %v", err)
	}
	c.updateReq(req, map[string]string{})
	res, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer res.Body.Close()
	type RebootResult struct {
		Result struct {
			Status    int    `json:"status"`
			StatusStr string `json:"status_str"`
		} `json:"_global_result"`
	}
	var apReboot RebootResult
	json.NewDecoder(res.Body).Decode(&apReboot)
	return strings.ToLower(apReboot.Result.StatusStr), nil
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
// var LED = map[string][]string{
// 	"actions": []string{
// 		"blink", "normal", "fault-disable", "fault-enable",
// 	},
// }

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

/*
{
  "Association Table": [
    {
      "Band steer moves (T/S)": "0/0",
      "Flags": "WVAB",
      "Name": "ap01.bsc.norton.cafb.al",
      "aid": "2",
      "assoc": "y",
      "assoc. time": "19m:0s",
      "auth": "y",
      "bssid": "a8:bd:27:5e:a8:12",
      "essid": "MyResNet-5G",
      "l-int": "20",
      "mac": "88:a4:79:cd:30:47",
      "num assoc": "1",
      "phy": "a-VHT-40sgi-2ss",
      "phy_cap": "a-HE-80-2ss-V",
      "tunnel-id": "0x10a35",
      "vlan-id": "1180"
	},
*/

// GetAp ...
func (c *Client) GetAp(apName string) (AP, error) {
	ap := AP{Name: apName}
	if c.cookie == nil {
		return ap, fmt.Errorf(loginWarning)
	}
	req, err := c.genGetReq("/configuration/showcommand")
	if err != nil {
		return ap, fmt.Errorf(loginWarning)
	}
	cmd := fmt.Sprintf("show ap details ap-name %s", apName)
	qs := map[string]string{"command": cmd}
	c.updateReq(req, qs)
	res, err := c.http.Do(req)
	if err != nil {
		return ap, fmt.Errorf(loginWarning)
	}
	defer res.Body.Close()
	type resData struct {
		Item  string `json:"Item"`
		Value string `json:"Value"`
	}
	type resResult map[string][]resData
	basicKey := fmt.Sprintf("AP %s Basic Information", apName)
	hwKey := fmt.Sprintf("AP %s Hardware Information", apName)
	var result resResult
	json.NewDecoder(res.Body).Decode(&result)
	for _, val := range result[basicKey] {
		switch {
		case val.Item == "LMS IP Address":
			ap.PrimaryWlc = val.Value
		case val.Item == "AP IP Address":
			ap.IPAddr = val.Value
		case val.Item == "Group":
			ap.Group = val.Value
		}
	}
	for _, val := range result[hwKey] {
		switch {
		case val.Item == "AP Type":
			ap.Model = val.Value
		case val.Item == "Wired MAC Address":
			ap.MacAddr = val.Value
		case val.Item == "Serial #":
			ap.Serial = val.Value
		}
	}
	return ap, nil
}

// GetApAssocCount returns the number of Clients Registered with a Specific AP
// Can only be run on the Controller the AP is Registered with
func (c *Client) GetApAssocCount(apName string) (int, error) {
	if c.cookie == nil {
		return 0, fmt.Errorf(loginWarning)
	}
	req, err := c.genGetReq("/configuration/showcommand")
	if err != nil {
		return 0, fmt.Errorf(loginWarning)
	}
	cmd := fmt.Sprintf("show ap association ap-name %s", apName)
	qs := map[string]string{"command": cmd}
	c.updateReq(req, qs)
	res, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf(loginWarning)
	}
	defer res.Body.Close()
	type ApAssoc struct {
		VlanID string `json:"vlan-id"`
	}
	type resResult map[string][]ApAssoc
	var result resResult
	json.NewDecoder(res.Body).Decode(&result)
	return len(result["Association Table"]), nil
}
