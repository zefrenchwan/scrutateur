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
