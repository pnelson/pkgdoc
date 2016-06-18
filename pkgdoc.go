// Package pkgdoc prepares package documentation from source.
package pkgdoc

import (
	"bytes"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"html/template"

	"golang.org/x/tools/go/loader"
)

// Package represents package documentation.
type Package struct {
	Name       string
	ImportPath string
	Doc        Doc
	Synopsis   string
	Constants  []Value
	Variables  []Value
	Functions  []Function
	Types      []Type
}

// New returns a new Package.
//
// The package is loaded from source and must exist on the file system.
func New(repo string) (Package, error) {
	conf := &loader.Config{
		TypeChecker: types.Config{Error: func(err error) {}},
		ParserMode:  parser.ParseComments,
	}
	conf.Import(repo)
	prog, err := conf.Load()
	if err != nil {
		return Package{}, err
	}
	pkg := prog.Package(repo)
	files := make(map[string]*ast.File)
	for _, file := range pkg.Files {
		name := prog.Fset.Position(file.Pos()).Filename
		files[name] = file
	}
	apkg, _ := ast.NewPackage(prog.Fset, files, nil, nil) // non-applicable error
	dpkg := doc.New(apkg, pkg.Pkg.Path(), 0)
	pdoc := Package{
		Name:       dpkg.Name,
		ImportPath: dpkg.ImportPath,
		Doc:        Doc(dpkg.Doc),
		Synopsis:   doc.Synopsis(dpkg.Doc),
		Constants:  pkgValues(dpkg.Consts, prog.Fset),
		Variables:  pkgValues(dpkg.Vars, prog.Fset),
		Functions:  pkgFunctions(dpkg.Funcs, prog.Fset),
		Types:      pkgTypes(dpkg.Types, prog.Fset),
	}
	return pdoc, nil
}

// Doc represents source code documentation.
type Doc string

// HTML returns the source code documentation as formatted HTML.
func (d Doc) HTML() template.HTML {
	var w bytes.Buffer
	doc.ToHTML(&w, string(d), nil)
	return template.HTML(w.String())
}

// Value represents the documentation for values.
type Value struct {
	Doc  Doc
	Decl string
}

// Function represents the documentation for functions.
type Function struct {
	Doc  Doc
	Name string
	Decl string
}

// Type represents the documentation for types.
type Type struct {
	Doc       Doc
	Name      string
	Decl      string
	Constants []Value
	Variables []Value
	Functions []Function
	Methods   []Function
}

func decl(v interface{}, fset *token.FileSet) string {
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, v)
	return buf.String()
}

func pkgValues(vs []*doc.Value, fset *token.FileSet) []Value {
	rv := make([]Value, len(vs))
	for i := range vs {
		rv[i] = newValue(vs[i], fset)
	}
	return rv
}

func newValue(v *doc.Value, fset *token.FileSet) Value {
	return Value{
		Doc:  Doc(v.Doc),
		Decl: decl(v.Decl, fset),
	}
}

func pkgFunctions(fs []*doc.Func, fset *token.FileSet) []Function {
	rv := make([]Function, len(fs))
	for i := range fs {
		rv[i] = newFunction(fs[i], fset)
	}
	return rv
}

func newFunction(f *doc.Func, fset *token.FileSet) Function {
	return Function{
		Doc:  Doc(f.Doc),
		Name: f.Name,
		Decl: decl(f.Decl, fset),
	}
}

func pkgTypes(ts []*doc.Type, fset *token.FileSet) []Type {
	rv := make([]Type, len(ts))
	for i := range ts {
		rv[i] = newType(ts[i], fset)
	}
	return rv
}

func newType(t *doc.Type, fset *token.FileSet) Type {
	return Type{
		Doc:       Doc(t.Doc),
		Name:      t.Name,
		Decl:      decl(t.Decl, fset),
		Constants: pkgValues(t.Consts, fset),
		Variables: pkgValues(t.Vars, fset),
		Functions: pkgFunctions(t.Funcs, fset),
		Methods:   pkgFunctions(t.Methods, fset),
	}
}
