package arubaos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// APRenameReq is the object definition needed for MM
// To Rename an AP from Default
type APRenameReq struct {
	MacAddr string `json:"wired-mac"`
	Name    string `json:"new-name"`
}

// APRegroupReq is the object definition needed for MM
// To Reassign an AP to appropriate Group
type APRegroupReq struct {
	MacAddr string `json:"wired-mac"`
	Group   string `json:"new-group"`
}

// APConfList used to Modify Multiple Objects at the same time
type APConfList struct {
	APRename  APRenameReq  `json:"ap_rename"`
	APRegroup APRegroupReq `json:"ap_regroup"`
}

// APProvision configures both APName and APGroup
type APProvision struct {
	APConfList []APConfList `json:"_list"`
}

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
	var apConf []APConfList
	for _, newAP := range newAPs {
		apConf = append(apConf, APConfList{
			APRename: APRenameReq{
				MacAddr: newAP.MacAddr,
				Name:    newAP.Name,
			},
			APRegroup: APRegroupReq{
				MacAddr: newAP.MacAddr,
				Group:   newAP.Group,
			},
		})
	}
	apProv := APProvision{
		APConfList: apConf,
	}
	jdata, _ := json.Marshal(apProv)
	body := strings.NewReader(string(jdata))
	err := c.login()
	if err != nil {
		return fmt.Errorf("%v", err)
	}
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
