//
// Copyright (c) 2023 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package utils

import (
	"testing"

	coreDtos "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/stretchr/testify/assert"
)

func TestSliceToMap(t *testing.T) {
	input := []coreDtos.Device{
		{Name: "D1"},
		{Name: "D2"},
		{Name: "D3"},
	}

	actual := SliceToMap(input, func(d coreDtos.Device) string { return d.Name })
	assert.Len(t, actual, len(input))
	for _, item := range input {
		assert.NotNil(t, actual[item.Name])
	}
}

func TestMapToSlice(t *testing.T) {
	input := map[string]*coreDtos.Device{
		"D1": {Name: "D1"},
		"D2": {Name: "D2"},
		"D3": {Name: "D3"},
	}

	actual := MapToSlice(input)
	assert.Len(t, actual, len(input))
	for _, item := range actual {
		assert.NotNil(t, input[item.Name])
	}
}
