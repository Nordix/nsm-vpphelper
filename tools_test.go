package vpphelper

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_NewVPPConfigFile(t *testing.T) {
	const expectedConfig = `500 + /root/dir`
	require.Equal(t, expectedConfig, NewVPPConfigFile(
		`{{ .DataSize }} + {{ .RootDir }}`,
		VPPConfigParameters{DataSize: 500, RootDir: `/root/dir`}))
}
