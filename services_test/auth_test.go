package services_test

import (
	"testing"

	"github.com/zefrenchwan/scrutateur.git/dto"
	"github.com/zefrenchwan/scrutateur.git/services"
)

func TestMatchesAccept(t *testing.T) {
	rule := dto.GrantAccessForResource{Operator: dto.OperatorMatches, Template: "/root/admin/*", UserRoles: []dto.GrantRole{dto.RoleRoot}}
	engine := services.AuthRulesEngine{Conditions: []dto.GrantAccessForResource{rule}}
	if access, roles, err := engine.CanAccessResource("/root/admin/popo"); err != nil {
		t.Fatal(err)
	} else if !access {
		t.Fail()
	} else if len(roles) != 1 {
		t.Fail()
	} else if roles[0] != dto.RoleRoot {
		t.Fail()
	}
}

func TestMatchesRefuse(t *testing.T) {
	rule := dto.GrantAccessForResource{Operator: dto.OperatorMatches, Template: "/root/admin/*", UserRoles: []dto.GrantRole{dto.RoleRoot}}
	engine := services.AuthRulesEngine{Conditions: []dto.GrantAccessForResource{rule}}
	if access, roles, err := engine.CanAccessResource("/root/admin/popo/"); err != nil {
		t.Fatal(err)
	} else if access {
		t.Fail()
	} else if roles != nil {
		t.Fail()
	}

	if access, roles, err := engine.CanAccessResource("/root/admin/"); err != nil {
		t.Fatal(err)
	} else if access {
		t.Fail()
	} else if roles != nil {
		t.Fail()
	}

	if access, roles, err := engine.CanAccessResource("/root/admin/popo/wawa"); err != nil {
		t.Fatal(err)
	} else if access {
		t.Fail()
	} else if roles != nil {
		t.Fail()
	}
}

func TestGrantInsufficientRole(t *testing.T) {
	adminRoles := []dto.GrantRole{dto.RoleAdmin, dto.RoleReader}
	rootRoles := []dto.GrantRole{dto.RoleAdmin, dto.RoleRoot}
	noAdmin := []dto.GrantRole{dto.RoleEditor, dto.RoleReader}

	// can't grant remote group
	accessRoles := map[string][]dto.GrantRole{"group": adminRoles}
	requestRoles := map[string][]dto.GrantRole{"other": adminRoles}
	if err := services.MayGrant(accessRoles, requestRoles); err == nil {
		t.Log("cannot grant to group we are not in")
		t.Fail()
	}

	// admin can't grant root
	accessRoles = map[string][]dto.GrantRole{
		"group": adminRoles,
	}

	requestRoles = map[string][]dto.GrantRole{"group": rootRoles}

	if err := services.MayGrant(accessRoles, requestRoles); err == nil {
		t.Log("admin cannot grant root")
		t.Fail()
	}

	// special case: not admin
	accessRoles = make(map[string][]dto.GrantRole)
	requestRoles = map[string][]dto.GrantRole{"group": rootRoles}

	if err := services.MayGrant(accessRoles, requestRoles); err == nil {
		t.Log("no admin role")
		t.Fail()
	}

	// need admin access
	accessRoles = map[string][]dto.GrantRole{"group": noAdmin}
	requestRoles = map[string][]dto.GrantRole{"group": noAdmin}
	if err := services.MayGrant(accessRoles, requestRoles); err == nil {
		t.Log("cannot grant to admin if not admin")
		t.Fail()
	}
}

func TestGrant(t *testing.T) {
	adminRoles := []dto.GrantRole{dto.RoleAdmin, dto.RoleReader}
	rootRoles := []dto.GrantRole{dto.RoleAdmin, dto.RoleRoot}
	noAdmin := []dto.GrantRole{dto.RoleEditor, dto.RoleReader}

	// admin can grant reader and editor
	accessRoles := map[string][]dto.GrantRole{"group": adminRoles}
	requestRoles := map[string][]dto.GrantRole{"group": noAdmin}

	if err := services.MayGrant(accessRoles, requestRoles); err != nil {
		t.Log("admin should grant reader and editor", err)
		t.Fail()
	}

	// root can grant reader and editor
	accessRoles = map[string][]dto.GrantRole{"group": rootRoles}
	requestRoles = map[string][]dto.GrantRole{"group": noAdmin}

	if err := services.MayGrant(accessRoles, requestRoles); err != nil {
		t.Log("root should grant reader and editor", err)
		t.Fail()
	}
}
