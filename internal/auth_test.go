package internal_test

import (
	"github.com/jtanza/post-pigeon/internal"
	"testing"
)

const (
	plaintextMessage = "HELLO"
	base64Signature  = "MIGIAkIA1kTl7BljHlrQ6uL04hGavPXWv+g1/NOBhPqRwldmg5pjPhC3YFxxnMtBNkfJcZJPxxNcsu9Ydr8KCej3wR+yHu4CQgH18fTvqze6qo3Z1q13m1Cjwz2BnFf9ZY6cPRLuIP6NIXsi0nbqeAHzcZqaayGa5Rm1ouzBCnCkAoxLn6hN0nT9vQ==\n"
	pubKey           = "-----BEGIN PUBLIC KEY-----\nMIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQAdI8T8Vfccs6rWACR3b5o3MuVkYjf\ngN2nnYAXYNC4fIVWgyfEeTYIGIjLxEB9BLquMld4Je+1vITaNQWfuRTD2HcBax6N\nRwxwcNGqwoJNWpCry9AXxRiDACkks9I2f08BIIHlOCLnPUfIWrASmuNGhyWtSUtA\nJrEKBzI+y/fyWp7z09U=\n-----END PUBLIC KEY-----"
)

func TestValidateSignatureVerifiesValidMessage(t *testing.T) {
	if err := internal.ValidateSignature(pubKey, base64Signature, plaintextMessage); err != nil {
		t.Error(err)
	}
}

func TestValidateSignatureFailsInvalidSignature(t *testing.T) {
	badSignature := base64Signature[5:]
	if err := internal.ValidateSignature(pubKey, badSignature, plaintextMessage); err == nil {
		t.Error(err)
	}
}

func TestValidateSignatureFailsInvalidKey(t *testing.T) {
	badKey := pubKey[5:]
	if err := internal.ValidateSignature(badKey, base64Signature, plaintextMessage); err == nil {
		t.Error(err)
	}
}

func TestValidateSignatureFailsInvalidMessage(t *testing.T) {
	badMessage := plaintextMessage[5:]
	if err := internal.ValidateSignature(pubKey, base64Signature, badMessage); err == nil {
		t.Error(err)
	}
}

func TestValidateSignatureVerifiesValidMessageNewline(t *testing.T) {
	m := "<pre> &lt;!DOCTYPE html&gt; &lt;html&gt; &lt;head&gt;\n&lt;title&gt;Hello World&lt;/title&gt; &lt;/head&gt; &lt;body&gt; &lt;p&gt;Lorem\nipsum dolor sit amet, consectetur adipiscing elit. Donec viverra nec nulla vitae\nmollis.&lt;/p&gt; &lt;/body&gt; &lt;/html&gt; </pre>"
	sig := "MIGIAkIB9Ll8gnRfPl6Z/FQnfRGcLAMeHbI9bbk6EZKUpnex9MxVczKVLiLNRR6cjzc0Rs4L9YSnRBP0E2N7CuOq8V+zWysCQgDhUtDTVBed2AnydbK4Qm+eY54EpjzRfTkUB9ksJ8slUdCHDXaJcWCLriqRZH5Dq2yfLHt6nlkfUv+R4YiBFzXyEQ=="
	if err := internal.ValidateSignature(pubKey, sig, m); err != nil {
		t.Error(err)
	}

	err := internal.ValidateSignature("-----BEGIN PUBLIC KEY-----\nMIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQAdI8T8Vfccs6rWACR3b5o3MuVkYjf\ngN2nnYAXYNC4fIVWgyfEeTYIGIjLxEB9BLquMld4Je+1vITaNQWfuRTD2HcBax6N\nRwxwcNGqwoJNWpCry9AXxRiDACkks9I2f08BIIHlOCLnPUfIWrASmuNGhyWtSUtA\nJrEKBzI+y/fyWp7z09U=\n-----END PUBLIC KEY-----", "MIGHAkIBTvNSt5cICA3K74uAOKXZqZWL3uSvi5tP2CB/oaTK1X/F1A5hd8WCcWWDKIXgDpLcAd4zh5s7qvqfDGkXoyK2nQgCQSraxRd+EpjkFgMswyxi4Oz7yWOxOh9Urq2aUoaYRTmmWAovLc4aDXkdihbPqDZP7USLajAZEEvU1qcPZ8WXnfIh", "<pre> &lt;!DOCTYPE html&gt; &lt;html&gt; &lt;head&gt;\n&lt;title&gt;Hello World&lt;/title&gt; &lt;/head&gt; &lt;body&gt; &lt;p&gt;Lorem\nipsum dolor sit amet, consectetur adipiscing elit. Donec viverra nec nulla vitae\nmollis.&lt;/p&gt; &lt;/body&gt; &lt;/html&gt; </pre>")
	if err != nil {
		t.Error(err)
	}

}
