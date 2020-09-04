package main

import (
	"fmt"
	"log"

	"github.com/ApogeeNetworking/arubaos"
)

func main() {
	wlclient := arubaos.New(
		"controller_ip",
		"your_user",
		"your_pass",
		true,
	)

	var aps []arubaos.AP
	aps, err := wlclient.GetApDB()
	if err != nil {
		log.Fatalf("%v", err)
	}
	fmt.Println(aps)
	wlclient.Logout()
}
