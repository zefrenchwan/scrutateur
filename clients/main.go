package main

import (
	"fmt"

	"github.com/zefrenchwan/scrutateur.git/clients/clients"
)

func main() {
	session, errConnection := clients.Connect("root", "root")
	if errConnection != nil {
		panic(errConnection)
	}

	identity, errDetails := session.GetUsername()
	if errDetails != nil {
		panic(errDetails)
	} else {
		fmt.Println("Hello " + identity)
	}
}
