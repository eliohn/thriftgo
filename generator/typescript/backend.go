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
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
)

// TypeScriptBackend generates TypeScript codes.
// The zero value of TypeScriptBackend is ready for use.
type TypeScriptBackend struct {
	err error
	tpl *template.Template
	req *plugin.Request
	res *plugin.Response
	log backend.LogFunc

	utils *CodeUtils
	funcs template.FuncMap
}

// Name implements the Backend interface.
func (t *TypeScriptBackend) Name() string {
	return "typescript"
}

// Lang implements the Backend interface.
func (t *TypeScriptBackend) Lang() string {
	return "TypeScript"
}

// Options implements the Backend interface.
func (t *TypeScriptBackend) Options() (opts []plugin.Option) {
	for _, p := range allParams {
		opts = append(opts, plugin.Option{
			Name: p.name,
			Desc: p.desc,
		})
	}
	return opts
}

// BuiltinPlugins implements the Backend interface.
func (t *TypeScriptBackend) BuiltinPlugins() []*plugin.Desc {
	return nil
}

// GetPlugin implements the Backend interface.
func (t *TypeScriptBackend) GetPlugin(desc *plugin.Desc) plugin.Plugin {
	return nil
}

// Generate implements the Backend interface.
func (t *TypeScriptBackend) Generate(req *plugin.Request, log backend.LogFunc) *plugin.Response {
	t.req = req
	t.res = plugin.NewResponse()
	t.log = log
	t.prepareUtilities()
	if t.err != nil {
		return t.buildResponse()
	}

	// 设置全局 AST，供模板函数使用
	SetGlobalAST(req.AST)

	t.prepareTemplates()
	t.fillRequisitions()
	t.executeTemplates()
	return t.buildResponse()
}

func (t *TypeScriptBackend) GetCoreUtils() *CodeUtils {
	return t.utils
}

func (t *TypeScriptBackend) prepareUtilities() {
	if t.err != nil {
		return
	}

	t.utils = NewCodeUtils(t.log)
	// 处理生成选项
	t.err = t.utils.HandleOptions(t.req.GeneratorParameters)
	if t.err != nil {
		return
	}

	t.funcs = t.utils.BuildFuncMap()
	t.funcs["Version"] = func() string { return t.req.Version }
}

func (t *TypeScriptBackend) prepareTemplates() {
	if t.err != nil {
		return
	}

	all := template.New("thrift").Funcs(t.funcs)
	tpls := Templates()

	for _, tpl := range tpls {
		all = template.Must(all.Parse(tpl))
	}
	t.tpl = all
}

func (t *TypeScriptBackend) fillRequisitions() {
	if t.err != nil {
		return
	}
}

func (t *TypeScriptBackend) executeTemplates() {
	if t.err != nil {
		return
	}

	processed := make(map[*parser.Thrift]bool)

	var trees chan *parser.Thrift
	if t.req.Recursive {
		trees = t.req.AST.DepthFirstSearch()
	} else {
		trees = make(chan *parser.Thrift, 1)
		trees <- t.req.AST
		close(trees)
	}

	for ast := range trees {
		if processed[ast] {
			continue
		}
		processed[ast] = true
		t.log.Info("Processing", ast.Filename)

		if t.err = t.renderOneFile(ast); t.err != nil {
			break
		}
	}
}

func (t *TypeScriptBackend) renderOneFile(ast *parser.Thrift) error {
	scope, err := BuildScope(t.utils, ast)
	if err != nil {
		return err
	}

	// 检查是否有 TypeScript namespace
	tsNamespace := t.utils.getTypeScriptNamespace(ast)
	if tsNamespace != "" {
		// 有 namespace，生成到对应文件夹
		path := t.utils.CombineOutputPath(t.req.OutputPath, ast)
		return t.renderSeparateFiles(scope, t.tpl, path, ast)
	} else {
		// 没有 namespace，生成到根目录
		path := t.utils.CombineOutputPath(t.req.OutputPath, ast)
		return t.renderByTemplate(scope, t.tpl, filepath.Join(path, t.utils.GetFilename(ast)))
	}
}

