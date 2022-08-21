package main

import (
	"fmt"
	"log"
	"os"

	"github.com/helgeolav/arubaos"
	"github.com/subosito/gotenv"
)

var host, user, pass string

func init() {
	gotenv.Load()
	host = os.Getenv("ARUBA_HOST")
	user = os.Getenv("ARUBA_USER")
	pass = os.Getenv("ARUBA_PASS")
}

func main() {
	client := arubaos.New(host, user, pass, true)
	err := client.Login()
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer client.Logout()

	// intf, _ := client.GetApPortStatus("d0:d3:e0:c1:3b:e0")
	wiredMac := "d0:d3:e0c1:3d:1e"
	intf, _ := client.GetApPortStatus(wiredMac)
	fmt.Println(intf)
	// lldp, err := client.GetApLLDPInfo("ap01.mst.jeffersonn.705.mo")
	// if err != nil {
	// 	log.Fatalf("%v", err)
	// }
	// fmt.Println(lldp)
	// ass, _ := client.GetApAssocCount("ap01.mst.jeffersonn.705.mo")
	// fmt.Println(ass)
}

func something(client *arubaos.Client) {
	f := arubaos.AFilter{
		Count:   2000,
		CfgPath: "/md/ResNet/Birmingham_Souhtern_College",
	}

	aps, _ := client.GetMMApDB(f)
	fmt.Println(len(aps))

	// ap, _ := client.GetAp("ap01.bsc.olin.102.al")
	// fmt.Println(ap)
	// apAssocs, _ := client.GetApAssocCount("ap01.bsc.hilltopapt.33-2-c.al")
	// fmt.Println(apAssocs)
}
