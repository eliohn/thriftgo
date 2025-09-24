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
	"bytes"
	"fmt"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
)

// OpenAPIBackend generates OpenAPI documentation from Thrift IDL.
// The zero value of OpenAPIBackend is ready for use.
type OpenAPIBackend struct {
	err error
	tpl *template.Template
	req *plugin.Request
	res *plugin.Response
	log backend.LogFunc

	utils *CodeUtils
	funcs template.FuncMap
}

// Name implements the Backend interface.
func (o *OpenAPIBackend) Name() string {
	return "openapi"
}

// Lang implements the Backend interface.
func (o *OpenAPIBackend) Lang() string {
	return "OpenAPI"
}

// Options implements the Backend interface.
func (o *OpenAPIBackend) Options() (opts []plugin.Option) {
	for _, p := range allParams {
		opts = append(opts, plugin.Option{
			Name: p.name,
			Desc: p.desc,
		})
	}
	return opts
}

// BuiltinPlugins implements the Backend interface.
func (o *OpenAPIBackend) BuiltinPlugins() []*plugin.Desc {
	return nil
}

// GetPlugin implements the Backend interface.
func (o *OpenAPIBackend) GetPlugin(desc *plugin.Desc) plugin.Plugin {
	return nil
}

// Generate implements the Backend interface.
func (o *OpenAPIBackend) Generate(req *plugin.Request, log backend.LogFunc) *plugin.Response {
	o.req = req
	o.res = plugin.NewResponse()
	o.log = log
	o.prepareUtilities()
	if o.err != nil {
		return o.buildResponse()
	}

	o.prepareTemplates()
	o.fillRequisitions()
	o.executeTemplates()
	return o.buildResponse()
}

func (o *OpenAPIBackend) GetCoreUtils() *CodeUtils {
	return o.utils
}

func (o *OpenAPIBackend) prepareUtilities() {
	if o.err != nil {
		return
	}

	o.utils = NewCodeUtils(o.log)
	o.err = o.utils.HandleOptions(o.req.GeneratorParameters)
	if o.err != nil {
		return
	}

	o.funcs = o.utils.BuildFuncMap()
	o.funcs["Version"] = func() string { return o.req.Version }
}

func (o *OpenAPIBackend) prepareTemplates() {
	if o.err != nil {
		return
	}

	all := template.New("openapi").Funcs(o.funcs)
	tpls := Templates()

	for _, tpl := range tpls {
		all = template.Must(all.Parse(tpl))
	}
	o.tpl = all
}

func (o *OpenAPIBackend) fillRequisitions() {
	if o.err != nil {
		return
	}
}

func (o *OpenAPIBackend) executeTemplates() {
	if o.err != nil {
		return
	}

	processed := make(map[*parser.Thrift]bool)

	var trees chan *parser.Thrift
	if o.req.Recursive {
		trees = o.req.AST.DepthFirstSearch()
	} else {
		trees = make(chan *parser.Thrift, 1)
		trees <- o.req.AST
		close(trees)
	}

	for ast := range trees {
		if processed[ast] {
			continue
		}
		processed[ast] = true
		o.log.Info("Processing", ast.Filename)

		if o.err = o.renderOneFile(ast); o.err != nil {
			break
		}
	}
}

func (o *OpenAPIBackend) renderOneFile(ast *parser.Thrift) error {
	scope, err := BuildScope(o.utils, ast)
	if err != nil {
		return err
	}

	path := o.utils.CombineOutputPath(o.req.OutputPath, ast)
	filename := filepath.Join(path, o.utils.GetFilename(ast))
	return o.renderByTemplateWithAST(scope, o.tpl, filename, ast)
}

var poolBuffer = sync.Pool{
	New: func() any {
		p := &bytes.Buffer{}
		p.Grow(100 << 10)
		return p
	},
}

func (o *OpenAPIBackend) renderByTemplate(scope *Scope, executeTpl *template.Template, filename string) error {
	if scope == nil {
		return nil
	}

	// if scope has no content, just skip and don't generate this file
	if o.utils.Features().SkipEmpty {
		if scope.IsEmpty() {
			return nil
		}
	}

	w := poolBuffer.Get().(*bytes.Buffer)
	defer poolBuffer.Put(w)

	w.Reset()

	o.utils.SetRootScope(scope)
	err := executeTpl.ExecuteTemplate(w, executeTpl.Name(), scope)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	o.res.Contents = append(o.res.Contents, &plugin.Generated{
		Content: w.String(),
		Name:    &filename,
	})
	return nil
}

func (o *OpenAPIBackend) renderByTemplateWithAST(scope *Scope, executeTpl *template.Template, filename string, ast *parser.Thrift) error {
	if scope == nil {
		return nil
	}

	// if scope has no content, just skip and don't generate this file
	if o.utils.Features().SkipEmpty {
		if scope.IsEmpty() {
			return nil
		}
	}

	w := poolBuffer.Get().(*bytes.Buffer)
	defer poolBuffer.Put(w)

	w.Reset()

	o.utils.SetRootScope(scope)
	o.utils.SetAST(ast)
	err := executeTpl.ExecuteTemplate(w, executeTpl.Name(), scope)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	o.res.Contents = append(o.res.Contents, &plugin.Generated{
		Content: w.String(),
		Name:    &filename,
	})
	return nil
}

func (o *OpenAPIBackend) buildResponse() *plugin.Response {
	if o.err != nil {
		return plugin.BuildErrorResponse(o.err.Error())
	}
	return o.res
}

// PostProcess implements the backend.PostProcessor interface to do
// source formatting before writing files out.
func (o *OpenAPIBackend) PostProcess(path string, content []byte) ([]byte, error) {
	// OpenAPI 文档通常不需要特殊的格式化，直接返回
	return content, nil
}
