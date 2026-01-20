package config

import (
	"testing"
)

func TestColorModeString(t *testing.T) {
	tests := []struct {
		mode ColorMode
		want string
	}{
		{Color16, "16 colors"},
		{Color256, "256 colors"},
		{ColorTrueColor, "TrueColor (24-bit)"},
		{ColorMode(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("ColorMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestShouldUseASCII(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name        string
		utf8Support bool
		override    *bool
		want        bool
	}{
		{"UTF8 supported, no override", true, nil, false},
		{"UTF8 not supported, no override", false, nil, true},
		{"UTF8 supported, override true", true, &trueVal, true},
		{"UTF8 supported, override false", true, &falseVal, false},
		{"UTF8 not supported, override false", false, &falseVal, false},
		{"UTF8 not supported, override true", false, &trueVal, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caps := &TermCapabilities{UTF8Support: tt.utf8Support}
			if got := caps.ShouldUseASCII(tt.override); got != tt.want {
				t.Errorf("ShouldUseASCII() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldUseTrueColor(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name      string
		colorMode ColorMode
		override  *bool
		want      bool
	}{
		{"TrueColor, no override", ColorTrueColor, nil, true},
		{"256 color, no override", Color256, nil, false},
		{"16 color, no override", Color16, nil, false},
		{"256 color, override true", Color256, &trueVal, true},
		{"TrueColor, override false", ColorTrueColor, &falseVal, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caps := &TermCapabilities{ColorMode: tt.colorMode}
			if got := caps.ShouldUseTrueColor(tt.override); got != tt.want {
				t.Errorf("ShouldUseTrueColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCapabilities(t *testing.T) {
	// Reset global state
	GlobalCapabilities = nil

	caps := GetCapabilities()
	if caps == nil {
		t.Fatal("GetCapabilities() returned nil")
	}

	// Should return same instance on subsequent calls
	caps2 := GetCapabilities()
	if caps != caps2 {
		t.Error("GetCapabilities() should return same instance")
	}
}

func TestInitCapabilities(t *testing.T) {
	// Reset global state
	GlobalCapabilities = nil

	InitCapabilities()

	if GlobalCapabilities == nil {
		t.Fatal("InitCapabilities() did not set GlobalCapabilities")
	}
}

func TestDetectCapabilities(t *testing.T) {
	caps := DetectCapabilities()

	if caps == nil {
		t.Fatal("DetectCapabilities() returned nil")
	}

	// Just verify the struct is populated (actual values depend on environment)
	// ColorMode should be a valid value
	if caps.ColorMode < Color16 || caps.ColorMode > ColorTrueColor {
		t.Errorf("DetectCapabilities().ColorMode = %d, out of valid range", caps.ColorMode)
	}
}
