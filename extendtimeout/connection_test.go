// Copyright (c) 2024 Cisco and/or its affiliates.
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

package extendtimeout_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.fd.io/govpp/api"

	"github.com/networkservicemesh/vpphelper/extendtimeout"
)

type testConn struct {
	api.Connection
	invoke func(ctx context.Context)
}

func (c *testConn) Invoke(ctx context.Context, req, reply api.Message) error {
	c.invoke(ctx)
	return nil
}

func TestSmallTimeout(t *testing.T) {
	testConn := &testConn{invoke: func(ctx context.Context) {
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		timeout := time.Until(deadline)
		require.Greater(t, timeout, time.Second)
	}}

	smallCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	err := extendtimeout.NewConnection(testConn, 2*time.Second).Invoke(smallCtx, nil, nil)
	require.NoError(t, err)
}

func TestBigTimeout(t *testing.T) {
	testConn := &testConn{invoke: func(ctx context.Context) {
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		timeout := time.Until(deadline)
		require.Greater(t, timeout, 7*time.Second)
	}}

	bigCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := extendtimeout.NewConnection(testConn, 2*time.Second).Invoke(bigCtx, nil, nil)
	require.NoError(t, err)
}
