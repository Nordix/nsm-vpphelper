// Copyright (c) 2023 Cisco and/or its affiliates.
// Copyright (c) 2025 OpenInfra Foundation Europe.
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

//go:build !arm && !arm64

package vpphelper

const (
	// For data-size the default 2048 was chosen because with the previous 3776 only
	// 520 buffers were allocated in a pool on VPP v24.10 as a result of buffer pool
	// allocation improvements.
	vppDefaultDataSize = 2048
)
