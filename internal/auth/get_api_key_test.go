package auth

import (
	"errors"
	"net/http"
	"testing"
)

func TestGetAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{name: "valid key", header: "ApiKey abc123", want: "abc123"},
		{name: "no header", header: "", wantErr: true},
		{name: "wrong scheme", header: "Bearer abc123", wantErr: true},
		{name: "missing key", header: "ApiKey", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			if tt.header != "" {
				headers.Set("Authorization", tt.header)
			}
			got, err := GetAPIKey(headers)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetAPIKeyNoHeaderError(t *testing.T) {
	_, err := GetAPIKey(http.Header{})
	if !errors.Is(err, ErrNoAuthHeaderIncluded) {
		t.Errorf("err = %v, want ErrNoAuthHeaderIncluded", err)
	}
}
