//go:build ignore

package main

import (
	"bufio"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const (
	templateModule  = "github.com/oto-lab/go-pkg-temp"
	templatePackage = "gopkgtemp"
)

type identity struct {
	module, pkg, description, author, license, repo string
}

func main() {
	if err := setup(); err != nil {
		fmt.Fprintln(os.Stderr, "Setup:", err)
		os.Exit(1)
	}
}

func setup() error {
	mod, err := os.ReadFile("go.mod")
	if err != nil {
		return fmt.Errorf("run setup from the module root: %w", err)
	}
	if lines := strings.Fields(string(mod)); len(lines) < 2 || lines[0] != "module" || lines[1] != templateModule {
		return errors.New("setup expects an unconfigured copy of the template")
	}

	answer, err := readIdentity()
	if err != nil {
		return err
	}

	edits, err := planEdits(answer)
	if err != nil {
		return err
	}
	for _, edit := range edits {
		if err := os.WriteFile(edit.path, edit.content, edit.mode.Perm()); err != nil {
			return err
		}
		if edit.newPath != edit.path {
			if err := os.Rename(edit.path, edit.newPath); err != nil {
				return err
			}
			fmt.Printf("Renamed %s -> %s\n", edit.path, edit.newPath)
		} else {
			fmt.Println("Updated", edit.path)
		}
	}
	for _, args := range [][]string{
		{"go", "mod", "edit", "-module=" + answer.module},
		{"gofmt", "-w", "."},
		{"go", "mod", "tidy"},
		{"go", "build", "./..."},
		{"go", "test", "./..."},
	} {
		fmt.Println("+", strings.Join(args, " "))
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Env = append(os.Environ(), "GOWORK=off")
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", strings.Join(args, " "), err)
		}
	}
	fmt.Printf("Configured %s (package %s). Replace the placeholder LICENSE before publishing.\n", answer.module, answer.pkg)
	return nil
}

func readIdentity() (identity, error) {
	remote, _ := exec.Command("git", "remote", "get-url", "origin").Output()
	origin := githubRepo(strings.TrimSpace(string(remote)))

	input := bufio.NewScanner(os.Stdin)
	ask := func(label, fallback string) (string, error) {
		if fallback != "" {
			fmt.Printf("%s [%s]: ", label, fallback)
		} else {
			fmt.Printf("%s: ", label)
		}
		if !input.Scan() {
			if err := input.Err(); err != nil {
				return "", err
			}
			return "", fmt.Errorf("missing answer for %s", label)
		}
		if value := strings.TrimSpace(input.Text()); value != "" {
			return value, nil
		}
		return fallback, nil
	}

	id := identity{}
	var err error
	var moduleDefault string
	if origin != "" {
		moduleDefault = "github.com/" + origin
	}
	id.module, err = ask("Module path", moduleDefault)
	if err != nil {
		return id, err
	}
	if id.module == templateModule || !validModulePath(id.module) {
		return id, fmt.Errorf("enter a new Go module path, got %q", id.module)
	}

	parts := strings.Split(id.module, "/")
	last := parts[len(parts)-1]
	id.pkg, err = ask("Package name", strings.NewReplacer("-", "", ".", "").Replace(last))
	if err != nil {
		return id, err
	}
	if !token.IsIdentifier(id.pkg) || id.pkg == "main" || strings.HasPrefix(id.pkg, "_") || strings.HasSuffix(id.pkg, "_test") {
		return id, fmt.Errorf("%q is not a usable library package name", id.pkg)
	}
	id.repo = githubRepo("https://" + id.module)
	if id.repo == "" {
		id.repo = origin
	}
	if id.repo == "" {
		return id, errors.New("configure a GitHub origin to fill in the repository links")
	}

	for _, field := range []struct {
		label string
		value *string
	}{
		{"Description", &id.description}, {"Author", &id.author}, {"License", &id.license},
	} {
		*field.value, err = ask(field.label, "")
		if err != nil {
			return id, err
		}
		if *field.value == "" {
			return id, fmt.Errorf("%s cannot be empty", field.label)
		}
	}
	return id, nil
}

var pathElement = regexp.MustCompile("^[A-Za-z0-9_-][A-Za-z0-9._~-]*$")
var repoParts = regexp.MustCompile("^[A-Za-z0-9-]+/[A-Za-z0-9_.-]+$")
var majorSuffix = regexp.MustCompile("^v([2-9]|[1-9][0-9]+)$")
var versionLikeSuffix = regexp.MustCompile("^v[0-9.]+$")

