package label

import "testing"

func TestIsLikelyTikTokOrderID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		expects bool
	}{
		{
			name:    "valid long numeric id",
			input:   "582581679994275029",
			expects: true,
		},
		{
			name:    "short numeric id",
			input:   "123456789012345",
			expects: false,
		},
		{
			name:    "alphanumeric shopee order",
			input:   "260213JTDNDKRS",
			expects: false,
		},
		{
			name:    "numeric with spaces",
			input:   " 582581679994275029 ",
			expects: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := isLikelyTikTokOrderID(tc.input)
			if actual != tc.expects {
				t.Fatalf("expected %v, got %v", tc.expects, actual)
			}
		})
	}
}
