package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ApogeeNetworking/arubaos"
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

	// ap, _ := client.GetAp("ap01.bsc.olin.102.al")
	// fmt.Println(ap)
	apAssocs, _ := client.GetApAssocCount("ap01.bsc.hilltopapt.33-2-c.al")
	fmt.Println(apAssocs)
}
