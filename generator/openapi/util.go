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

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/parser"
)

// CodeUtils provides utility functions for OpenAPI code generation.
type CodeUtils struct {
	features  *Features
	options   map[string]string
	log       backend.LogFunc
	rootScope *Scope
	ast       *parser.Thrift
}

// Features contains feature flags for OpenAPI generation.
type Features struct {
	SkipEmpty bool
	Version   string
	Title     string
	BasePath  string
}

// NewCodeUtils creates a new CodeUtils instance.
func NewCodeUtils(log backend.LogFunc) *CodeUtils {
	return &CodeUtils{
		features: &Features{
			SkipEmpty: false,
			Version:   "3.0.0",
			Title:     "Thrift API",
			BasePath:  "/api",
		},
		options: make(map[string]string),
		log:     log,
	}
}

// HandleOptions processes generator options.
func (u *CodeUtils) HandleOptions(args []string) error {
	var name, value string
	for _, a := range args {
		parts := strings.SplitN(a, "=", 2)
		switch len(parts) {
		case 0:
			continue
		case 1:
			name, value = parts[0], ""
		case 2:
			name, value = parts[0], parts[1]
		}

		switch name {
		case "skip_empty":
			u.features.SkipEmpty = value == "true"
		case "version":
			u.features.Version = value
		case "title":
			u.features.Title = value
		case "base_path":
			u.features.BasePath = value
		case "description":
			u.options["description"] = value
		case "contact_name":
			u.options["contact_name"] = value
		case "contact_email":
			u.options["contact_email"] = value
		case "contact_url":
			u.options["contact_url"] = value
		case "license_name":
			u.options["license_name"] = value
		case "license_url":
			u.options["license_url"] = value
		case "server_url":
			u.options["server_url"] = value
		case "server_description":
			u.options["server_description"] = value
		}
		u.options[name] = value
	}
	return nil
}

// Features returns the current features configuration.
func (u *CodeUtils) Features() *Features {
	return u.features
}

// GetFilename generates the output filename for a Thrift file.
func (u *CodeUtils) GetFilename(ast *parser.Thrift) string {
	base := strings.TrimSuffix(filepath.Base(ast.Filename), ".thrift")
	return base + ".yaml"
}

// CombineOutputPath combines the output path with the Thrift file path.
func (u *CodeUtils) CombineOutputPath(outputPath string, ast *parser.Thrift) string {
	if outputPath == "" {
		return "."
	}
	return outputPath
}

// SetRootScope sets the root scope for the current generation.
func (u *CodeUtils) SetRootScope(scope *Scope) {
	u.rootScope = scope
}

// SetAST sets the AST for the current generation.
func (u *CodeUtils) SetAST(ast *parser.Thrift) {
	u.ast = ast
}

// BuildFuncMap creates a template function map for OpenAPI generation.
func (u *CodeUtils) BuildFuncMap() template.FuncMap {
	return template.FuncMap{
		"ToOpenAPIType":     u.ToOpenAPIType,
		"ToOpenAPIFormat":   u.ToOpenAPIFormat,
		"ToOpenAPIMethod":   u.ToOpenAPIMethod,
		"ToOpenAPIPath":     u.ToOpenAPIPath,
		"GetSchemaName":     u.GetSchemaName,
		"GetServiceName":    u.GetServiceName,
		"GetOperationId":    u.GetOperationId,
		"GetDescription":    u.GetDescription,
		"GetExample":        u.GetExample,
		"GetDefaultValue":   u.GetDefaultValue,
		"IsRequired":        u.IsRequired,
		"GetEnumValues":     u.GetEnumValues,
		"GetStructFields":   u.GetStructFields,
		"GetServiceMethods": u.GetServiceMethods,
		"IsExpandField":     isExpandField,
		"IsExpandableStruct": isExpandableStruct,
		"IsFieldExpanded":   u.IsFieldExpanded,
		"GetExpandedFields": u.GetExpandedFields,
		"GetExpandedFieldNames": u.GetExpandedFieldNames,
	}
}

