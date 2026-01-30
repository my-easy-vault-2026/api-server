package services

import (
	"context"
)

type TestService struct {
}

func NewTestService() *TestService {

	return &TestService{}
}

func (cs *TestService) ForTest(ctx context.Context) error {

	return nil
}
