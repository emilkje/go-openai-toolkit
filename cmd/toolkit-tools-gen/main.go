package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

var (
	path   = flag.String("path", "", "path to scan for tools")
	suffix = flag.String("suffix", "_gen.go", "the suffix to use for generated files")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of toolkit-tools-gen:\n")
	fmt.Fprintf(os.Stderr, "\ttoolkit-tools-gen -path <directory>[,...] # directory to scan for tools \n")
	fmt.Fprintf(os.Stderr, "\ttoolkit-tools-gen -path './tools' -suffix '_gen.go' # specifies the generated file suffix\n")
	fmt.Fprintf(os.Stderr, "For more information, see:\n")
	fmt.Fprintf(os.Stderr, "\thttps://pkg.go.dev/emilkje/go-openai-toolkit/cmd/toolkit-tools-gen\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {

	flag.Usage = Usage
	flag.Parse()

	if len(*path) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	scanner := NewScanner(*path)
	err := scanner.ScanTools()

	if err != nil {
		fmt.Println("error scanning tools", err)
		os.Exit(1)
	}

	if len(scanner.GetTools()) == 0 {
		fmt.Println("👀 no tools found")
		os.Exit(0)
	}

	// scan for arguments
	scanner.ScanArguments()

	generator := NewGenerator(scanner)
	err = generator.Generate(*suffix)

	if err != nil {
		fmt.Println("error generating files", err)
		os.Exit(1)

	}
}

// ################### SCANNER ###################

var jsonSchemaTypes = map[string]string{
	"string":  "jsonschema.String",
	"int":     "jsonschema.Integer",
	"int64":   "jsonschema.Integer",
	"float64": "jsonschema.Number",
	"bool":    "jsonschema.Boolean",
	"Object":  "jsonschema.Object",
	"Array":   "jsonschema.Array",
}

type Tool struct {
	SourcePath   string
	Name         string
	TypeName     string
	Description  string
	ArgumentType string
	Arguments    *ToolArguments
	PackageName  string
}

func (t *Tool) GetArguments() []Arg {
	// return sorted list of Arg
	keys := make([]string, 0, len(t.Arguments.arguments))
	args := make([]Arg, 0, len(t.Arguments.arguments))
	for i := range t.Arguments.arguments {
		keys = append(keys, i)
	}
	sort.Strings(keys)

	for i := range keys {
		args = append(args, t.Arguments.arguments[keys[i]])
	}

	return args
}

func NewTool(sourcePath, typeName string) *Tool {
	return &Tool{
		SourcePath: sourcePath,
		Arguments:  NewToolArguments(),
		TypeName:   typeName,
	}
}

type ToolArguments struct {
	arguments map[string]Arg
}

func NewToolArguments() *ToolArguments {
	return &ToolArguments{
		arguments: make(map[string]Arg),
	}
}
func (t *ToolArguments) Add(key string, arg Arg) {
	if _, ok := t.arguments[key]; !ok {
		t.arguments[key] = arg
	}
}

type Arg struct {
	Name        string
	Description string
	Required    bool
	Type        string
}

func NewArg(fieldName string) Arg {
	return Arg{
		Name:     fieldName,
		Required: true,
	}
}

type ToolScanner struct {
	path  string
	tools map[string]*Tool
}

func NewScanner(path string) *ToolScanner {
	return &ToolScanner{
		path:  path,
		tools: make(map[string]*Tool),
	}
}
func (s *ToolScanner) Add(toolType string, t *Tool) {
	if _, ok := s.tools[toolType]; !ok {
		s.tools[toolType] = t
	}
}

func (s *ToolScanner) ScanTools() error {
	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {

		if !strings.HasSuffix(path, ".go") {
			slog.Debug("skipping invalid file", "file", path)
			return nil
		}

		slog.Debug("walking", "file", path)

		fileset := token.NewFileSet()
		file, err := parser.ParseFile(fileset, path, nil, parser.ParseComments)

		if err != nil {
			return err
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.GenDecl:
				if !hasMarker(x) {
					return true
				}

				slog.Debug("found marker", "file", path)

				for _, spec := range x.Specs {
					if typeSpec, isTypeSpec := spec.(*ast.TypeSpec); isTypeSpec {
						specTypeName := typeSpec.Name.Name
						tool := NewTool(path, specTypeName)
						typeLogger := slog.With("tool", specTypeName)
						typeLogger.Debug("extracting comments")
						for _, comment := range x.Doc.List {
							if strings.HasPrefix(comment.Text, "// +tool:name=") {
								tool.Name = strings.TrimPrefix(comment.Text, "// +tool:name=")
							}
							if strings.HasPrefix(comment.Text, "// +tool:description=") {
								tool.Description = strings.TrimPrefix(comment.Text, "// +tool:description=")
							}
						}

						structType, isStructType := typeSpec.Type.(*ast.StructType)
						if !isStructType {
							typeLogger.Error("tool is not a struct", "type", typeSpec.Type)
							return false
						}

						typeLogger.Debug("found struct", "fields", structType.Fields, "struct", structType.Struct)

						if structType.Fields.NumFields() == 0 {
							typeLogger.Debug("no arguments required")
							s.Add(specTypeName, tool)
							return false
						}

						// set package name
						tool.PackageName = file.Name.Name
						toolArgsField := structType.Fields.List[0]

						selector := toolArgsField.Type.(*ast.IndexExpr).X.(*ast.SelectorExpr)
						if selector.Sel.Name != "Tool" {
							typeLogger.Error("tools need a base type of toolkit.Tool")
							return false
						}

						tool.ArgumentType = toolArgsField.Type.(*ast.IndexExpr).Index.(*ast.Ident).Name
						typeLogger.Debug("found expected argument", "argument_type", tool.ArgumentType)
						s.Add(specTypeName, tool)
					}
				}
			case *ast.Comment:
				// if the function receiver is a tool
				//slog.Info("found comment", "comment", x.Text)
			}
			// continue traversal
			return true
		})

		return nil
	})

	return err
}

