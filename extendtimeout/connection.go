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

	"go.fd.io/govpp/api"
)

type extendedConnection struct {
	api.Connection
	contextTimeout time.Duration
}

// NewConnection - creates a wrapper for vpp connection that uses extended context timeout for all operations
func NewConnection(vppConn api.Connection, contextTimeout time.Duration) api.Connection {
	return &extendedConnection{
		Connection:     vppConn,
		contextTimeout: contextTimeout,
	}
}

func (c *extendedConnection) Invoke(ctx context.Context, req, reply api.Message) error {
	ctx, cancel := c.withExtendedTimeoutContext(ctx)
	err := c.Connection.Invoke(ctx, req, reply)
	cancel()
	return err
}

func (c *extendedConnection) withExtendedTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	var cancelContext, cancel = context.WithCancel(context.Background())
	var timeoutContext, timeoutCancel = context.WithTimeout(cancelContext, c.contextTimeout)
	go func() {
		<-timeoutContext.Done()
		timeoutCancel()
		select {
		case <-cancelContext.Done():
			return
		case <-ctx.Done():
			cancel()
			return
		}
	}()

	return cancelContext, cancel
}
