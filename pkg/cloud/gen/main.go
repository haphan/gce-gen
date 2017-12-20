/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Generator for GCE compute wrapper code. You must regenerate the code after
// modifying this file:
//
//   $ go run gen/main.go > gen.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"text/template"

	"github.com/bowei/gce-gen/pkg/cloud/meta"
	"github.com/golang/glog"
)

const (
	gofmt = "gofmt"

	// readOnly specifies that the given resource is read-only and should not
	// have insert() or delete() methods generated for the wrapper.
	readOnly = iota
)

var flags = struct {
	gofmt bool
	mode  string
}{}

func init() {
	flag.BoolVar(&flags.gofmt, "gofmt", true, "run output through gofmt")
	flag.StringVar(&flags.mode, "mode", "src", "content to generate: src, test, dummy")
}

func gofmtContent(r io.Reader) string {
	cmd := exec.Command(gofmt, "-s")
	out := &bytes.Buffer{}
	cmd.Stdin = r
	cmd.Stdout = out
	cmdErr := &bytes.Buffer{}
	cmd.Stderr = cmdErr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, cmdErr.String())
		panic(err)
	}
	return out.String()
}

// genHeader generate the header for the file.
func genHeader(wr io.Writer) {
	var hasGA, hasAlpha, hasBeta bool
	for _, s := range meta.AllServices {
		switch s.Version() {
		case meta.VersionGA:
			hasGA = true
		case meta.VersionAlpha:
			hasAlpha = true
		case meta.VersionBeta:
			hasBeta = true
		}
	}

	fmt.Fprintln(wr, `/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file was generated by "go run gen/main.go > gen.go". Do not edit
// directly.

package cloud

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"google.golang.org/api/googleapi"
	"github.com/bowei/gce-gen/pkg/cloud/meta"
`)
	if hasAlpha {
		fmt.Fprintln(wr, `alpha "google.golang.org/api/compute/v0.alpha"`)
	}
	if hasBeta {
		fmt.Fprintln(wr, `beta "google.golang.org/api/compute/v0.beta"`)
	}
	if hasGA {
		fmt.Fprintln(wr, `ga "google.golang.org/api/compute/v1"`)
	}
	fmt.Fprintln(wr, ")")
}

// genStubs generates the interface and wrapper stubs.
func genStubs(wr io.Writer) {
	const text = `// Cloud is an interface for the GCE compute API.
type Cloud interface { {{range .}}
	{{.WrapType}}() {{.WrapType}}{{end}}
}

// NewMockGCE returns a new mock for GCE.
func NewMockGCE() *MockGCE {
	mock := &MockGCE{ {{range .}}
		{{.MockField}}: New{{.MockWrapType}}(),{{end}}
	}
	return mock
}

// MockGCE implements Cloud.
var _ Cloud = (*MockGCE)(nil)

// MockGCE is the mock for the compute API.
type MockGCE struct { {{range .}}
	{{.MockField}} *{{.MockWrapType}}{{end}}
}
{{range .}}
func (mock *MockGCE) {{.WrapType}}() {{.WrapType}} {
	return mock.{{.MockField}}
}
{{end}}

// NewGCE returns a GCE.
func NewGCE(s *Service) *GCE {
	g := &GCE{ {{range .}}
		{{.Field}}: &{{.GCEWrapType}}{s},{{end}}
	}
	return g
}

// GCE implements Cloud.
var _ Cloud = (*GCE)(nil)

// GCE is the golang adapter for the compute APIs.
type GCE struct { {{range .}}
	{{.Field}} *{{.GCEWrapType}}{{end}}
}
{{range .}}
func (gce *GCE) {{.WrapType}}() {{.WrapType}} {
	return gce.{{.Field}}
}
{{end}}
`
	tmpl := template.Must(template.New("interface").Parse(text))
	if err := tmpl.Execute(wr, meta.AllServices); err != nil {
		panic(err)
	}
}