func (s *ToolScanner) GetTools() map[string]*Tool {
	return s.tools
}

func (s *ToolScanner) ScanArguments() {
	for _, tool := range s.tools {
		toolArgs, err := s.findToolArguments(tool.ArgumentType)
		if err != nil {
			slog.Error("error finding tool arguments", "tool", tool.Name, "err", err)
			continue
		}
		tool.Arguments = toolArgs
	}
}

func (s *ToolScanner) findToolArguments(attrName string) (*ToolArguments, error) {
	toolArgs := NewToolArguments()
	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {

		if !strings.HasSuffix(path, ".go") {
			slog.Debug("skipping invalid file", "file", path)
			return nil
		}

		fileset := token.NewFileSet()
		file, err := parser.ParseFile(fileset, path, nil, parser.ParseComments)

		if err != nil {
			return err
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.GenDecl:
				// check if type name matches the argument type
				for _, spec := range x.Specs {
					if typeSpec, isTypeSpec := spec.(*ast.TypeSpec); isTypeSpec {
						if typeSpec.Name.Name == attrName {
							structType, isStructType := typeSpec.Type.(*ast.StructType)
							if !isStructType {
								return false
							}

							for _, field := range structType.Fields.List {
								slog.Debug("found field", "field", field.Names[0].Name, "type", field.Type)
								tag, err := strconv.Unquote(field.Tag.Value)
								if err != nil {
									slog.Error("error unquoting tag", "tag", field.Tag.Value, "err", err)
									return false
								}

								slog.Debug("found tag", "tag", tag)

								// extract the `desc` and `optional` directive
								toolArg := NewArg(field.Names[0].Name)
								tagParts, err := parseTagString(tag)
								if err != nil {
									slog.Error("error parsing tag string", "tag", tag, "err", err)
									return false
								}

								// set argument type
								validatedJsonSchemaType, ok := jsonSchemaTypes[field.Type.(*ast.Ident).Name]
								if !ok {
									slog.Error("invalid json schema type", "type", field.Type.(*ast.Ident).Name)
									return false
								}

								toolArg.Type = validatedJsonSchemaType

								// hijack the json tag for the argument name
								if json, ok := tagParts["json"]; ok {
									toolArg.Name = json
								}

								if desc, ok := tagParts["desc"]; ok {
									toolArg.Description = desc
								}

								if optional, ok := tagParts["optional"]; ok {
									if optional == "true" {
										toolArg.Required = false
									}
								}

								toolArgs.Add(field.Names[0].Name, toolArg)
							}
						}
					}
				}
				return true
			}
			return true
		})

		return nil
	})

	return toolArgs, err
}