// ToOpenAPIType converts Thrift types to OpenAPI types.
func (u *CodeUtils) ToOpenAPIType(typ *parser.Type) string {
	if typ == nil {
		return "string"
	}

	switch typ.Category {
	case parser.Category_Bool:
		return "boolean"
	case parser.Category_Byte, parser.Category_I16, parser.Category_I32:
		return "integer"
	case parser.Category_I64:
		return "integer"
	case parser.Category_Double:
		return "number"
	case parser.Category_String:
		return "string"
	case parser.Category_Binary:
		return "string"
	case parser.Category_List:
		return "array"
	case parser.Category_Map:
		return "object"
	case parser.Category_Set:
		return "array"
	case parser.Category_Enum:
		return "string"
	case parser.Category_Struct, parser.Category_Union, parser.Category_Exception:
		return "object"
	default:
		return "string"
	}
}

// ToOpenAPIFormat returns the OpenAPI format for a Thrift type.
func (u *CodeUtils) ToOpenAPIFormat(typ *parser.Type) string {
	if typ == nil {
		return ""
	}

	switch typ.Category {
	case parser.Category_Byte:
		return "int8"
	case parser.Category_I16:
		return "int16"
	case parser.Category_I32:
		return "int32"
	case parser.Category_I64:
		return "int64"
	case parser.Category_Double:
		return "double"
	case parser.Category_String:
		return ""
	case parser.Category_Binary:
		return "binary"
	default:
		return ""
	}
}

// ToOpenAPIMethod converts Thrift function names to HTTP methods.
func (u *CodeUtils) ToOpenAPIMethod(funcName string) string {
	// 简单的命名约定：以 Get 开头使用 GET，以 Create/Add 开头使用 POST，以 Update 开头使用 PUT，以 Delete 开头使用 DELETE
	lower := strings.ToLower(funcName)
	if strings.HasPrefix(lower, "get") || strings.HasPrefix(lower, "find") || strings.HasPrefix(lower, "list") {
		return "get"
	}
	if strings.HasPrefix(lower, "create") || strings.HasPrefix(lower, "add") || strings.HasPrefix(lower, "insert") {
		return "post"
	}
	if strings.HasPrefix(lower, "update") || strings.HasPrefix(lower, "modify") {
		return "put"
	}
	if strings.HasPrefix(lower, "delete") || strings.HasPrefix(lower, "remove") {
		return "delete"
	}
	// 默认使用 POST
	return "post"
}

// ToOpenAPIPath converts Thrift service and function names to OpenAPI paths.
func (u *CodeUtils) ToOpenAPIPath(serviceName, funcName string) string {
	basePath := u.features.BasePath
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	if !strings.HasSuffix(basePath, "/") {
		basePath = basePath + "/"
	}

	// 将驼峰命名转换为小写并用连字符分隔
	servicePath := u.toKebabCase(serviceName)
	funcPath := u.toKebabCase(funcName)

	return fmt.Sprintf("%s%s/%s", basePath, servicePath, funcPath)
}

// toKebabCase converts camelCase to kebab-case.
func (u *CodeUtils) toKebabCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// GetSchemaName returns the schema name for a type.
func (u *CodeUtils) GetSchemaName(typ *parser.Type) string {
	if typ == nil {
		return "Unknown"
	}

	if typ.Name != "" {
		return typ.Name
	}

	switch typ.Category {
	case parser.Category_List:
		return u.GetSchemaName(typ.ValueType) + "List"
	case parser.Category_Map:
		keyType := u.GetSchemaName(typ.KeyType)
		valueType := u.GetSchemaName(typ.ValueType)
		return fmt.Sprintf("%sTo%sMap", keyType, valueType)
	case parser.Category_Set:
		return u.GetSchemaName(typ.ValueType) + "Set"
	default:
		return "Unknown"
	}
}

// GetServiceName returns the service name.
func (u *CodeUtils) GetServiceName(service *parser.Service) string {
	if service == nil {
		return "UnknownService"
	}
	return service.Name
}

// GetOperationId generates an operation ID for a service method.
func (u *CodeUtils) GetOperationId(service *parser.Service, function *parser.Function) string {
	if service == nil || function == nil {
		return "unknown"
	}
	return fmt.Sprintf("%s_%s", service.Name, function.Name)
}

