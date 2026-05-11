package adapters

import (
	"testing"
)

// TestInstallOpts_Struct verifies the InstallOpts struct used by the adapter interface.
func TestInstallOpts_Struct(t *testing.T) {
	t.Parallel()

	t.Run("Language field is present and settable", func(t *testing.T) {
		opts := InstallOpts{Language: "es"}
		if opts.Language != "es" {
			t.Errorf("expected Language='es', got Language=%q", opts.Language)
		}
	})

	t.Run("zero value has empty Language", func(t *testing.T) {
		opts := InstallOpts{}
		if opts.Language != "" {
			t.Errorf("expected empty Language, got %q", opts.Language)
		}
	})

	t.Run("different language values", func(t *testing.T) {
		tests := []struct {
			lang string
		}{
			{"en"},
			{"es"},
			{"pt-BR"},
			{"fr"},
			{"de"},
		}
		for _, tc := range tests {
			t.Run(tc.lang, func(t *testing.T) {
				opts := InstallOpts{Language: tc.lang}
				if opts.Language != tc.lang {
					t.Errorf("InstallOpts.Language = %q, want %q", opts.Language, tc.lang)
				}
			})
		}
	})
}
