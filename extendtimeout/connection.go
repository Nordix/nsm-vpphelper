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

// Package extendtimeout - provides a wrapper for vpp connection that uses extended timeout for all vpp operations
package extendtimeout

import (
	"context"
	"time"

	"github.com/edwarnicke/log"
	"go.fd.io/govpp/api"
)

type extendedConnection struct {
	api.Connection
	contextTimeout time.Duration
}

type extendedContext struct {
	context.Context
	valuesContext context.Context
}

func (ec *extendedContext) Value(key interface{}) interface{} {
	return ec.valuesContext.Value(key)
}

// NewConnection - creates a wrapper for vpp connection that uses extended context timeout for all operations
func NewConnection(vppConn api.Connection, contextTimeout time.Duration) api.Connection {
	return &extendedConnection{
		Connection:     vppConn,
		contextTimeout: contextTimeout,
	}
}

func (c *extendedConnection) Invoke(ctx context.Context, req, reply api.Message) error {
	ctx, cancel := c.withExtendedTimeoutCtx(ctx)
	err := c.Connection.Invoke(ctx, req, reply)
	cancel()
	return err
}

func (c *extendedConnection) cancelMonitorCtx(ctx context.Context) (cancelMonitorCtx context.Context, cancel func()) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	cancelCh := make(chan struct{})
	go func() {
		select {
		case <-time.After(c.contextTimeout):
			<-ctx.Done()
		case <-cancelCh:
		}
		cancelFunc()
	}()

	cancelMonitorCtx = &extendedContext{
		Context:       cancelCtx,
		valuesContext: ctx,
	}
	cancel = func() {
		cancelCh <- struct{}{}
		close(cancelCh)
	}
	return
}

func (c *extendedConnection) withExtendedTimeoutCtx(ctx context.Context) (extendedCtx context.Context, cancel func()) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return c.cancelMonitorCtx(ctx)
	}

	minDeadline := time.Now().Add(c.contextTimeout)
	if minDeadline.Before(deadline) {
		return c.cancelMonitorCtx(ctx)
	}
	log.Entry(ctx).Warnf("Context deadline has been extended by extendtimeout from %v to %v", deadline, minDeadline)
	deadline = minDeadline
	postponedCtx, cancel := context.WithDeadline(context.Background(), deadline)
	return &extendedContext{
		Context:       postponedCtx,
		valuesContext: ctx,
	}, cancel
}
