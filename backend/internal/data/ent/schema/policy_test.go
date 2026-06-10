package schema

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
	"testing"

	"entgo.io/ent"
)

// protected is the explicit list of schemas wired into the authorization
// layer. TestEverySchemaIsProtected cross-checks it against the structs that
// actually embed ent.Schema in this package, so adding a new entity without
// adding it here (and giving it a Policy and Interceptors) fails CI.
var protected = map[string]ent.Interface{
	"AccessGrant":          AccessGrant{},
	"APIKey":               APIKey{},
	"Attachment":           Attachment{},
	"AuthRoles":            AuthRoles{},
	"AuthTokens":           AuthTokens{},
	"Entity":               Entity{},
	"EntityField":          EntityField{},
	"EntityTemplate":       EntityTemplate{},
	"EntityType":           EntityType{},
	"Export":               Export{},
	"Group":                Group{},
	"GroupInvitationToken": GroupInvitationToken{},
	"MaintenanceEntry":     MaintenanceEntry{},
	"Notifier":             Notifier{},
	"PasswordResetTokens":  PasswordResetTokens{},
	"PermissionGroup":      PermissionGroup{},
	"Tag":                  Tag{},
	"TemplateField":        TemplateField{},
	"User":                 User{},
	"UserGroup":            UserGroup{},
}

type policyHolder interface{ Policy() ent.Policy }
type interceptorHolder interface{ Interceptors() []ent.Interceptor }

func TestEverySchemaIsProtected(t *testing.T) {
	declared := schemaStructNames(t)

	for name := range declared {
		s, ok := protected[name]
		if !ok {
			t.Errorf("schema %s is not in the protected list: give it a Policy() and Interceptors() and add it to policy_test.go", name)
			continue
		}
		if _, ok := s.(policyHolder); !ok {
			t.Errorf("schema %s has no Policy(): every schema must default-deny via authzrules.NewPolicy", name)
		}
		if _, ok := s.(interceptorHolder); !ok {
			t.Errorf("schema %s has no Interceptors(): every schema must filter reads", name)
		}
	}

	for name := range protected {
		if _, ok := declared[name]; !ok {
			t.Errorf("protected list contains %s but no such schema struct exists", name)
		}
	}
}

// schemaStructNames parses this package and returns the names of all struct
// types that embed ent.Schema (i.e. real entity schemas, not mixins).
func schemaStructNames(t *testing.T) map[string]struct{} {
	t.Helper()

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parsing schema package: %v", err)
	}

	names := make(map[string]struct{})
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				ts, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					return true
				}
				for _, f := range st.Fields.List {
					if len(f.Names) != 0 {
						continue // named field, not an embed
					}
					if sel, ok := f.Type.(*ast.SelectorExpr); ok {
						if x, ok := sel.X.(*ast.Ident); ok && x.Name == "ent" && sel.Sel.Name == "Schema" {
							names[ts.Name.Name] = struct{}{}
						}
					}
				}
				return true
			})
		}
	}
	if len(names) == 0 {
		t.Fatal("found no schema structs; parser broken?")
	}
	return names
}
