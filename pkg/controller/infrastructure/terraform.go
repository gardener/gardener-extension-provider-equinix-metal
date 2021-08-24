// Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
