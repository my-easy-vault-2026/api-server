package services

import (
	"context"

	"github.com/my-easy-vault-2026/api-server/lib"
)

type TestService struct {
	logger lib.Logger
}

func NewTestService(logger lib.Logger) *TestService {

	return &TestService{logger: logger}
}

func (cs *TestService) ForTest(ctx context.Context) error {

	return nil
}
