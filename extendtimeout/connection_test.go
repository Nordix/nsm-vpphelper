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
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.fd.io/govpp/api"
	"go.uber.org/goleak"

	"github.com/networkservicemesh/vpphelper/extendtimeout"
)

type testConn struct {
	api.Connection
	invokeBody func(ctx context.Context)
}

func (c *testConn) Invoke(ctx context.Context, req, reply api.Message) error {
	c.invokeBody(ctx)
	return nil
}

type key struct{}

const value = "value"

func TestSmallTimeout(t *testing.T) {
	testConn := &testConn{invokeBody: func(ctx context.Context) {
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		timeout := time.Until(deadline)
		require.Greater(t, timeout, time.Second)
		require.Equal(t, ctx.Value(&key{}), value)
	}}

	smallCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	smallCtx = context.WithValue(smallCtx, &key{}, value)
	defer cancel()

	err := extendtimeout.NewConnection(testConn, 2*time.Second).Invoke(smallCtx, nil, nil)
	require.NoError(t, err)
}

func TestBigTimeout(t *testing.T) {
	testConn := &testConn{invokeBody: func(ctx context.Context) {
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		timeout := time.Until(deadline)
		require.Greater(t, timeout, 7*time.Second)
		require.Equal(t, ctx.Value(&key{}), value)
	}}

	bigCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	bigCtx = context.WithValue(bigCtx, &key{}, value)
	defer cancel()

	err := extendtimeout.NewConnection(testConn, 2*time.Second).Invoke(bigCtx, nil, nil)
	require.NoError(t, err)
}

func TestNoTimeout(t *testing.T) {
	testConn := &testConn{invokeBody: func(ctx context.Context) {
		_, ok := ctx.Deadline()
		require.False(t, ok)
	}}

	emptyCtx := context.Background()
	err := extendtimeout.NewConnection(testConn, 2*time.Second).Invoke(emptyCtx, nil, nil)
	require.NoError(t, err)
}

func TestCanceledContext(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t)
	})

	counter := new(atomic.Int32)
	ch := make(chan struct{}, 1)
	defer close(ch)
	testConn := &testConn{invokeBody: func(ctx context.Context) {
		select {
		case <-ctx.Done():
			return
		case <-ch:
			counter.Add(1)
		}
	}}

	cancelCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		err := extendtimeout.NewConnection(testConn, 20*time.Second).Invoke(cancelCtx, nil, nil)
		require.NoError(t, err)
	}()

	cancel()
	ch <- struct{}{}

	require.Eventually(t, func() bool {
		return counter.Load() == 1
	}, time.Second, 100*time.Millisecond)
}
