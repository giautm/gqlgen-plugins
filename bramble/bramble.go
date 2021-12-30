package bramble

import (
	"fmt"
	"path/filepath"

	"github.com/99designs/gqlgen/codegen"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/99designs/gqlgen/plugin"
	"github.com/vektah/gqlparser/v2/ast"
)

type BrambleConfig struct {
	Name    string
	Version string
}

type BramblePlugin struct {
	brambleCfg *BrambleConfig
	Filename   string
}

var (
	_ plugin.Plugin              = (*BramblePlugin)(nil)
	_ plugin.CodeGenerator       = (*BramblePlugin)(nil)
	_ plugin.ConfigMutator       = (*BramblePlugin)(nil)
	_ plugin.EarlySourceInjector = (*BramblePlugin)(nil)
	_ plugin.LateSourceInjector  = (*BramblePlugin)(nil)
)

func New(serviceName, version string) *BramblePlugin {
	return &BramblePlugin{
		brambleCfg: &BrambleConfig{
			Name:    serviceName,
			Version: version,
		},
		Filename: "bramble.go",
	}
}

func (*BramblePlugin) Name() string {
	return "bramble"
}

// MutateConfig mutates the configuration
func (f *BramblePlugin) MutateConfig(cfg *config.Config) error {
	builtins := config.TypeMap{
		"Service": {
			Model: config.StringList{
				"giautm.dev/gqlgen-plugins/bramble/runtime.Service",
			},
		},
	}
	for typeName, entry := range builtins {
		if cfg.Models.Exists(typeName) {
			return fmt.Errorf("%v already exists which must be reserved when Bramble is enabled", typeName)
		}
		cfg.Models[typeName] = entry
	}

	cfg.Directives["boundary"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["namespace"] = config.DirectiveConfig{SkipRuntime: true}

	return nil
}

func (*BramblePlugin) InjectSourceEarly() *ast.Source {
	return &ast.Source{
		Name:    "bramble/directives.graphql",
		BuiltIn: true,
		Input: `
directive @boundary on OBJECT | FIELD_DEFINITION
directive @namespace on OBJECT
`,
	}
}

func (f *BramblePlugin) InjectSourceLate(schema *ast.Schema) *ast.Source {
	return &ast.Source{
		Name:    "bramble/service.graphql",
		BuiltIn: true,
		Input: `
# The Service type provides the gateway a schema
# to merge into the graph and a name/version to
# reference the service with
type Service {
	name: String!
	version: String!
	schema: String!
}

extend type Query {
  # The service query is used by the gateway when
  # the service is first registered
  service: Service!
}
`,
	}
}

func (f *BramblePlugin) GenerateCode(data *codegen.Data) error {
	if data.QueryRoot != nil {
		for _, f := range data.QueryRoot.Fields {
			if f.Name == "service" {
				f.GoFieldType = codegen.GoFieldMethod
				f.GoReceiverName = "ec"
				f.GoFieldName = "__resolve__bramble__service"
				f.IsResolver = false
				f.MethodHasContext = true
				break
			}
		}
	}

	dir := data.Config.Exec.DirName
	if dir == "" && data.Config.Exec.Filename != "" {
		dir = filepath.Dir(data.Config.Exec.Filename)
	}

	return templates.Render(templates.Options{
		Data:            f.brambleCfg,
		Filename:        filepath.Join(dir, f.Filename),
		GeneratedHeader: true,
		PackageName:     data.Config.Exec.Package,
		Packages:        data.Config.Packages,
	})
}
