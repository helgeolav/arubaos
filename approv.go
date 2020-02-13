package arubaos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// ApProv the Type Needed to by the ProvAPs Method
type ApProv struct {
	MacAddr string
	Name    string
	Group   string
}

/*
	{
		"_list": [
				{ ap_rename, ap_regroup },
				{ ap_rename, ap_regroup }
		]
	}
*/

// ProvAPs provisions the AP Name and AP Group
// This can only be performed using the MM
func (c *Client) ProvAPs(newAPs []ApProv) error {
	if c.cookie == nil {
		return fmt.Errorf(loginWarning)
	}
	type apRenameReq struct {
		MacAddr string `json:"wired-mac"`
		Name    string `json:"new-name"`
	}
	type apRegroupReq struct {
		MacAddr string `json:"wired-mac"`
		Group   string `json:"new-group"`
	}
	type apConfList struct {
		APRename  apRenameReq  `json:"ap_rename"`
		APRegroup apRegroupReq `json:"ap_regroup"`
	}
	type apProvision struct {
		APConfList []apConfList `json:"_list"`
	}

	var apConf []apConfList
	for _, newAP := range newAPs {
		apConf = append(apConf, apConfList{
			APRename: apRenameReq{
				MacAddr: newAP.MacAddr,
				Name:    newAP.Name,
			},
			APRegroup: apRegroupReq{
				MacAddr: newAP.MacAddr,
				Group:   newAP.Group,
			},
		})
	}
	apProv := apProvision{APConfList: apConf}

	jdata, _ := json.Marshal(apProv)
	body := strings.NewReader(string(jdata))

	endpoint := "/configuration/object"
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, body)

	// Set Appropriate Values needed for the Req to Succeed
	c.updateReq(req, map[string]string{})

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	result, err := ioutil.ReadAll(res.Body)
	fmt.Println(string(result))
	defer res.Body.Close()
	return nil
}
