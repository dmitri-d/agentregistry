package declarative

import (
	"context"

	cliCommon "github.com/agentregistry-dev/agentregistry/internal/cli/common"
	"github.com/agentregistry-dev/agentregistry/pkg/api/v1alpha1"
	"github.com/agentregistry-dev/agentregistry/pkg/cli/declarative"
	"github.com/agentregistry-dev/agentregistry/pkg/cli/scheme"
	"github.com/agentregistry-dev/agentregistry/pkg/client"
)

var apiClient *client.Client

// SetAPIClient sets the API client used by all declarative commands.
// Called by pkg/cli/root.go's PersistentPreRunE.
func SetAPIClient(c *client.Client) {
	apiClient = c
}

func init() {
	scheme.Register(declarative.TypedKind(
		"agent", "agents", []string{"Agent"},
		[]scheme.Column{
			{Header: "NAME"}, {Header: "VERSION"}, {Header: "FRAMEWORK"},
			{Header: "LANGUAGE"}, {Header: "PROVIDER"}, {Header: "MODEL"},
		},
		v1alpha1.KindAgent,
		func() *v1alpha1.Agent { return &v1alpha1.Agent{} },
		&apiClient,
		agentRow,
	))

	scheme.Register(declarative.TypedKind(
		"mcp", "mcps", []string{"MCPServer", "mcpserver", "mcp-server", "mcpservers"},
		[]scheme.Column{{Header: "NAME"}, {Header: "VERSION"}, {Header: "DESCRIPTION"}},
		v1alpha1.KindMCPServer,
		func() *v1alpha1.MCPServer { return &v1alpha1.MCPServer{} },
		&apiClient,
		mcpRow,
	))

	scheme.Register(declarative.TypedKind(
		"skill", "skills", []string{"Skill"},
		[]scheme.Column{
			{Header: "NAME"}, {Header: "VERSION"}, {Header: "DESCRIPTION"},
		},
		v1alpha1.KindSkill,
		func() *v1alpha1.Skill { return &v1alpha1.Skill{} },
		&apiClient,
		skillRow,
	))

	scheme.Register(declarative.TypedKind(
		"prompt", "prompts", []string{"Prompt"},
		[]scheme.Column{{Header: "NAME"}, {Header: "VERSION"}, {Header: "DESCRIPTION"}},
		v1alpha1.KindPrompt,
		func() *v1alpha1.Prompt { return &v1alpha1.Prompt{} },
		&apiClient,
		promptRow,
	))

	scheme.Register(declarative.TypedKind(
		"provider", "providers", []string{"Provider"},
		[]scheme.Column{{Header: "NAME"}, {Header: "PLATFORM"}},
		v1alpha1.KindProvider,
		func() *v1alpha1.Provider { return &v1alpha1.Provider{} },
		&apiClient,
		providerRow,
	))

	scheme.Register(declarative.TypedKind(
		"remote-mcp", "remote-mcps", []string{
			"RemoteMCPServer", "remotemcpserver", "remote-mcp-server", "remotemcpservers",
		},
		[]scheme.Column{{Header: "NAME"}, {Header: "VERSION"}, {Header: "TYPE"}, {Header: "URL"}},
		v1alpha1.KindRemoteMCPServer,
		func() *v1alpha1.RemoteMCPServer { return &v1alpha1.RemoteMCPServer{} },
		&apiClient,
		remoteMCPServerRow,
	))

	// Deployment is registered manually because its Get/Delete dispatch
	// does NOT key on the v1alpha1 metadata identity (namespace/name/
	// version). Users address deployments by the underlying target's name
	// — `arctl get deployment <agent-or-mcp-name>` — and the CLI walks the
	// /v0/deployments listing to find the matching row. The typed
	// helper assumes (kind, namespace, name, version) lookup, which is
	// the wrong shape for this dispatch.
	scheme.Register(&scheme.Kind{
		Kind:    "deployment",
		Plural:  "deployments",
		Aliases: []string{"Deployment"},
		Get: func(_ context.Context, name, _ string) (any, error) {
			return getDeploymentByTarget(context.Background(), name)
		},
		Delete: func(_ context.Context, name, version string, force bool) error {
			return deleteDeploymentByTarget(context.Background(), name, version, force)
		},
		ListFunc: func(_ context.Context) ([]any, error) {
			return listDeploymentAny(context.Background())
		},
		RowFunc: func(item any) []string {
			deployment, ok := item.(*cliCommon.DeploymentRecord)
			if !ok {
				return []string{"<invalid>"}
			}
			return deploymentRow(deployment)
		},
		ToYAMLFunc: func(item any) any {
			deployment, ok := item.(*cliCommon.DeploymentRecord)
			if !ok {
				return nil
			}
			return deploymentToDocument(deployment)
		},
		TableColumns: []scheme.Column{
			{Header: "ID"}, {Header: "NAME"}, {Header: "VERSION"},
			{Header: "TYPE"}, {Header: "PROVIDER"}, {Header: "STATUS"},
		},
	})
}
