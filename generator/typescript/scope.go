// Copyright 2024 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package typescript

import (
	"path/filepath"
	"strings"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/parser"
)

// Scope 表示 TypeScript 代码生成的作用域
type Scope struct {
	// 文件信息
	Filename string
	Package  string

	// 导入
	Imports []ImportInfo

	// 定义
	Constants  []*parser.Constant
	Typedefs   []*parser.Typedef
	Enums      []*parser.Enum
	Structs    []*parser.StructLike
	Unions     []*parser.StructLike
	Exceptions []*parser.StructLike
	Services   []*parser.Service

	// 展开的结构体（用于处理 expandable 注解）
	ExpandedStructs map[string]*ExpandedStruct

	// 工具
	utils *CodeUtils
}

// ExpandedStruct 表示展开后的结构体
type ExpandedStruct struct {
	OriginalStruct     *parser.StructLike
	ExpandedFields     []*parser.Field
	ExpandedFieldNames map[string]bool // 记录哪些字段被展开了
}

// ImportInfo 表示导入信息
type ImportInfo struct {
	Module string
	Types  []string
	Path   string
}

// IsEmpty 检查作用域是否为空
func (s *Scope) IsEmpty() bool {
	return len(s.Constants) == 0 &&
		len(s.Typedefs) == 0 &&
		len(s.Enums) == 0 &&
		len(s.Structs) == 0 &&
		len(s.Unions) == 0 &&
		len(s.Exceptions) == 0 &&
		len(s.Services) == 0
}

// GetPackageName 获取包名
func (s *Scope) GetPackageName() string {
	if s.Package != "" {
		return s.Package
	}
	// 从文件名生成包名
	name := filepath.Base(s.Filename)
	name = strings.TrimSuffix(name, ".thrift")
	return strings.ToLower(name)
}

// GetFileName 获取生成的文件名
func (s *Scope) GetFileName() string {
	base := filepath.Base(s.Filename)
	name := strings.TrimSuffix(base, ".thrift")
	return name + ".ts"
}

// GetSourceThriftFile 获取来源的 Thrift 文件路径
func (s *Scope) GetSourceThriftFile() string {
	return s.Filename
}

// BuildScope 构建 TypeScript 作用域
func BuildScope(utils *CodeUtils, ast *parser.Thrift) (*Scope, error) {
	scope := &Scope{
		Filename:        ast.Filename,
		Package:         "", // Thrift 结构没有 Name 字段
		utils:           utils,
		ExpandedStructs: make(map[string]*ExpandedStruct),
	}

	// 收集常量
	for _, constant := range ast.Constants {
		scope.Constants = append(scope.Constants, constant)
	}

	// 收集类型定义
	for _, typedef := range ast.Typedefs {
		scope.Typedefs = append(scope.Typedefs, typedef)
	}

	// 收集枚举
	for _, enum := range ast.Enums {
		scope.Enums = append(scope.Enums, enum)
	}

	// 收集结构体
	for _, structLike := range ast.Structs {
		scope.Structs = append(scope.Structs, structLike)
	}

	// 收集联合体
	for _, union := range ast.Unions {
		scope.Unions = append(scope.Unions, union)
	}

	// 收集异常
	for _, exception := range ast.Exceptions {
		scope.Exceptions = append(scope.Exceptions, exception)
	}

	// 收集服务
	for _, service := range ast.Services {
		scope.Services = append(scope.Services, service)
	}

	// 处理展开的结构体
	scope.processExpandedStructs(ast)

	// 收集导入信息（在展开结构体之后，以便收集展开字段的导入）
	scope.collectImports(ast)

	return scope, nil
}

