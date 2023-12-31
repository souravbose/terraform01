// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validate

import (
	"strings"
	"testing"
)

func TestDNSZoneSOARecordEmail(t *testing.T) {
	cases := []struct {
		Value  string
		Errors int
	}{
		{
			Value:  "",
			Errors: 1,
		},
		{
			Value:  "a..com",
			Errors: 1,
		},
		{
			Value:  ".a.com",
			Errors: 1,
		},
		{
			Value:  "a.com.",
			Errors: 1,
		},
		{
			Value:  "a",
			Errors: 1,
		},
		{
			Value:  "a@.com.",
			Errors: 1,
		},
		{
			Value:  "a.com",
			Errors: 0,
		},
		{
			Value:  strings.Repeat("a.", 33) + "com",
			Errors: 0,
		},
		{
			Value:  strings.Repeat("a.", 34) + "com",
			Errors: 1,
		},
		{
			Value:  "a-b.com",
			Errors: 0,
		},
		{
			Value:  strings.Repeat("s", 63) + ".com",
			Errors: 0,
		},
		{
			Value:  strings.Repeat("s", 64) + ".com",
			Errors: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Value, func(t *testing.T) {
			_, errors := DnsZoneSOARecordEmail(tc.Value, "email")

			if len(errors) != tc.Errors {
				t.Fatalf("Expected DNSZoneSOARecordEmail to return %d error(s) not %d", tc.Errors, len(errors))
			}
		})
	}
}
