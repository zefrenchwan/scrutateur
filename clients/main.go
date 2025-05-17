package main

import (
	"fmt"
	"time"

	"github.com/zefrenchwan/scrutateur.git/clients/clients"
)

func main() {
	currentPassword := "root"
	connectionStart := time.Now()
	session, errConnection := clients.Connect("root", currentPassword)
	if errConnection != nil {
		panic(errConnection)
	}

	identity, errDetails := session.GetUsername()
	if errDetails != nil {
		panic(errDetails)
	} else {
		fmt.Println("Hello "+identity, "(took ", time.Since(connectionStart), ")")
	}

	connectionStart = time.Now()
	newPassword := "popo"
	if err := session.SetUserPassword(newPassword); err != nil {
		panic(err)
	} else if err := session.SetUserPassword(currentPassword); err != nil {
		panic(err)
	} else {
		fmt.Println("Changed password twice (took ", time.Since(connectionStart), ")")
	}

	connectionStart = time.Now()
	var username = "other"
	if err := session.AddUser(username, "secret"); err != nil {
		panic(err)
	} else {
		fmt.Println("Created new user (took ", time.Since(connectionStart), ")")
	}

}
