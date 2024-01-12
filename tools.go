// Copyright (c) 2023-2024 Cisco and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vpphelper

import (
	"strings"
	"text/template"
)

// VPPConfigParameters - custom parameters used by VPP config
type VPPConfigParameters struct {
	DataSize int
	RootDir  string
}

// NewVPPConfigFile creates new VPP config based on parameters
func NewVPPConfigFile(configTemplate string, params VPPConfigParameters) string {
	vppConfigBuilder := new(strings.Builder)

	t := template.Must(template.New("vppConfig").Parse(configTemplate))
	err := t.Execute(vppConfigBuilder, params)
	if err != nil {
		panic(err)
	}
	return vppConfigBuilder.String()
}