// renderSeparateFiles 为每个类型生成单独的文件
func (t *TypeScriptBackend) renderSeparateFiles(scope *Scope, executeTpl *template.Template, basePath string, ast *parser.Thrift) error {
	// 生成 index.ts 文件（包含所有导入和导出）
	if err := t.renderIndexFile(scope, executeTpl, basePath); err != nil {
		return err
	}

	// 为每个枚举生成单独文件
	for _, enum := range scope.Enums {
		if err := t.renderEnumFile(scope, executeTpl, basePath, enum); err != nil {
			return err
		}
	}

	// 为每个结构体生成单独文件
	for _, structLike := range scope.Structs {
		if err := t.renderStructFile(scope, executeTpl, basePath, structLike, ast); err != nil {
			return err
		}
		// 检查是否需要生成 fields.ts 文件
		if ShouldGenerateFieldsFile(structLike) {
			if err := t.renderFieldsFile(scope, executeTpl, basePath, structLike); err != nil {
				return err
			}
		}
	}

	// 为每个联合体生成单独文件
	for _, union := range scope.Unions {
		if err := t.renderStructFile(scope, executeTpl, basePath, union, ast); err != nil {
			return err
		}
	}

	// 为每个异常生成单独文件
	for _, exception := range scope.Exceptions {
		if err := t.renderStructFile(scope, executeTpl, basePath, exception, ast); err != nil {
			return err
		}
	}

	// 为每个服务生成单独文件
	for _, service := range scope.Services {
		if err := t.renderServiceFile(scope, executeTpl, basePath, service); err != nil {
			return err
		}
	}

	// 生成简化版服务实现类文件（如果有服务的话）
	if len(scope.Services) > 0 {
		if err := t.renderSimpleServiceImplementationFiles(scope, executeTpl, basePath); err != nil {
			return err
		}
	}

	return nil
}

var poolBuffer = sync.Pool{
	New: func() any {
		p := &bytes.Buffer{}
		p.Grow(100 << 10)
		return p
	},
}

func (t *TypeScriptBackend) renderByTemplate(scope *Scope, executeTpl *template.Template, filename string) error {
	if scope == nil {
		return nil
	}

	// if scope has no content, just skip and don't generate this file
	if t.utils.Features().SkipEmpty {
		if scope.IsEmpty() {
			return nil
		}
	}

	w := poolBuffer.Get().(*bytes.Buffer)
	defer poolBuffer.Put(w)

	w.Reset()

	t.utils.SetRootScope(scope)
	err := executeTpl.ExecuteTemplate(w, executeTpl.Name(), scope)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	t.res.Contents = append(t.res.Contents, &plugin.Generated{
		Content: w.String(),
		Name:    &filename,
	})
	return nil
}

func (t *TypeScriptBackend) buildResponse() *plugin.Response {
	if t.err != nil {
		return plugin.BuildErrorResponse(t.err.Error())
	}
	return t.res
}

// PostProcess implements the backend.PostProcessor interface to do
// source formatting before writing files out.
func (t *TypeScriptBackend) PostProcess(path string, content []byte) ([]byte, error) {
	// TypeScript 不需要特殊的格式化，直接返回
	return content, nil
}

// renderIndexFile 生成 index.ts 文件
func (t *TypeScriptBackend) renderIndexFile(scope *Scope, executeTpl *template.Template, basePath string) error {
	filename := filepath.Join(basePath, "index.ts")

	// index.ts 只导入外部文件，不导入本地文件
	externalImports := []ImportInfo{}
	for _, imp := range scope.Imports {
		externalImports = append(externalImports, imp)
	}

	// 创建只包含外部导入的 scope
	indexScope := &Scope{
		Filename:   scope.Filename,
		Package:    scope.Package,
		Imports:    externalImports,
		Enums:      scope.Enums,
		Structs:    scope.Structs,
		Unions:     scope.Unions,
		Exceptions: scope.Exceptions,
		Services:   scope.Services,
		utils:      scope.utils,
	}

	return t.renderByTemplateWithTemplate(indexScope, executeTpl, filename, "index")
}

// renderEnumFile 生成枚举文件
func (t *TypeScriptBackend) renderEnumFile(scope *Scope, executeTpl *template.Template, basePath string, enum *parser.Enum) error {
	filename := filepath.Join(basePath, strings.ToLower(enum.Name)+".ts")

	// 为枚举单独收集导入（枚举通常不需要外部导入）
	enumImports := []ImportInfo{}

	// 创建只包含该枚举的 scope
	enumScope := &Scope{
		Filename: scope.Filename,
		Package:  scope.Package,
		Imports:  enumImports,
		Enums:    []*parser.Enum{enum},
		utils:    scope.utils,
	}

	return t.renderByTemplateWithTemplate(enumScope, executeTpl, filename, "singleEnum")
}

