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
	"go.uber.org/goleak"

	"github.com/networkservicemesh/vpphelper/extendtimeout"
)

type testConn struct {
	api.Connection
	invokeBody func(ctx context.Context) error
}

func (c *testConn) Invoke(ctx context.Context, req, reply api.Message) error {
	return c.invokeBody(ctx)
}

func TestTinyTimeout(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t)
	})

	testConn := &testConn{invokeBody: func(ctx context.Context) error {
		time.Sleep(20 * time.Millisecond)
		return ctx.Err()
	}}

	cancelCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	err := extendtimeout.NewConnection(testConn, time.Second).Invoke(cancelCtx, nil, nil)
	require.NoError(t, err)
}

func TestOriginalContextCanceled(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t)
	})

	ch := make(chan struct{}, 1)
	defer close(ch)
	testConn := &testConn{invokeBody: func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ch:
			return nil
		}
	}}

	cancelCtx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	defer close(done)
	var err error
	go func() {
		err = extendtimeout.NewConnection(testConn, time.Second).Invoke(cancelCtx, nil, nil)
		done <- struct{}{}
	}()

	cancel()
	ch <- struct{}{}
	<-done
	require.NoError(t, err)
}

func TestLongSuccessfulOperation(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t)
	})

	testConn := &testConn{invokeBody: func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return ctx.Err()
	}}

	cancelCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan struct{})
	defer close(done)
	var err error
	go func() {
		err = extendtimeout.NewConnection(testConn, 100*time.Millisecond).Invoke(cancelCtx, nil, nil)
		done <- struct{}{}
	}()

	<-done
	require.NoError(t, err)
}

func TestLongUnsuccessfulOperation(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t)
	})

	testConn := &testConn{invokeBody: func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return ctx.Err()
	}}

	cancelCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	defer close(done)

	var err error
	go func() {
		err = extendtimeout.NewConnection(testConn, 50*time.Millisecond).Invoke(cancelCtx, nil, nil)
		done <- struct{}{}
	}()

	<-done
	require.Error(t, err)
}
