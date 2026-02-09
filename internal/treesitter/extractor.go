package treesitter

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

// SemanticUnit represents a semantic unit of code (function, class, etc.)
type SemanticUnit struct {
	Name      string
	Type      string // function, class, method, interface, etc.
	Content   string
	StartLine int
	EndLine   int
	Language  string
}

// Extractor extracts semantic units from AST
type Extractor struct {
	language string
	queries  map[string]string
}

// NewExtractor creates a new extractor for a language
func NewExtractor(language string) *Extractor {
	return &Extractor{
		language: language,
		queries:  getQueriesForLanguage(language),
	}
}

// Extract extracts semantic units from the AST
func (e *Extractor) Extract(root *sitter.Node, content []byte) []SemanticUnit {
	units := make([]SemanticUnit, 0)

	switch e.language {
	case "python":
		units = e.extractPython(root, content)
	case "javascript", "typescript":
		units = e.extractJavaScript(root, content)
	case "go":
		units = e.extractGo(root, content)
	case "php":
		units = e.extractPHP(root, content)
	case "java":
		units = e.extractJava(root, content)
	case "rust":
		units = e.extractRust(root, content)
	case "c":
		units = e.extractC(root, content)
	case "cpp":
		units = e.extractCPP(root, content)
	case "ruby":
		units = e.extractRuby(root, content)
	case "csharp":
		units = e.extractCSharp(root, content)
	default:
		// Fallback: extract top-level nodes
		units = e.extractGeneric(root, content)
	}

	return units
}

// extractPython extracts Python functions and classes
func (e *Extractor) extractPython(root *sitter.Node, content []byte) []SemanticUnit {
	units := make([]SemanticUnit, 0)

	// Query for classes and functions
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	e.walkNode(cursor, content, func(node *sitter.Node) {
		nodeType := node.Type()

		switch nodeType {
		case "function_definition":
			units = append(units, e.extractPythonFunction(node, content))
		case "class_definition":
			units = append(units, e.extractPythonClass(node, content))
		}
	})

	return units
}

// extractPythonFunction extracts a Python function
func (e *Extractor) extractPythonFunction(node *sitter.Node, content []byte) SemanticUnit {
	name := "anonymous"

	// Find function name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			name = child.Content(content)
			break
		}
	}

	return SemanticUnit{
		Name:      name,
		Type:      "function",
		Content:   node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Language:  "python",
	}
}

// extractPythonClass extracts a Python class
func (e *Extractor) extractPythonClass(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousClass"

	// Find class name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			name = child.Content(content)
			break
		}
	}

	return SemanticUnit{
		Name:      name,
		Type:      "class",
		Content:   node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Language:  "python",
	}
}

// extractJavaScript extracts JavaScript/TypeScript functions and classes
func (e *Extractor) extractJavaScript(root *sitter.Node, content []byte) []SemanticUnit {
	units := make([]SemanticUnit, 0)

	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	e.walkNode(cursor, content, func(node *sitter.Node) {
		nodeType := node.Type()

		switch nodeType {
		case "function_declaration", "function":
			units = append(units, e.extractJSFunction(node, content, "function"))
		case "arrow_function":
			units = append(units, e.extractJSFunction(node, content, "arrow_function"))
		case "method_definition":
			units = append(units, e.extractJSFunction(node, content, "method"))
		case "class_declaration", "class":
			units = append(units, e.extractJSClass(node, content))
		}
	})

	return units
}

// extractJSFunction extracts a JavaScript function
func (e *Extractor) extractJSFunction(node *sitter.Node, content []byte, funcType string) SemanticUnit {
	name := "anonymous"

	// Try to find function name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" || child.Type() == "property_identifier" {
			name = child.Content(content)
			break
		}
	}

	return SemanticUnit{
		Name:      name,
		Type:      funcType,
		Content:   node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Language:  e.language,
	}
}

// extractJSClass extracts a JavaScript class
func (e *Extractor) extractJSClass(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousClass"

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" || child.Type() == "type_identifier" {
			name = child.Content(content)
			break
		}
	}

	return SemanticUnit{
		Name:      name,
		Type:      "class",
		Content:   node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Language:  e.language,
	}
}