// renderStructFile 生成结构体文件
func (t *TypeScriptBackend) renderStructFile(scope *Scope, executeTpl *template.Template, basePath string, structLike *parser.StructLike, ast *parser.Thrift) error {
	filename := filepath.Join(basePath, strings.ToLower(structLike.Name)+".ts")

	// 为单个结构体收集导入
	structImports := t.collectImportsForStruct(scope, structLike, ast)

	// 创建只包含该结构体的 scope
	structScope := &Scope{
		Filename:        scope.Filename,
		Package:         scope.Package,
		Imports:         structImports,
		Structs:         []*parser.StructLike{structLike},
		ExpandedStructs: scope.ExpandedStructs,
		utils:           scope.utils,
	}

	return t.renderByTemplateWithTemplate(structScope, executeTpl, filename, "singleStruct")
}

// renderServiceFile 生成服务文件
func (t *TypeScriptBackend) renderServiceFile(scope *Scope, executeTpl *template.Template, basePath string, service *parser.Service) error {
	filename := filepath.Join(basePath, strings.ToLower(service.Name)+".ts")

	// 创建只包含该服务的 scope
	serviceScope := &Scope{
		Filename:        scope.Filename,
		Package:         scope.Package,
		Imports:         []ImportInfo{}, // 不继承父 scope 的导入信息
		Services:        []*parser.Service{service},
		Structs:         scope.Structs,         // 复制结构体信息
		Unions:          scope.Unions,          // 复制联合体信息
		Exceptions:      scope.Exceptions,      // 复制异常信息
		Enums:           scope.Enums,           // 复制枚举信息
		Typedefs:        scope.Typedefs,        // 复制类型别名信息
		ExpandedStructs: scope.ExpandedStructs, // 复制展开结构体信息
		utils:           scope.utils,
	}

	// 为服务接口文件单独收集导入信息
	ast := GetGlobalAST()
	if ast != nil {
		// 在分离文件模式下，服务文件需要导入其他类型文件
		serviceScope.collectImportsForService(ast, service.Name)
	}

	return t.renderByTemplateWithTemplate(serviceScope, executeTpl, filename, "singleService")
}

// collectImportsForStruct 为单个结构体收集导入
func (t *TypeScriptBackend) collectImportsForStruct(scope *Scope, structLike *parser.StructLike, ast *parser.Thrift) []ImportInfo {
	importMap := make(map[string][]string)
	localTypes := make(map[string]bool)

	// 创建一个临时的 scope，只包含当前结构体，用于正确识别自引用
	tempScope := &Scope{
		Filename:        scope.Filename,
		Package:         scope.Package,
		Structs:         []*parser.StructLike{structLike},
		ExpandedStructs: scope.ExpandedStructs,
		utils:           scope.utils,
	}

	// 获取展开字段名映射
	expandedFieldNames := make(map[string]bool)
	if expandedStruct, exists := scope.ExpandedStructs[structLike.Name]; exists {
		expandedFieldNames = expandedStruct.ExpandedFieldNames
	}

	// 只收集未被展开的字段的导入
	for _, field := range structLike.Fields {
		if !expandedFieldNames[field.Name] {
			tempScope.collectImportsFromTypeWithCurrentFile(field.Type, importMap, ast, structLike.Name)
			// 检查字段类型及其容器类型中的本地类型引用
			t.collectLocalTypesFromType(field.Type, localTypes)
		}
	}

	// 收集展开字段的导入
	if expandedStruct, exists := scope.ExpandedStructs[structLike.Name]; exists {
		for _, expandedField := range expandedStruct.ExpandedFields {
			tempScope.collectImportsFromTypeWithCurrentFile(expandedField.Type, importMap, ast, structLike.Name)
			// 检查字段类型及其容器类型中的本地类型引用
			t.collectLocalTypesFromType(expandedField.Type, localTypes)
		}
	}

	// 获取当前文件的 TypeScript namespace
	// 在分离文件模式下，需要根据结构体的实际位置确定 namespace
	currentNamespace := t.utils.getTypeScriptNamespace(ast)

	// 转换为 ImportInfo 列表，并去重
	importSet := make(map[string]ImportInfo)
	for module, types := range importMap {
		if len(types) > 0 {
			// 计算相对路径
			relativePath := scope.calculateRelativePath(currentNamespace, module)

			// 创建导入键，用于去重
			importKey := relativePath

			// 如果已经存在相同路径的导入，合并类型
			if existingImport, exists := importSet[importKey]; exists {
				// 合并类型列表，去重
				existingTypes := make(map[string]bool)
				for _, t := range existingImport.Types {
					existingTypes[t] = true
				}
				for _, t := range types {
					if !existingTypes[t] {
						existingImport.Types = append(existingImport.Types, t)
					}
				}
				importSet[importKey] = existingImport
			} else {
				importSet[importKey] = ImportInfo{
					Module: module,
					Types:  types,
					Path:   relativePath,
				}
			}
		}
	}

	// 添加本地类型导入
	for typeName := range localTypes {
		// 在分离文件模式下，所有本地类型引用都需要导入
		// 因为每个类型都会生成到单独的文件中
		localPath := "./" + strings.ToLower(typeName)
		importKey := localPath

		// 如果已经存在相同路径的导入，合并类型
		if existingImport, exists := importSet[importKey]; exists {
			// 检查类型是否已存在
			found := false
			for _, t := range existingImport.Types {
				if t == typeName {
					found = true
					break
				}
			}
			if !found {
				existingImport.Types = append(existingImport.Types, typeName)
				importSet[importKey] = existingImport
			}
		} else {
			importSet[importKey] = ImportInfo{
				Module: ".",
				Types:  []string{typeName},
				Path:   localPath,
			}
		}
	}

	// 将去重后的导入添加到列表中，过滤掉自引用的导入
	var imports []ImportInfo
	for _, importInfo := range importSet {
		// 过滤掉自引用的导入
		if !isSelfReferenceImport(importInfo, structLike.Name) {
			imports = append(imports, importInfo)
		}
	}

	return imports
}

