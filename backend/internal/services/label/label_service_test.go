package label

import "testing"

func TestResolveTikTokDocumentType(t *testing.T) {
	tests := []struct {
		name    string
		options LabelOptions
		expects string
	}{
		{
			name:    "default shipping label",
			options: LabelOptions{},
			expects: "SHIPPING_LABEL",
		},
		{
			name: "include products enabled",
			options: LabelOptions{
				IncludeProducts: true,
			},
			expects: "SHIPPING_LABEL_AND_PACKING_SLIP",
		},
		{
			name: "explicit document type wins",
			options: LabelOptions{
				IncludeProducts:    true,
				TikTokDocumentType: "PACKING_SLIP",
			},
			expects: "PACKING_SLIP",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := resolveTikTokDocumentType(tc.options)
			if actual != tc.expects {
				t.Fatalf("expected %s, got %s", tc.expects, actual)
			}
		})
	}
}
