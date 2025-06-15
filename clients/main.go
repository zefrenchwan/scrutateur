package main

import (
	"fmt"
	"time"

	"github.com/zefrenchwan/scrutateur.git/clients/clients"
)

const USER_PASSWORD = "root"

// prove basic features (connection, user management)
func validateUserAuthSystem(session clients.ClientSession) {
	connectionStart := time.Now()
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
	} else if err := session.SetUserPassword(USER_PASSWORD); err != nil {
		panic(err)
	} else {
		fmt.Println("Changed password twice (took ", time.Since(connectionStart), ")")
		fmt.Println()
	}

	connectionStart = time.Now()
	access := map[string][]string{
		"self":       {"reader", "editor", "admin", "root"},
		"management": {},
	}
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

	// to put as final content
	fmt.Println()
	fmt.Println("RELEASE NOTES")
	connectionStart = time.Now()
	if body, err := clients.LoadStaticContent("changelog.txt"); err != nil {
		panic(err)
	} else {
		fmt.Println(string(body))
		fmt.Println()
		fmt.Println("Loaded constant file (took ", time.Since(connectionStart), ")")
	}
}

func validateUsersGroups(session clients.ClientSession) {
	connectionStart := time.Now()
	if err := session.CreateGroupOfUsers("developers"); err != nil {
		panic(err)
	}

	if values, err := session.GetCurrentUserGroups(); err != nil {
		panic(err)
	} else {
		fmt.Println("Current user groups:", values)
	}

	if err := session.DeleteGroupOfUsers("developers"); err != nil {
		panic(err)
	}

	fmt.Println("Created and deleted group (took ", time.Since(connectionStart), ")")
}

func main() {
	session, errConnection := clients.Connect("root", USER_PASSWORD)
	if errConnection != nil {
		panic(errConnection)
	}

	validateUserAuthSystem(session)
	validateUsersGroups(session)
}
