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

package typescript

import (
	"fmt"
	"strings"
	"sync"

	"github.com/cloudwego/thriftgo/parser"
)

// 全局 AST 缓存，用于模板函数访问
var (
	globalAST *parser.Thrift
	astMutex  sync.RWMutex
)

// SetGlobalAST 设置全局 AST
func SetGlobalAST(ast *parser.Thrift) {
	astMutex.Lock()
	defer astMutex.Unlock()
	globalAST = ast
}

// GetGlobalAST 获取全局 AST
func GetGlobalAST() *parser.Thrift {
	astMutex.RLock()
	defer astMutex.RUnlock()
	return globalAST
}

// TypeMapping 定义 Thrift 类型到 TypeScript 类型的映射
type TypeMapping struct {
	ThriftType     string
	TypeScriptType string
	IsPrimitive    bool
}

var typeMappings = []TypeMapping{
	// 基本类型
	{"bool", "boolean", true},
	{"byte", "number", true},
	{"i8", "number", true},
	{"i16", "number", true},
	{"i32", "number", true},
	{"i64", "number", true},
	{"double", "number", true},
	{"string", "string", true},
	{"binary", "Uint8Array", true},

	// 容器类型
	{"list", "Array", false},
	{"set", "Set", false},
	{"map", "Map", false},
}

// GetTypeScriptType 将 Thrift 类型转换为 TypeScript 类型
func GetTypeScriptType(thriftType *parser.Type) string {
	if thriftType == nil {
		return "any"
	}

	// 处理基本类型
	if isPrimitiveType(thriftType.Category) {
		for _, mapping := range typeMappings {
			if mapping.ThriftType == thriftType.Name && mapping.IsPrimitive {
				return mapping.TypeScriptType
			}
		}
		return "any"
	}

	// 处理容器类型
	switch thriftType.Category {
	case parser.Category_List:
		elementType := GetTypeScriptType(thriftType.ValueType)
		return fmt.Sprintf("Array<%s>", elementType)
	case parser.Category_Set:
		elementType := GetTypeScriptType(thriftType.ValueType)
		return fmt.Sprintf("Set<%s>", elementType)
	case parser.Category_Map:
		keyType := GetTypeScriptType(thriftType.KeyType)
		valueType := GetTypeScriptType(thriftType.ValueType)
		// 在 TypeScript 中，Map 类型应该使用对象类型语法
		return fmt.Sprintf("{ [key: %s]: %s }", keyType, valueType)
	case parser.Category_Enum:
		return getSimpleTypeName(thriftType.Name)
	case parser.Category_Struct, parser.Category_Union, parser.Category_Exception:
		return getSimpleTypeName(thriftType.Name)
	case parser.Category_Typedef:
		return GetTypeScriptType(thriftType.ValueType)
	default:
		return "any"
	}
}

