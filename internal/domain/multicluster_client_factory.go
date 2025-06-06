package domain

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DynamicMulticlusterFactory interface {
	GetCachedClientFor(ctx context.Context, serverURL string, k8sClient client.Client) (client.Client, error)
}
