// Copyright 2024 CloudWeGo Authors
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

package openapi

// Templates returns all OpenAPI templates.
func Templates() []string {
	return []string{
		openapiTemplate,
	}
}

const openapiTemplate = `
openapi: {{.GetOpenAPIVersion}}
info:
  title: {{.GetAPITitle}}
  description: {{.GetAPIDescription}}
  version: 1.0.0
paths:
{{range .GetAllServices}}
  {{$service := .}}
  {{range .Functions}}
  {{ToOpenAPIPath $service.Name .Name}}:
    {{ToOpenAPIMethod .Name}}:
      tags:
        - {{$service.Name}}
      summary: {{GetDescription .}}
      operationId: {{GetOperationId $service .}}
      {{if .Arguments}}
      parameters:
        {{range .Arguments}}
        - name: {{.Name}}
          in: query
          required: {{IsRequired .}}
          schema:
            type: {{ToOpenAPIType .Type}}
            {{if ToOpenAPIFormat .Type}}
            format: {{ToOpenAPIFormat .Type}}
            {{end}}
            {{if GetExample .Type}}
            example: {{GetExample .Type}}
            {{end}}
          description: {{GetDescription .}}
        {{end}}
      {{end}}
      {{if .FunctionType}}
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/{{GetSchemaName .FunctionType}}'
      {{end}}
      responses:
        '200':
          description: 成功响应
          {{if .FunctionType}}
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/{{GetSchemaName .FunctionType}}'
          {{end}}
        '400':
          description: 请求错误
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                    description: 错误信息
        '500':
          description: 服务器错误
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                    description: 错误信息
  {{end}}
{{end}}
components:
  schemas:
{{range .GetAllSchemas}}
    {{if eq (printf "%T" .) "*parser.Enum"}}
    {{.Name}}:
      type: string
      enum:
        {{range .Values}}
        - {{.Name}}
        {{end}}
      description: {{GetDescription .}}
    {{else if eq (printf "%T" .) "*parser.StructLike"}}
    {{.Name}}:
      type: object
      description: {{GetDescription .}}
      {{if .Fields}}
      properties:
        {{range .Fields}}
        {{if IsFieldExpanded .}}
        {{/* 展开字段：显示展开后的字段 */}}
        {{else}}
        {{.Name}}:
          type: {{ToOpenAPIType .Type}}
          {{if ToOpenAPIFormat .Type}}
          format: {{ToOpenAPIFormat .Type}}
          {{end}}
          {{if GetExample .Type}}
          example: {{GetExample .Type}}
          {{end}}
          description: {{GetDescription .}}
        {{end}}
        {{end}}
        {{/* 添加展开的字段 */}}
        {{if GetExpandedFields .}}
        {{range GetExpandedFields .}}
        {{.Name}}:
          type: {{ToOpenAPIType .Type}}
          {{if ToOpenAPIFormat .Type}}
          format: {{ToOpenAPIFormat .Type}}
          {{end}}
          {{if GetExample .Type}}
          example: {{GetExample .Type}}
          {{end}}
          description: {{GetDescription .}}
        {{end}}
        {{end}}
      required:
        {{range .Fields}}
        {{if not (IsFieldExpanded .)}}
        {{if IsRequired .}}
        - {{.Name}}
        {{end}}
        {{end}}
        {{end}}
        {{/* 添加展开字段的必需字段 */}}
        {{if GetExpandedFields .}}
        {{range GetExpandedFields .}}
        {{if IsRequired .}}
        - {{.Name}}
        {{end}}
        {{end}}
        {{end}}
      {{end}}
    {{end}}
{{end}}
`