// getSimpleTypeName 获取简单的类型名（去掉前缀）
func getSimpleTypeName(typeName string) string {
	if strings.Contains(typeName, ".") {
		parts := strings.Split(typeName, ".")
		return parts[len(parts)-1]
	}
	return typeName
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

// isPrimitiveType 检查是否为基本类型
func isPrimitiveType(category parser.Category) bool {
	switch category {
	case parser.Category_Bool, parser.Category_Byte, parser.Category_I16,
		parser.Category_I32, parser.Category_I64, parser.Category_Double,
		parser.Category_String, parser.Category_Binary:
		return true
	default:
		return false
	}
}

// GetFieldType 获取字段的 TypeScript 类型
func GetFieldType(field *parser.Field) string {
	tsType := GetTypeScriptType(field.Type)

	// 处理可选字段
	if field.Requiredness == parser.FieldType_Optional {
		tsType += " | undefined"
	}

	return tsType
}

// GetMethodSignature 获取方法的 TypeScript 签名
func GetMethodSignature(method *parser.Function) string {
	var params []string
	var returnType string

	// 处理参数
	for _, param := range method.Arguments {
		paramType := GetFieldType(param)
		paramName := param.Name
		if param.Requiredness == parser.FieldType_Optional {
			paramName += "?"
		}
		params = append(params, fmt.Sprintf("%s: %s", paramName, paramType))
	}

	// 处理返回值
	if method.FunctionType != nil {
		returnType = GetTypeScriptType(method.FunctionType)
	} else {
		returnType = "void"
	}

	return fmt.Sprintf("(%s): %s", strings.Join(params, ", "), returnType)
}

// GetAsyncMethodSignature 获取异步方法的 TypeScript 签名
func GetAsyncMethodSignature(method *parser.Function) string {
	var params []string
	var returnType string

	// 处理参数
	for _, param := range method.Arguments {
		paramType := GetFieldType(param)
		paramName := param.Name
		if param.Requiredness == parser.FieldType_Optional {
			paramName += "?"
		}
		params = append(params, fmt.Sprintf("%s: %s", paramName, paramType))
	}

	// 处理返回值 - 异步方法返回 Promise
	if method.FunctionType != nil {
		returnType = fmt.Sprintf("Promise<%s>", GetTypeScriptType(method.FunctionType))
	} else {
		returnType = "Promise<void>"
	}

	return fmt.Sprintf("(%s): %s", strings.Join(params, ", "), returnType)
}

// GetInterfaceName 获取接口名称
func GetInterfaceName(name string) string {
	return strings.Title(name)
}

// GetClassName 获取类名称
func GetClassName(name string) string {
	return strings.Title(name)
}

// GetEnumName 获取枚举名称
func GetEnumName(name string) string {
	return strings.Title(name)
}

// GetEnumValueName 获取枚举值名称
func GetEnumValueName(name string) string {
	return strings.ToUpper(name)
}

// GetPropertyName 获取属性名称
func GetPropertyName(name string) string {
	// 默认使用 JSON 名称格式，如果与字段名相同则使用原字段名
	return name
}

// GetPropertyNameWithStyle 根据命名风格获取属性名称
func GetPropertyNameWithStyle(name string, features *Features) string {
	if features.SnakeStylePropertyName {
		return snakify(name)
	} else if features.LowerCamelCasePropertyName {
		return lowerCamelCase(name)
	}
	// 默认使用原始名称
	return name
}

// snakify 将字符串转换为 snake_case
func snakify(id string) string {
	if id == "" {
		return id
	}
	
	var result []rune
	for i, r := range id {
		if i > 0 && isUpper(r) {
			result = append(result, '_')
		}
		result = append(result, toLower(r))
	}
	return string(result)
}

// lowerCamelCase 将字符串转换为 lowerCamelCase
func lowerCamelCase(id string) string {
	if id == "" {
		return id
	}
	
	// 转换为 camelCase
	var result []rune
	nextUpper := false
	for i, r := range id {
		if i == 0 {
			result = append(result, toLower(r))
		} else if r == '_' {
			nextUpper = true
		} else if nextUpper {
			result = append(result, toUpper(r))
			nextUpper = false
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// isUpper 检查字符是否为大写
func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// isLower 检查字符是否为小写
func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

// toUpper 将字符转换为大写
func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
}

// toLower 将字符转换为小写
func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}

// GetConstantName 获取常量名称
func GetConstantName(name string) string {
	// 将 snake_case 转换为 UPPER_CASE
	return strings.ToUpper(name)
}

// IsOptional 检查字段是否可选
func IsOptional(field *parser.Field) bool {
	return field.Requiredness != parser.FieldType_Required
}

// GetDefaultValue 获取字段的默认值
func GetDefaultValue(field *parser.Field) string {
	if field.Default == nil || field.Default.TypedValue == nil {
		return ""
	}

	switch field.Type.Category {
	case parser.Category_Bool:
		if field.Default.TypedValue.Literal != nil && *field.Default.TypedValue.Literal == "true" {
			return "true"
		}
		return "false"
	case parser.Category_String:
		if field.Default.TypedValue.Literal != nil {
			return fmt.Sprintf(`"%s"`, *field.Default.TypedValue.Literal)
		}
		return `""`
	case parser.Category_Byte, parser.Category_I16, parser.Category_I32, parser.Category_I64, parser.Category_Double:
		if field.Default.TypedValue.Int != nil {
			return fmt.Sprintf("%d", *field.Default.TypedValue.Int)
		}
		if field.Default.TypedValue.Double != nil {
			return fmt.Sprintf("%f", *field.Default.TypedValue.Double)
		}
		return "0"
	default:
		return "null"
	}
}

// GetDefaultValueForType 获取类型的默认值
func GetDefaultValueForType(typ *parser.Type) string {
	if typ == nil {
		return "null"
	}

	switch typ.Category {
	case parser.Category_Bool:
		return "false"
	case parser.Category_String:
		return `""`
	case parser.Category_Byte, parser.Category_I16, parser.Category_I32, parser.Category_I64, parser.Category_Double:
		return "0"
	case parser.Category_List:
		return "[]"
	case parser.Category_Set:
		return "new Set()"
	case parser.Category_Map:
		return "{}"
	case parser.Category_Enum:
		// 对于枚举，返回第一个值
		return "0"
	case parser.Category_Struct, parser.Category_Union, parser.Category_Exception:
		// 对于结构体类型，返回 null
		return "null"
	default:
		return "null"
	}
}

// GetConstantValue 获取常量的值
func GetConstantValue(constant *parser.Constant) string {
	if constant == nil || constant.Value == nil || constant.Value.TypedValue == nil {
		return "null"
	}

	switch constant.Type.Category {
	case parser.Category_Bool:
		if constant.Value.TypedValue.Literal != nil && *constant.Value.TypedValue.Literal == "true" {
			return "true"
		}
		return "false"
	case parser.Category_String:
		if constant.Value.TypedValue.Literal != nil {
			return fmt.Sprintf(`"%s"`, *constant.Value.TypedValue.Literal)
		}
		return `""`
	case parser.Category_Byte, parser.Category_I16, parser.Category_I32, parser.Category_I64, parser.Category_Double:
		if constant.Value.TypedValue.Int != nil {
			return fmt.Sprintf("%d", *constant.Value.TypedValue.Int)
		}
		if constant.Value.TypedValue.Double != nil {
			return fmt.Sprintf("%f", *constant.Value.TypedValue.Double)
		}
		return "0"
	default:
		return "null"
	}
}

// GetStructFields 获取结构体的字段列表
// 这个函数需要在模板中通过其他方式调用，因为需要 AST 信息
func GetStructFields(field *parser.Field) []*parser.Field {
	if field == nil || field.Type == nil || !field.Type.Category.IsStructLike() {
		return nil
	}

	// 这里需要从 AST 中查找结构体定义
	// 由于模板中无法直接访问 AST，我们需要通过其他方式
	// 暂时返回空，后续可以通过其他方式实现
	return nil
}

// GetStructFieldAnnotations 获取结构体字段的注解信息
// 从 AST 中动态获取结构体字段的注解信息，包括展开字段
func GetStructFieldAnnotations(field *parser.Field, ast *parser.Thrift) map[string]map[string]string {
	annotations := make(map[string]map[string]string)

	if field == nil || field.Type == nil || !field.Type.Category.IsStructLike() {
		return annotations
	}

	// 查找结构体定义
	structLike := findStructLikeByName(field.Type.Name, ast)
	if structLike == nil {
		return annotations
	}

	// 遍历结构体字段，收集注解信息
	for _, structField := range structLike.Fields {
		fieldAnnotations := make(map[string]string)

		// 检查 api.path 注解
		if pathAnno := structField.Annotations.Get("api.path"); len(pathAnno) > 0 {
			fieldAnnotations["api.path"] = pathAnno[0]
		}

		// 检查 api.query 注解
		if queryAnno := structField.Annotations.Get("api.query"); len(queryAnno) > 0 {
			fieldAnnotations["api.query"] = queryAnno[0]
		}

		// 检查 api.body 注解
		if bodyAnno := structField.Annotations.Get("api.body"); len(bodyAnno) > 0 {
			fieldAnnotations["api.body"] = bodyAnno[0]
		}

		// 添加所有字段，包括没有注解的字段
		annotations[structField.Name] = fieldAnnotations
	}

	// 检查是否有展开字段需要处理
	// 这里需要检查字段是否有 thrift.expand 注解
	if expandAnno := field.Annotations.Get("thrift.expand"); len(expandAnno) > 0 && expandAnno[0] == "true" {
		// 如果字段被展开，我们需要获取被展开结构体的字段注解
		// 但这里我们已经在上面处理了结构体字段，所以展开字段的注解应该已经包含在内
		// 不过我们需要确保展开字段的注解被正确保留
		// 由于展开字段的注解处理在 golang 生成器中，这里我们主要处理的是模板中的使用
	}

	return annotations
}

// findStructLikeByName 根据名称查找结构体定义
func findStructLikeByName(name string, ast *parser.Thrift) *parser.StructLike {
	// 提取结构体的实际名称（去掉命名空间前缀）
	actualName := name
	if lastDot := strings.LastIndex(name, "."); lastDot != -1 {
		actualName = name[lastDot+1:]
	}

	// 在当前文件中查找
	for _, structLike := range ast.Structs {
		if structLike.Name == actualName {
			return structLike
		}
	}

	// 在包含的文件中查找
	for _, include := range ast.Includes {
		if include.Reference != nil {
			for _, structLike := range include.Reference.Structs {
				if structLike.Name == actualName {
					return structLike
				}
			}
		}
	}

	return nil
}

// GetStructFieldAnnotationsForTemplate 模板中使用的结构体字段注解获取函数
// 使用全局 AST 缓存来获取结构体字段的注解信息
func GetStructFieldAnnotationsForTemplate(field *parser.Field) map[string]map[string]string {
	ast := GetGlobalAST()
	if ast == nil {
		return make(map[string]map[string]string)
	}
	return GetStructFieldAnnotations(field, ast)
}

// GetStructFieldByName 根据字段名获取结构体字段
func GetStructFieldByName(structField *parser.Field, fieldName string) *parser.Field {
	if structField == nil || structField.Type == nil || !structField.Type.Category.IsStructLike() {
		return nil
	}

	ast := GetGlobalAST()
	if ast == nil {
		return nil
	}

	structLike := findStructLikeByName(structField.Type.Name, ast)
	if structLike == nil {
		return nil
	}

	// 查找字段
	for _, field := range structLike.Fields {
		if field.Name == fieldName {
			return field
		}
	}

	return nil
}

// GetFieldExpandedFields 获取字段对应的展开字段
// 如果字段是结构体类型且被展开，返回展开的字段列表
func GetFieldExpandedFields(field *parser.Field) []*parser.Field {
	if field == nil || field.Type == nil || !field.Type.Category.IsStructLike() {
		return nil
	}

	// 检查字段是否有展开注解
	shouldExpand := false
	
	// 检查 thrift.expand 注解
	if expandAnno := field.Annotations.Get("thrift.expand"); len(expandAnno) > 0 && expandAnno[0] == "true" {
		shouldExpand = true
	}
	
	// 检查引用的结构体是否可展开
	if !shouldExpand {
		ast := GetGlobalAST()
		if ast != nil {
			structLike := findStructLikeByName(field.Type.Name, ast)
			if structLike != nil {
				if isExpandableStruct(structLike) {
					shouldExpand = true
				}
			}
		}
	}

	if shouldExpand {
		// 获取被展开的结构体定义
		ast := GetGlobalAST()
		if ast == nil {
			return nil
		}

		structLike := findStructLikeByName(field.Type.Name, ast)
		if structLike == nil {
			return nil
		}

		// 返回结构体的字段作为展开字段
		return structLike.Fields
	}

	return nil
}

// IsStructField 检查字段是否为结构体类型
func IsStructField(field *parser.Field) bool {
	if field == nil || field.Type == nil {
		return false
	}

	return field.Type.Category.IsStructLike()
}

// FormatCommentForJSDoc 将 Thrift 注释格式化为 TypeScript JSDoc 格式
func FormatCommentForJSDoc(comment string) string {
	if comment == "" {
		return ""
	}

	// 清理注释内容
	comment = strings.TrimSpace(comment)

	// 移除 Thrift 注释标记
	comment = strings.TrimPrefix(comment, "//")
	comment = strings.TrimPrefix(comment, "/*")
	comment = strings.TrimSuffix(comment, "*/")
	comment = strings.TrimSpace(comment)

	// 如果注释为空，返回空字符串
	if comment == "" {
		return ""
	}

	// 将多行注释转换为 JSDoc 格式
	lines := strings.Split(comment, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 移除行首的注释标记
		line = strings.TrimPrefix(line, "//")
		line = strings.TrimPrefix(line, "/*")
		line = strings.TrimSuffix(line, "*/")
		line = strings.TrimSpace(line)

		// 移除行首的 * 符号（如果存在）
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)

		if line != "" {
			result = append(result, " * "+line)
		}
	}

	if len(result) == 0 {
		return ""
	}

	return "\n/**\n" + strings.Join(result, "\n") + "\n */"
}

// GetStructComment 获取结构体的注释
func GetStructComment(structLike *parser.StructLike) string {
	if structLike == nil {
		return ""
	}
	return FormatCommentForJSDoc(structLike.ReservedComments)
}

// GetFieldComment 获取字段的注释
func GetFieldComment(field *parser.Field) string {
	if field == nil {
		return ""
	}
	return FormatCommentForJSDoc(field.ReservedComments)
}

// GetEnumComment 获取枚举的注释
func GetEnumComment(enum *parser.Enum) string {
	if enum == nil {
		return ""
	}
	return FormatCommentForJSDoc(enum.ReservedComments)
}

// GetEnumValueComment 获取枚举值的注释
func GetEnumValueComment(enumValue *parser.EnumValue) string {
	if enumValue == nil {
		return ""
	}
	return FormatCommentForJSDoc(enumValue.ReservedComments)
}

// GetServiceComment 获取服务的注释
func GetServiceComment(service *parser.Service) string {
	if service == nil {
		return ""
	}
	return FormatCommentForJSDoc(service.ReservedComments)
}

// GetFunctionComment 获取函数的注释
func GetFunctionComment(function *parser.Function) string {
	if function == nil {
		return ""
	}
	return FormatCommentForJSDoc(function.ReservedComments)
}
