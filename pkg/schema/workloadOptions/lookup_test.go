package workloadOptions

import "testing"

func TestLocalOptionsForMatchesTheOrgLocationLink(t *testing.T) {
	opts := LocalOptions{
		{Location: "/org/my-org/location/aws-us-west-2-extra"},
		{Location: "/org/my-org/location/aws-us-west-2"},
		{Location: "gcp-us-east1"},
	}
	cases := map[string]LocationLink{
		"aws-us-west-2": "/org/my-org/location/aws-us-west-2",
		"gcp-us-east1":  "gcp-us-east1",
	}
	for location, want := range cases {
		got := opts.For(location)
		if got == nil || got.Location != want {
			t.Fatalf("For(%q) = %v, want %q", location, got, want)
		}
	}
	for _, location := range []string{"", "aws-us-west", "west-2"} {
		if got := opts.For(location); got != nil {
			t.Fatalf("For(%q) = %v, want nil", location, got.Location)
		}
	}
}
