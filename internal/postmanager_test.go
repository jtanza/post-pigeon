package internal

import (
	"testing"
)

const pubKey = "-----BEGIN PUBLIC KEY-----\nMIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQAdI8T8Vfccs6rWACR3b5o3MuVkYjf\ngN2nnYAXYNC4fIVWgyfEeTYIGIjLxEB9BLquMld4Je+1vITaNQWfuRTD2HcBax6N\nRwxwcNGqwoJNWpCry9AXxRiDACkks9I2f08BIIHlOCLnPUfIWrASmuNGhyWtSUtA\nJrEKBzI+y/fyWp7z09U=\n-----END PUBLIC KEY-----"

func TestGenerateDeterministicUUID(t *testing.T) {
	uuid, err := GenerateDeterministicUUID(pubKey, "hello world")
	if err != nil {
		t.Error(err)
	}
	expected := "259b8c45-ef29-5616-9c81-3c4f1dcc70ba"
	if uuid != expected {
		t.Error("non matching uuid")
	}

	uuid, err = GenerateDeterministicUUID(pubKey, "hello world")
	if err != nil {
		t.Error(err)
	}
	if uuid != expected {
		t.Error("non matching uuid")
	}
}

func TestGenerateDeterministicUUIDUnique(t *testing.T) {
	uuid, err := GenerateDeterministicUUID(pubKey, "hello world")
	if err != nil {
		t.Error(err)
	}

	uuid2, err := GenerateDeterministicUUID(pubKey, "not hello world")
	if err != nil {
		t.Error(err)
	}

	if uuid == uuid2 {
		t.Error("uuids are not unique for different inputs")
	}
}

func TestParseExpiration(t *testing.T) {
	if expiration := ParseExpiration("junk"); expiration != nil {
		t.Error("non standard expirations should return nil")
	}
}
