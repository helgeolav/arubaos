package main

import (
	"fmt"
	"log"

	"github.com/drkchiloll/go-arubaos"
)

func main() {
	wlclient := arubaos.New(
		"controller_ip",
		"your_user",
		"your_pass",
		true,
	)

	var aps arubaos.APDatabase
	aps, err := wlclient.GetAPDatabase()
	if err != nil {
		log.Fatalf("%v", err)
	}
	fmt.Println(aps)
	wlclient.Logout()
}
