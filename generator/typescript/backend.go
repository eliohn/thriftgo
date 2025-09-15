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
	return t.renderByTemplateWithTemplate(scope, executeTpl, filename, "index")
}

// renderEnumFile 生成枚举文件
func (t *TypeScriptBackend) renderEnumFile(scope *Scope, executeTpl *template.Template, basePath string, enum *parser.Enum) error {
	filename := filepath.Join(basePath, strings.ToLower(enum.Name)+".ts")

	// 创建只包含该枚举的 scope
	enumScope := &Scope{
		Filename: scope.Filename,
		Package:  scope.Package,
		Imports:  scope.Imports,
		Enums:    []*parser.Enum{enum},
		utils:    scope.utils,
	}

	return t.renderByTemplateWithTemplate(enumScope, executeTpl, filename, "singleEnum")
}

// renderStructFile 生成结构体文件
func (t *TypeScriptBackend) renderStructFile(scope *Scope, executeTpl *template.Template, basePath string, structLike *parser.StructLike) error {
	filename := filepath.Join(basePath, strings.ToLower(structLike.Name)+".ts")

	// 创建只包含该结构体的 scope
	structScope := &Scope{
		Filename:        scope.Filename,
		Package:         scope.Package,
		Imports:         scope.Imports,
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
