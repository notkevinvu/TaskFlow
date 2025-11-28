package validation

import (
	"strings"
	"testing"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
		},
		{
			name:    "invalid format - no @",
			email:   "userexample.com",
			wantErr: true,
		},
		{
			name:    "invalid format - no domain",
			email:   "user@",
			wantErr: true,
		},
		{
			name:    "too long",
			email:   strings.Repeat("a", 250) + "@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				// Ensure it returns a ValidationError
				if _, ok := err.(*domain.ValidationError); !ok {
					t.Errorf("ValidateEmail() returned error type %T, want *domain.ValidationError", err)
				}
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		password string
		wantErr bool
	}{
		{
			name:     "valid password",
			password: "Test1234",
			wantErr:  false,
		},
		{
			name:     "valid complex password",
			password: "MySecure123Password!",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Test12",
			wantErr:  true,
		},
		{
			name:     "no uppercase",
			password: "test1234",
			wantErr:  true,
		},
		{
			name:     "no lowercase",
			password: "TEST1234",
			wantErr:  true,
		},
		{
			name:     "no number",
			password: "TestPassword",
			wantErr:  true,
		},
		{
			name:     "too long",
			password: strings.Repeat("Test1", 25), // 125 chars
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				// Ensure it returns a ValidationError
				if _, ok := err.(*domain.ValidationError); !ok {
					t.Errorf("ValidatePassword() returned error type %T, want *domain.ValidationError", err)
				}
			}
		})
	}
}

func TestValidatePriority(t *testing.T) {
	tests := []struct {
		name     string
		priority int
		wantErr  bool
	}{
		{name: "valid min", priority: 1, wantErr: false},
		{name: "valid mid", priority: 5, wantErr: false},
		{name: "valid max", priority: 10, wantErr: false},
		{name: "too low", priority: 0, wantErr: true},
		{name: "too high", priority: 11, wantErr: true},
		{name: "negative", priority: -1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePriority(tt.priority)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePriority() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCategory(t *testing.T) {
	validCategory := "Work"
	invalidCategory := "Invalid@Category!"
	tooLong := strings.Repeat("a", 51)
	empty := ""

	tests := []struct {
		name     string
		category *string
		wantErr  bool
		wantNil  bool
	}{
		{name: "nil category", category: nil, wantErr: false, wantNil: true},
		{name: "empty string", category: &empty, wantErr: false, wantNil: true},
		{name: "valid category", category: &validCategory, wantErr: false, wantNil: false},
		{name: "invalid chars", category: &invalidCategory, wantErr: true},
		{name: "too long", category: &tooLong, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateCategory(tt.category)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCategory() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantNil && result != nil {
				t.Errorf("ValidateCategory() result = %v, want nil", result)
			}
		})
	}
}

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxLength int
		want      string
		wantErr   bool
	}{
		{
			name:      "normal text",
			text:      "Hello World",
			maxLength: 100,
			want:      "Hello World",
			wantErr:   false,
		},
		{
			name:      "trim whitespace",
			text:      "  Hello World  ",
			maxLength: 100,
			want:      "Hello World",
			wantErr:   false,
		},
		{
			name:      "too long",
			text:      strings.Repeat("a", 101),
			maxLength: 100,
			want:      "",
			wantErr:   true,
		},
		{
			name:      "control character",
			text:      "Hello\x00World",
			maxLength: 100,
			want:      "",
			wantErr:   true,
		},
		{
			name:      "unicode characters",
			text:      "Hello 世界",
			maxLength: 100,
			want:      "Hello 世界",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeText(tt.text, tt.maxLength, "test_field")
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SanitizeText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateRequiredText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{name: "valid text", text: "Hello", wantErr: false},
		{name: "empty", text: "", wantErr: true},
		{name: "only spaces", text: "   ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateRequiredText(tt.text, 100, "test_field")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequiredText() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		maxItems int
		wantErr  bool
	}{
		{
			name:     "valid slice",
			slice:    []string{"one", "two", "three"},
			maxItems: 5,
			wantErr:  false,
		},
		{
			name:     "too many items",
			slice:    []string{"one", "two", "three", "four", "five", "six"},
			maxItems: 5,
			wantErr:  true,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			maxItems: 5,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateStringSlice(tt.slice, 100, tt.maxItems, "test_field")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
