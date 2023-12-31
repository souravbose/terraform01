// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"encoding/json"
	"fmt"
)

func NormalizeJson(jsonString interface{}) string {
	if jsonString == nil || jsonString == "" {
		return ""
	}
	var j interface{}

	if err := json.Unmarshal([]byte(jsonString.(string)), &j); err != nil {
		return fmt.Sprintf("Error parsing JSON: %+v", err)
	}
	b, _ := json.Marshal(j)
	return string(b)
}