// isSelfReferenceImport 检查导入是否是自引用
func isSelfReferenceImport(importInfo ImportInfo, currentStructName string) bool {
	// 检查导入路径是否是当前结构体的路径
	expectedPath := "./" + strings.ToLower(currentStructName)

	if importInfo.Path == expectedPath {
		// 检查导入的类型是否包含当前结构体名称
		for _, typeName := range importInfo.Types {
			if typeName == currentStructName {
				return true
			}
		}
	}
	return false
}

// findModuleForType 查找类型所在的模块
func (t *TypeScriptBackend) findModuleForType(scope *Scope, typeName string) string {
	// 这里简化处理，假设类型在根目录的对应文件中
	// 实际实现中应该根据 AST 信息查找
	return strings.ToLower(typeName)
}

// isTypeInSameThriftFile 检查类型是否在同一个 thrift 文件中定义
func (t *TypeScriptBackend) isTypeInSameThriftFile(scope *Scope, typeName string) bool {
	// 在分离文件模式下，即使类型在同一个 thrift 文件中定义，
	// 也会生成到不同的 ts 文件中，所以需要导入
	return true
}

// isPrimitiveType 检查是否为基本类型
func (t *TypeScriptBackend) isPrimitiveType(typ *parser.Type) bool {
	if typ == nil {
		return false
	}
	switch typ.Category {
	case parser.Category_Bool, parser.Category_Byte, parser.Category_I16,
		parser.Category_I32, parser.Category_I64, parser.Category_Double,
		parser.Category_String, parser.Category_Binary:
		return true
	case parser.Category_List, parser.Category_Map, parser.Category_Set:
		return true
	default:
		return false
	}
}

// collectLocalTypesFromType 从类型中收集本地类型引用
func (t *TypeScriptBackend) collectLocalTypesFromType(typ *parser.Type, localTypes map[string]bool) {
	if typ == nil {
		return
	}

	// 处理容器类型
	if typ.ValueType != nil {
		t.collectLocalTypesFromType(typ.ValueType, localTypes)
	}
	if typ.KeyType != nil {
		t.collectLocalTypesFromType(typ.KeyType, localTypes)
	}

	// 检查是否是本地类型引用（不包含点号的类型名且不是基本类型）
	if typ.Name != "" && !strings.Contains(typ.Name, ".") && !t.isPrimitiveType(typ) {
		localTypes[typ.Name] = true
	}
}

// isTypeDefinedInFile 检查类型是否在文件中定义
func (t *TypeScriptBackend) isTypeDefinedInFile(scope *Scope, typeName string) bool {
	// 检查枚举
	for _, enum := range scope.Enums {
		if enum.Name == typeName {
			return true
		}
	}

	// 检查结构体
	for _, structLike := range scope.Structs {
		if structLike.Name == typeName {
			return true
		}
	}

	// 检查联合体
	for _, union := range scope.Unions {
		if union.Name == typeName {
			return true
		}
	}

	// 检查异常
	for _, exception := range scope.Exceptions {
		if exception.Name == typeName {
			return true
		}
	}

	// 检查服务
	for _, service := range scope.Services {
		if service.Name == typeName {
			return true
		}
	}

	// 检查类型定义
	for _, typedef := range scope.Typedefs {
		if typedef.Alias == typeName {
			return true
		}
	}

	return false
}

