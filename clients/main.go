package main

import (
	"fmt"

	"github.com/zefrenchwan/scrutateur.git/clients/clients"
)

func main() {
	fmt.Println("Starting a connection")
	session, errConnection := clients.Connect("root", "root")
	if errConnection != nil {
		panic(errConnection)
	}

	details, errDetails := session.GetUserDetails()
	if errDetails != nil {
		panic(errDetails)
	} else {
		fmt.Println(details)
	}
}
