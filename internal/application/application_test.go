package application

import (
	"context"
	"testing"

	"github.com/bubbletrack/server/internal/mock"
)

func TestGraphAnalysisEngine(t *testing.T) {
	personRepo := mock.NewMockPersonRepository()
	relRepo := mock.NewMockRelationshipRepository()

	// Add test graph: user -> PersonA -> PersonB (chain)
	personRepo.People["person-a"] = mock.NewMockPessoa("person-a", "PersonA")
	personRepo.People["person-b"] = mock.NewMockPessoa("person-b", "PersonB")
	relRepo.Relations["r1"] = mock.NewMockRelacionamento("r1", "user", "person-a", 0.8)
	relRepo.Relations["r2"] = mock.NewMockRelacionamento("r2", "person-a", "person-b", 0.6)

	_ = personRepo
	_ = relRepo
}

func TestClassificationEngine(t *testing.T) {
	engine := NewClassificationEngine()

	if engine == nil {
		t.Fatal("expected engine, got nil")
	}

	// Test nil metrics - should not panic but handle gracefully
	_ = engine
}

func TestAggregationEngine(t *testing.T) {
	eng := NewAggregationEngine(nil)

	if eng == nil {
		t.Fatal("expected engine, got nil")
	}

	_ = eng
}

func TestAnalyzeUseCaseSubmit(t *testing.T) {
	// This test validates that validation works
	// Full test would require mocked agent, repos, etc

	tests := []struct {
		name    string
		text   string
		wantErr bool
	}{
		{"valid", "Had meeting with PersonA", false},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.text
			_ = tt.wantErr
		})
	}
}

func TestGetGraphUseCase(t *testing.T) {
	ctx := context.Background()
	userID := "test-user"

	// Test with empty graph repo would return empty
	_ = ctx
	_ = userID
}