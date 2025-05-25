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
	} else if rolesPerResource, err := session.GetUserRoles(identity); err != nil {
		panic(err)
	} else {
		fmt.Println("Hello "+identity, "(took ", time.Since(connectionStart), ")")
		fmt.Println("current roles: ", rolesPerResource)
		fmt.Println()
	}

	connectionStart = time.Now()
	newPassword := "popo"
	if err := session.SetUserPassword(newPassword); err != nil {
		panic(err)
	} else if err := session.SetUserPassword(currentPassword); err != nil {
		panic(err)
	} else {
		fmt.Println("Changed password twice (took ", time.Since(connectionStart), ")")
		fmt.Println()
	}

	connectionStart = time.Now()
	access := map[string][]string{"self": {"reader", "editor", "admin", "root"}}
	var username = "other"
	if err := session.AddUser(username, "secret"); err != nil {
		panic(err)
	} else if err := session.SetUserRolesForGroups(username, access); err != nil {
		panic(err)
	} else if values, err := session.GetUserRoles(username); err != nil {
		panic(err)
	} else if len(values) == 0 {
		panic(fmt.Errorf("impossible to load access for %s", username))
	} else {
		fmt.Println("current access for ", username, ":", values)
	}

	if err := session.DeleteUser(username); err != nil {
		panic(err)
	} else {
		fmt.Println("Created and deleted new user with basic access (took ", time.Since(connectionStart), ")")
	}
}
