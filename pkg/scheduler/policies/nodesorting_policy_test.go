/*
 Licensed to the Apache Software Foundation (ASF) under one
 or more contributor license agreements.  See the NOTICE file
 distributed with this work for additional information
 regarding copyright ownership.  The ASF licenses this file
 to you under the Apache License, Version 2.0 (the
 "License"); you may not use this file except in compliance
 with the License.  You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package policies

import (
	"testing"
)

func TestSortingPolicyFromString(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    SortingPolicy
		wantErr bool
	}{
		{"EmptyString", "", FairnessPolicy, false},
		{"FairString", "fair", FairnessPolicy, false},
		{"BinString", "binpacking", BinPackingPolicy, false},
		{"AgingString", "fairwithaging", FairWithAgingNodePolicy, false},
		{"UnknownString", "unknown", FairnessPolicy, true},
	}
	for _, tt := range tests {
		got, err := SortingPolicyFromString(tt.arg)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s unexpected error returned, expected error: %t, got error '%v'", tt.name, tt.wantErr, err)
			return
		}
		if got != tt.want {
			t.Errorf("%s unexpected string returned, expected string: '%s', got string '%v'", tt.name, tt.want, got)
		}
	}
}

func TestSortingPolicyToString(t *testing.T) {
	var someSP SortingPolicy // since SortingPolicy is an iota it defaults to first in the list
	tests := []struct {
		name string
		sp   SortingPolicy
		want string
	}{
		{"FairString", FairnessPolicy, "fair"},
		{"BinString", BinPackingPolicy, "binpacking"},
		{"AgingString", FairWithAgingNodePolicy, "fairwithaging"},
		{"NoneString", someSP, "binpacking"},
	}
	for _, tt := range tests {
		if got := tt.sp.String(); got != tt.want {
			t.Errorf("%s unexpected string returned, expected = '%s', got '%v'", tt.name, tt.want, got)
		}
	}
}
