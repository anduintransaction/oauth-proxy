//go:generate goru packer template template -out=.
package template

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"html/template"
	"os"
	"reflect"
	"strings"
	"text/template/parse"

	"gottb.io/goru/packer"
)

type Error struct {
	File    string
	Msg     string
	Context string
}

func (e *Error) Error() string {
	if e.Context == "" {
		return fmt.Sprintf("%s: %s", e.File, e.Msg)
	}
	return fmt.Sprintf("%s: %s\n%s", e.File, e.Msg, e.Context)
}

type Arg struct {
	Name    string
	DotName string
	Type    string
}

type Template struct {
	HTMLTemplate *template.Template
	Args         []*Arg
	Imports      map[string]string
	Content      []byte
	signature    *types.Signature
	sigName      string
}

func (t *Template) generateSignatureFile(name string) error {
	w, err := os.Create("generated/check/" + name + ".go")
	if err != nil {
		return err
	}
	defer w.Close()
	tmpl, err := packer.LoadTemplate(VVVPack, "sig.go.tmpl")
	if err != nil {
		return err
	}
	args := []string{}
	for _, arg := range t.Args {
		args = append(args, fmt.Sprintf("%s %s", arg.Name, arg.Type))
	}
	tmplData := &struct {
		Imports     map[string]string
		Declaration string
		Name        string
	}{
		Imports:     t.Imports,
		Declaration: strings.Join(args, ", "),
		Name:        name,
	}
	return tmpl.Execute(w, tmplData)
}

var Templates = make(map[string]*Template)
var sigMap = make(map[string]*Template)

func Check() error {
	err := os.MkdirAll("generated/check", 0755)
	if err != nil {
		return err
	}
	defer os.RemoveAll("generated/check")
	err = generateVariableFiles()
	if err != nil {
		return err
	}
	err = generateFunctionsFile()
	if err != nil {
		return err
	}
	err = generateHelpersFile()
	if err != nil {
		return err
	}
	objects, err := buildObjects()
	if err != nil {
		return err
	}
	funcs := make(map[string]types.Type)
	helpers := make(map[string]types.Type)
	for name, object := range objects {
		if strings.HasPrefix(name, "fff_") {
			name = strings.TrimPrefix(name, "fff_")
			funcs[name] = object
		} else if strings.HasPrefix(name, "hhh_") {
			name = strings.TrimPrefix(name, "hhh_")
			helpers[name] = object
		} else if strings.HasPrefix(name, "vvv_") {
			name = strings.TrimPrefix(name, "vvv_")
			t := sigMap[name]
			signature, ok := object.(*types.Signature)
			if !ok {
				return fmt.Errorf("template signature is not a function: %s", name)
			}
			t.signature = signature
		}
	}
	for name, t := range Templates {
		treeSet, err := parse.Parse(name, string(t.Content), "", "", Funcs)
		if err != nil {
			return err
		}
		if len(treeSet) > 1 {
			return &Error{name, "define, block, template is not allowed", ""}
		}
		rootScope := newScope()
		for name, object := range funcs {
			rootScope.symbols[name] = object
		}
		params := t.signature.Params()
		for i := 0; i < params.Len(); i++ {
			param := params.At(i)
			rootScope.symbols["$"+param.Name()] = param.Type()
		}
		ch := &checker{
			name:         name,
			tree:         t.HTMLTemplate.Tree,
			argLen:       len(t.Args),
			rootScope:    rootScope,
			currentScope: rootScope,
			argSkipped:   0,
			context:      "",
			helpers:      helpers,
		}
		err = ch.check()
		if err != nil {
			return err
		}
	}
	return nil
}

func generateVariableFiles() error {
	i := 0
	for _, t := range Templates {
		i++
		sigName := fmt.Sprintf("sig%d", i)
		err := t.generateSignatureFile(sigName)
		if err != nil {
			return err
		}
		t.sigName = sigName
		sigMap[sigName] = t
	}
	return nil
}

func generateFunctionsFile() error {
	w, err := os.Create("generated/check/functions.go")
	if err != nil {
		return err
	}
	defer w.Close()
	f := newFormatter()
	funcs := make(map[string]string)
	for name, fun := range Funcs {
		f.context = name
		signature, err := f.sprint(reflect.TypeOf(fun))
		if err != nil {
			return err
		}
		funcs[name] = signature
	}
	tmpl, err := packer.LoadTemplate(VVVPack, "functions.go.tmpl")
	if err != nil {
		return err
	}
	tmplData := &struct {
		Funcs   map[string]string
		Imports map[string]string
	}{
		Funcs:   funcs,
		Imports: f.imports,
	}
	return tmpl.Execute(w, tmplData)
}