// collectImports 收集导入信息
func (s *Scope) collectImports(ast *parser.Thrift) {
	importMap := make(map[string][]string)

	// 遍历所有结构体，收集外部类型引用
	for _, structLike := range ast.Structs {
		s.collectImportsFromStruct(structLike, importMap, ast)
	}

	for _, union := range ast.Unions {
		s.collectImportsFromStruct(union, importMap, ast)
	}

	for _, exception := range ast.Exceptions {
		s.collectImportsFromStruct(exception, importMap, ast)
	}

	// 遍历所有服务，收集外部类型引用
	for _, service := range ast.Services {
		s.collectImportsFromService(service, importMap, ast)
	}

	// 获取当前文件的 TypeScript namespace
	currentNamespace := s.utils.getTypeScriptNamespace(ast)

	// 转换为 ImportInfo 列表
	for module, types := range importMap {
		if len(types) > 0 {
			// 计算相对路径
			relativePath := s.calculateRelativePath(currentNamespace, module)

			s.Imports = append(s.Imports, ImportInfo{
				Module: module,
				Types:  types,
				Path:   relativePath,
			})
		}
	}
}

// collectImportsFromStruct 从结构体中收集导入信息
func (s *Scope) collectImportsFromStruct(structLike *parser.StructLike, importMap map[string][]string, ast *parser.Thrift) {
	// 获取展开字段名映射
	expandedFieldNames := make(map[string]bool)
	if expandedStruct, exists := s.ExpandedStructs[structLike.Name]; exists {
		expandedFieldNames = expandedStruct.ExpandedFieldNames
	}

	// 只收集未被展开的字段的导入信息
	for _, field := range structLike.Fields {
		if !expandedFieldNames[field.Name] {
			s.collectImportsFromType(field.Type, importMap, ast)
		}
	}

	// 收集展开字段的导入信息
	if expandedStruct, exists := s.ExpandedStructs[structLike.Name]; exists {
		for _, expandedField := range expandedStruct.ExpandedFields {
			s.collectImportsFromType(expandedField.Type, importMap, ast)
		}
	}
}

// collectImportsFromService 从服务中收集导入信息
func (s *Scope) collectImportsFromService(service *parser.Service, importMap map[string][]string, ast *parser.Thrift) {
	for _, function := range service.Functions {
		// 收集函数参数的导入
		for _, arg := range function.Arguments {
			s.collectImportsFromType(arg.Type, importMap, ast)
		}

		// 收集函数返回类型的导入
		if function.FunctionType != nil {
			s.collectImportsFromType(function.FunctionType, importMap, ast)
		}
	}
}

// collectImportsFromType 从类型中收集导入信息
func (s *Scope) collectImportsFromType(typ *parser.Type, importMap map[string][]string, ast *parser.Thrift) {
	if typ == nil {
		return
	}

	// 处理容器类型
	if typ.ValueType != nil {
		s.collectImportsFromType(typ.ValueType, importMap, ast)
	}
	if typ.KeyType != nil {
		s.collectImportsFromType(typ.KeyType, importMap, ast)
	}

	// 处理外部类型引用
	if typ.Name != "" && typ.Category >= parser.Category_Enum {
		// 检查是否是外部引用（包含点号）
		if strings.Contains(typ.Name, ".") {
			parts := strings.Split(typ.Name, ".")
			if len(parts) == 2 {
				module := parts[0]
				typeName := parts[1]

				// 检查是否是当前文件中定义的类型
				if s.isTypeDefinedInCurrentFile(typeName) {
					return
				}
				// 映射模块名到实际的 namespace
				actualModule := s.mapModuleToNamespace(module, ast)

				// 避免重复添加
				types := importMap[actualModule]
				found := false
				for _, t := range types {
					if t == typeName {
						found = true
						break
					}
				}
				if !found {
					importMap[actualModule] = append(types, typeName)
				}
			}
		} else {
			// 处理本地类型引用（不包含点号）
			// 检查是否是当前文件中定义的类型
			if !s.isTypeDefinedInCurrentFile(typ.Name) {
				// 对于本地类型引用，即使不在当前文件中定义，
				// 在分离文件模式下也需要导入（因为每个类型都会生成到单独的文件中）
				// 这里不需要特殊处理，让调用方处理导入逻辑
				return
			}
		}
	}
}

