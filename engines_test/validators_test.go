package services_test

import (
	"testing"

	"github.com/zefrenchwan/scrutateur.git/engines"
)

func TestValidUsername(t *testing.T) {
	if !engines.ValidateUsernameFormat("popo") {
		t.Log("valid with 4 chars at least")
		t.Fail()
	}

	if !engines.ValidateUsernameFormat("other") {
		t.Log("valid with 5 chars > 4 chars")
		t.Fail()
	}

	if !engines.ValidateUsernameFormat("popo01") {
		t.Log("Valid with number at the end")
		t.Fail()
	}
}

func TestInvalidUsername(t *testing.T) {
	if engines.ValidateUsernameFormat("p") {
		t.Log("Should be at least 4 chars")
		t.Fail()
	}

	if engines.ValidateUsernameFormat("01popo01") {
		t.Log("Should refuse first number")
		t.Fail()
	}
}