// extractGo extracts Go functions and types
func (e *Extractor) extractGo(root *sitter.Node, content []byte) []SemanticUnit {
	units := make([]SemanticUnit, 0)

	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	e.walkNode(cursor, content, func(node *sitter.Node) {
		nodeType := node.Type()

		switch nodeType {
		case "function_declaration":
			units = append(units, e.extractGoFunction(node, content))
		case "method_declaration":
			units = append(units, e.extractGoMethod(node, content))
		case "type_declaration":
			units = append(units, e.extractGoType(node, content))
		}
	})

	return units
}

// extractGoFunction extracts a Go function
func (e *Extractor) extractGoFunction(node *sitter.Node, content []byte) SemanticUnit {
	name := "anonymous"

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			name = child.Content(content)
			break
		}
	}

	return SemanticUnit{
		Name:      name,
		Type:      "function",
		Content:   node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Language:  "go",
	}
}

// extractGoMethod extracts a Go method
func (e *Extractor) extractGoMethod(node *sitter.Node, content []byte) SemanticUnit {
	name := "anonymous"

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "field_identifier" {
			name = child.Content(content)
			break
		}
	}

	return SemanticUnit{
		Name:      name,
		Type:      "method",
		Content:   node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Language:  "go",
	}
}

// extractGoType extracts a Go type
func (e *Extractor) extractGoType(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousType"

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "type_spec" {
			// Get identifier from type_spec
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "type_identifier" {
					name = subChild.Content(content)
					break
				}
			}
			break
		}
	}

	return SemanticUnit{
		Name:      name,
		Type:      "type",
		Content:   node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Language:  "go",
	}
}

// extractPHP extracts PHP functions and classes
func (e *Extractor) extractPHP(root *sitter.Node, content []byte) []SemanticUnit {
	units := make([]SemanticUnit, 0)

	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	e.walkNode(cursor, content, func(node *sitter.Node) {
		nodeType := node.Type()

		switch nodeType {
		case "function_definition":
			units = append(units, e.extractPHPFunction(node, content))
		case "method_declaration":
			units = append(units, e.extractPHPMethod(node, content))
		case "class_declaration":
			units = append(units, e.extractPHPClass(node, content))
		}
	})

	return units
}

// extractPHPFunction extracts a PHP function
func (e *Extractor) extractPHPFunction(node *sitter.Node, content []byte) SemanticUnit {
	name := "anonymous"

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "name" {
			name = child.Content(content)
			break
		}
	}

	return SemanticUnit{
		Name:      name,
		Type:      "function",
		Content:   node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Language:  "php",
	}
}

// extractPHPMethod extracts a PHP method
func (e *Extractor) extractPHPMethod(node *sitter.Node, content []byte) SemanticUnit {
	name := "anonymous"

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "name" {
			name = child.Content(content)
			break
		}
	}

	return SemanticUnit{
		Name:      name,
		Type:      "method",
		Content:   node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Language:  "php",
	}
}

// extractPHPClass extracts a PHP class
func (e *Extractor) extractPHPClass(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousClass"

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "name" {
			name = child.Content(content)
			break
		}
	}

	return SemanticUnit{
		Name:      name,
		Type:      "class",
		Content:   node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Language:  "php",
	}
}

// extractGeneric extracts top-level nodes as fallback
func (e *Extractor) extractGeneric(root *sitter.Node, content []byte) []SemanticUnit {
	units := make([]SemanticUnit, 0)

	// Just extract top-level children as chunks
	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if !child.IsNamed() {
			continue
		}

		units = append(units, SemanticUnit{
			Name:      fmt.Sprintf("block_%d", i),
			Type:      child.Type(),
			Content:   child.Content(content),
			StartLine: int(child.StartPoint().Row) + 1,
			EndLine:   int(child.EndPoint().Row) + 1,
			Language:  e.language,
		})
	}

	return units
}

// walkNode walks the AST and calls the visitor function for each node
func (e *Extractor) walkNode(cursor *sitter.TreeCursor, content []byte, visitor func(*sitter.Node)) {
	e.walkNodeRecursive(cursor, content, visitor)
}

func (e *Extractor) walkNodeRecursive(cursor *sitter.TreeCursor, content []byte, visitor func(*sitter.Node)) {
	node := cursor.CurrentNode()
	visitor(node)

	if cursor.GoToFirstChild() {
		e.walkNodeRecursive(cursor, content, visitor)
		cursor.GoToParent()
	}

	if cursor.GoToNextSibling() {
		e.walkNodeRecursive(cursor, content, visitor)
	}
}

