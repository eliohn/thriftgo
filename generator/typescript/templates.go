// Copyright 2021 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package typescript

import (
	"github.com/cloudwego/thriftgo/generator/typescript/templates"
)

// Templates 返回所有 TypeScript 模板
func Templates() []string {
	return []string{
		templates.FileTemplate,
		templates.IndexTemplate,
		templates.ImportsTemplate,
		templates.EnumTemplate,
		templates.StructTemplate,
		templates.ServiceTemplate,
		templates.ConstantTemplate,
		templates.TypedefTemplate,
		templates.UnionTemplate,
		templates.ExceptionTemplate,
		templates.SingleEnumTemplate,
		templates.SingleStructTemplate,
		templates.SingleServiceTemplate,
		templates.SimpleServiceImplementationTemplate,
	}
}
