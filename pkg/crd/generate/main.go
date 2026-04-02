//go:build codegen
// +build codegen

package main

import (
	"fmt"
	"github.com/gertd/go-pluralize"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"slices"
	"strings"
)

var forceIncludeTypes = []string{"Deployment", "ContainerStatus", "DeploymentVersion", "JobExecutionStatus", "VolumeSetStatusLocation", "PersistentVolumeStatus"}

var kindSpecialCases = map[string]string{
	"AuditContext": "auditctx",
}

func main() {
	schemaDir := "../schema"
	prefix := "github.com/controlplane-com/libs-go/pkg/schema/"
	code, err := GenerateCRDCode(schemaDir, prefix)
	if err != nil {
		log.Fatalf("Error generating code for CRDs: %v\n", err)
	}

	// We'll write this to ./generated/crds.go
	outputFile := "./output/crds.go"
	if err := os.MkdirAll("./output/yaml", 0755); err != nil {
		log.Fatalf("Error creating generated directory: %v\n", err)
	}
	if err = os.WriteFile(outputFile, []byte(code), 0644); err != nil {
		log.Fatalf("Failed to write generated runner: %v\n", err)
	}

	fmt.Printf("Successfully wrote %s\n", outputFile)
}

func GenerateCRDCode(dir, prefix string) (string, error) {
	if prefix == "" {
		prefix = "github.com/controlplane-com/libs-go/pkg/schema/"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read dir %q: %w", dir, err)
	}

	var importLines []string
	var generationLines []string

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		folderName := e.Name()
		if strings.HasPrefix(folderName, ".") {
			continue
		}

		// We'll build an import alias and path
		importAlias := folderName
		importPath := prefix + folderName

		lines, err := buildGenerationLines(importAlias, fmt.Sprintf("%s/%s", dir, folderName))
		if err != nil {
			return "", fmt.Errorf("failed to build generation lines: %w", err)
		}
		if len(lines) > 0 {
			importLines = append(importLines, fmt.Sprintf("\t\"%s\"", importPath))
			generationLines = append(generationLines, lines...)
		}
	}

	// Build the final code
	sb := new(strings.Builder)
	sb.WriteString("// +build crdgen\n\npackage main\n\n")
	sb.WriteString("import (\n\t\"github.com/controlplane-com/libs-go/pkg/crd\"\n\tapiextensionsv1 \"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1\"\n)\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"fmt\"\n")
	sb.WriteString("\t\"os\"\n\n")
	for _, line := range importLines {
		sb.WriteString(line + "\n")
	}
	sb.WriteString(")\n\n")
	sb.WriteString("func main() {\n")
	sb.WriteString("\tvar (\n\t\tc *apiextensionsv1.CustomResourceDefinition\n\t\tyaml string\n\t\terr error\n\t)\n")
	for _, gl := range generationLines {
		sb.WriteString("\t" + strings.ReplaceAll(gl, "\n", "\n\t"))
		sb.WriteString("\n")
	}
	sb.WriteString("\tfmt.Println(\"All CRD YAML generated.\")\n")
	sb.WriteString("}\n")

	return sb.String(), nil
}

func buildGenerationLines(importAlias string, folderName string) ([]string, error) {
	crdStructs, err := findCRDStructs(folderName)
	if err != nil {
		return nil, err
	}

	p := pluralize.NewClient()
	var lines []string
	for _, c := range crdStructs {
		kind := c
		apiKind := kind
		if sc, ok := kindSpecialCases[kind]; ok {
			apiKind = sc
		}
		plural := strings.ToLower(p.Plural(c))
		yamlFile := fmt.Sprintf("./output/yaml/%s.yaml", kind)

		lines = append(lines, fmt.Sprintf(
			`c, err = crd.ConvertStructToCRD(
	&%s.%s{}, 
	"cpln.io", 
	"v1", 
	"%s", 
	"%s",
)
if err != nil {
	panic(err)
}
yaml, err = crd.CRDToYAML(c)
if err != nil{
	panic(err)
}
err = os.WriteFile("%s", []byte(yaml + "\n"), 0644)
if err != nil {
	panic(err)
}
fmt.Printf("Wrote CRD YAML for %s to %s\n")
`, importAlias, kind, apiKind, plural, yamlFile, kind, yamlFile))
	}
	return lines, nil
}

// findCRDStructs scans the given directory for .go files and returns the names
// of all struct types that have *all three* fields: Name, Kind, and Version.
func findCRDStructs(dir string) ([]string, error) {
	fset := token.NewFileSet()

	// ParseDir parses the entire directory of Go code into an AST.
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		return nil, err
	}

	var result []string

	// Walk through each package in that directory
	for _, pkg := range pkgs {
		// Inspect each file in the package
		for fname, file := range pkg.Files {
			fmt.Printf("Inspecting file: %s\n", fname)

			// Use ast.Inspect to walk the AST
			ast.Inspect(file, func(n ast.Node) bool {
				// We only care about type declarations: `type MyType struct { ... }`
				typeSpec, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}

				// We only care if the underlying type is a struct
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					return true
				}

				if slices.Contains(forceIncludeTypes, typeSpec.Name.Name) {
					result = append(result, typeSpec.Name.Name)
					return true
				}

				// Check if this struct has fields Name, Kind, Version
				if structHasAllFields(structType, "Name", "Kind", "Version") {
					// If yes, store the struct's identifier (typeSpec.Name.Name)
					result = append(result, typeSpec.Name.Name)
				}

				return true
			})
		}
	}

	return result, nil
}

// structHasAllFields checks if the given AST struct has all of the specified field names.
func structHasAllFields(st *ast.StructType, fieldNames ...string) bool {
	need := make(map[string]bool)
	for _, f := range fieldNames {
		need[f] = true
	}

	// Look through all fields declared in the struct
	for _, field := range st.Fields.List {
		// Each field can have multiple names (e.g., "A, B int")
		for _, name := range field.Names {
			// If this name is one we need, remove it from 'need'
			if need[name.Name] {
				delete(need, name.Name)
				// If we've found all required fields, we're done
				if len(need) == 0 {
					return true
				}
			}
		}
	}

	return len(need) == 0 // or false, depending on whether we consider partial matches
}