// extractJava extracts Java classes, methods, and interfaces
func (e *Extractor) extractJava(root *sitter.Node, content []byte) []SemanticUnit {
	units := make([]SemanticUnit, 0)
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	e.walkNode(cursor, content, func(node *sitter.Node) {
		switch node.Type() {
		case "class_declaration":
			units = append(units, e.extractJavaClass(node, content))
		case "method_declaration":
			units = append(units, e.extractJavaMethod(node, content))
		case "interface_declaration":
			units = append(units, e.extractJavaInterface(node, content))
		}
	})
	return units
}

func (e *Extractor) extractJavaClass(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousClass"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "identifier" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "class", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "java",
	}
}

func (e *Extractor) extractJavaMethod(node *sitter.Node, content []byte) SemanticUnit {
	name := "anonymous"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "identifier" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "method", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "java",
	}
}

func (e *Extractor) extractJavaInterface(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousInterface"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "identifier" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "interface", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "java",
	}
}

// extractRust extracts Rust functions, structs, traits, impl blocks
func (e *Extractor) extractRust(root *sitter.Node, content []byte) []SemanticUnit {
	units := make([]SemanticUnit, 0)
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	e.walkNode(cursor, content, func(node *sitter.Node) {
		switch node.Type() {
		case "function_item":
			units = append(units, e.extractRustFunction(node, content))
		case "struct_item":
			units = append(units, e.extractRustStruct(node, content))
		case "trait_item":
			units = append(units, e.extractRustTrait(node, content))
		case "impl_item":
			units = append(units, e.extractRustImpl(node, content))
		}
	})
	return units
}

func (e *Extractor) extractRustFunction(node *sitter.Node, content []byte) SemanticUnit {
	name := "anonymous"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "identifier" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "function", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "rust",
	}
}

func (e *Extractor) extractRustStruct(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousStruct"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "type_identifier" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "struct", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "rust",
	}
}

func (e *Extractor) extractRustTrait(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousTrait"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "type_identifier" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "trait", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "rust",
	}
}

func (e *Extractor) extractRustImpl(node *sitter.Node, content []byte) SemanticUnit {
	name := "impl"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "type_identifier" {
			name = "impl " + child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "impl", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "rust",
	}
}

// extractC extracts C functions and structs
func (e *Extractor) extractC(root *sitter.Node, content []byte) []SemanticUnit {
	units := make([]SemanticUnit, 0)
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	e.walkNode(cursor, content, func(node *sitter.Node) {
		switch node.Type() {
		case "function_definition":
			units = append(units, e.extractCFunction(node, content))
		case "struct_specifier":
			units = append(units, e.extractCStruct(node, content))
		}
	})
	return units
}

func (e *Extractor) extractCFunction(node *sitter.Node, content []byte) SemanticUnit {
	name := "anonymous"
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "function_declarator" {
			for j := 0; j < int(child.ChildCount()); j++ {
				if gchild := child.Child(j); gchild.Type() == "identifier" {
					name = gchild.Content(content)
					break
				}
			}
		}
	}
	return SemanticUnit{
		Name: name, Type: "function", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "c",
	}
}

func (e *Extractor) extractCStruct(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousStruct"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "type_identifier" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "struct", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "c",
	}
}

// extractCPP extracts C++ classes, functions, namespaces
func (e *Extractor) extractCPP(root *sitter.Node, content []byte) []SemanticUnit {
	units := make([]SemanticUnit, 0)
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	e.walkNode(cursor, content, func(node *sitter.Node) {
		switch node.Type() {
		case "function_definition":
			units = append(units, e.extractCPPFunction(node, content))
		case "class_specifier":
			units = append(units, e.extractCPPClass(node, content))
		case "namespace_definition":
			units = append(units, e.extractCPPNamespace(node, content))
		}
	})
	return units
}

func (e *Extractor) extractCPPFunction(node *sitter.Node, content []byte) SemanticUnit {
	name := "anonymous"
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "function_declarator" {
			for j := 0; j < int(child.ChildCount()); j++ {
				if gchild := child.Child(j); gchild.Type() == "identifier" || gchild.Type() == "field_identifier" {
					name = gchild.Content(content)
					break
				}
			}
		}
	}
	return SemanticUnit{
		Name: name, Type: "function", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "cpp",
	}
}

func (e *Extractor) extractCPPClass(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousClass"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "type_identifier" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "class", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "cpp",
	}
}

func (e *Extractor) extractCPPNamespace(node *sitter.Node, content []byte) SemanticUnit {
	name := "global"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "identifier" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "namespace", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "cpp",
	}
}

