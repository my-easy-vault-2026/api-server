package services

import (
	"api-server/lib"
	"context"
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
