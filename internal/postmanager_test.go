package internal

import (
	"github.com/bluele/gcache"
	"github.com/jtanza/post-pigeon/internal/model"
	"html/template"
	"strings"
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

func TestMarkdownParses(t *testing.T) {
	pm := NewPostManager(DB{nil}, gcache.New(1).LRU().Build())

	data, err := pm.formatRequestData(model.PostRequest{
		Title:      "Foo",
		Body:       "# This is a title",
		PublicKey:  pubKey,
		Signature:  "",
		Expiration: "",
	})
	if err != nil {
		t.Error(err)
	}

	actual := strings.TrimRight(string(data["Body"].(template.HTML)), "\n")
	expected := "<h1 id=\"this-is-a-title\">This is a title</h1>"
	if actual != expected {
		t.Errorf("markdown does not match expected\n got: %s wanted: %s", actual, expected)
	}
}

func TestMarkdownClearsBuffer(t *testing.T) {
	pm := NewPostManager(DB{nil}, gcache.New(1).LRU().Build())

	data, err := pm.formatRequestData(model.PostRequest{
		Title:      "Foo",
		Body:       "a",
		PublicKey:  pubKey,
		Signature:  "",
		Expiration: "",
	})
	if err != nil {
		t.Error(err)
	}

	actual := strings.TrimRight(string(data["Body"].(template.HTML)), "\n")
	expected := "<p>a</p>"
	if actual != expected {
		t.Errorf("markdown does not match expected\n got: %s wanted: %s", actual, expected)
	}

	data2, err := pm.formatRequestData(model.PostRequest{
		Title:      "Foo",
		Body:       "b",
		PublicKey:  pubKey,
		Signature:  "",
		Expiration: "",
	})
	if err != nil {
		t.Error(err)
	}

	actual2 := strings.TrimRight(string(data2["Body"].(template.HTML)), "\n")
	expected2 := "<p>b</p>"
	if actual2 != expected2 {
		t.Errorf("markdown does not match expected\n got: %s wanted: %s", actual2, expected2)
	}
}