// processExpandedStructs 处理展开的结构体
func (s *Scope) processExpandedStructs(ast *parser.Thrift) {
	// 遍历所有结构体，检查是否有展开的字段
	for _, structLike := range ast.Structs {
		expandedFields, expandedFieldNames := s.collectExpandedFields(structLike, ast)
		if len(expandedFields) > 0 {
			s.ExpandedStructs[structLike.Name] = &ExpandedStruct{
				OriginalStruct:     structLike,
				ExpandedFields:     expandedFields,
				ExpandedFieldNames: expandedFieldNames,
			}
		}
	}
}

// collectExpandedFields 收集展开的字段
func (s *Scope) collectExpandedFields(structLike *parser.StructLike, ast *parser.Thrift) ([]*parser.Field, map[string]bool) {
	var expandedFields []*parser.Field
	expandedFieldNames := make(map[string]bool)

	for _, field := range structLike.Fields {
		// 检查字段是否应该展开
		shouldExpand := isExpandField(field)
		// 检查引用的结构体是否可展开
		referencedStruct := s.getReferencedStruct(field, ast)
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

// getReferencedStruct 获取引用的结构体
func (s *Scope) getReferencedStruct(field *parser.Field, ast *parser.Thrift) *parser.StructLike {
	if field.Type == nil || !field.Type.Category.IsStructLike() {
		return nil
	}

	typeName := field.Type.Name
	if !strings.Contains(typeName, ".") {
		// 在当前 AST 中查找结构体
		for _, structLike := range ast.Structs {
			if structLike.Name == typeName {
				return structLike
			}
		}

		for _, union := range ast.Unions {
			if union.Name == typeName {
				return union
			}
		}

		for _, exception := range ast.Exceptions {
			if exception.Name == typeName {
				return exception
			}
		}
		return nil
	}

	// 处理跨文件引用，需要在所有包含的文件中查找
	parts := strings.Split(typeName, ".")
	if len(parts) != 2 {
		return nil
	}

	moduleName := parts[0]
	structName := parts[1]

	// 遍历所有包含的文件
	for _, include := range ast.Includes {
		// 检查路径是否匹配（支持相对路径）
		includePath := include.Path
		expectedPath := moduleName + ".thrift"

		// 检查完整路径匹配
		if includePath == expectedPath {
			// 完整路径匹配
		} else {
			// 检查文件名匹配（处理相对路径）
			includeFileName := filepath.Base(includePath)
			if includeFileName != expectedPath {
				continue
			}
		}

		if include.Reference != nil {
			// 在包含的文件中查找结构体
			for _, structLike := range include.Reference.Structs {
				if structLike.Name == structName {
					return structLike
				}
			}

			for _, union := range include.Reference.Unions {
				if union.Name == structName {
					return union
				}
			}

			for _, exception := range include.Reference.Exceptions {
				if exception.Name == structName {
					return exception
				}
			}
		}
	}

	return nil
}

// CodeUtils TypeScript 代码生成工具
type CodeUtils struct {
	features  *Features
	log       backend.LogFunc
	rootScope *Scope
}

// Features TypeScript 生成特性
type Features struct {
	SkipEmpty          bool
	GenerateInterfaces bool
	GenerateClasses    bool
	UseStrictMode      bool
	UseES6Modules      bool
	// 命名风格选项
	SnakeStylePropertyName     bool // 使用 snake_case 命名属性
	LowerCamelCasePropertyName bool // 使用 lowerCamelCase 命名属性（默认）
}

// NewCodeUtils 创建新的代码工具
func NewCodeUtils(log backend.LogFunc) *CodeUtils {
	return &CodeUtils{
		features: &Features{
			SkipEmpty:                  false,
			GenerateInterfaces:         true,
			GenerateClasses:            false,
			UseStrictMode:              true,
			UseES6Modules:              true,
			SnakeStylePropertyName:     false,
			LowerCamelCasePropertyName: true, // 默认使用小驼峰命名
		},
		log: log,
	}
}

// Features 获取特性配置
func (u *CodeUtils) Features() *Features {
	return u.features
}

// SetRootScope 设置根作用域
func (u *CodeUtils) SetRootScope(scope *Scope) {
	u.rootScope = scope
}

// GetRootScope 获取根作用域
func (u *CodeUtils) GetRootScope() *Scope {
	return u.rootScope
}

// HandleOptions 处理生成选项
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
		case "snake_style_property_name":
			if value == "true" {
				u.features.SnakeStylePropertyName = true
				u.features.LowerCamelCasePropertyName = false
			}
		case "lower_camel_case_property_name":
			if value == "true" {
				u.features.LowerCamelCasePropertyName = true
				u.features.SnakeStylePropertyName = false
			}
		}
	}
	return nil
}

