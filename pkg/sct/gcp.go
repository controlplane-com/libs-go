package sct

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"context"
	"fmt"
	"github.com/googleapis/gax-go/v2/apierror"
	"google.golang.org/grpc/codes"
)

type GcpTokenRepository[T any] struct {
	projectId string
	template  *secretmanagerpb.Secret
	client    *secretmanager.Client
}

func NewGcpTokenManager[T any](projectId string, client *secretmanager.Client, secretTemplate *secretmanagerpb.Secret) *TokenManager[T] {
	if secretTemplate == nil {
		secretTemplate = &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		}
	}
	tokenReader := GcpTokenRepository[T]{
		client:    client,
		projectId: projectId,
		template:  secretTemplate,
	}
	return NewTokenManager[T]().WithRepository(&tokenReader).WithFormatter(JsonFormatter[T]{})
}

func (g *GcpTokenRepository[T]) Write(ctx context.Context, t *Token[T], f TokenFormatter[T]) error {
	b, err := f.Marshal(t)
	if err != nil {
		return err
	}
	secretName := g.GetSecretName(t)
	_, err = g.client.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent:  secretName,
		Payload: &secretmanagerpb.SecretPayload{Data: b},
	})
	if g.isNotFound(err) {
		err = g.createSecret(ctx, t)
		if err != nil {
			return err
		}
		_, err = g.client.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
			Parent:  secretName,
			Payload: &secretmanagerpb.SecretPayload{Data: b},
		})
	}
	return err
}

func (g *GcpTokenRepository[T]) Read(ctx context.Context, t *Token[T], f TokenFormatter[T]) error {
	s, err := g.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: g.GetSecretLatestSecretVersionName(t),
	})
	if err != nil {
		if g.isNotFound(err) {
			return TokenNotFound
		}
		return err
	}
	err = f.Unmarshal(s.Payload.Data, t)
	return err
}

func (g *GcpTokenRepository[T]) Delete(ctx context.Context, t *Token[T]) error {
	return g.client.DeleteSecret(ctx, &secretmanagerpb.DeleteSecretRequest{
		Name: g.GetSecretName(t),
	})
}

func (g *GcpTokenRepository[T]) Exists(ctx context.Context, t *Token[T]) (bool, error) {
	s, err := g.client.GetSecret(ctx, &secretmanagerpb.GetSecretRequest{Name: g.GetSecretName(t)})
	if apiErr, ok := err.(*apierror.APIError); ok {
		if apiErr.GRPCStatus().Code() == codes.NotFound {
			return false, nil
		}
	}
	if err != nil {
		return false, err
	}
	return s != nil, nil
}

func (g *GcpTokenRepository[T]) Close() error {
	return nil
}

func (g *GcpTokenRepository[T]) GetSecretName(t *Token[T]) string {
	return fmt.Sprintf("projects/%s/secrets/sct-%s", g.projectId, t.Id)
}

func (g *GcpTokenRepository[T]) GetProjectPath() string {
	return fmt.Sprintf("project/%s", g.projectId)
}

func (g *GcpTokenRepository[T]) GetSecretId(t *Token[T]) string {
	return fmt.Sprintf("sct-%s", t.Id)
}

func (g *GcpTokenRepository[T]) GetSecretLatestSecretVersionName(t *Token[T]) string {
	return fmt.Sprintf("projects/%s/secrets/sct-%s/versions/latest", g.projectId, t.Id)
}

func (g *GcpTokenRepository[T]) createSecret(ctx context.Context, t *Token[T]) error {
	_, err := g.client.CreateSecret(ctx, &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", g.projectId),
		SecretId: g.GetSecretId(t),
		Secret:   g.getSecretFromTemplate(),
	})
	return err
}

func (g *GcpTokenRepository[T]) getSecretFromTemplate() *secretmanagerpb.Secret {
	if g.template == nil {
		g.template = &secretmanagerpb.Secret{}
	}
	//goland:noinspection GoVetCopyLock
	secret := *g.template
	if secret.Replication == nil {
		secret.Replication = &secretmanagerpb.Replication{
			Replication: &secretmanagerpb.Replication_Automatic_{
				Automatic: &secretmanagerpb.Replication_Automatic{},
			},
		}
	}
	return &secret
}

func (g *GcpTokenRepository[T]) isNotFound(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*apierror.APIError); ok {
		if apiErr.GRPCStatus().Code() == codes.NotFound {
			return true
		}
	}
	return false
}