func validModulePath(path string) bool {
	host, rest, ok := strings.Cut(path, "/")
	if !ok || !strings.Contains(host, ".") || strings.ToLower(host) != host || strings.ContainsAny(host, "_~") {
		return false
	}
	for part := range strings.SplitSeq(path, "/") {
		if !pathElement.MatchString(part) || strings.HasSuffix(part, ".") {
			return false
		}
	}
	last := filepath.Base(rest)
	return !versionLikeSuffix.MatchString(last) || majorSuffix.MatchString(last)
}

func githubRepo(remote string) string {
	if scpPath, ok := strings.CutPrefix(remote, "git@github.com:"); ok {
		remote = "ssh://git@github.com/" + scpPath
	}
	u, err := url.Parse(remote)
	if err != nil || !strings.EqualFold(u.Hostname(), "github.com") {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 && !(len(parts) == 3 && majorSuffix.MatchString(parts[2])) {
		return ""
	}
	repo := parts[0] + "/" + strings.TrimSuffix(parts[1], ".git")
	if !repoParts.MatchString(repo) {
		return ""
	}
	return repo
}

type fileEdit struct {
	path, newPath string
	content       []byte
	mode          fs.FileMode
}

func planEdits(id identity) ([]fileEdit, error) {
	text := strings.NewReplacer(
		"<package-name>", id.pkg,
		"<package-description>", id.description,
		"<package-repo>", "https://github.com/"+id.repo,
		"<package-git>", "https://github.com/"+id.repo+".git",
		"<package-license>", id.license,
		"<package-author>", id.author,
		"<module-path>", id.module,
		"<repository-slug>", id.repo,
	)
	var edits []fileEdit
	err := fs.WalkDir(os.DirFS("."), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != "." && path != ".github" && (strings.HasPrefix(entry.Name(), ".") || slices.Contains([]string{"vendor", "testdata"}, entry.Name())) {
				return fs.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || path == "scripts/setup.go" {
			return nil
		}
		isGo := strings.HasSuffix(path, ".go")
		isTemplateText := path == "README.md" || path == "CONTRIBUTING.md" || strings.HasPrefix(path, ".github/")
		if !isGo && !isTemplateText {
			return nil
		}

		original, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(original)
		newPath := path
		if isGo {
			content, err = rewriteGo(path, original, id)
			if err != nil {
				return err
			}
			if path == templatePackage+".go" || path == templatePackage+"_test.go" {
				newPath = id.pkg + strings.TrimPrefix(path, templatePackage)
				if newPath != path {
					if _, err := os.Lstat(newPath); !errors.Is(err, fs.ErrNotExist) {
						return fmt.Errorf("cannot rename %s to %s: destination is occupied or inaccessible", path, newPath)
					}
				}
			}
		} else {
			content = text.Replace(content)
		}
		if content != string(original) || newPath != path {
			info, err := entry.Info()
			if err != nil {
				return err
			}
			edits = append(edits, fileEdit{path, newPath, []byte(content), info.Mode()})
		}
		return nil
	})
	return edits, err
}

func rewriteGo(path string, source []byte, id identity) (string, error) {
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, path, source, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return "", err
	}
	offset := files.File(file.Pos()).Offset
	type replacement struct {
		start, end int
		text       string
	}
	var changes []replacement
	replace := func(node ast.Node, text string) {
		changes = append(changes, replacement{offset(node.Pos()), offset(node.End()), text})
	}

	if file.Name.Name == templatePackage || file.Name.Name == templatePackage+"_test" {
		replace(file.Name, id.pkg+strings.TrimPrefix(file.Name.Name, templatePackage))
	}
	if file.Doc != nil {
		for _, comment := range file.Doc.List {
			const prefix = "// Package " + templatePackage + " "
			if strings.HasPrefix(comment.Text, prefix) {
				replace(comment, "// Package "+id.pkg+" "+strings.TrimPrefix(comment.Text, prefix))
			}
		}
	}

	renameQualifier := false
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return "", err
		}
		if path == templateModule || strings.HasPrefix(path, templateModule+"/") {
			replace(spec.Path, strconv.Quote(id.module+strings.TrimPrefix(path, templateModule)))
		}
		if path == templateModule && (spec.Name == nil || spec.Name.Name == templatePackage) {
			renameQualifier = true
			if spec.Name != nil {
				replace(spec.Name, id.pkg)
			}
		}
	}
	if renameQualifier {
		ast.Inspect(file, func(node ast.Node) bool {
			if selector, ok := node.(*ast.SelectorExpr); ok {
				if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == templatePackage {
					replace(ident, id.pkg)
				}
			}
			return true
		})
	}

	// Apply edits from the end so earlier source offsets stay valid.
	slices.SortFunc(changes, func(a, b replacement) int { return b.start - a.start })
	result := string(source)
	for _, change := range changes {
		result = result[:change.start] + change.text + result[change.end:]
	}
	return result, nil
}
