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

	"github.com/cloudwego/thriftgo/parser"
)

// Scope represents the scope for OpenAPI generation.
type Scope struct {
	Filename        string
	Package         string
	Imports         []ImportInfo
	Enums           []*parser.Enum
	Structs         []*parser.StructLike
	Unions          []*parser.StructLike
	Exceptions      []*parser.StructLike
	Services        []*parser.Service
	Typedefs        []*parser.Typedef
	Constants       []*parser.Constant
	ExpandedStructs map[string]*ExpandedStruct
	utils           *CodeUtils
}

// ExpandedStruct 表示展开的结构体信息
type ExpandedStruct struct {
	StructLike          *parser.StructLike
	ExpandedFields      []*parser.Field
	ExpandedFieldNames  map[string]bool
}

// ImportInfo represents import information.
type ImportInfo struct {
	Module string
	Path   string
}

// BuildScope builds a scope from a Thrift AST.
func BuildScope(utils *CodeUtils, ast *parser.Thrift) (*Scope, error) {
	scope := &Scope{
		Filename:        ast.Filename,
		Package:         getPackageName(ast),
		Imports:         buildImports(ast),
		Enums:           ast.Enums,
		Structs:         ast.Structs,
		Unions:          ast.Unions,
		Exceptions:      ast.Exceptions,
		Services:        ast.Services,
		Typedefs:        ast.Typedefs,
		Constants:       ast.Constants,
		ExpandedStructs: make(map[string]*ExpandedStruct),
		utils:           utils,
	}

	// 处理结构体展开
	processExpandedStructs(scope, ast)

	return scope, nil
}

// processExpandedStructs 处理结构体展开
func processExpandedStructs(scope *Scope, ast *parser.Thrift) {
	// 处理结构体
	for _, structLike := range ast.Structs {
		expandedFields, expandedFieldNames := scope.utils.collectExpandedFields(structLike, ast)
		if len(expandedFields) > 0 {
			scope.ExpandedStructs[structLike.Name] = &ExpandedStruct{
				StructLike:         structLike,
				ExpandedFields:     expandedFields,
				ExpandedFieldNames: expandedFieldNames,
			}
		}
	}

	// 处理联合体
	for _, union := range ast.Unions {
		expandedFields, expandedFieldNames := scope.utils.collectExpandedFields(union, ast)
		if len(expandedFields) > 0 {
			scope.ExpandedStructs[union.Name] = &ExpandedStruct{
				StructLike:         union,
				ExpandedFields:     expandedFields,
				ExpandedFieldNames: expandedFieldNames,
			}
		}
	}

	// 处理异常
	for _, exception := range ast.Exceptions {
		expandedFields, expandedFieldNames := scope.utils.collectExpandedFields(exception, ast)
		if len(expandedFields) > 0 {
			scope.ExpandedStructs[exception.Name] = &ExpandedStruct{
				StructLike:         exception,
				ExpandedFields:     expandedFields,
				ExpandedFieldNames: expandedFieldNames,
			}
		}
	}
}

// getPackageName extracts the package name from the Thrift file.
func getPackageName(ast *parser.Thrift) string {
	// 从文件名中提取包名
	base := filepath.Base(ast.Filename)
	name := strings.TrimSuffix(base, ".thrift")
	return name
}

// buildImports builds import information from the Thrift AST.
func buildImports(ast *parser.Thrift) []ImportInfo {
	var imports []ImportInfo
	for _, imp := range ast.Includes {
		imports = append(imports, ImportInfo{
			Module: imp.Path,
			Path:   imp.Path,
		})
	}
	return imports
}

// IsEmpty checks if the scope is empty.
func (s *Scope) IsEmpty() bool {
	return len(s.Enums) == 0 &&
		len(s.Structs) == 0 &&
		len(s.Unions) == 0 &&
		len(s.Exceptions) == 0 &&
		len(s.Services) == 0 &&
		len(s.Typedefs) == 0 &&
		len(s.Constants) == 0
}

// GetOpenAPIVersion returns the OpenAPI version.
func (s *Scope) GetOpenAPIVersion() string {
	return s.utils.Features().Version
}

// GetAPITitle returns the API title.
func (s *Scope) GetAPITitle() string {
	return s.utils.Features().Title
}

// GetBasePath returns the base path.
func (s *Scope) GetBasePath() string {
	return s.utils.Features().BasePath
}

// GetAPIDescription returns the API description.
func (s *Scope) GetAPIDescription() string {
	// 可以从注释或配置中获取
	return fmt.Sprintf("基于 %s 生成的 API 文档", s.Package)
}

// GetContactInfo returns contact information.
func (s *Scope) GetContactInfo() map[string]string {
	contact := make(map[string]string)
	if name, ok := s.utils.options["contact_name"]; ok {
		contact["name"] = name
	}
	if email, ok := s.utils.options["contact_email"]; ok {
		contact["email"] = email
	}
	if url, ok := s.utils.options["contact_url"]; ok {
		contact["url"] = url
	}
	return contact
}

// GetLicenseInfo returns license information.
func (s *Scope) GetLicenseInfo() map[string]string {
	license := make(map[string]string)
	if name, ok := s.utils.options["license_name"]; ok {
		license["name"] = name
	}
	if url, ok := s.utils.options["license_url"]; ok {
		license["url"] = url
	}
	return license
}

// GetServerInfo returns server information.
func (s *Scope) GetServerInfo() map[string]string {
	server := make(map[string]string)
	if url, ok := s.utils.options["server_url"]; ok {
		server["url"] = url
	}
	if desc, ok := s.utils.options["server_description"]; ok {
		server["description"] = desc
	}
	return server
}

// GetAllSchemas returns all schemas defined in the scope.
func (s *Scope) GetAllSchemas() []interface{} {
	var schemas []interface{}
	
	// 添加枚举
	for _, enum := range s.Enums {
		schemas = append(schemas, enum)
	}
	
	// 添加结构体
	for _, structLike := range s.Structs {
		schemas = append(schemas, structLike)
	}
	
	// 添加联合体
	for _, union := range s.Unions {
		schemas = append(schemas, union)
	}
	
	// 添加异常
	for _, exception := range s.Exceptions {
		schemas = append(schemas, exception)
	}
	
	return schemas
}

// GetAllServices returns all services defined in the scope.
func (s *Scope) GetAllServices() []*parser.Service {
	return s.Services
}

// GetSchemaByName returns a schema by name.
func (s *Scope) GetSchemaByName(name string) interface{} {
	// 查找枚举
	for _, enum := range s.Enums {
		if enum.Name == name {
			return enum
		}
	}
	
	// 查找结构体
	for _, structLike := range s.Structs {
		if structLike.Name == name {
			return structLike
		}
	}
	
	// 查找联合体
	for _, union := range s.Unions {
		if union.Name == name {
			return union
		}
	}
	
	// 查找异常
	for _, exception := range s.Exceptions {
		if exception.Name == name {
			return exception
		}
	}
	
	return nil
}

// GetServiceByName returns a service by name.
func (s *Scope) GetServiceByName(name string) *parser.Service {
	for _, service := range s.Services {
		if service.Name == name {
			return service
		}
	}
	return nil
}