// genTypes generates the type wrappers.
func genTypes(wr io.Writer) {
	const text = `// {{.WrapType}} is an interface that allows for mocking of {{.Service}}.
type {{.WrapType}} interface { {{if .GenerateCustomOps}}
	// {{.WrapTypeOps}} is an interface with additional non-CRUD type methods.
	// This interface is expected to be implemented by hand (non-autogenerated).
	{{.WrapTypeOps}}
{{end}}
	Get(ctx context.Context, key meta.Key) (*{{.FQObjectType}}, error) {{if .KeyIsGlobal}}
	List(ctx context.Context) ([]*{{.FQObjectType}}, error){{end}}{{if .KeyIsRegional}}
	List(ctx context.Context, region string) ([]*{{.FQObjectType}}, error){{end}}{{if .KeyIsZonal}}
	List(ctx context.Context, zone string) ([]*{{.FQObjectType}}, error){{end}} {{if .GenerateMutations}}
	Insert(ctx context.Context, key meta.Key, obj *{{.FQObjectType}}) error
	Delete(ctx context.Context, key meta.Key) error {{end}}
{{with .Methods}}{{range .}}
	{{.InterfaceFunc}}{{end}}
{{end}}
}

// New{{.MockWrapType}} returns a new mock for {{.Service}}.
func New{{.MockWrapType}}() *{{.MockWrapType}} {
	mock := &{{.MockWrapType}}{
		Objects: map[meta.Key]*{{.FQObjectType}}{},
		GetError: map[meta.Key]error{},
		InsertError: map[meta.Key]error{},
		DeleteError: map[meta.Key]error{},
	}
	return mock
}

// {{.MockWrapType}} is the mock for {{.Service}}.
type {{.MockWrapType}} struct {
	Lock sync.Mutex

	// Objects maintained by the mock.
	Objects map[meta.Key]*{{.FQObjectType}}

	// If an entry exists for the given key and operation, then the error
	// will be returned instead of the operation.
	GetError map[meta.Key]error
	ListError *error
	InsertError map[meta.Key]error
	DeleteError map[meta.Key]error

	// GetHook, ListHook, InsertHook, DeleteHook allow you to intercept the
	// standard processing of the mock in order to add your own logic.
	// Return (true, _, _) to prevent the normal execution flow of the
	// mock. Return (false, nil, nil) to continue with normal mock behavior
	// after the hook function executes.
	GetHook func(m *{{.MockWrapType}}, ctx context.Context, key meta.Key) (bool, *{{.FQObjectType}}, error)
	ListHook func(m *{{.MockWrapType}}, ctx context.Context) (bool, []*{{.FQObjectType}}, error)
	InsertHook func(m *{{.MockWrapType}}, ctx context.Context, key meta.Key, obj *{{.FQObjectType}}) (bool, error)
	DeleteHook func(m *{{.MockWrapType}}, ctx context.Context, key meta.Key) (bool, error)
{{with .Methods}}{{range .}}
	{{.MockHook}}{{end}}{{end}}

	// X is extra state that can be used as part of the mock. Generated code
	// will not use this field.
	X interface{}
}

// Get returns the object from the mock.
func (m *{{.MockWrapType}}) Get(ctx context.Context, key meta.Key) (*{{.FQObjectType}}, error) {
	if m.GetHook != nil {
		if intercept, obj, err := m.GetHook(m, ctx, key);  intercept {
			return obj, err
		}
	}

	m.Lock.Lock()
	defer m.Lock.Unlock()

	if err, ok := m.GetError[key]; ok {
		return nil, err
	}
	if obj, ok := m.Objects[key]; ok {
		return obj, nil
	}
	return nil, &googleapi.Error{
		Code: http.StatusNotFound,
		Message: fmt.Sprintf("{{.MockWrapType}} %v not found", key),
	}
}

{{if .KeyIsGlobal}}
// List all of the objects in the mock.
func (m *{{.MockWrapType}}) List(ctx context.Context) ([]*{{.FQObjectType}}, error) { {{end}}{{if .KeyIsRegional}}
// List all of the objects in the mock in the given region.
func (m *{{.MockWrapType}}) List(ctx context.Context, region string) ([]*{{.FQObjectType}}, error) { {{end}}{{if .KeyIsZonal}}
// List all of the objects in the mock in the given zone.
func (m *{{.MockWrapType}}) List(ctx context.Context, zone string) ([]*{{.FQObjectType}}, error) { {{end}}
	if m.ListHook != nil {
		if intercept, objs, err := m.ListHook(m, ctx);  intercept {
			return objs, err
		}
	}

	m.Lock.Lock()
	defer m.Lock.Unlock()

	if m.ListError != nil {
		return nil, *m.ListError
	}

	var objs []*{{.FQObjectType}} {{if .KeyIsGlobal}}
	for _, obj := range m.Objects { {{else}}
	for key, obj := range m.Objects { {{end}}{{if .KeyIsRegional}}
		if key.Region != region {
			continue
		}
{{end}}{{if .KeyIsZonal}}
		if key.Zone != zone {
			continue
		}
{{end}}
		objs = append(objs, obj)
	}

	return objs, nil
}
{{if .GenerateMutations}}
// Insert is a mock for inserting/creating a new object.
func (m *{{.MockWrapType}}) Insert(ctx context.Context, key meta.Key, obj *{{.FQObjectType}}) error {
	if m.InsertHook != nil {
		if intercept, err := m.InsertHook(m, ctx, key, obj);  intercept {
			return err
		}
	}

	m.Lock.Lock()
	defer m.Lock.Unlock()

	if err, ok := m.InsertError[key]; ok {
		return err
	}
	if _, ok := m.Objects[key]; ok {
		return &googleapi.Error{
			Code: http.StatusConflict,
			Message: fmt.Sprintf("{{.MockWrapType}} %v exists", key),
		}
	}

	m.Objects[key] = obj
	return nil
}

// Delete is a mock for deleting the object.
func (m *{{.MockWrapType}}) Delete(ctx context.Context, key meta.Key) error {
	if m.DeleteHook != nil {
		if intercept, err := m.DeleteHook(m, ctx, key);  intercept {
			return err
		}
	}

	m.Lock.Lock()
	defer m.Lock.Unlock()

	if err, ok := m.DeleteError[key]; ok {
		return err
	}
	if _, ok := m.Objects[key]; !ok {
		return &googleapi.Error{
			Code: http.StatusNotFound,
			Message: fmt.Sprintf("{{.MockWrapType}} %v not found", key),
		}
	}

	delete(m.Objects, key)
	return nil
}
{{end}}
{{with .Methods}}{{range .}}
func (m *{{.MockWrapType}}) {{.FcnArgs}} {
{{if eq .ReturnType "Operation"}}
	if m.{{.MockHookName}} != nil {
		return m.{{.MockHookName}}(m, ctx, key {{.CallArgs}})
	}
	return nil
{{else}}
	if m.{{.MockHookName}} != nil {
		return m.{{.MockHookName}}(m, ctx, key {{.CallArgs}})
	}
	return nil, fmt.Errorf("{{.MockHookName}} must be set")
{{end}} }
{{end}}{{end}}
// {{.GCEWrapType}} is a simplifying adapter for the GCE {{.Service}}.
type {{.GCEWrapType}} struct {
	s *Service
}

// Get the {{.Object}} named by key.
func (g *{{.GCEWrapType}}) Get(ctx context.Context, key meta.Key) (*{{.FQObjectType}}, error) {
	rk := &RateLimitKey{
		Operation: "Get",
		Version: meta.Version("{{.Version}}"),
		Target: "{{.Object}}",
	}
	g.s.RateLimiter.Accept(ctx, rk)
	projectID := g.s.ProjectRouter.ProjectID(ctx, "{{.Version}}", "{{.Service}}")
{{if .KeyIsGlobal}}
	call := g.s.{{.VersionField}}.{{.Service}}.Get(projectID, key.Name)
{{end}}{{if .KeyIsRegional}}
	call := g.s.{{.VersionField}}.{{.Service}}.Get(projectID, key.Region, key.Name)
{{end}}{{if .KeyIsZonal}}
	call := g.s.{{.VersionField}}.{{.Service}}.Get(projectID, key.Zone, key.Name)
{{end}}
	call.Context(ctx)

	return call.Do()
}

// List all {{.Object}} objects.{{if .KeyIsGlobal}}
func (g *{{.GCEWrapType}}) List(ctx context.Context) ([]*{{.FQObjectType}}, error) { {{end}}{{if .KeyIsRegional}}
func (g *{{.GCEWrapType}}) List(ctx context.Context, region string) ([]*{{.FQObjectType}}, error) { {{end}}{{if .KeyIsZonal}}
func (g *{{.GCEWrapType}}) List(ctx context.Context, zone string) ([]*{{.FQObjectType}}, error) { {{end}}
	rk := &RateLimitKey{
		Operation: "List",
		Version: meta.Version("{{.Version}}"),
		Target: "{{.Object}}",
	}
	g.s.RateLimiter.Accept(ctx, rk)
	projectID := g.s.ProjectRouter.ProjectID(ctx, "{{.Version}}", "{{.Service}}")
{{if .KeyIsGlobal}}
	call := g.s.{{.VersionField}}.{{.Service}}.List(projectID)
{{end}}{{if .KeyIsRegional}}
	call := g.s.{{.VersionField}}.{{.Service}}.List(projectID, region)
{{end}}{{if .KeyIsZonal}}
	call := g.s.{{.VersionField}}.{{.Service}}.List(projectID, zone)
{{end}}
	var all []*{{.FQObjectType}}
	f := func(l *{{.ObjectListType}}) error {
		all = append(all, l.Items...)
		return nil
	}
	if err := call.Pages(ctx, f); err != nil {
		return nil, err
	}
	return all, nil
}
{{if .GenerateMutations}}
// Insert {{.Object}} with key of value obj.
func (g *{{.GCEWrapType}}) Insert(ctx context.Context, key meta.Key, obj *{{.FQObjectType}}) error {
	rk := &RateLimitKey{
		Operation: "Insert",
		Version: meta.Version("{{.Version}}"),
		Target: "{{.Object}}",
	}
	g.s.RateLimiter.Accept(ctx, rk)
	projectID := g.s.ProjectRouter.ProjectID(ctx, "{{.Version}}", "{{.Service}}")
	obj.Name = key.Name
{{if .KeyIsGlobal}}
	call := g.s.{{.VersionField}}.{{.Service}}.Insert(projectID, obj)
{{end}}{{if .KeyIsRegional}}
	call := g.s.{{.VersionField}}.{{.Service}}.Insert(projectID, key.Region, obj)
{{end}}{{if .KeyIsZonal}}
	call := g.s.{{.VersionField}}.{{.Service}}.Insert(projectID, key.Zone, obj)
{{end}}
	call.Context(ctx)

	op, err := call.Do()
	if err != nil {
		return err
	}
	return g.s.WaitForCompletion(ctx, op)
}

// Delete the {{.Object}} referenced by key.
func (g *{{.GCEWrapType}}) Delete(ctx context.Context, key meta.Key) error {
	rk := &RateLimitKey{
		Operation: "Delete",
		Version: meta.Version("{{.Version}}"),
		Target: "{{.Object}}",
	}
	g.s.RateLimiter.Accept(ctx, rk)
	projectID := g.s.ProjectRouter.ProjectID(ctx, "{{.Version}}", "{{.Service}}")
{{if .KeyIsGlobal}}
	call := g.s.{{.VersionField}}.{{.Service}}.Delete(projectID, key.Name)
{{end}}{{if .KeyIsRegional}}
	call := g.s.{{.VersionField}}.{{.Service}}.Delete(projectID, key.Region, key.Name)
{{end}}{{if .KeyIsZonal}}
	call := g.s.{{.VersionField}}.{{.Service}}.Delete(projectID, key.Zone, key.Name)
{{end}}
	call.Context(ctx)

	op, err := call.Do()
	if err != nil {
		return err
	}
	return g.s.WaitForCompletion(ctx, op)
}
{{end}}
{{with .Methods}}{{range .}}
func (g *{{.GCEWrapType}}) {{.FcnArgs}} {
	rk := &RateLimitKey{
		Operation: "{{.Name}}",
		Version: meta.Version("{{.Version}}"),
		Target: "{{.Object}}",
	}
	g.s.RateLimiter.Accept(ctx, rk)
	projectID := g.s.ProjectRouter.ProjectID(ctx, "{{.Version}}", "{{.Service}}")
{{if .KeyIsGlobal}}
	call := g.s.{{.VersionField}}.{{.Service}}.{{.Name}}(projectID, key.Name {{.CallArgs}})
{{end}}{{if .KeyIsRegional}}
	call := g.s.{{.VersionField}}.{{.Service}}.{{.Name}}(projectID, key.Region, key.Name {{.CallArgs}})
{{end}}{{if .KeyIsZonal}}
	call := g.s.{{.VersionField}}.{{.Service}}.{{.Name}}(projectID, key.Zone, key.Name {{.CallArgs}})
{{end}}
	call.Context(ctx)
{{if eq .ReturnType "Operation"}}
	op, err := call.Do()
	if err != nil {
		return err
	}
	return g.s.WaitForCompletion(ctx, op)
{{else}} return call.Do() {{end}}
}
{{end}}{{end}}
`
	tmpl := template.Must(template.New("interface").Parse(text))
	for _, s := range meta.AllServices {
		if err := tmpl.Execute(wr, s); err != nil {
			panic(err)
		}
	}
}

func genDummy(wr io.Writer) {
	fmt.Fprintln(wr, `
package cloud
`)
	for _, s := range meta.AllServices {
		if s.GenerateCustomOps() {
			fmt.Fprintf(wr, "type %v interface {}\n", s.WrapTypeOps())
		}
	}
}

func main() {
	flag.Parse()

	out := &bytes.Buffer{}

	switch flags.mode {
	case "src":
		genHeader(out)
		genStubs(out)
		genTypes(out)
	case "test":
	case "dummy":
		genDummy(out)
	default:
		glog.Fatalf("Invalid -mode: %q", flags.mode)
	}

	if flags.gofmt {
		fmt.Print(gofmtContent(out))
	} else {
		fmt.Print(out.String())
	}
}
