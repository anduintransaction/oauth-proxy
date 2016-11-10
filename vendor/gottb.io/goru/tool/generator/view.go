package generator

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gottb.io/goru"
	"gottb.io/goru/errors"
	"gottb.io/goru/packer"
	"gottb.io/goru/utils"
	"gottb.io/goru/view/template"
)

type viewParseError struct {
	filename string
	reason   string
	line     string
}

func (e *viewParseError) Error() string {
	if e.line != "" {
		return fmt.Sprintf("error parsing view file %s, %s: %s", e.filename, e.reason, e.line)
	} else {
		return fmt.Sprintf("error parsing view file %s, %s", e.filename, e.reason)
	}
}

type viewArg struct {
	name    string
	dotName string
	argType string
}

type viewImports map[string]string

func (v viewImports) exists(name string) bool {
	_, ok := v[name]
	return ok
}

type viewData struct {
	args         map[string]*viewArg
	argSorted    []string
	imports      viewImports
	relativePath string
	argsContent  *bytes.Buffer
	content      *bytes.Buffer
}

func (v *viewData) collectBody(line string) {
	fmt.Fprintln(v.content, line)
}

var viewImportRegex = regexp.MustCompile("^//import\\s+([a-zA-Z0-9\\.]+\\s+)?\"([a-zA-Z0-9/_\\-\\.]+)\"\\s*$")

func (v *viewData) collectImport(line string) error {
	matches := viewImportRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return errors.Wrap(&viewParseError{v.relativePath, "invalid import line", line})
	}
	importName := strings.TrimSpace(matches[1])
	pkgPath := matches[2]
	if importName == "" {
		importName = filepath.Base(pkgPath)
	}
	oldPath, ok := v.imports[importName]
	if ok && pkgPath != oldPath {
		return errors.Wrap(&viewParseError{v.relativePath, "duplicate imports: " + importName, ""})
	}
	v.imports[importName] = pkgPath
	return nil
}

func (v *viewData) collectArgs(line string) {
	funcLine := strings.TrimPrefix(line, "//")
	fmt.Fprintln(v.argsContent, funcLine)
}

func (v *viewData) buildArgs() error {
	argsContent := v.argsContent.String()
	expr, err := parser.ParseExpr(argsContent)
	if err != nil {
		return errors.Wrap(&viewParseError{v.relativePath, err.Error(), ""})
	}
	fun, ok := expr.(*ast.FuncType)
	if !ok {
		return errors.Wrap(&viewParseError{v.relativePath, "invalid argument declaration", ""})
	}
	args := fun.Params.List
	fset := token.NewFileSet()
	for _, arg := range args {
		names := arg.Names
		if len(names) != 1 {
			return errors.Wrap(&viewParseError{v.relativePath, "invalid argument declaration", ""})
		}
		argName := names[0].Name
		b := &bytes.Buffer{}
		err = printer.Fprint(b, fset, arg.Type)
		if err != nil {
			return errors.Wrap(&viewParseError{v.relativePath, err.Error(), ""})
		}
		_, ok := v.args[argName]
		if ok {
			return errors.Wrap(&viewParseError{v.relativePath, "arg has been declared: " + argName, ""})
		}
		v.args[argName] = &viewArg{
			name:    argName,
			dotName: makePublicVariableName(argName),
			argType: b.String(),
		}
		v.argSorted = append(v.argSorted, argName)
	}
	return nil
}

func (v *viewData) writeArgsDeclare() error {
	for _, arg := range v.args {
		fieldName := makePublicVariableName(arg.name)
		fmt.Fprintf(v.content, "{{$%s := .%s}}", arg.name, fieldName)
	}
	return nil
}

type scanMode int

const (
	scanModeBegin scanMode = iota
	scanModeImport
	scanModeFunc
	scanModeBody
)

type viewGenerator struct {
	viewDatas map[string]*viewData
}

