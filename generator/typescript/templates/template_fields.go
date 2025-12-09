// Copyright 2021 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package templates

// Fields 文件模板
const FieldsTemplate = `
{{- define "fields" -}}
{{- $struct := index .Structs 0 }}
{{- if $struct }}
/**
 * {{ GetInterfaceName $struct.Name }} 字段常量
 * 此文件用于导出 {{ GetInterfaceName $struct.Name }} 接口的字段常量
 * 使用 satisfies 确保所有字段都是 {{ GetInterfaceName $struct.Name }} 的键
 * 如果字段不存在于 {{ GetInterfaceName $struct.Name }} 中，TypeScript 会在编译时报错
 * 
 * 注意：此文件不会被自动生成覆盖，可以安全地手动维护
 */

import type { {{ GetInterfaceName $struct.Name }} } from './{{ ToLower $struct.Name }}';

/**
 * {{ GetInterfaceName $struct.Name }} 的所有字段常量
 * 使用 satisfies 确保类型安全
 */
export const {{ GetInterfaceName $struct.Name }}Fields = {
{{- $fieldNames := GetStructFieldNames $struct }}
{{- range $index, $fieldName := $fieldNames }}
{{- if $index }},{{ end }}
  {{ $fieldName }}: '{{ $fieldName }}'
{{- end }}
} as const satisfies Record<string, keyof {{ GetInterfaceName $struct.Name }}>;

/**
 * 类型：{{ GetInterfaceName $struct.Name }} 的字段名
 */
export type {{ GetInterfaceName $struct.Name }}Field = keyof {{ GetInterfaceName $struct.Name }};
{{- end }}
{{- end -}}
`

