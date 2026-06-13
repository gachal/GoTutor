package verifier

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// CheckError is a user-facing verifier rejection. The Reason string is
// safe to return in the SubmitResult output — it explains *why* the
// code was rejected (forbidden import, listen call, syntax error) in
// language a learner can act on.
type CheckError struct {
	Reason string
}

func (e *CheckError) Error() string { return e.Reason }

// ASTCheck parses the user's source and rejects anything that violates
// the chapter policy. Runs BEFORE compilation so we fail fast on
// forbidden imports (os/exec, unsafe, etc.) without spawning `go test`.
//
// Returns nil on success; otherwise a *CheckError whose message is safe
// to surface to the learner.
func ASTCheck(userCode string, policy Policy) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", userCode, parser.AllErrors)
	if err != nil {
		return &CheckError{Reason: "syntax error: " + err.Error()}
	}

	// 1. Imports: every import path must be in the policy whitelist.
	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if !policy.IsImportAllowed(path) {
			return &CheckError{
				Reason: fmt.Sprintf(
					"import %q is not allowed in this chapter. Allowed packages are limited to a safe stdlib subset (see docs/SECURITY.md).",
					path,
				),
			}
		}
	}

	// 2. Walk call sites to reject net.Listen / http.ListenAndServe
	//    unless the chapter explicitly opts in via AllowListen.
	if !policy.AllowListen {
		var viol string
		ast.Inspect(f, func(n ast.Node) bool {
			if viol != "" {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				pkg, _ := exprName(sel.X)
				name := sel.Sel.Name
				if (pkg == "net" && name == "Listen") ||
					(pkg == "http" && (name == "ListenAndServe" || name == "ListenAndServeTLS" || name == "Serve")) {
					viol = fmt.Sprintf("%s.%s is not allowed — chapters can only make outbound HTTP requests, not open listeners", pkg, name)
					return false
				}
			}
			return true
		})
		if viol != "" {
			return &CheckError{Reason: viol}
		}
	}

	return nil
}

// exprName extracts the literal name of an *ast.Ident or the package
// portion of a nested selector. Returns ("", false) for non-identifiers.
func exprName(e ast.Expr) (string, bool) {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name, true
	case *ast.SelectorExpr:
		return exprName(v.X)
	}
	return "", false
}
