package tgsession

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractShortSession(t *testing.T) {
	fullSessionJSON := `{
  "Version": 1,
  "Data": {
    "Config": {
      "BlockedMode": false,
      "ForceTryIpv6": false,
      "Date": 1753814180,
      "Expires": 1753817434,
      "TestMode": false,
      "ThisDC": 4,
      "DCOptions": [
        {
          "Flags": 0,
          "Ipv6": false,
          "MediaOnly": false,
          "TCPObfuscatedOnly": false,
          "CDN": false,
          "Static": false,
          "ThisPortOnly": false,
          "ID": 1,
          "IPAddress": "149.154.175.54",
          "Port": 443,
          "Secret": null
        },
        {
          "Flags": 16,
          "Ipv6": false,
          "MediaOnly": false,
          "TCPObfuscatedOnly": false,
          "CDN": false,
          "Static": true,
          "ThisPortOnly": false,
          "ID": 2,
          "IPAddress": "149.154.167.41",
          "Port": 443,
          "Secret": null
        }
      ],
      "DCTxtDomainName": "apv3.stel.com",
      "TmpSessions": 0,
      "WebfileDCID": 4
    },
    "DC": 4,
    "Addr": "149.154.167.92:443",
    "AuthKey": "dGVzdC1hdXRoLWtleQ==",
    "AuthKeyID": "12345678901234567890",
    "Salt": 123456789
  }
}`

	expectedShortSessionJSON := `{
  "Version": 1,
  "Data": {
    "DC": 4,
    "Addr": "149.154.167.92:443",
    "AuthKey": "dGVzdC1hdXRoLWtleQ==",
    "AuthKeyID": "12345678901234567890",
    "Salt": 123456789
  }
}`

	result, err := TrimSession([]byte(fullSessionJSON))

	require.NoError(t, err, "Should not return an error for valid JSON")
	assert.JSONEq(t, expectedShortSessionJSON, string(result), "Result should match expected short session JSON exactly")
}

func TestExtractShortSession_InvalidJSON(t *testing.T) {
	invalidJSON := `{"Version": 1, "Data": {invalid json}}`

	result, err := TrimSession([]byte(invalidJSON))

	assert.Error(t, err, "Should return error for invalid JSON")
	assert.Nil(t, result, "Result should be nil for invalid JSON")
}