// renderByTemplateWithTemplate 使用指定模板渲染文件
func (t *TypeScriptBackend) renderByTemplateWithTemplate(scope *Scope, executeTpl *template.Template, filename string, templateName string) error {
	if scope == nil {
		return nil
	}

	// if scope has no content, just skip and don't generate this file
	if t.utils.Features().SkipEmpty {
		if scope.IsEmpty() {
			return nil
		}
	}

	w := poolBuffer.Get().(*bytes.Buffer)
	defer poolBuffer.Put(w)

	w.Reset()

	t.utils.SetRootScope(scope)
	err := executeTpl.ExecuteTemplate(w, templateName, scope)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	t.res.Contents = append(t.res.Contents, &plugin.Generated{
		Content: w.String(),
		Name:    &filename,
	})
	return nil
}

// renderServiceImplementationFiles 生成服务实现类文件
func (t *TypeScriptBackend) renderServiceImplementationFiles(scope *Scope, executeTpl *template.Template, basePath string) error {
	// 为每个服务生成实现类文件
	for _, service := range scope.Services {
		if err := t.renderServiceImplementationFile(scope, executeTpl, basePath, service); err != nil {
			return err
		}
	}
	return nil
}

// renderServiceImplementationFile 生成单个服务的实现类文件
func (t *TypeScriptBackend) renderServiceImplementationFile(scope *Scope, executeTpl *template.Template, basePath string, service *parser.Service) error {
	filename := filepath.Join(basePath, strings.ToLower(service.Name)+"impl.ts")

	// 创建只包含该服务的 scope
	serviceScope := &Scope{
		Filename: scope.Filename,
		Package:  scope.Package,
		Imports:  scope.Imports,
		Services: []*parser.Service{service},
		utils:    scope.utils,
	}

	return t.renderByTemplateWithTemplate(serviceScope, executeTpl, filename, "serviceImplementation")
}

// renderSimpleServiceImplementationFiles 生成简化版服务实现类文件
func (t *TypeScriptBackend) renderSimpleServiceImplementationFiles(scope *Scope, executeTpl *template.Template, basePath string) error {
	// 为每个服务生成简化版实现类文件
	for _, service := range scope.Services {
		if err := t.renderSimpleServiceImplementationFile(scope, executeTpl, basePath, service); err != nil {
			return err
		}
	}
	return nil
}

// renderSimpleServiceImplementationFile 生成单个服务的简化版实现类文件
func (t *TypeScriptBackend) renderSimpleServiceImplementationFile(scope *Scope, executeTpl *template.Template, basePath string, service *parser.Service) error {
	filename := filepath.Join(basePath, strings.ToLower(service.Name)+"client.ts")

	// 创建只包含该服务的 scope
	serviceScope := &Scope{
		Filename:        scope.Filename,
		Package:         scope.Package,
		Imports:         []ImportInfo{}, // 不继承父 scope 的导入信息
		Services:        []*parser.Service{service},
		Structs:         scope.Structs,         // 复制结构体信息
		Unions:          scope.Unions,          // 复制联合体信息
		Exceptions:      scope.Exceptions,      // 复制异常信息
		Enums:           scope.Enums,           // 复制枚举信息
		Typedefs:        scope.Typedefs,        // 复制类型别名信息
		ExpandedStructs: scope.ExpandedStructs, // 复制展开结构体信息
		utils:           scope.utils,
	}

	// 为客户端文件单独收集导入信息
	ast := GetGlobalAST()
	if ast != nil {
		// 在分离文件模式下，客户端文件需要导入其他类型文件
		serviceScope.collectImportsForService(ast, service.Name)
	}

	return t.renderByTemplateWithTemplate(serviceScope, executeTpl, filename, "simpleServiceImplementation")
}

// renderFieldsFile 生成 fields.ts 文件
func (t *TypeScriptBackend) renderFieldsFile(scope *Scope, executeTpl *template.Template, basePath string, structLike *parser.StructLike) error {
	filename := filepath.Join(basePath, GetFieldsFileName(structLike))

	// 创建只包含该结构体的 scope
	structScope := &Scope{
		Filename:        scope.Filename,
		Package:         scope.Package,
		Imports:         []ImportInfo{}, // fields.ts 文件只需要导入对应的接口文件
		Structs:         []*parser.StructLike{structLike},
		ExpandedStructs: scope.ExpandedStructs,
		utils:           scope.utils,
	}

	return t.renderByTemplateWithTemplate(structScope, executeTpl, filename, "fields")
}
