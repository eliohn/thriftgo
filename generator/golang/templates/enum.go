// Copyright 2021 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package templates

// Enum .
var Enum = `
{{define "Enum"}}
{{- $EnumType := .GoName}}
{{InsertionPoint "enum" .Name}}
{{- if and Features.ReserveComments .ReservedComments}}
{{.ReservedComments}}
{{- end}}

{{- $enumType := "64"}}
{{- range .Annotations}}
{{- if eq .Key "go.type"}}
{{- $typeValue := index .Values 0}}
{{- if eq $typeValue "int8"}}
{{- $enumType = "8"}}
{{- else if eq $typeValue "int16"}}
{{- $enumType = "16"}}
{{- else if eq $typeValue "int32"}}
{{- $enumType = "32"}}
{{- else if eq $typeValue "int64"}}
{{- $enumType = "64"}}
{{- end}}
{{- end}}
{{- end}}

type {{$EnumType}} int{{if eq $enumType "8"}}8{{else if eq $enumType "16"}}16{{else if eq $enumType "32"}}32{{else}}64{{end}}

const (
	{{- range .Values}}
	{{- if and Features.ReserveComments .ReservedComments}}
	{{.ReservedComments}}{{end}}
	{{.GoName}} {{$EnumType}} = {{.Value}}
	{{- end}}
)

func (p {{$EnumType}}) String() string {
	switch p {
	{{- range .Values}}
	case {{.GoName}}:
		return "{{.GoLiteral}}"
	{{- end}}
	}
	return "<UNSET>"
}

func {{$EnumType}}FromString(s string) ({{$EnumType}}, error) {
	switch s {
	{{- range .Values}}
	case "{{.GoLiteral}}":
		return {{.GoName}}, nil
	{{- end}}
	}
	{{- UseStdLibrary "fmt"}}
	return {{$EnumType}}(0), fmt.Errorf("not a valid {{$EnumType}} string")
}

func {{$EnumType}}Ptr(v {{$EnumType}} ) *{{$EnumType}}  { return &v }

// 获取枚举的原始值
func (p {{$EnumType}}) ToInt() {{if eq $enumType "8"}}int8{{else if eq $enumType "16"}}int16{{else if eq $enumType "32"}}int32{{else}}int64{{end}} {
	return {{if eq $enumType "8"}}int8(p){{else if eq $enumType "16"}}int16(p){{else if eq $enumType "32"}}int32(p){{else}}int64(p){{end}}
}

{{- if or Features.MarshalEnumToText Features.MarshalEnum}}

func (p {{$EnumType}}) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

{{end}}{{/* if or Features.MarshalEnumToText Features.MarshalEnum */}}

{{- if or Features.MarshalEnumToText Features.UnmarshalEnum}}

func (p *{{$EnumType}}) UnmarshalText(text []byte) error {
	q, err := {{$EnumType}}FromString(string(text))
	if err != nil {
		return err
	}
	*p = q
	return nil
}
{{end}}{{/* if or Features.MarshalEnumToText Features.UnmarshalEnum */}}

{{- if Features.ScanValueForEnum}}
{{- UseStdLibrary "sql" "driver"}}
{{- $enumType := "64"}}
{{- range .Annotations}}
{{- if eq .Key "go.type"}}
{{- $typeValue := index .Values 0}}
{{- if eq $typeValue "int8"}}
{{- $enumType = "8"}}
{{- else if eq $typeValue "int16"}}
{{- $enumType = "16"}}
{{- else if eq $typeValue "int32"}}
{{- $enumType = "32"}}
{{- else if eq $typeValue "int64"}}
{{- $enumType = "64"}}
{{- end}}
{{- end}}
{{- end}}

func (p *{{$EnumType}}) Scan(value interface{}) (err error) {
	{{- if eq $enumType "8"}}
	var result sql.NullInt64
	err = result.Scan(value)
	*p = {{$EnumType}}(int8(result.Int64))
	{{- else if eq $enumType "16"}}
	var result sql.NullInt64
	err = result.Scan(value)
	*p = {{$EnumType}}(int16(result.Int64))
	{{- else if eq $enumType "32"}}
	var result sql.NullInt32
	err = result.Scan(value)
	*p = {{$EnumType}}(result.Int32)
	{{- else}}
	var result sql.NullInt64
	err = result.Scan(value)
	*p = {{$EnumType}}(result.Int64)
	{{- end}}
	return
}

func (p *{{$EnumType}}) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	{{- if eq $enumType "8"}}
	return int64(*p), nil
	{{- else if eq $enumType "16"}}
	return int64(*p), nil
	{{- else if eq $enumType "32"}}
	return int32(*p), nil
	{{- else}}
	return int64(*p), nil
	{{- end}}
}
{{- end}}{{/* if .Features.ScanValueForEnum */}}

{{- if Features.GetEnumAnnotation}}
var annotations_{{$EnumType}} = map[{{$EnumType}}]map[string][]string{
    {{- range .Values}}
    {{.GoName}}: map[string][]string{
        {{genAnnotations .}}
    },
    {{- end}}
}

func (p {{$EnumType}}) GetAnnotation(key string) []string {
    switch p {
    {{- range .Values}}
    case {{.GoName}}:
        return annotations_{{$EnumType}}[{{.GoName}}][key]
    {{- end}}
    }
    return nil
}
{{- end}}{{/* if Features.GenGetEnumAnnotation */}}
{{end}}
`