func generateHelpersFile() error {
	w, err := os.Create("generated/check/helpers.go")
	if err != nil {
		return err
	}
	defer w.Close()
	tmpl, err := packer.LoadTemplate(VVVPack, "helpers.go.tmpl")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, nil)
}

func buildObjects() (map[string]types.Type, error) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, "generated/check", nil, 0)
	if err != nil {
		return nil, err
	}
	checkPackage, ok := packages["check"]
	if !ok {
		return nil, fmt.Errorf("no check package found")
	}
	files := []*ast.File{}
	for _, file := range checkPackage.Files {
		files = append(files, file)
	}
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("check", fset, files, nil)
	if err != nil {
		return nil, err
	}
	names := pkg.Scope().Names()
	symbols := make(map[string]types.Type)
	for _, name := range names {
		obj := pkg.Scope().Lookup(name)
		if obj == nil {
			continue
		}
		symbols[name] = obj.Type()
	}
	return symbols, nil
}

type checker struct {
	name         string
	tree         *parse.Tree
	argLen       int
	rootScope    *scope
	currentScope *scope
	argSkipped   int
	context      string
	helpers      map[string]types.Type
}

func (c *checker) check() error {
	for _, node := range c.tree.Root.Nodes {
		c.context = node.String()
		_, err := c.checkNode(node)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *checker) checkNode(node parse.Node) (types.Type, error) {
	switch node := node.(type) {
	case *parse.StringNode:
		return types.Typ[types.String], nil
	case *parse.BoolNode:
		return types.Typ[types.Bool], nil
	case *parse.NumberNode:
		switch {
		case node.IsInt || node.IsUint:
			return types.Typ[types.UntypedInt], nil
		case node.IsFloat:
			return types.Typ[types.UntypedFloat], nil
		case node.IsComplex:
			return types.Typ[types.UntypedComplex], nil
		}
		return nil, &Error{c.name, fmt.Sprintf("not allowed: %T", node), c.context}
	case *parse.NilNode:
		return types.Typ[types.UntypedNil], nil
	case *parse.ActionNode:
		return c.checkActionNode(node)
	case *parse.VariableNode:
		return c.checkVariableNode(node)
	case *parse.PipeNode:
		return c.checkPipeNode(node)
	case *parse.ChainNode:
		return c.checkChainNode(node)
	case *parse.RangeNode:
		return c.checkRangeNode(node)
	case *parse.IfNode:
		return c.checkIfNode(node)
	case *parse.WithNode:
		return c.checkWithNode(node)
	case *parse.TemplateNode:
		return nil, &Error{c.name, "define, block, template are not allowed in goru template", ""}
	case *parse.TextNode:
		return nil, nil
	default:
		return nil, &Error{c.name, fmt.Sprintf("not allowed: %T", node), c.context}
	}
}

func (c *checker) checkVariableNode(node *parse.VariableNode) (types.Type, error) {
	if len(node.Ident) == 0 {
		return nil, &Error{c.name, "empty variable name", c.context}
	}
	varName := node.Ident[0]
	varType, ok := c.currentScope.findSymbol(varName)
	if !ok {
		return nil, &Error{c.name, fmt.Sprintf("variable not found: %s", varName), c.context}
	}
	return c.inferType(varType, node.Ident[1:])
}

func (c *checker) checkChainNode(node *parse.ChainNode) (types.Type, error) {
	varType, err := c.checkNode(node.Node)
	if err != nil {
		return nil, err
	}
	return c.inferType(varType, node.Field)
}

func (c *checker) checkActionNode(node *parse.ActionNode) (types.Type, error) {
	if c.argSkipped < c.argLen {
		c.argSkipped++
		return nil, nil
	}
	if len(node.Pipe.Decl) > 0 {
		return c.checkDeclarationNode(node.Pipe)
	}
	return c.checkPipeNode(node.Pipe)

}

func (c *checker) checkDeclarationNode(node *parse.PipeNode) (types.Type, error) {
	argName, err := c.getVarName(node.Decl[0])
	if err != nil {
		return nil, err
	}
	if _, ok := c.currentScope.findSymbol(argName); ok {
		return nil, &Error{c.name, fmt.Sprintf("%s was declared in this scope", argName), c.context}
	}
	t, err := c.checkPipeNode(node)
	if err != nil {
		return nil, err
	}
	c.currentScope.symbols[argName] = t
	return t, nil
}

func (c *checker) checkPipeNode(node *parse.PipeNode) (types.Type, error) {
	var lastArg types.Type = nil
	var err error
	for _, cmdNode := range node.Cmds {
		lastArg, err = c.checkCommandNode(cmdNode, lastArg)
		if err != nil {
			return nil, err
		}
	}
	return lastArg, nil
}

func (c *checker) checkCommandNode(node *parse.CommandNode, lastArg types.Type) (types.Type, error) {
	args := node.Args
	if len(args) == 0 {
		return nil, &Error{c.name, "empty command", c.context}
	}
	switch firstArgs := args[0].(type) {
	case *parse.StringNode, *parse.BoolNode, *parse.NumberNode:
		if lastArg != nil {
			return nil, &Error{c.name, "invalid function call", c.context}
		}
		return c.checkNode(firstArgs)
	case *parse.VariableNode, *parse.ChainNode, *parse.PipeNode:
		varType, err := c.checkNode(firstArgs)
		if err != nil {
			return nil, err
		}
		signature, ok := varType.(*types.Signature)
		if ok {
			if signature.Recv() == nil {
				if len(args[1:]) > 0 || lastArg != nil {
					return nil, &Error{c.name, fmt.Sprintf("%s is a function but cannot be invoked, use \"call\" instead", firstArgs), c.context}
				}
				return varType, nil
			}
			return c.checkFuncCall(firstArgs.String(), signature, args[1:], lastArg)
		}
		if len(args[1:]) > 0 || lastArg != nil {
			return nil, &Error{c.name, fmt.Sprintf("%s is not a function", firstArgs), c.context}
		}
		return varType, nil
	case *parse.IdentifierNode:
		return c.checkCommandStartWithIdentifier(firstArgs, args[1:], lastArg)
	default:
		return nil, &Error{c.name, fmt.Sprintf("not allowed: %T", firstArgs), c.context}
	}
}

func (c *checker) checkCommandStartWithIdentifier(ident *parse.IdentifierNode, args []parse.Node, lastArg types.Type) (types.Type, error) {
	if ident.Ident == "index" {
		return c.checkIndexCall(args, lastArg)
	} else if ident.Ident == "call" {
		return c.checkCallCall(args, lastArg)
	} else if ident.Ident == "include" {
		return c.checkIncludeCall(args, lastArg)
	}
	funcType, ok := c.currentScope.findSymbol(ident.Ident)
	if !ok {
		return nil, &Error{c.name, fmt.Sprintf("function not found: %s", ident.Ident), c.context}
	}
	signature, ok := funcType.(*types.Signature)
	if !ok {
		return nil, &Error{c.name, fmt.Sprintf("%s is not a function", ident.Ident), c.context}
	}
	return c.checkFuncCall(ident.Ident, signature, args, lastArg)
}

func (c *checker) checkRangeNode(node *parse.RangeNode) (types.Type, error) {
	c.context = node.Pipe.String()
	if len(node.Pipe.Decl) == 0 {
		return nil, &Error{c.name, "not enough variable in range", c.context}
	}
	pipeType, err := c.checkPipeNode(node.Pipe)
	if err != nil {
		return nil, err
	}
	scope := newScope()
	scope.parent = c.currentScope
	switch pipeType := pipeType.(type) {
	case *types.Map:
		if len(node.Pipe.Decl) == 1 {
			valueName, err := c.getVarName(node.Pipe.Decl[0])
			if err != nil {
				return nil, err
			}
			if _, ok := c.currentScope.findSymbol(valueName); ok {
				return nil, &Error{c.name, fmt.Sprintf("%s was declared in this scope", valueName), c.context}
			}
			scope.symbols[valueName] = pipeType.Elem()
		} else {
			keyName, err := c.getVarName(node.Pipe.Decl[0])
			if err != nil {
				return nil, err
			}
			if _, ok := c.currentScope.findSymbol(keyName); ok {
				return nil, &Error{c.name, fmt.Sprintf("%s was declared in this scope", keyName), c.context}
			}
			scope.symbols[keyName] = pipeType.Key()
			valueName, err := c.getVarName(node.Pipe.Decl[1])
			if err != nil {
				return nil, err
			}
			if _, ok := c.currentScope.findSymbol(valueName); ok {
				return nil, &Error{c.name, fmt.Sprintf("%s was declared in this scope", valueName), c.context}
			}
			scope.symbols[valueName] = pipeType.Elem()
		}
	case *types.Array, *types.Slice:
		var elem types.Type
		if t, ok := pipeType.(*types.Array); ok {
			elem = t.Elem()
		} else if t, ok := pipeType.(*types.Slice); ok {
			elem = t.Elem()
		}
		if len(node.Pipe.Decl) == 1 {
			valueName, err := c.getVarName(node.Pipe.Decl[0])
			if err != nil {
				return nil, err
			}
			if _, ok := c.currentScope.findSymbol(valueName); ok {
				return nil, &Error{c.name, fmt.Sprintf("%s was declared in this scope", valueName), c.context}
			}
			scope.symbols[valueName] = elem
		} else {
			keyName, err := c.getVarName(node.Pipe.Decl[0])
			if err != nil {
				return nil, err
			}
			if _, ok := c.currentScope.findSymbol(keyName); ok {
				return nil, &Error{c.name, fmt.Sprintf("%s was declared in this scope", keyName), c.context}
			}
			scope.symbols[keyName] = types.Typ[types.Int]
			valueName, err := c.getVarName(node.Pipe.Decl[1])
			if err != nil {
				return nil, err
			}
			if _, ok := c.currentScope.findSymbol(valueName); ok {
				return nil, &Error{c.name, fmt.Sprintf("%s was declared in this scope", valueName), c.context}
			}
			scope.symbols[valueName] = elem
		}
	default:
		return nil, &Error{c.name, fmt.Sprintf("cannot range over %s", pipeType), c.context}
	}
	c.currentScope = scope
	oldContext := c.context
	if node.List != nil {
		for _, childNode := range node.List.Nodes {
			c.context = childNode.String()
			_, err = c.checkNode(childNode)
			if err != nil {
				return nil, err
			}
		}
	}
	c.currentScope = scope.parent
	if node.ElseList != nil {
		for _, childNode := range node.ElseList.Nodes {
			c.context = childNode.String()
			_, err = c.checkNode(childNode)
			if err != nil {
				return nil, err
			}
		}
	}
	c.context = oldContext
	return nil, nil
}

func (c *checker) checkIfNode(node *parse.IfNode) (types.Type, error) {
	c.context = node.Pipe.String()
	_, err := c.checkNode(node.Pipe)
	if err != nil {
		return nil, err
	}
	if node.List != nil {
		for _, childNode := range node.List.Nodes {
			c.context = childNode.String()
			_, err = c.checkNode(childNode)
			if err != nil {
				return nil, err
			}
		}
	}
	if node.ElseList != nil {
		for _, childNode := range node.ElseList.Nodes {
			c.context = childNode.String()
			_, err = c.checkNode(childNode)
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

func (c *checker) checkWithNode(node *parse.WithNode) (types.Type, error) {
	c.context = node.Pipe.String()
	if len(node.Pipe.Decl) == 0 {
		return nil, &Error{c.name, "not enough variable in range", c.context}
	}
	scope := newScope()
	scope.parent = c.currentScope
	c.currentScope = scope
	_, err := c.checkDeclarationNode(node.Pipe)
	if err != nil {
		return nil, err
	}
	if node.List != nil {
		for _, childNode := range node.List.Nodes {
			c.context = childNode.String()
			_, err = c.checkNode(childNode)
			if err != nil {
				return nil, err
			}
		}
	}
	c.currentScope = scope.parent
	if node.ElseList != nil {
		for _, childNode := range node.ElseList.Nodes {
			c.context = childNode.String()
			_, err = c.checkNode(childNode)
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

func (c *checker) checkIndexCall(args []parse.Node, lastArg types.Type) (types.Type, error) {
	if len(args) == 0 || (len(args) == 1 && lastArg == nil) {
		return nil, &Error{c.name, "not enough argument in call to index", c.context}
	}
	obj, err := c.checkNode(args[0])
	if err != nil {
		return nil, err
	}
	_ = obj
	argTypes := []types.Type{}
	for _, arg := range args[1:] {
		t, err := c.checkNode(arg)
		if err != nil {
			return nil, err
		}
		argTypes = append(argTypes, t)
	}
	if lastArg != nil {
		argTypes = append(argTypes, lastArg)
	}
	for _, argType := range argTypes {
		obj, err = c.checkIndex(obj, argType)
		if err != nil {
			return nil, err
		}
	}
	return obj, nil
}

func (c *checker) checkIndex(obj types.Type, key types.Type) (types.Type, error) {
	switch obj := obj.(type) {
	case *types.Map:
		keyType := obj.Key()
		if !c.assignableTo(key, keyType) {
			return nil, &Error{c.name, fmt.Sprintf("cannot use type %s as type %s of map index", key, keyType), c.context}
		}
		return obj.Elem(), nil
	case *types.Array, *types.Slice:
		keyType, ok := key.(*types.Basic)
		if !ok {
			return nil, &Error{c.name, fmt.Sprintf("non integer type in slice index: %s", key), c.context}
		}
		if (keyType.Kind() < types.Int || keyType.Kind() > types.Uint64) && keyType.Kind() != types.UntypedInt {
			return nil, &Error{c.name, fmt.Sprintf("non integer type in slice index: %s", key), c.context}
		}
		if objType, ok := obj.(*types.Array); ok {
			return objType.Elem(), nil
		} else if objType, ok := obj.(*types.Slice); ok {
			return objType.Elem(), nil
		}
		return nil, &Error{c.name, fmt.Sprintf("cannot take index of %s", obj), c.context}
	default:
		return nil, &Error{c.name, fmt.Sprintf("cannot take index of %s", obj), c.context}
	}
	return nil, nil
}

func (c *checker) checkCallCall(args []parse.Node, lastArg types.Type) (types.Type, error) {
	if len(args) == 0 {
		return nil, &Error{c.name, "not enough argument in call", c.context}
	}
	fun, err := c.checkNode(args[0])
	if err != nil {
		return nil, err
	}
	signature, ok := fun.(*types.Signature)
	if !ok {
		return nil, &Error{c.name, fmt.Sprintf("%s is not a function", args[0]), c.context}
	}
	if signature.Recv() != nil {
		return nil, &Error{c.name, fmt.Sprintf("%s is a method receiver, not allowed in call", args[0]), c.context}
	}
	return c.checkFuncCall(args[0].String(), signature, args[1:], lastArg)
}

func (c *checker) checkIncludeCall(args []parse.Node, lastArg types.Type) (types.Type, error) {
	if len(args) == 0 {
		return nil, &Error{c.name, "not enough argument in include", c.context}
	}
	node, ok := args[0].(*parse.StringNode)
	if !ok {
		return nil, &Error{c.name, fmt.Sprintf("first argument in include must be a string"), c.context}
	}
	templateName := node.Text
	if !strings.HasSuffix(templateName, ".tmpl.html") {
		templateName += ".tmpl.html"
	}
	tmpl, ok := Templates[templateName]
	if !ok {
		return nil, &Error{c.name, fmt.Sprintf("template not found: %s", templateName), c.context}
	}
	return c.checkFuncCall(templateName, tmpl.signature, args[1:], lastArg)
}

func (c *checker) checkFuncCall(name string, f *types.Signature, args []parse.Node, lastArg types.Type) (types.Type, error) {
	params := f.Params()
	results := f.Results()
	if results.Len() < 1 || results.Len() > 2 {
		return nil, &Error{c.name, fmt.Sprintf("invalid function: %s - %s (invalid number of returned values)", name, f), c.context}
	}
	if results.Len() == 2 {
		if !c.assignableTo(results.At(1).Type(), c.helpers["err"]) {
			return nil, &Error{c.name, fmt.Sprintf("invalid function: %s - %s (invalid type of second returned value)", name, f), c.context}
		}
	}
	typeArgs := []types.Type{}
	for _, arg := range args {
		t, err := c.checkNode(arg)
		if err != nil {
			return nil, err
		}
		if signature, ok := t.(*types.Signature); ok {
			if signature.Recv() != nil {
				t, err = c.checkFuncCall(arg.String(), signature, nil, nil)
				if err != nil {
					return nil, err
				}
			}
		}
		typeArgs = append(typeArgs, t)
	}
	if lastArg != nil {
		typeArgs = append(typeArgs, lastArg)
	}
	if params.Len() > len(typeArgs) && !f.Variadic() {
		return nil, &Error{c.name, fmt.Sprintf("not enough arguments in call to %s", name), c.context}
	}
	for i, typeArg := range typeArgs {
		if typeArg == nil {
			return nil, nil
		}
		if i < params.Len()-1 {
			param := params.At(i).Type()
			if !c.assignableTo(typeArg, param) {
				return nil, &Error{c.name, fmt.Sprintf("cannot use type %s as type %s in argument to %s", typeArg, param, name), c.context}
			}
		} else if i == params.Len()-1 {
			param := params.At(i).Type()
			if !f.Variadic() {
				if !c.assignableTo(typeArg, param) {
					return nil, &Error{c.name, fmt.Sprintf("cannot use type %s as type %s in argument to %s", typeArg, param, name), c.context}
				}
			} else {
				s, ok := param.(*types.Slice)
				if !ok {
					return nil, &Error{c.name, fmt.Sprintf("invalid function: %s - %s", name, f), c.context}
				}
				if !c.assignableTo(typeArg, s.Elem()) {
					return nil, &Error{c.name, fmt.Sprintf("cannot use type %s as type %s in argument to %s", typeArg, s.Elem(), name), c.context}
				}
			}
		} else {
			if !f.Variadic() {
				return nil, &Error{c.name, fmt.Sprintf("too many arguments in call to %s", name), c.context}
			}
			param := params.At(params.Len() - 1).Type()
			s, ok := param.(*types.Slice)
			if !ok {
				return nil, &Error{c.name, fmt.Sprintf("invalid function: %s - %s", name, f), c.context}
			}
			if !c.assignableTo(typeArg, s.Elem()) {
				return nil, &Error{c.name, fmt.Sprintf("cannot use type %s as type %s in argument to %s", typeArg, s.Elem(), name), c.context}
			}
		}
	}
	return results.At(0).Type(), nil
}

func (c *checker) getVarName(variable *parse.VariableNode) (string, error) {
	if len(variable.Ident) != 1 {
		return "", &Error{c.name, fmt.Sprintf("invalid variable declaration: %s", variable), c.context}
	}
	return variable.Ident[0], nil
}

func (c *checker) inferType(varType types.Type, idents []string) (types.Type, error) {
	if len(idents) == 0 {
		return varType, nil
	}
	if t, ok := varType.(*types.Pointer); ok {
		return c.inferType(t.Elem(), idents)
	}
	t, ok := varType.(*types.Named)
	if !ok {
		return nil, &Error{c.name, fmt.Sprintf("type %s has no field or method", varType), c.context}
	}
	underlying, ok := t.Underlying().(*types.Struct)
	if !ok {
		return nil, &Error{c.name, fmt.Sprintf("type %s has no field or method", varType), c.context}
	}
	ident := idents[0]
	var subType types.Type
	for i := 0; i < t.NumMethods(); i++ {
		method := t.Method(i)
		if method.Name() == ident && method.Exported() {
			signature, ok := method.Type().(*types.Signature)
			if !ok {
				return nil, &Error{c.name, fmt.Sprintf("invalid signature: %s", method.Type()), c.context}
			}
			subType = signature
		}
	}
	for i := 0; i < underlying.NumFields(); i++ {
		field := underlying.Field(i)
		if field.Name() == ident && field.Exported() {
			subType = underlying.Field(i).Type()
		}
	}
	if subType == nil {
		return nil, &Error{c.name, fmt.Sprintf("type %s has no field or method %s", varType, ident), c.context}
	}
	if len(idents[1:]) == 0 {
		return subType, nil
	}
	signature, ok := subType.(*types.Signature)
	if !ok {
		return c.inferType(subType, idents[1:])
	}
	subType, err := c.checkFuncCall(ident, signature, nil, nil)
	if err != nil {
		return nil, err
	}
	return c.inferType(subType, idents[1:])
}

func (c *checker) assignableTo(V, T types.Type) bool {
	if types.AssignableTo(V, T) {
		return true
	}
	if _, ok := V.(*types.Basic); ok {
		t, ok := T.(*types.Basic)
		if !ok {
			return false
		}
		switch V {
		case types.Typ[types.UntypedInt]:
			return t.Kind() >= types.Int && t.Kind() <= types.Complex128
		case types.Typ[types.UntypedFloat]:
			return t.Kind() >= types.Float32 && t.Kind() <= types.Complex128
		case types.Typ[types.UntypedComplex]:
			return t.Kind() >= types.Complex64 && t.Kind() <= types.Complex128
		}
	}
	return false
}

type scope struct {
	parent  *scope
	symbols map[string]types.Type
}

func newScope() *scope {
	return &scope{
		parent:  nil,
		symbols: make(map[string]types.Type),
	}
}

func (s *scope) findSymbol(name string) (types.Type, bool) {
	current := s
	for current != nil {
		t, ok := current.symbols[name]
		if ok {
			return t, true
		}
		current = current.parent
	}
	return nil, false
}
