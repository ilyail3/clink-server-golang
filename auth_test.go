package server

import "testing"

func TestEncryptPassword(t *testing.T) {
	password := "random password"

	bad_password := "random password2"

	encrypted := EncryptPassword(password)

	res, err := ComparePassword(encrypted, password)

	if err != nil {
		t.Errorf("Compare failed:%s", err.Error())
		t.Fail()
	}

	if !res {
		t.Error("Compare failed")
		t.Fail()
	}

	res,err = ComparePassword(encrypted, bad_password)

	if err != nil {
		t.Errorf("Compare failed:%s", err.Error())
		t.Fail()
	}

	if res {
		t.Error("Compare failed, considered equal")
		t.Fail()
	}
}
