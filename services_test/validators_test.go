package services_test

import (
	"testing"

	"github.com/zefrenchwan/scrutateur.git/services"
)

func TestValidUsername(t *testing.T) {
	if !services.ValidateUsernameFormat("popo") {
		t.Log("valid with 4 chars at least")
		t.Fail()
	}

	if !services.ValidateUsernameFormat("other") {
		t.Log("valid with 5 chars > 4 chars")
		t.Fail()
	}

	if !services.ValidateUsernameFormat("popo01") {
		t.Log("Valid with number at the end")
		t.Fail()
	}
}

func TestInvalidUsername(t *testing.T) {
	if services.ValidateUsernameFormat("p") {
		t.Log("Should be at least 4 chars")
		t.Fail()
	}

	if services.ValidateUsernameFormat("01popo01") {
		t.Log("Should refuse first number")
		t.Fail()
	}
}