// BuildFuncMap 构建模板函数映射
func (u *CodeUtils) BuildFuncMap() map[string]interface{} {
	return map[string]interface{}{
		"GetTypeScriptType":        GetTypeScriptType,
		"GetFieldType":             GetFieldType,
		"GetMethodSignature":       GetMethodSignature,
		"GetAsyncMethodSignature":  GetAsyncMethodSignature,
		"GetInterfaceName":         GetInterfaceName,
		"GetClassName":             GetClassName,
		"GetEnumName":              GetEnumName,
		"GetEnumValueName":         GetEnumValueName,
		"GetPropertyName":          GetPropertyName,
		"GetPropertyNameWithStyle": func(name string) string { return GetPropertyNameWithStyle(name, u.features) },
		"GetConstantName":          GetConstantName,
		"IsOptional":               IsOptional,
		"GetDefaultValue":          GetDefaultValue,
		"GetDefaultValueForType":   GetDefaultValueForType,
		"GetConstantValue":         GetConstantValue,
		"IsExpandField":            isExpandField,
		"IsExpandableStruct":       isExpandableStruct,
		"GetExpandedFields":        func(structLike *parser.StructLike) []*parser.Field { return u.getExpandedFields(structLike) },
		"GetExpandedFieldNames":    func(structLike *parser.StructLike) map[string]bool { return u.getExpandedFieldNames(structLike) },
		"GetFieldExpandedFields":   GetFieldExpandedFields,
		"IsFieldExpanded": func(field *parser.Field, expandedFields []*parser.Field) bool {
			// 检查字段是否应该展开
			shouldExpand := isExpandField(field)
			if shouldExpand {
				return true
			}

			// 检查引用的结构体是否可展开
			if field.Type != nil && field.Type.Category.IsStructLike() {
				// 检查引用的结构体是否可展开
				// 如果该字段引用的结构体是可展开的，则该字段应该被展开
				if u.rootScope != nil {
					// 使用完整的 getReferencedStruct 实现
					referencedStruct := u.getReferencedStructFromAST(field)
					if referencedStruct != nil && isExpandableStruct(referencedStruct) {
						return true
					}
				}
			}

			return false
		},
		"GetStructFields":                      GetStructFields,
		"IsStructField":                        IsStructField,
		"GetStructFieldAnnotations":            GetStructFieldAnnotations,
		"GetStructFieldAnnotationsForTemplate": GetStructFieldAnnotationsForTemplate,
		"GetStructFieldByName":                 GetStructFieldByName,
		"GetStructComment":                     GetStructComment,
		"GetFieldComment":                      GetFieldComment,
		"GetEnumComment":                       GetEnumComment,
		"GetEnumValueComment":                  GetEnumValueComment,
		"GetServiceComment":                    GetServiceComment,
		"GetFunctionComment":                   GetFunctionComment,
		"GetPackageName":                       func(s *Scope) string { return s.GetPackageName() },
		"GetFileName":                          func(s *Scope) string { return s.GetFileName() },
		"GetSourceThriftFile":                  func(s *Scope) string { return s.GetSourceThriftFile() },
		"ToTitle":                              strings.Title,
		"ToLower":                              strings.ToLower,
		"ToUpper":                              strings.ToUpper,
		"HasSuffix":                            strings.HasSuffix,
	}
}

// CombineOutputPath 组合输出路径
func (u *CodeUtils) CombineOutputPath(basePath string, ast *parser.Thrift) string {
	if basePath == "" {
		basePath = "."
	}

	// 查找 TypeScript namespace
	tsNamespace := u.getTypeScriptNamespace(ast)
	if tsNamespace != "" {
		return filepath.Join(basePath, tsNamespace)
	}

	// 没有 TypeScript namespace，生成到根目录
	return basePath
}