// extractRuby extracts Ruby classes, modules, methods
func (e *Extractor) extractRuby(root *sitter.Node, content []byte) []SemanticUnit {
	units := make([]SemanticUnit, 0)
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	e.walkNode(cursor, content, func(node *sitter.Node) {
		switch node.Type() {
		case "class":
			units = append(units, e.extractRubyClass(node, content))
		case "module":
			units = append(units, e.extractRubyModule(node, content))
		case "method":
			units = append(units, e.extractRubyMethod(node, content))
		}
	})
	return units
}

func (e *Extractor) extractRubyClass(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousClass"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "constant" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "class", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "ruby",
	}
}

func (e *Extractor) extractRubyModule(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousModule"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "constant" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "module", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "ruby",
	}
}

func (e *Extractor) extractRubyMethod(node *sitter.Node, content []byte) SemanticUnit {
	name := "anonymous"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "identifier" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "method", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "ruby",
	}
}

// extractCSharp extracts C# classes, methods, namespaces
func (e *Extractor) extractCSharp(root *sitter.Node, content []byte) []SemanticUnit {
	units := make([]SemanticUnit, 0)
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	e.walkNode(cursor, content, func(node *sitter.Node) {
		switch node.Type() {
		case "class_declaration":
			units = append(units, e.extractCSharpClass(node, content))
		case "method_declaration":
			units = append(units, e.extractCSharpMethod(node, content))
		case "namespace_declaration":
			units = append(units, e.extractCSharpNamespace(node, content))
		case "interface_declaration":
			units = append(units, e.extractCSharpInterface(node, content))
		}
	})
	return units
}

func (e *Extractor) extractCSharpClass(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousClass"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "identifier" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "class", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "csharp",
	}
}

func (e *Extractor) extractCSharpMethod(node *sitter.Node, content []byte) SemanticUnit {
	name := "anonymous"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "identifier" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "method", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "csharp",
	}
}

func (e *Extractor) extractCSharpNamespace(node *sitter.Node, content []byte) SemanticUnit {
	name := "global"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "identifier" || child.Type() == "qualified_name" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "namespace", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "csharp",
	}
}

func (e *Extractor) extractCSharpInterface(node *sitter.Node, content []byte) SemanticUnit {
	name := "AnonymousInterface"
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child.Type() == "identifier" {
			name = child.Content(content)
			break
		}
	}
	return SemanticUnit{
		Name: name, Type: "interface", Content: node.Content(content),
		StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1, Language: "csharp",
	}
}

// getQueriesForLanguage returns Tree-sitter queries for a language
func getQueriesForLanguage(language string) map[string]string {
	queries := make(map[string]string)

	switch language {
	case "python":
		queries["functions"] = "(function_definition) @function"
		queries["classes"] = "(class_definition) @class"
	case "javascript", "typescript":
		queries["functions"] = "(function_declaration) @function"
		queries["classes"] = "(class_declaration) @class"
	case "go":
		queries["functions"] = "(function_declaration) @function"
		queries["methods"] = "(method_declaration) @method"
		queries["types"] = "(type_declaration) @type"
	case "php":
		queries["functions"] = "(function_definition) @function"
		queries["methods"] = "(method_declaration) @method"
		queries["classes"] = "(class_declaration) @class"
	case "java":
		queries["classes"] = "(class_declaration) @class"
		queries["methods"] = "(method_declaration) @method"
		queries["interfaces"] = "(interface_declaration) @interface"
	case "rust":
		queries["functions"] = "(function_item) @function"
		queries["structs"] = "(struct_item) @struct"
		queries["traits"] = "(trait_item) @trait"
		queries["impls"] = "(impl_item) @impl"
	case "c":
		queries["functions"] = "(function_definition) @function"
		queries["structs"] = "(struct_specifier) @struct"
	case "cpp":
		queries["functions"] = "(function_definition) @function"
		queries["classes"] = "(class_specifier) @class"
		queries["namespaces"] = "(namespace_definition) @namespace"
	case "ruby":
		queries["classes"] = "(class) @class"
		queries["modules"] = "(module) @module"
		queries["methods"] = "(method) @method"
	case "csharp":
		queries["classes"] = "(class_declaration) @class"
		queries["methods"] = "(method_declaration) @method"
		queries["namespaces"] = "(namespace_declaration) @namespace"
		queries["interfaces"] = "(interface_declaration) @interface"
	}

	return queries
}
