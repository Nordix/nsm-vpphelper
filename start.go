// Copyright (c) 2020-2024 Cisco and/or its affiliates.
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

// Package vpphelper provides a simple Start function that will start up a local vpp,
// dial it, and return the grpc.ClientConnInterface
package vpphelper

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/edwarnicke/exechelper"
	"go.fd.io/govpp/api"

	"github.com/edwarnicke/log"
)

// StartAndDialContext - starts vpp
// Stdout and Stderr for vpp are set to be log.Entry(ctx).Writer().
func StartAndDialContext(ctx context.Context, opts ...Option) (conn api.Connection, errCh <-chan error) {
	o := &option{
		rootDir:   DefaultRootDir,
		vppConfig: DefaultVPPConfTemplate,
	}
	for _, opt := range opts {
		opt(o)
	}

	if err := writeDefaultConfigFiles(ctx, o); err != nil {
		errCh := make(chan error, 1)
		errCh <- err
		close(errCh)
		return nil, errCh
	}
	// We need to reset time in logger to make sure that
	// we don't use a static timestamp for a long-running process
	logWriter := log.Entry(ctx).WithField("cmd", "vpp").WithTime(time.Time{}).Writer()
	vppErrCh := exechelper.Start("vpp -c "+filepath.Join(o.rootDir, vppConfFilename),
		exechelper.WithContext(ctx),
		exechelper.WithStdout(logWriter),
		exechelper.WithStderr(logWriter),
	)
	select {
	case err := <-vppErrCh:
		errCh := make(chan error, 1)
		errCh <- err
		close(errCh)
		return nil, errCh
	default:
	}

	return DialContext(ctx, filepath.Join(o.rootDir, "/var/run/vpp/api.sock")), vppErrCh
}

func writeDefaultConfigFiles(ctx context.Context, o *option) error {
	configFiles := map[string]string{
		vppConfFilename: NewVPPConfigFile(o.vppConfig, VPPConfigParameters{RootDir: o.rootDir, DataSize: vppDefaultDataSize}),
	}
	for filename, contents := range configFiles {
		filename = filepath.Join(o.rootDir, filename)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.Entry(ctx).Infof("Configuration file: %q not found, using defaults", filename)
			if err := os.MkdirAll(path.Dir(filename), 0o700); err != nil {
				return err
			}
			if err := os.WriteFile(filename, []byte(contents), 0o600); err != nil {
				return err
			}
		}
	}
	if err := os.MkdirAll(filepath.Join(o.rootDir, "/var/run/vpp"), 0o700); os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Join(o.rootDir, "/var/log/vpp"), 0o700); os.IsNotExist(err) {
		return err
	}
	return nil
}
