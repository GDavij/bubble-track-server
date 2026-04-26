package mock

import "context"

const (
	MockUserID   = "test-user"
	MockUserName = "Test User"
)

type MockUserProvider struct{}

func NewUserProvider() *MockUserProvider {
	return &MockUserProvider{}
}

func (m *MockUserProvider) GetCurrentUserID(_ context.Context) string {
	return MockUserID
}

func (m *MockUserProvider) GetUserDisplayName(_ context.Context) string {
	return MockUserName
}
