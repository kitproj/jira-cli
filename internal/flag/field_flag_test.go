package flag

import (
	"testing"
)

func TestFieldFlag_String(t *testing.T) {
	f := make(FieldFlag)
	if f.String() != "" {
		t.Errorf("String() should return empty string, got %q", f.String())
	}
}

func TestFieldFlag_Set(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantKey   string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid field=value",
			value:     "Effort Estimate=0",
			wantKey:   "Effort Estimate",
			wantValue: "0",
			wantErr:   false,
		},
		{
			name:      "valid with spaces",
			value:     "Story Points=5",
			wantKey:   "Story Points",
			wantValue: "5",
			wantErr:   false,
		},
		{
			name:      "valid with equals in value",
			value:     "field=value=with=equals",
			wantKey:   "field",
			wantValue: "value=with=equals",
			wantErr:   false,
		},
		{
			name:    "missing equals",
			value:   "noequals",
			wantErr: true,
		},
		{
			name:    "empty value",
			value:   "field=",
			wantKey: "field",
			wantValue: "",
			wantErr: false,
		},
		{
			name:    "empty key",
			value:   "=value",
			wantKey: "",
			wantValue: "value",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := make(FieldFlag)
			err := f.Set(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotValue, ok := f[tt.wantKey]; !ok {
					t.Errorf("Set() key %q not found in map", tt.wantKey)
				} else if gotValue != tt.wantValue {
					t.Errorf("Set() value = %q, want %q", gotValue, tt.wantValue)
				}
			}
		})
	}
}

func TestFieldFlag_Set_Multiple(t *testing.T) {
	f := make(FieldFlag)

	values := []string{
		"Effort Estimate=0",
		"Story Points=5",
		"Priority=High",
	}

	for _, v := range values {
		if err := f.Set(v); err != nil {
			t.Errorf("Set(%q) error = %v", v, err)
		}
	}

	if len(f) != len(values) {
		t.Errorf("Expected %d entries, got %d", len(values), len(f))
	}

	if f["Effort Estimate"] != "0" {
		t.Errorf("Effort Estimate = %q, want %q", f["Effort Estimate"], "0")
	}
	if f["Story Points"] != "5" {
		t.Errorf("Story Points = %q, want %q", f["Story Points"], "5")
	}
	if f["Priority"] != "High" {
		t.Errorf("Priority = %q, want %q", f["Priority"], "High")
	}
}

