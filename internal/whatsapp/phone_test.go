package whatsapp

import "testing"

func TestNormalizeIndian(t *testing.T) {
	valid := map[string]string{
		"9876543210":       "919876543210",
		"+91 98765 43210":  "919876543210",
		"919876543210":     "919876543210",
		"09876543210":      "919876543210",
		"98765-43210":      "919876543210",
		"(+91) 6123456789": "916123456789",
	}
	for in, want := range valid {
		got, err := NormalizeIndian(in)
		if err != nil || got != want {
			t.Errorf("NormalizeIndian(%q) = %q, %v; want %q", in, got, err, want)
		}
	}

	invalid := map[string]error{
		"":              ErrNoPhone,
		"  - ":          ErrNoPhone,
		"12345":         ErrInvalidPhone,
		"5876543210":    ErrInvalidPhone,
		"98765432101":   ErrInvalidPhone,
		"+1 4155552671": ErrInvalidPhone,
	}
	for in, want := range invalid {
		if _, err := NormalizeIndian(in); err != want {
			t.Errorf("NormalizeIndian(%q) error = %v; want %v", in, err, want)
		}
	}
}
