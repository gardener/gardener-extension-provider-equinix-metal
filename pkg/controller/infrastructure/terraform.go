// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package infrastructure

import (
	_ "embed"
	"text/template"

	"github.com/Masterminds/sprig"
)

var (
	//go:embed templates/main.tpl.tf
	tplContentMainTF string
	tplNameMainTF    = "main.tf"
	tplMainTF        *template.Template
)

func init() {
	var err error
	tplMainTF, err = template.
		New(tplNameMainTF).
		Funcs(sprig.TxtFuncMap()).
		Parse(tplContentMainTF)
	if err != nil {
		panic(err)
	}
}

const (
	terraformTFVars = `# New line is needed! Do not remove this comment.
`
	variablesTF = `variable "EQXM_API_KEY" {
  description = "API Key"
  type        = string
}

variable "EQXM_PROJECT_ID" {
  description = "Project ID"
  type        = string
}`
)