func parseTagString(tag string) (map[string]string, error) {
	result := make(map[string]string)
	// Regular expression to match key:"value" pairs, allowing for spaces within the values
	re := regexp.MustCompile(`(\w+):"([^"]+)"`)
	matches := re.FindAllStringSubmatch(tag, -1)

	for _, match := range matches {
		if len(match) == 3 { // [0] is the entire match, [1] is the key, [2] is the value
			key := match[1]
			value := match[2]
			result[key] = value
		}
	}

	return result, nil
}

func hasMarker(decl *ast.GenDecl) bool {
	if decl.Doc != nil {
		for _, comment := range decl.Doc.List {
			if strings.HasPrefix(comment.Text, "// +tool:") {
				return true
			}
		}
	}

	return false
}

// ################### GENERATOR ###################

type Generator struct {
	scanner *ToolScanner
}

func NewGenerator(scanner *ToolScanner) *Generator {
	return &Generator{
		scanner: scanner,
	}
}

func (g *Generator) Generate(suffix string) error {

	// Generate source files for each tool
	for _, tool := range g.scanner.tools {
		err := g.generateToolFile(tool, suffix)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateToolFile(tool *Tool, suffix string) error {

	// Generate the tool file
	content, err := g.generateToolFileContent(tool)
	if err != nil {
		return err
	}

	// Save the file
	originalPath := tool.SourcePath
	newPath := strings.TrimSuffix(originalPath, ".go") + suffix

	// Write the file to disc using writer
	file, err := os.Create(newPath) // #nosec
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

var tmpl = `// Code generated with go-openai-toolkit. DO NOT EDIT.

package {{.PackageName}}

import (
	"github.com/emilkje/go-openai-toolkit/toolkit"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func ({{.ReceiverName}} *{{.TypeName}}) Definition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name: "{{.Name}}",
		Description: "{{.Description}}",
		Parameters: jsonschema.Definition{
			Type:        jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				{{- range .Arguments}}
				"{{.Name}}": {
					Type:        {{.Type}},
					Description: "{{.Description}}",
				},
				{{- end}}
			},
			Required: []string { {{ join "\", \"" .RequiredArgs "\"" }} },
		},
	}
}

func New{{.TypeName}}() *{{.TypeName}} {
	return &{{.TypeName}}{&toolkit.ToolArgs[{{.ArgumentType}}]{}}
}
`

type Definition struct {
	ReceiverName string
	Name         string
	TypeName     string
	Description  string
	Arguments    []Arg
	ArgumentType string
	RequiredArgs []string
	PackageName  string
}

func join(sep string, s []string, surroundingStr string) string {
	if len(s) == 0 {
		return ""
	}
	if surroundingStr == "" {
		return strings.Join(s, sep)
	}
	return surroundingStr + strings.Join(s, sep) + surroundingStr
}

func (g *Generator) generateToolFileContent(tool *Tool) (string, error) {

	// Create a new template
	t := template.New("tool").Funcs(template.FuncMap{"join": join})

	// Parse the template
	t, err := t.Parse(tmpl)
	if err != nil {
		return "", err
	}

	// Execute the template
	var buf bytes.Buffer

	var requiredTools []string
	for _, arg := range tool.GetArguments() {
		if arg.Required {
			requiredTools = append(requiredTools, arg.Name)
		}
	}

	def := Definition{
		ReceiverName: strings.ToLower(tool.TypeName[:1]),
		Name:         tool.Name,
		TypeName:     tool.TypeName,
		Description:  tool.Description,
		Arguments:    tool.GetArguments(),
		RequiredArgs: requiredTools,
		PackageName:  tool.PackageName,
		ArgumentType: tool.ArgumentType,
	}

	err = t.Execute(&buf, def)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
