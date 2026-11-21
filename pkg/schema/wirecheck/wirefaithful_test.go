package wirecheck

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Two rules, both about a field's ability to say "absent" without lying.
//
// 1. An optional field is dropped when it holds its zero value. That is only correct when
//    the zero value cannot itself be a legal value: bool `false`, number `0` and struct `{}`
//    all are, so those fields must be pointers.
// 2. The option must be omitzero, not omitempty. omitempty additionally drops an empty
//    slice or map, so a stored `[]` or `{}` comes back as an absent key.

type typeInfo struct {
	isStruct   bool
	underlying string // for named non-struct types: the underlying expression, e.g. "float32"
}

func schemaRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// index maps pkg -> type name -> typeInfo, over every non-test .go file in each package.
func index(t *testing.T, root string) map[string]map[string]typeInfo {
	t.Helper()
	out := map[string]map[string]typeInfo{}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		files, _ := filepath.Glob(filepath.Join(root, e.Name(), "*.go"))
		types := map[string]typeInfo{}
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			file, err := parser.ParseFile(token.NewFileSet(), f, nil, 0)
			if err != nil {
				t.Fatalf("parse %s: %v", f, err)
			}
			for _, d := range file.Decls {
				gd, ok := d.(*ast.GenDecl)
				if !ok || gd.Tok != token.TYPE {
					continue
				}
				for _, s := range gd.Specs {
					ts := s.(*ast.TypeSpec)
					if _, isStruct := ts.Type.(*ast.StructType); isStruct {
						types[ts.Name.Name] = typeInfo{isStruct: true}
					} else if id, ok := ts.Type.(*ast.Ident); ok {
						types[ts.Name.Name] = typeInfo{underlying: id.Name}
					} else {
						types[ts.Name.Name] = typeInfo{underlying: "?"}
					}
				}
			}
		}
		if len(types) > 0 {
			out[e.Name()] = types
		}
	}
	return out
}

var zeroIsLegal = map[string]bool{
	"bool": true, "float32": true, "float64": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
}

// needsPointer reports whether an omitempty field of this type can lie about presence.
func needsPointer(idx map[string]map[string]typeInfo, pkg string, expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		if zeroIsLegal[e.Name] {
			return true
		}
		return resolve(idx, pkg, e.Name, 0)
	case *ast.SelectorExpr:
		id, ok := e.X.(*ast.Ident)
		if !ok {
			return false
		}
		return resolve(idx, id.Name, e.Sel.Name, 0)
	}
	// pointers, slices, maps, interfaces and generics already have a nil form.
	return false
}

func resolve(idx map[string]map[string]typeInfo, pkg, name string, depth int) bool {
	if depth > 8 {
		return false
	}
	ti, ok := idx[pkg][name]
	if !ok {
		return false
	}
	if ti.isStruct {
		return true
	}
	if zeroIsLegal[ti.underlying] {
		return true
	}
	return resolve(idx, pkg, ti.underlying, depth+1)
}

func TestOptionalFieldsCanSayAbsent(t *testing.T) {
	root := schemaRoot(t)
	idx := index(t, root)
	if len(idx) < 20 {
		t.Fatalf("indexed only %d schema packages, expected the whole tree", len(idx))
	}

	var bad []string
	for pkg := range idx {
		files, _ := filepath.Glob(filepath.Join(root, pkg, "*.go"))
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, f, nil, 0)
			if err != nil {
				t.Fatalf("parse %s: %v", f, err)
			}
			ast.Inspect(file, func(n ast.Node) bool {
				st, ok := n.(*ast.StructType)
				if !ok {
					return true
				}
				for _, fld := range st.Fields.List {
					if fld.Tag == nil {
						continue
					}
					tag, err := strconv.Unquote(fld.Tag.Value)
					if err != nil {
						continue
					}
					js := reflect.StructTag(tag).Get("json")
					name := "<embedded>"
					if len(fld.Names) > 0 {
						name = fld.Names[0].Name
					}
					where := filepath.Base(filepath.Dir(f)) + "/" + filepath.Base(f) + ":" +
						strconv.Itoa(fset.Position(fld.Pos()).Line) + " " + name
					if strings.Contains(js, ",omitempty") {
						bad = append(bad, where+" (omitempty drops an empty slice/map; use omitzero)")
						return true
					}
					if !strings.Contains(js, ",omitzero") {
						continue
					}
					if !needsPointer(idx, pkg, fld.Type) {
						continue
					}
					bad = append(bad, where+" (zero value is a legal wire value; needs a pointer)")
				}
				return true
			})
		}
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		show := bad
		if len(show) > 25 {
			show = show[:25]
		}
		t.Fatalf("%d schema fields cannot say \"absent\" without lying; regenerate with joi-to-go:\n  %s",
			len(bad), strings.Join(show, "\n  "))
	}
}
