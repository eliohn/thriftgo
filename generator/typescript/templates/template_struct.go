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

// 结构体模板
const StructTemplate = `
{{- define "struct" -}}
export interface {{ GetInterfaceName .Name }} {
{{- $expandedFields := GetExpandedFields . }}
{{- $expandedFieldNames := GetExpandedFieldNames . }}
// DEBUG: expandedFieldNames = {{ $expandedFieldNames }}
{{- range .Fields }}
{{- $isExpanded := index $expandedFieldNames .Name }}
{{- if not $isExpanded }}
  {{ GetPropertyName .Name }}{{ if IsOptional . }}?{{ end }}: {{ GetFieldType . }};
{{- else }}
  // {{ GetPropertyName .Name }} is expanded ({{ $isExpanded }})
{{- end }}
{{- end }}
{{- if $expandedFields }}
{{- range $expandedFields }}
  {{ GetPropertyName .Name }}{{ if IsOptional . }}?{{ end }}: {{ GetFieldType . }};
{{- end }}
{{- end }}
}
{{- end -}}
`
