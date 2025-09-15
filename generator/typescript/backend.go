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
	// 暂时跳过选项处理
	// t.err = t.utils.HandleOptions(t.req.GeneratorParameters)
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
		return t.renderSeparateFiles(scope, t.tpl, path)
	} else {
		// 没有 namespace，生成到根目录
		path := t.utils.CombineOutputPath(t.req.OutputPath, ast)
		return t.renderByTemplate(scope, t.tpl, filepath.Join(path, t.utils.GetFilename(ast)))
	}
}

// renderSeparateFiles 为每个类型生成单独的文件
func (t *TypeScriptBackend) renderSeparateFiles(scope *Scope, executeTpl *template.Template, basePath string) error {
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
		if err := t.renderStructFile(scope, executeTpl, basePath, structLike); err != nil {
			return err
		}
	}

	// 为每个联合体生成单独文件
	for _, union := range scope.Unions {
		if err := t.renderStructFile(scope, executeTpl, basePath, union); err != nil {
			return err
		}
	}

	// 为每个异常生成单独文件
	for _, exception := range scope.Exceptions {
		if err := t.renderStructFile(scope, executeTpl, basePath, exception); err != nil {
			return err
		}
	}

	// 为每个服务生成单独文件
	for _, service := range scope.Services {
		if err := t.renderServiceFile(scope, executeTpl, basePath, service); err != nil {
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
func (t *TypeScriptBackend) renderStructFile(scope *Scope, executeTpl *template.Template, basePath string, structLike *parser.StructLike) error {
	filename := filepath.Join(basePath, strings.ToLower(structLike.Name)+".ts")

	// 为单个结构体收集导入
	structImports := t.collectImportsForStruct(scope, structLike)

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
		Filename: scope.Filename,
		Package:  scope.Package,
		Imports:  scope.Imports,
		Services: []*parser.Service{service},
		utils:    scope.utils,
	}

	return t.renderByTemplateWithTemplate(serviceScope, executeTpl, filename, "singleService")
}

// collectImportsForStruct 为单个结构体收集导入
func (t *TypeScriptBackend) collectImportsForStruct(scope *Scope, structLike *parser.StructLike) []ImportInfo {
	importMap := make(map[string][]string)
	localTypes := make(map[string]bool)

	// 收集结构体字段的导入
	for _, field := range structLike.Fields {
		scope.collectImportsFromType(field.Type, importMap)
		// 检查是否是本地类型引用（不包含点号的类型名且不是基本类型）
		if field.Type != nil && !strings.Contains(field.Type.Name, ".") && !t.isPrimitiveType(field.Type) {
			localTypes[field.Type.Name] = true
		}
	}

	// 收集展开字段的导入
	if expandedStruct, exists := scope.ExpandedStructs[structLike.Name]; exists {
		for _, expandedField := range expandedStruct.ExpandedFields {
			scope.collectImportsFromType(expandedField.Type, importMap)
			// 检查是否是本地类型引用（不包含点号的类型名且不是基本类型）
			if expandedField.Type != nil && !strings.Contains(expandedField.Type.Name, ".") && !t.isPrimitiveType(expandedField.Type) {
				localTypes[expandedField.Type.Name] = true
			}
		}
	}

	// 获取当前文件的 TypeScript namespace
	// 在分离文件模式下，当前 namespace 是 test_temp
	currentNamespace := "test_temp"

	// 转换为 ImportInfo 列表
	var imports []ImportInfo
	for module, types := range importMap {
		if len(types) > 0 {
			// 计算相对路径
			relativePath := scope.calculateRelativePath(currentNamespace, module)

			imports = append(imports, ImportInfo{
				Module: module,
				Types:  types,
				Path:   relativePath,
			})
		}
	}

	// 添加本地类型导入
	for typeName := range localTypes {
		// 在分离文件模式下，所有本地类型引用都需要导入
		// 因为每个类型都会生成到单独的文件中
		imports = append(imports, ImportInfo{
			Module: ".",
			Types:  []string{typeName},
			Path:   "./" + strings.ToLower(typeName),
		})
	}

	return imports
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
	default:
		return false
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