// GetDescription returns a description for a type or function.
func (u *CodeUtils) GetDescription(item interface{}) string {
	// 这里可以根据需要添加注释解析逻辑
	switch v := item.(type) {
	case *parser.Function:
		if v.Annotations != nil {
			for _, ann := range v.Annotations {
				if ann.Key == "description" && len(ann.Values) > 0 {
					return ann.Values[0]
				}
			}
		}
		return fmt.Sprintf("调用 %s 方法", v.Name)
	case *parser.Field:
		if v.Annotations != nil {
			for _, ann := range v.Annotations {
				if ann.Key == "description" && len(ann.Values) > 0 {
					return ann.Values[0]
				}
			}
		}
		return fmt.Sprintf("字段 %s", v.Name)
	case *parser.StructLike:
		if v.Annotations != nil {
			for _, ann := range v.Annotations {
				if ann.Key == "description" && len(ann.Values) > 0 {
					return ann.Values[0]
				}
			}
		}
		return fmt.Sprintf("结构体 %s", v.Name)
	case *parser.Enum:
		if v.Annotations != nil {
			for _, ann := range v.Annotations {
				if ann.Key == "description" && len(ann.Values) > 0 {
					return ann.Values[0]
				}
			}
		}
		return fmt.Sprintf("枚举 %s", v.Name)
	default:
		return ""
	}
}

// GetExample returns an example value for a type.
func (u *CodeUtils) GetExample(typ *parser.Type) interface{} {
	if typ == nil {
		return ""
	}

	switch typ.Category {
	case parser.Category_Bool:
		return true
	case parser.Category_Byte, parser.Category_I16, parser.Category_I32, parser.Category_I64:
		return 42
	case parser.Category_Double:
		return 3.14
	case parser.Category_String:
		return "example"
	case parser.Category_Binary:
		return "base64encodedstring"
	case parser.Category_List:
		return []interface{}{u.GetExample(typ.ValueType)}
	case parser.Category_Map:
		return map[string]interface{}{
			"key": u.GetExample(typ.ValueType),
		}
	case parser.Category_Set:
		return []interface{}{u.GetExample(typ.ValueType)}
	default:
		return "example"
	}
}

// GetDefaultValue returns the default value for a field.
func (u *CodeUtils) GetDefaultValue(field *parser.Field) interface{} {
	if field == nil {
		return nil
	}

	if field.Default != nil {
		return field.Default
	}

	return u.GetExample(field.Type)
}

// IsRequired determines if a field is required.
func (u *CodeUtils) IsRequired(field *parser.Field) bool {
	if field == nil {
		return false
	}

	// 在 Thrift 中，required 字段是必需的
	return field.Requiredness == parser.FieldType_Required
}

// GetEnumValues returns the values of an enum.
func (u *CodeUtils) GetEnumValues(enum *parser.Enum) []string {
	if enum == nil {
		return nil
	}

	var values []string
	for _, value := range enum.Values {
		values = append(values, value.Name)
	}
	return values
}

// GetStructFields returns the fields of a struct-like type.
func (u *CodeUtils) GetStructFields(structLike *parser.StructLike) []*parser.Field {
	if structLike == nil {
		return nil
	}
	return structLike.Fields
}

// GetServiceMethods returns the methods of a service.
func (u *CodeUtils) GetServiceMethods(service *parser.Service) []*parser.Function {
	if service == nil {
		return nil
	}
	return service.Functions
}

// isExpandField 检查字段是否应该展开
func isExpandField(field *parser.Field) bool {
	return annotationContainsTrue(field.Annotations, "thrift.expand")
}

// isExpandableStruct 检查结构体是否可展开
func isExpandableStruct(structLike *parser.StructLike) bool {
	// 检查 Expandable 字段（由 expandable = "true" 注解解析而来）
	return structLike.Expandable != nil && *structLike.Expandable
}

// annotationContainsTrue 检查注解是否包含 true 值
func annotationContainsTrue(annos parser.Annotations, anno string) bool {
	vals := annos.Get(anno)
	if len(vals) == 0 {
		return false
	}
	if len(vals) > 1 {
		return false
	}
	return vals[0] == "true"
}

