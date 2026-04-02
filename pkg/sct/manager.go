package sct

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const tokenPrefix = "sct"
const minSecretByteLength = 32
const minIdByteLength = 16

type TokenManager[T any] struct {
	repository TokenRepository[T]
	formatter  TokenFormatter[T]
}

var SecretChallengeFailure = errors.New("Invalid Secret")
var TokenNotFound = errors.New("Secret Not Found")
var NoRepositoryError = errors.New("no TokenRepository assigned. Call WithRepository to assign a TokenRepository")

func NewTokenManager[T any]() *TokenManager[T] {
	return &TokenManager[T]{}
}

func (tm *TokenManager[T]) NewToken(ref string, challengeSecretName string) (*Token[T], error) {
	ctx := context.Background()
	p, err := tm.parseRef(ref)
	if err != nil {
		return nil, err
	}
	t := &Token[T]{
		Id: p[1],
	}
	err = tm.Read(ctx, t)
	if err == TokenNotFound {
		t.Secrets = make(map[string]string)
		t.Secrets[challengeSecretName] = p[2]
		return t, nil
	}
	if err != nil {
		return nil, err
	}
	if err := tm.ValidateTokenRef(p[2], challengeSecretName, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (tm *TokenManager[T]) ValidateTokenRef(refSecret string, challengeSecretName string, t *Token[T]) error {
	challengeSecret := t.GetSecret(challengeSecretName)
	if challengeSecret != refSecret {
		return SecretChallengeFailure
	}
	return nil
}

func (tm *TokenManager[T]) parseRef(ref string) ([]string, error) {
	pieces := strings.Split(ref, ".")
	if len(pieces) != 3 {
		return nil, errors.New("invalid token reference. Token references should be of the form: sct.{Id}.{Secret}")
	}
	if pieces[0] != tokenPrefix {
		return nil, errors.New(fmt.Sprintf("invalid token reference. Token references must begin with \"%s\", and be of the form %s.{Id}.{Secret}", tokenPrefix, tokenPrefix))
	}
	if len([]byte(pieces[2])) < minIdByteLength {
		return nil, errors.New(fmt.Sprintf("invalid token reference. Token ids must be at least %d bytes in length", minSecretByteLength))
	}
	if len([]byte(pieces[2])) < minSecretByteLength {
		return nil, errors.New(fmt.Sprintf("invalid token reference. Token secrets must be at least %d bytes in length", minSecretByteLength))
	}
	return pieces, nil
}
func (tm *TokenManager[T]) WithRepository(writer TokenRepository[T]) *TokenManager[T] {
	tm.repository = writer
	return tm
}

func (tm *TokenManager[T]) WithFormatter(formatter TokenFormatter[T]) *TokenManager[T] {
	tm.formatter = formatter
	return tm
}

func (tm *TokenManager[T]) Close() error {
	if tm.repository == nil {
		return NoRepositoryError
	}
	return tm.repository.Close()
}

// Read reads the token data into this instance using the configured Tokenrepository
func (tm *TokenManager[T]) Read(ctx context.Context, t *Token[T]) error {
	if tm.repository == nil {
		return NoRepositoryError
	}

	return tm.repository.Read(ctx, t, tm.formatter)
}

// Write writes the token data in this instance using the configured TokenRepository
func (tm *TokenManager[T]) Write(ctx context.Context, t *Token[T]) error {
	if tm.repository == nil {
		return NoRepositoryError
	}
	err := t.Validate()
	if err != nil {
		return err
	}
	return tm.repository.Write(ctx, t, tm.formatter)
}

func (tm *TokenManager[T]) Delete(ctx context.Context, t *Token[T]) error {
	if tm.repository == nil {
		return NoRepositoryError
	}
	return tm.repository.Delete(ctx, t)
}

func (tm *TokenManager[T]) Exists(ctx context.Context, t *Token[T]) (bool, error) {
	if tm.repository == nil {
		return false, NoRepositoryError
	}
	return tm.repository.Exists(ctx, t)
}

func (tm *TokenManager[T]) Bytes(t *Token[T]) ([]byte, error) {
	if tm.formatter == nil {
		return nil, errors.New("no TokenFormatter assigned. Call WithReader to assign a TokenFormatter")
	}
	b, err := tm.formatter.Marshal(t)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (tm *TokenManager[T]) String(t *Token[T]) (string, error) {
	b, err := tm.Bytes(t)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
