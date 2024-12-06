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

func TestOriginalContextCanceled(t *testing.T) {
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

	cancelCtx, cancel := context.WithCancel(context.Background())

	go func() {
		err := extendtimeout.NewConnection(testConn, 10*time.Second).Invoke(cancelCtx, nil, nil)
		require.NoError(t, err)
	}()

	cancel()
	time.Sleep(50 * time.Millisecond)
	ch <- struct{}{}

	require.Eventually(t, func() bool {
		return counter.Load() == 1
	}, 200*time.Millisecond, 10*time.Millisecond)
}
