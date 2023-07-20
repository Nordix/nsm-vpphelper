package vpphelper

import (
	"strings"
	"text/template"
)

type VPPConfigParameters struct {
	DataSize int
	RootDir  string
}

func NewVPPConfigFile(confTemplate string, params VPPConfigParameters) string {
	vppConfigBuilder := new(strings.Builder)

	t := template.Must(template.New("vppConfig").Parse(confTemplate))
	err := t.Execute(vppConfigBuilder, params)
	if err != nil {
		panic(err)
	}
	return vppConfigBuilder.String()
}
