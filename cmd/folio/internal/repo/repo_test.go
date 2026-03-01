package repo

import (
	"errors"
	"testing"
)

func TestValidateCommitMessage(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		wantErr bool
	}{
		// Valid
		{"basic feat", "feat(srm): add spike sources", false},
		{"docs update", "docs(folio): update readme", false},
		{"auto codegen", "auto(codegen): graphqlme", false},
		{"fix with dots in scope", "fix(ben.payroll): handle nil", false},
		{"chore with hyphen scope", "chore(ci-lint): add step", false},
		{"multi-line with body", "feat(folio): add validation\n\nMore details here.", false},

		// Invalid
		{"empty", "", true},
		{"old default", "folio: update", true},
		{"no scope", "feat: add thing", true},
		{"uppercase description", "feat(bar): Thing", true},
		{"trailing period", "docs(folio): update readme.", true},
		{"bare text", "update stuff", true},
		{"bad type", "yolo(foo): do thing", true},
		{"uppercase scope", "feat(Foo): bar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommitMessage(tt.msg)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for %q, got nil", tt.msg)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for %q: %v", tt.msg, err)
			}
			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidCommitMessage) {
				t.Errorf("expected ErrInvalidCommitMessage, got: %v", err)
			}
		})
	}
}
