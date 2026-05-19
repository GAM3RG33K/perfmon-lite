package styles

import (
	"strings"
	"testing"

	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
)

func TestDebugBadge(t *testing.T) {
	if DebugBadge == "" {
		t.Error("DebugBadge should not be empty")
	}
	if !strings.Contains(DebugBadge, "DEBUG") {
		t.Errorf("DebugBadge should contain 'DEBUG', got %q", DebugBadge)
	}
}

func TestReleaseBadge(t *testing.T) {
	if ReleaseBadge == "" {
		t.Error("ReleaseBadge should not be empty")
	}
	if !strings.Contains(ReleaseBadge, "RELEASE") {
		t.Errorf("ReleaseBadge should contain 'RELEASE', got %q", ReleaseBadge)
	}
}

func TestBuildBadge(t *testing.T) {
	tests := []struct {
		buildType engine.BuildType
		wantText  string
	}{
		{engine.BuildDebug, "DEBUG"},
		{engine.BuildRelease, "RELEASE"},
		{engine.BuildUnknown, "UNKNOWN"},
		{engine.BuildType(""), "UNKNOWN"},
	}

	for _, tt := range tests {
		got := BuildBadge(tt.buildType)
		if got == "" {
			t.Errorf("BuildBadge(%s) returned empty string", tt.buildType)
		}
		if !strings.Contains(got, tt.wantText) {
			t.Errorf("BuildBadge(%s) should contain %q, got %q", tt.buildType, tt.wantText, got)
		}
	}
}

func TestPlatformBadge(t *testing.T) {
	tests := []struct {
		platform engine.Platform
		wantText string
	}{
		{engine.PlatformAndroid, "android"},
		{engine.PlatformIOS, "ios"},
		{engine.PlatformMock, "mock"},
		{engine.Platform("unknown"), "unknown"},
	}

	for _, tt := range tests {
		got := PlatformBadge(tt.platform)
		if got == "" {
			t.Errorf("PlatformBadge(%s) returned empty string", tt.platform)
		}
		if !strings.Contains(got, tt.wantText) {
			t.Errorf("PlatformBadge(%s) should contain %q, got %q", tt.platform, tt.wantText, got)
		}
	}
}