// getTypeScriptNamespace 获取 TypeScript namespace
func (u *CodeUtils) getTypeScriptNamespace(ast *parser.Thrift) string {
	for _, ns := range ast.Namespaces {
		if ns.Language == "ts" || ns.Language == "typescript" {
			// 将点号转换为路径分隔符
			return strings.ReplaceAll(ns.Name, ".", "/")
		}
	}
	return ""
}

// mapModuleToNamespace 映射模块名到实际的 namespace
func (s *Scope) mapModuleToNamespace(module string, ast *parser.Thrift) string {
	// 根据 include 的文件名映射到实际的 namespace
	// 从 AST 中的 include 信息来动态映射

	// 优先选择相对路径的 include 文件（更可能是相关的）
	var relativeIncludes []*parser.Include
	var absoluteIncludes []*parser.Include

	for _, include := range ast.Includes {
		if include.Reference == nil {
			continue
		}

		// 提取文件名（去掉 .thrift 扩展名）
		fileName := strings.TrimSuffix(filepath.Base(include.Path), ".thrift")

		if fileName == module {
			if strings.HasPrefix(include.Path, "../") || strings.HasPrefix(include.Path, "./") {
				relativeIncludes = append(relativeIncludes, include)
			} else {
				absoluteIncludes = append(absoluteIncludes, include)
			}
		}
	}

	// 如果没有找到直接包含的文件，递归查找间接包含的文件
	return s.findModuleNamespaceRecursively(module, ast, make(map[string]bool))
}

// findModuleNamespaceRecursively 递归查找模块的 namespace
func (s *Scope) findModuleNamespaceRecursively(module string, ast *parser.Thrift, visited map[string]bool) string {
	// 防止循环引用
	if visited[ast.Filename] {
		return module
	}
	visited[ast.Filename] = true

	// 遍历所有包含的文件
	for _, include := range ast.Includes {
		if include.Reference == nil {
			continue
		}

		// 提取文件名（去掉 .thrift 扩展名）
		fileName := strings.TrimSuffix(filepath.Base(include.Path), ".thrift")

		if fileName == module {
			// 查找被引用文件的 TypeScript namespace
			for _, ns := range include.Reference.Namespaces {
				if ns.Language == "ts" || ns.Language == "typescript" {
					// 将点号转换为路径分隔符
					result := strings.ReplaceAll(ns.Name, ".", "/")
					return result
				}
			}
			// 如果没有找到 TypeScript namespace，使用文件名
			fileName := strings.TrimSuffix(filepath.Base(include.Path), ".thrift")
			return fileName
		}

		// 递归查找间接包含的文件
		if result := s.findModuleNamespaceRecursively(module, include.Reference, visited); result != module {
			return result
		}
	}

	// 如果没有找到对应的 include，返回原始模块名
	return module
}

// calculateRelativePath 计算相对路径
func (s *Scope) calculateRelativePath(currentNamespace, targetModule string) string {
	// 如果当前文件没有 namespace，目标文件也没有 namespace，使用相对路径
	if currentNamespace == "" {
		return "./" + targetModule

	}

	currentParts := strings.Split(currentNamespace, "/")
	targetParts := strings.Split(targetModule, "/")

	// 检查是否是兄弟目录（有相同的父目录）
	// 例如：common.base 到 common.enums 需要 ../enums
	if len(currentParts) > 1 && len(targetParts) > 1 {
		currentParent := strings.Join(currentParts[:len(currentParts)-1], "/")
		targetParent := strings.Join(targetParts[:len(targetParts)-1], "/")

		if currentParent == targetParent {
			// 兄弟目录，使用 ../ 前缀
			return "../" + targetParts[len(targetParts)-1]
		}
	}

	// 计算需要向上几级目录
	// 例如：从 domain/merchantVO 到 common/base 需要向上 2 级
	currentDepth := len(currentParts)
	// 计算向上级数
	upLevels := currentDepth
	// 构建相对路径
	var pathParts []string
	for i := 0; i < upLevels; i++ {
		pathParts = append(pathParts, "..")
	}
	pathParts = append(pathParts, targetParts...)
	return strings.Join(pathParts, "/")
}

