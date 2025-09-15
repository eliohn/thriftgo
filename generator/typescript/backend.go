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
	path := t.utils.CombineOutputPath(t.req.OutputPath, ast)
	filename := filepath.Join(path, t.utils.GetFilename(ast))
	scope, err := BuildScope(t.utils, ast)
	if err != nil {
		return err
	}
	return t.renderByTemplate(scope, t.tpl, filename)
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
