package declarative

import (
	"context"

	"github.com/agentregistry-dev/agentregistry/pkg/api/v1alpha1"
	"github.com/agentregistry-dev/agentregistry/pkg/cli/scheme"
	"github.com/agentregistry-dev/agentregistry/pkg/client"
)

// typedKind builds a scheme.Kind whose Get / List / Delete dispatch
// closures all wire through the typed v1alpha1 client helpers
// (client.GetTyped[T] / client.ListAllTyped[T] / apiClient.Delete) for
// the canonical kind. Per-kind callers supply the user-facing name +
// aliases, the table layout, and a row formatter that takes the typed
// envelope T directly. RowFunc shape-checks the input via T-assertion
// so the registry's `any` API stays internal.
func TypedKind[T v1alpha1.Object](
	cliName, plural string,
	aliases []string,
	columns []scheme.Column,
	canonicalKind string,
	newObj func() T,
	apiClient **client.Client,
	row func(T) []string,
) *scheme.Kind {
	return &scheme.Kind{
		Kind:         cliName,
		Plural:       plural,
		Aliases:      aliases,
		TableColumns: columns,
		ToYAMLFunc:   func(item any) any { return item },
		RowFunc: func(item any) []string {
			t, ok := item.(T)
			if !ok {
				return []string{"<invalid>"}
			}
			return row(t)
		},
		Get: func(ctx context.Context, name, _ string) (any, error) {
			return client.GetTyped(ctx, *apiClient, canonicalKind, v1alpha1.DefaultNamespace, name, "", newObj)
		},
		ListFunc: func(ctx context.Context) ([]any, error) {
			return ListLatestAny(ctx, *apiClient, canonicalKind, newObj)
		},
		Delete: func(ctx context.Context, name, version string, force bool) error {
			return DeleteAny(ctx, *apiClient, canonicalKind, name, version, force, newObj)
		},
	}
}

func DeleteAny[T v1alpha1.Object](ctx context.Context, apiClient *client.Client, kind, name, version string, force bool, newObj func() T) error {
	targetVersion := version
	if targetVersion == "" {
		obj, err := client.GetTyped(ctx, apiClient, kind, v1alpha1.DefaultNamespace, name, "", newObj)
		if err != nil {
			return err
		}
		targetVersion = obj.GetMetadata().Version
	}
	return apiClient.Delete(ctx, kind, v1alpha1.DefaultNamespace, name, targetVersion, client.DeleteOpts{Force: force})
}

func ListLatestAny[T v1alpha1.Object](ctx context.Context, apiClient *client.Client, kind string, newObj func() T) ([]any, error) {
	items, err := client.ListAllTyped(
		ctx,
		apiClient,
		kind,
		client.ListOpts{
			Namespace:  v1alpha1.DefaultNamespace,
			LatestOnly: true,
			Limit:      200,
		},
		newObj,
	)
	if err != nil {
		return nil, err
	}

	out := make([]any, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out, nil
}
