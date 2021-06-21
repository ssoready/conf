package conf

import "testing"

func TestEnvify(t *testing.T) {
	// These test cases are adapted from:
	//
	// https://github.com/segmentio/conf/blob/master/snakecase_test.go
	testCases := []struct {
		in  string
		out string
	}{
		{"", ""},
		{"A", "A"},
		{"HelloWorld", "HELLO_WORLD"},
		{"HELLOWorld", "HELLO_WORLD"},
		{"Hello1World2", "HELLO1_WORLD2"},
		{"123_", "123_"},
		{"_", "_"},
		{"___", "___"},
		{"HELLO_WORLD", "HELLO_WORLD"},
		{"HelloWORLD", "HELLO_WORLD"},
		{"test_P_x", "TEST_P_X"},
		{"__hello_world__", "__HELLO_WORLD__"},
		{"__Hello_World__", "__HELLO_WORLD__"},
		{"__Hello__World__", "__HELLO__WORLD__"},
		{"hello-world", "HELLO_WORLD"},
	}

	for _, tt := range testCases {
		t.Run(tt.in, func(t *testing.T) {
			got := envify(tt.in)
			if got != tt.out {
				t.Fatalf("want != got, want: %v, got: %v", tt.out, got)
			}
		})
	}
}