func (v *viewGenerator) collectViewData(cwd, path string) (*viewData, error) {
	relativePath := strings.TrimPrefix(path, cwd+"/")
	vData := &viewData{
		args:         make(map[string]*viewArg),
		imports:      make(viewImports),
		relativePath: relativePath,
		argsContent:  &bytes.Buffer{},
		content:      &bytes.Buffer{},
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	mode := scanModeBegin
	for scanner.Scan() {
		line := scanner.Text()
		switch mode {
		case scanModeBegin:
			if !strings.HasPrefix(line, "//import") && !strings.HasPrefix(line, "//func") {
				vData.collectBody(line)
				mode = scanModeBody
			} else if strings.HasPrefix(line, "//import") {
				err = vData.collectImport(line)
				if err != nil {
					return nil, err
				}
				mode = scanModeImport
			} else {
				vData.collectArgs(line)
				mode = scanModeFunc
			}
		case scanModeImport:
			if !strings.HasPrefix(line, "//import") && !strings.HasPrefix(line, "//func") {
				return nil, &viewParseError{relativePath, "imports declared but no argument found", ""}
			} else if strings.HasPrefix(line, "//import") {
				err = vData.collectImport(line)
				if err != nil {
					return nil, err
				}
			} else {
				vData.collectArgs(line)
				mode = scanModeFunc
			}
		case scanModeFunc:
			if strings.HasPrefix(line, "//import") || strings.HasPrefix(line, "//func") {
				return nil, &viewParseError{relativePath, "import or func declared in body", ""}
			} else if strings.HasPrefix(line, "//") {
				vData.collectArgs(line)
			} else {
				err = vData.buildArgs()
				if err != nil {
					return nil, err
				}
				err = vData.writeArgsDeclare()
				if err != nil {
					return nil, err
				}
				vData.collectBody(line)
				mode = scanModeBody
			}
		case scanModeBody:
			vData.collectBody(line)
		}
	}
	err = scanner.Err()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	v.viewDatas[relativePath] = vData
	return vData, nil
}

func (v *viewGenerator) copyViewFile(r io.Reader, cwd, path, tmpFolder string) error {
	relativePath := strings.TrimPrefix(path, cwd+"/")
	dir := filepath.Dir(relativePath)
	basename := filepath.Base(relativePath)
	outputFolder := filepath.Join(tmpFolder, dir)
	err := os.MkdirAll(outputFolder, 0755)
	if err != nil {
		return errors.Wrap(err)
	}
	w, err := os.Create(filepath.Join(outputFolder, basename))
	if err != nil {
		return errors.Wrap(err)
	}
	defer w.Close()
	_, err = io.Copy(w, r)
	if err != nil {
		return errors.Wrap(err)
	}
	w.Chmod(0644)
	return nil
}

func (v *viewGenerator) packViewFolder(root, pkgName string) (err error) {
	generatedFolder := filepath.Join(generatedRoot, "views")
	err = os.MkdirAll(generatedFolder, 0755)
	if err != nil {
		return errors.Wrap(err)
	}
	defer func() {
		if err != nil {
			os.RemoveAll(generatedFolder)
		}
	}()
	generatedFile, err := os.Create(filepath.Join(generatedFolder, generatedRootFile))
	if err != nil {
		return errors.Wrap(err)
	}
	defer generatedFile.Close()
	err = packer.Generate(pkgName, root, false, true, generatedFile)
	if err != nil {
		return err
	}
	return errors.Wrap(generatedFile.Chmod(0644))
}

func (v *viewGenerator) createViewInitFile() (err error) {
	pkgName := "views"
	initFilename := filepath.Join(generatedRoot, "views.go")
	initFile, err := os.Create(initFilename)
	if err != nil {
		return errors.Wrap(err)
	}
	defer func() {
		if err != nil {
			os.RemoveAll(initFilename)
		} else {
			initFile.Close()
		}
	}()
	pkgPath, err := utils.PkgPath(initFilename)
	if err != nil {
		return err
	}

	tmplData := &struct {
		TargetPkg     string
		PkgName       string
		TargetPkgName string
	}{
		TargetPkg:     pkgPath + "/views",
		PkgName:       pkgName,
		TargetPkgName: pkgName,
	}

	tmpl, err := packer.LoadTemplate(VVVPack, "import.go.tmpl")
	if err != nil {
		return err
	}
	err = tmpl.Execute(initFile, tmplData)
	if err != nil {
		return errors.Wrap(err)
	}
	return errors.Wrap(initFile.Chmod(0644))
}

func (v *viewGenerator) generateViewCode(cwd string, vData *viewData) (err error) {
	base := strings.TrimSuffix(filepath.Base(vData.relativePath), ".html") + ".go"
	dir := filepath.Join(cwd, filepath.Dir(vData.relativePath))
	pkgName := filepath.Base(dir)
	filename := filepath.Join(dir, base)
	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err)
	}
	defer func() {
		if err != nil {
			os.RemoveAll(filename)
		} else {
			f.Close()
		}
	}()

	typeName := strings.ToLower(typeRegex.ReplaceAllString(strings.Replace(base, ".tmpl.go", "", -1), "")) + "View"
	varName := makePublicVariableName(strings.Replace(base, ".tmpl.go", "", -1))
	args := []*template.Arg{}
	declarations := []string{}
	calls := []string{}
	for _, argName := range vData.argSorted {
		arg := vData.args[argName]
		args = append(args, &template.Arg{
			Name:    arg.name,
			DotName: arg.dotName,
			Type:    arg.argType,
		})
		declarations = append(declarations, arg.name+" "+arg.argType)
		calls = append(calls, arg.name)
	}
	tmplData := &struct {
		PkgName     string
		Imports     map[string]string
		Path        string
		TypeName    string
		VarName     string
		Args        []*template.Arg
		Declaration string
		Call        string
	}{
		PkgName:     pkgName,
		Imports:     map[string]string(vData.imports),
		Path:        vData.relativePath,
		TypeName:    typeName,
		VarName:     varName,
		Args:        args,
		Declaration: strings.Join(declarations, ", "),
		Call:        strings.Join(calls, ", "),
	}

	tmpl, err := packer.LoadTemplate(VVVPack, "view.go.tmpl")
	if err != nil {
		return err
	}
	err = tmpl.Execute(f, tmplData)
	if err != nil {
		return errors.Wrap(err)
	}
	return errors.Wrap(f.Chmod(0644))
}

