package mwr

import (
	"bajetapp/model"
	"encoding/json"
	"testing"
)

func TestUnmarshalGoogleUserInfo(t *testing.T) {
	jsonData := `{
        "id": "104146954444783557932",
        "email": "ahmadfarisfs@gmail.com",
        "verified_email": true,
        "name": "Ahmad Faris",
        "given_name": "Ahmad",
        "family_name": "Faris",
        "picture": "https://lh3.googleusercontent.com/a/ACg8ocI7m5ynsuBEcdpiyllo479nvaMtM2j70HfR2NFnMA9pujf91SnUgw=s96-c"
    }`

	var userInfo model.GoogleUserInfo
	err := json.Unmarshal([]byte(jsonData), &userInfo)
	if err != nil {
		t.Fatalf("failed to unmarshal user info: %v", err)
	}

	expected := model.GoogleUserInfo{
		ID:            "104146954444783557932",
		Email:         "ahmadfarisfs@gmail.com",
		VerifiedEmail: true,
		Name:          "Ahmad Faris",
		GivenName:     "Ahmad",
		FamilyName:    "Faris",
		Picture:       "https://lh3.googleusercontent.com/a/ACg8ocI7m5ynsuBEcdpiyllo479nvaMtM2j70HfR2NFnMA9pujf91SnUgw=s96-c",
	}

	if userInfo.ID != expected.ID ||
		userInfo.Email != expected.Email ||
		userInfo.VerifiedEmail != expected.VerifiedEmail ||
		userInfo.Name != expected.Name ||
		userInfo.GivenName != expected.GivenName ||
		userInfo.FamilyName != expected.FamilyName ||
		userInfo.Picture != expected.Picture {
		t.Errorf("unmarshaled user info does not match expected value. Got %+v, expected %+v", userInfo, expected)
	}
}
