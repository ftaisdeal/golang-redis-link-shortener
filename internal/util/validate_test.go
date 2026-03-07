package util

import "testing"

func TestValidateCustomCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{name: "valid", code: "abc-123_DEF", wantErr: false},
		{name: "too short", code: "ab", wantErr: true},
		{name: "invalid char", code: "abc$123", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCustomCode(tc.code)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateCustomCode() error = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateAndNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "valid", url: "https://example.com/hello", wantErr: false},
		{name: "invalid scheme", url: "ftp://example.com", wantErr: true},
		{name: "localhost blocked", url: "http://localhost:8080", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateAndNormalizeURL(tc.url)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateAndNormalizeURL() error = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}