// isTypeDefinedInCurrentFile 检查类型是否在当前文件中定义
func (s *Scope) isTypeDefinedInCurrentFile(typeName string) bool {
	// 检查枚举
	for _, enum := range s.Enums {
		if enum.Name == typeName {
			return true
		}
	}

	// 检查结构体
	for _, structLike := range s.Structs {
		if structLike.Name == typeName {
			return true
		}
	}

	// 检查联合体
	for _, union := range s.Unions {
		if union.Name == typeName {
			return true
		}
	}

	// 检查异常
	for _, exception := range s.Exceptions {
		if exception.Name == typeName {
			return true
		}
	}

	// 检查类型别名
	for _, typedef := range s.Typedefs {
		if typedef.Alias == typeName {
			return true
		}
	}

	return false
}

// isTypeDefinedInOtherFiles 检查类型是否在其他文件中定义
func (s *Scope) isTypeDefinedInOtherFiles(typeName string, ast *parser.Thrift) bool {
	// 遍历所有包含的文件
	for _, include := range ast.Includes {
		if include.Reference == nil {
			continue
		}

		// 检查枚举
		for _, enum := range include.Reference.Enums {
			if enum.Name == typeName {
				return true
			}
		}

		// 检查结构体
		for _, structLike := range include.Reference.Structs {
			if structLike.Name == typeName {
				return true
			}
		}

		// 检查联合体
		for _, union := range include.Reference.Unions {
			if union.Name == typeName {
				return true
			}
		}

		// 检查异常
		for _, exception := range include.Reference.Exceptions {
			if exception.Name == typeName {
				return true
			}
		}

		// 检查类型别名
		for _, typedef := range include.Reference.Typedefs {
			if typedef.Alias == typeName {
				return true
			}
		}
	}

	return false
}

// GetFilename 获取生成的文件名
func (u *CodeUtils) GetFilename(ast *parser.Thrift) string {
	base := filepath.Base(ast.Filename)
	name := strings.TrimSuffix(base, ".thrift")
	return name + ".ts"
}

// getExpandedFields 获取结构体的展开字段（无调试信息）
func (u *CodeUtils) getExpandedFields(structLike *parser.StructLike) []*parser.Field {
	if u.rootScope == nil {
		return nil
	}

	expandedStruct, exists := u.rootScope.ExpandedStructs[structLike.Name]
	if !exists {
		return nil
	}

	return expandedStruct.ExpandedFields
}

// getExpandedFieldNames 获取结构体的展开字段名映射
func (u *CodeUtils) getExpandedFieldNames(structLike *parser.StructLike) map[string]bool {
	if u.rootScope == nil {
		return nil
	}

	expandedStruct, exists := u.rootScope.ExpandedStructs[structLike.Name]
	if !exists {
		return nil
	}

	return expandedStruct.ExpandedFieldNames
}

// getReferencedStructFromAST 获取引用的结构体（使用完整 AST 信息）
func (u *CodeUtils) getReferencedStructFromAST(field *parser.Field) *parser.StructLike {
	if field.Type == nil || !field.Type.Category.IsStructLike() {
		return nil
	}

	typeName := field.Type.Name
	if !strings.Contains(typeName, ".") {
		return nil
	}

	// 处理跨文件引用
	parts := strings.Split(typeName, ".")
	if len(parts) != 2 {
		return nil
	}

	_ = parts[0] // moduleName
	_ = parts[1] // structName

	// 遍历所有包含的文件
	if u.rootScope != nil {
		// 从当前作用域获取 AST 信息
		// 这里需要从其他地方获取 AST，暂时返回 nil
		// 实际实现中需要从 backend 或其他地方获取完整的 AST
		return nil
	}

	return nil
}

// getReferencedStruct 获取引用的结构体（简化版本）
func (u *CodeUtils) getReferencedStruct(field *parser.Field) *parser.StructLike {
	return u.getReferencedStructFromAST(field)
}
