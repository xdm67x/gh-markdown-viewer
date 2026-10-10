package ghpr

import (
	"testing"
)

func TestParseRef(t *testing.T) {
	tests := []struct {
		arg          string
		defaultOwner string
		defaultRepo  string
		wantOwner    string
		wantRepo     string
		wantNumber   int
		wantErr      bool
	}{
		{
			arg:        "https://github.com/facebook/react/pull/1234",
			wantOwner:  "facebook",
			wantRepo:   "react",
			wantNumber: 1234,
		},
		{
			arg:        "golang/go#54321",
			wantOwner:  "golang",
			wantRepo:   "go",
			wantNumber: 54321,
		},
		{
			arg:          "42",
			defaultOwner: "owner",
			defaultRepo:  "repo",
			wantOwner:    "owner",
			wantRepo:     "repo",
			wantNumber:   42,
		},
		{
			arg:     "invalid-ref",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		got, err := ParseRef(tt.arg, tt.defaultOwner, tt.defaultRepo)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseRef(%q) expected error, got nil", tt.arg)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseRef(%q) unexpected error: %v", tt.arg, err)
			continue
		}
		if got.Owner != tt.wantOwner || got.Repo != tt.wantRepo || got.Number != tt.wantNumber {
			t.Errorf("ParseRef(%q) = %+v, want %s/%s#%d",
				tt.arg, got, tt.wantOwner, tt.wantRepo, tt.wantNumber)
		}
	}
}

func TestSplitOwnerRepo(t *testing.T) {
	owner, repo, err := SplitOwnerRepo("octocat/Hello-World")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owner != "octocat" || repo != "Hello-World" {
		t.Errorf("got %s/%s, want octocat/Hello-World", owner, repo)
	}

	_, _, err = SplitOwnerRepo("invalid")
	if err == nil {
		t.Errorf("expected error for 'invalid', got nil")
	}
}
