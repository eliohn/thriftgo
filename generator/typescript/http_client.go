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
	"text/template"

	"github.com/cloudwego/thriftgo/plugin"
)

// renderHttpClientFiles 生成 HTTP 客户端文件
func (t *TypeScriptBackend) renderHttpClientFiles(scope *Scope, executeTpl *template.Template, basePath string) error {
	// 生成 Fetch HTTP 客户端
	if err := t.renderHttpClientFile(scope, executeTpl, basePath, "httpClient", "HttpClient.ts"); err != nil {
		return err
	}

	// 生成 Axios HTTP 客户端
	if err := t.renderHttpClientFile(scope, executeTpl, basePath, "httpClientAxios", "HttpClientAxios.ts"); err != nil {
		return err
	}

	return nil
}

// renderHttpClientFile 生成单个 HTTP 客户端文件
func (t *TypeScriptBackend) renderHttpClientFile(scope *Scope, executeTpl *template.Template, basePath, templateName, filename string) error {
	w := poolBuffer.Get().(*bytes.Buffer)
	defer poolBuffer.Put(w)

	w.Reset()

	t.utils.SetRootScope(scope)
	err := executeTpl.ExecuteTemplate(w, templateName, scope)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}

	fullPath := filepath.Join(basePath, filename)
	t.res.Contents = append(t.res.Contents, &plugin.Generated{
		Content: w.String(),
		Name:    &fullPath,
	})
	return nil
}
