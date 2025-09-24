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

import "github.com/cloudwego/thriftgo/parser"

// OpenAPISchema represents an OpenAPI schema.
type OpenAPISchema struct {
	Type        string                 `json:"type,omitempty"`
	Format      string                 `json:"format,omitempty"`
	Description string                 `json:"description,omitempty"`
	Example     interface{}            `json:"example,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Enum        []string               `json:"enum,omitempty"`
	Items       interface{}            `json:"items,omitempty"`
	Ref         string                 `json:"$ref,omitempty"`
}

// OpenAPIParameter represents an OpenAPI parameter.
type OpenAPIParameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Required    bool        `json:"required,omitempty"`
	Description string      `json:"description,omitempty"`
	Schema      interface{} `json:"schema,omitempty"`
}

// OpenAPIResponse represents an OpenAPI response.
type OpenAPIResponse struct {
	Description string                 `json:"description"`
	Content     map[string]interface{} `json:"content,omitempty"`
}

// OpenAPIOperation represents an OpenAPI operation.
type OpenAPIOperation struct {
	Tags        []string                 `json:"tags,omitempty"`
	Summary     string                   `json:"summary,omitempty"`
	Description string                   `json:"description,omitempty"`
	OperationId string                   `json:"operationId,omitempty"`
	Parameters  []OpenAPIParameter       `json:"parameters,omitempty"`
	RequestBody map[string]interface{}   `json:"requestBody,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses"`
}

// OpenAPIPathItem represents an OpenAPI path item.
type OpenAPIPathItem struct {
	Get    *OpenAPIOperation `json:"get,omitempty"`
	Post   *OpenAPIOperation `json:"post,omitempty"`
	Put    *OpenAPIOperation `json:"put,omitempty"`
	Delete *OpenAPIOperation `json:"delete,omitempty"`
}

// OpenAPIInfo represents OpenAPI info section.
type OpenAPIInfo struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Version     string                 `json:"version"`
	Contact     map[string]string      `json:"contact,omitempty"`
	License     map[string]string      `json:"license,omitempty"`
}

// OpenAPIServer represents an OpenAPI server.
type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// OpenAPIDocument represents the complete OpenAPI document.
type OpenAPIDocument struct {
	OpenAPI    string                        `json:"openapi"`
	Info       OpenAPIInfo                   `json:"info"`
	Servers    []OpenAPIServer               `json:"servers,omitempty"`
	Paths      map[string]OpenAPIPathItem    `json:"paths"`
	Components map[string]map[string]interface{} `json:"components,omitempty"`
}

// ConvertToOpenAPISchema converts a Thrift type to OpenAPI schema.
func ConvertToOpenAPISchema(typ *parser.Type) OpenAPISchema {
	if typ == nil {
		return OpenAPISchema{Type: "string"}
	}

	schema := OpenAPISchema{}

	switch typ.Category {
	case parser.Category_Bool:
		schema.Type = "boolean"
	case parser.Category_Byte, parser.Category_I16, parser.Category_I32:
		schema.Type = "integer"
		schema.Format = "int32"
	case parser.Category_I64:
		schema.Type = "integer"
		schema.Format = "int64"
	case parser.Category_Double:
		schema.Type = "number"
		schema.Format = "double"
	case parser.Category_String:
		schema.Type = "string"
	case parser.Category_Binary:
		schema.Type = "string"
		schema.Format = "binary"
	case parser.Category_List:
		schema.Type = "array"
		schema.Items = ConvertToOpenAPISchema(typ.ValueType)
	case parser.Category_Map:
		schema.Type = "object"
		// Map 的键值对类型
		schema.Properties = map[string]interface{}{
			"key":   ConvertToOpenAPISchema(typ.KeyType),
			"value": ConvertToOpenAPISchema(typ.ValueType),
		}
	case parser.Category_Set:
		schema.Type = "array"
		schema.Items = ConvertToOpenAPISchema(typ.ValueType)
		schema.Properties = map[string]interface{}{
			"uniqueItems": true,
		}
	case parser.Category_Enum:
		schema.Type = "string"
		// 枚举值需要从 AST 中获取
	case parser.Category_Struct, parser.Category_Union, parser.Category_Exception:
		schema.Type = "object"
		schema.Ref = "#/components/schemas/" + typ.Name
	default:
		schema.Type = "string"
	}

	return schema
}

// ConvertStructToOpenAPISchema converts a Thrift struct to OpenAPI schema.
func ConvertStructToOpenAPISchema(structLike *parser.StructLike) OpenAPISchema {
	schema := OpenAPISchema{
		Type:        "object",
		Properties:  make(map[string]interface{}),
		Required:    []string{},
	}

	for _, field := range structLike.Fields {
		fieldSchema := ConvertToOpenAPISchema(field.Type)
		schema.Properties[field.Name] = fieldSchema

		if field.Requiredness == parser.FieldType_Required {
			schema.Required = append(schema.Required, field.Name)
		}
	}

	return schema
}

// ConvertEnumToOpenAPISchema converts a Thrift enum to OpenAPI schema.
func ConvertEnumToOpenAPISchema(enum *parser.Enum) OpenAPISchema {
	schema := OpenAPISchema{
		Type: "string",
		Enum: []string{},
	}

	for _, value := range enum.Values {
		schema.Enum = append(schema.Enum, value.Name)
	}

	return schema
}
