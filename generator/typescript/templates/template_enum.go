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

// 枚举模板
const EnumTemplate = `
{{- define "enum" -}}
{{- $enumComment := GetEnumComment . }}
{{- if $enumComment }}
{{ $enumComment }}
{{- end }}
export enum {{ GetInterfaceName .Name }} {
{{- range .Values }}
{{- $valueComment := GetEnumValueComment . }}
{{- if $valueComment }}
{{ $valueComment }}
{{- end }}
  {{ GetEnumValueName .Name }} = {{ .Value }},
{{- end }}
}

{{- if HasEnumValueWithTag . }}
// 枚举选项数组，用于下拉菜单等场景
export const {{ GetInterfaceName .Name }}Options = [
{{- range GetEnumOptions . }}
  {
    label: '{{ .label }}',
    value: {{ .value }},{{- if .color }}
    color: '{{ .color }}',{{- end }}
  },
{{- end }}
] as const;
{{- end }}
{{- end -}}
`