// getReferencedStruct 获取引用的结构体
func (u *CodeUtils) getReferencedStruct(field *parser.Field, ast *parser.Thrift) *parser.StructLike {
	if field == nil || field.Type == nil {
		return nil
	}
	
	if !field.Type.Category.IsStructLike() {
		return nil
	}
	
	// 查找引用的结构体
	typeName := field.Type.Name
	if typeName == "" {
		return nil
	}
	
	// 处理命名空间
	var actualTypeName string
	if strings.Contains(typeName, ".") {
		parts := strings.Split(typeName, ".")
		actualTypeName = parts[len(parts)-1]
	} else {
		actualTypeName = typeName
	}
	
	// 在当前 AST 中查找结构体
	for _, structLike := range ast.Structs {
		if structLike.Name == actualTypeName {
			return structLike
		}
	}
	
	for _, union := range ast.Unions {
		if union.Name == actualTypeName {
			return union
		}
	}
	
	for _, exception := range ast.Exceptions {
		if exception.Name == actualTypeName {
			return exception
		}
	}
	
	return nil
}

// collectExpandedFields 收集展开的字段
func (u *CodeUtils) collectExpandedFields(structLike *parser.StructLike, ast *parser.Thrift) ([]*parser.Field, map[string]bool) {
	var expandedFields []*parser.Field
	expandedFieldNames := make(map[string]bool)

	for _, field := range structLike.Fields {
		// 检查字段是否应该展开
		shouldExpand := isExpandField(field)
		// 检查引用的结构体是否可展开
		referencedStruct := u.getReferencedStruct(field, ast)
		structIsExpandable := referencedStruct != nil && isExpandableStruct(referencedStruct)

		if shouldExpand || structIsExpandable {
			// 记录原始字段被展开了
			expandedFieldNames[field.Name] = true

			// 展开字段，直接使用引用结构体的字段名，不添加前缀
			if referencedStruct != nil {
				for _, refField := range referencedStruct.Fields {
					expandedField := &parser.Field{
						Name:             refField.Name, // 直接使用原始字段名
						Type:             refField.Type,
						ID:               refField.ID,
						Requiredness:     refField.Requiredness,
						Default:          refField.Default,
						Annotations:      refField.Annotations,
						ReservedComments: refField.ReservedComments, // 复制注释
					}
					expandedFields = append(expandedFields, expandedField)
				}
			}
		}
	}
	return expandedFields, expandedFieldNames
}

// GetExpandedFields 获取展开的字段
func (u *CodeUtils) GetExpandedFields(structLike *parser.StructLike) []*parser.Field {
	if u.rootScope == nil || u.ast == nil {
		return nil
	}
	
	// 从 ExpandedStructs 中获取展开字段
	if expandedStruct, exists := u.rootScope.ExpandedStructs[structLike.Name]; exists {
		return expandedStruct.ExpandedFields
	}
	
	return nil
}

// GetExpandedFieldNames 获取展开的字段名映射
func (u *CodeUtils) GetExpandedFieldNames(structLike *parser.StructLike) map[string]bool {
	if u.rootScope == nil || u.ast == nil {
		return nil
	}
	
	// 从 ExpandedStructs 中获取展开字段名映射
	if expandedStruct, exists := u.rootScope.ExpandedStructs[structLike.Name]; exists {
		return expandedStruct.ExpandedFieldNames
	}
	
	return nil
}

// IsFieldExpanded 检查字段是否被展开
func (u *CodeUtils) IsFieldExpanded(field *parser.Field) bool {
	// 检查字段是否应该展开
	shouldExpand := isExpandField(field)
	if shouldExpand {
		return true
	}

	// 检查引用的结构体是否可展开
	if field.Type != nil && field.Type.Category.IsStructLike() && u.ast != nil {
		referencedStruct := u.getReferencedStruct(field, u.ast)
		if referencedStruct != nil && isExpandableStruct(referencedStruct) {
			return true
		}
	}

	return false
}
