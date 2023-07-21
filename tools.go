package vpphelper

import (
	"strings"
	"text/template"
)

type VPPConfigParameters struct {
	DataSize int
	RootDir  string
}

func NewVPPConfigFile(configTemplate string, params VPPConfigParameters) string {
	vppConfigBuilder := new(strings.Builder)

	t := template.Must(template.New("vppConfig").Parse(configTemplate))
	err := t.Execute(vppConfigBuilder, params)
	if err != nil {
		panic(err)
	}
	return vppConfigBuilder.String()
}
