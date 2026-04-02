package sct

import (
	"context"
	"errors"
	"fmt"
)

type TokenRepository[T any] interface {
	Exists(ctx context.Context, t *Token[T]) (bool, error)
	Read(ctx context.Context, t *Token[T], f TokenFormatter[T]) error
	//TODO: Figure out whether we need a Query method
	//TODO: If we do need one, figure out if this is a sensible signature
	//Query(ctx context.Context, labelFilters map[string]string, f TokenFormatter[T]) ([]*Token[T], error)

	Write(ctx context.Context, t *Token[T], f TokenFormatter[T]) error
	Delete(ctx context.Context, t *Token[T]) error
	Close() error
}

type TokenFormatter[T any] interface {
	Marshal(t *Token[T]) ([]byte, error)
	Unmarshal(b []byte, t *Token[T]) error
}

type DefaultTokenPayload map[string]string

type Token[T any] struct {
	Content T                 `json:"content"`
	Name    string            `json:"name"`
	Id      string            `json:"id"`
	Scopes  []string          `json:"scopes"`
	Secrets map[string]string `json:"secrets"`
	Labels  map[string]string `json:"labels"`
}

type TokenQuery struct {
	Name               string
	ScopeFilters       []string
	SecretNameFilters  []string
	SecretValueFilters []string
}

func (t *Token[T]) GetSecret(secretName string) string {
	if t.Secrets == nil {
		return ""
	}
	if s, ok := t.Secrets[secretName]; !ok {
		return ""
	} else {
		return s
	}
}

func (t *Token[T]) Validate() error {
	if t.Name == "" {
		return errors.New("Name must not be empty")
	}
	if t.Id == "" {
		return errors.New("Id must not be empty")
	}
	if len(t.Id) < minIdByteLength {
		return errors.New(fmt.Sprintf("Id is too short. The minimum length is %d bytes", minIdByteLength))
	}
	if t.Secrets == nil {
		return errors.New("Secrets must not be nil")
	}
	for secretName, secret := range t.Secrets {
		if len([]byte(secret)) < minSecretByteLength {
			return errors.New(fmt.Sprintf("secret %s is too short. The minimum length is %d bytes", secretName, minSecretByteLength))
		}
	}
	return nil
}