func (v *viewGenerator) generateViewInitCode(cwd string, vData *viewData) error {
	base := strings.TrimSuffix(fmt.Sprintf("%x.%s", md5.Sum([]byte(vData.relativePath)), strings.Replace(vData.relativePath, "/", "_", -1)), ".html") + ".go"
	filename := filepath.Join(cwd, generatedRoot, "views", base)
	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err)
	}
	defer func() {
		if err != nil {
			os.RemoveAll(filename)
		} else {
			f.Close()
		}
	}()
	pkgPath, err := utils.PkgPath(filepath.Join(cwd, vData.relativePath))
	if err != nil {
		return err
	}
	name := strings.Replace(filepath.Base(vData.relativePath), ".tmpl.html", "", -1)
	tmplData := &struct {
		PackagePath string
		Path        string
		VarName     string
	}{
		PackagePath: pkgPath,
		Path:        vData.relativePath,
		VarName:     makePublicVariableName(name),
	}

	tmpl, err := packer.LoadTemplate(VVVPack, "view_init.go.tmpl")
	if err != nil {
		return err
	}
	err = tmpl.Execute(f, tmplData)
	if err != nil {
		return errors.Wrap(err)
	}
	return f.Chmod(0644)
}

var typeRegex = regexp.MustCompile("[^a-zA-Z0-9_]")
var slugRegex = regexp.MustCompile("(\\s|-|_)+")

func makePublicVariableName(name string) string {
	return strings.Replace(strings.Title(slugRegex.ReplaceAllString(name, " ")), " ", "", -1)
}

func createMainInitFile() (err error) {
	_, err = os.Stat(generatedRootFile)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return errors.Wrap(err)
	}
	goru.Printf(goru.ColorWhite, "-> Generating main initializer\n")
	mainInitFile, err := os.Create(generatedRootFile)
	if err != nil {
		return errors.Wrap(err)
	}
	defer func() {
		if err != nil {
			os.RemoveAll(generatedRootFile)
		} else {
			mainInitFile.Close()
		}
	}()
	pkgPath, err := utils.PkgPath(generatedRootFile)
	if err != nil {
		return err
	}
	tmplData := &struct {
		PkgName       string
		TargetPkgName string
	}{
		PkgName:       "main",
		TargetPkgName: pkgPath + "/generated",
	}
	tmpl, err := packer.LoadTemplate(VVVPack, "main_init.go.tmpl")
	if err != nil {
		return err
	}
	err = tmpl.Execute(mainInitFile, tmplData)
	if err != nil {
		return errors.Wrap(err)
	}
	return errors.Wrap(mainInitFile.Chmod(0644))
}

func (v *viewGenerator) generate(args []string, options map[string]string) {
	goru.Println(goru.ColorWhite, "-> Generating views")
	cwd, err := os.Getwd()
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", errors.Wrap(err))
		os.Exit(1)
	}

	tmpFolder := filepath.Join(generatedRoot, "tmp", "views")
	err = os.MkdirAll(tmpFolder, 0755)
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", errors.Wrap(err))
		os.Exit(1)
	}
	defer func() {
		os.RemoveAll(filepath.Join(generatedRoot, "tmp"))
	}()

	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(path, filepath.Join(cwd, "generated")) {
			return nil
		}
		if !strings.HasSuffix(path, ".tmpl.html") {
			return nil
		}

		vData, err := v.collectViewData(cwd, path)
		if err != nil {
			return err
		}

		err = v.copyViewFile(vData.content, cwd, path, tmpFolder)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", err)
		os.Exit(1)
	}

	err = v.packViewFolder(tmpFolder, "views")
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", err)
		os.Exit(1)
	}

	for _, vData := range v.viewDatas {
		err = v.generateViewCode(cwd, vData)
		if err != nil {
			goru.ErrPrintln(goru.ColorRed, "-> ", err)
			os.Exit(1)
		}
		err = v.generateViewInitCode(cwd, vData)
		if err != nil {
			goru.ErrPrintln(goru.ColorRed, "-> ", err)
			os.Exit(1)
		}
	}

	err = v.createViewInitFile()
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", err)
		os.Exit(1)
	}
	err = createMainInitFile()
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", err)
		os.Exit(1)
	}
}

func ViewGenerator(args []string, options map[string]string) {
	v := &viewGenerator{
		viewDatas: make(map[string]*viewData),
	}
	v.generate(args, options)
}
