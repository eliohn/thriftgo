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

	"github.com/cloudwego/thriftgo/parser"
)

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
	// 将 snake_case 转换为 camelCase
	parts := strings.Split(name, "_")
	if len(parts) == 1 {
		return parts[0]
	}

	result := parts[0]
	for _, part := range parts[1:] {
		result += strings.Title(part)
	}
	return result
}

// GetConstantName 获取常量名称
func GetConstantName(name string) string {
	// 将 snake_case 转换为 UPPER_CASE
	return strings.ToUpper(name)
}

// IsOptional 检查字段是否可选
func IsOptional(field *parser.Field) bool {
	return field.Requiredness == parser.FieldType_Optional
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
