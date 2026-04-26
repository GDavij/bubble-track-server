package domain

import (
	"testing"
	"time"
)

func TestValidateRelacionamento(t *testing.T) {
	tests := []struct {
		name    string
		rel     *Relacionamento
		wantErr bool
	}{
		{
			name: "valid relationship",
			rel: &Relacionamento{
				SourcePersonID: "person-1",
				TargetPersonID: "person-2",
				Strength:       0.7,
				ReciprocityIndex: 0.5,
			},
			wantErr: false,
		},
		{
			name:    "missing source",
			rel:     &Relacionamento{TargetPersonID: "person-2"},
			wantErr: true,
		},
		{
			name:    "missing target",
			rel:     &Relacionamento{SourcePersonID: "person-1"},
			wantErr: true,
		},
		{
			name: "same source and target",
			rel: &Relacionamento{
				SourcePersonID: "person-1",
				TargetPersonID: "person-1",
			},
			wantErr: true,
		},
		{
			name: "strength too high",
			rel: &Relacionamento{
				SourcePersonID: "person-1",
				TargetPersonID: "person-2",
				Strength:       1.5,
			},
			wantErr: true,
		},
		{
			name: "strength negative",
			rel: &Relacionamento{
				SourcePersonID: "person-1",
				TargetPersonID: "person-2",
				Strength:       -0.1,
			},
			wantErr: true,
		},
		{
			name: "reciprocity too high",
			rel: &Relacionamento{
				SourcePersonID:  "person-1",
				TargetPersonID:  "person-2",
				ReciprocityIndex: 1.5,
			},
			wantErr: true,
		},
		{
			name: "boundary strength 0",
			rel: &Relacionamento{
				SourcePersonID: "person-1",
				TargetPersonID: "person-2",
				Strength:       0.0,
				ReciprocityIndex: 0.0,
			},
			wantErr: false,
		},
		{
			name: "boundary strength 1",
			rel: &Relacionamento{
				SourcePersonID: "person-1",
				TargetPersonID: "person-2",
				Strength:       1.0,
				ReciprocityIndex: 1.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRelacionamento(tt.rel)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRelacionamento() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateInteracao(t *testing.T) {
	tests := []struct {
		name    string
		inter   *Interacao
		wantErr bool
	}{
		{
			name: "valid interaction",
			inter: &Interacao{
				UserID:  "user-1",
				RawText: "Met Person1 today at the location",
			},
			wantErr: false,
		},
		{
			name:    "missing user_id",
			inter:   &Interacao{RawText: "some text"},
			wantErr: true,
		},
		{
			name:    "missing raw_text",
			inter:   &Interacao{UserID: "user-1"},
			wantErr: true,
		},
		{
			name:    "empty raw_text",
			inter:   &Interacao{UserID: "user-1", RawText: ""},
			wantErr: true,
		},
		{
			name: "text too long",
			inter: &Interacao{
				UserID:  "user-1",
				RawText: string(make([]byte, 10001)),
			},
			wantErr: true,
		},
		{
			name: "text at boundary",
			inter: &Interacao{
				UserID:  "user-1",
				RawText: string(make([]byte, 10000)),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInteracao(tt.inter)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInteracao() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNotFoundError(t *testing.T) {
	err := &NotFoundError{Entity: "person", ID: "123"}
	want := "person 123 not found"
	if err.Error() != want {
		t.Errorf("NotFoundError.Error() = %q, want %q", err.Error(), want)
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{Field: "name", Message: "required"}
	want := "validation error on name: required"
	if err.Error() != want {
		t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), want)
	}
}

func TestPessoa(t *testing.T) {
	now := time.Now().UTC()
	p := Pessoa{
		ID:          "123",
		DisplayName: "Person1",
		Aliases:     []string{"P1", "One"},
		Notes:       "Met at event",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if p.ID != "123" {
		t.Errorf("Pessoa.ID = %q, want %q", p.ID, "123")
	}
	if p.DisplayName != "Person1" {
		t.Errorf("Pessoa.DisplayName = %q, want %q", p.DisplayName, "Person1")
	}
	if len(p.Aliases) != 2 {
		t.Errorf("Pessoa.Aliases length = %d, want %d", len(p.Aliases), 2)
	}
}

func TestRelacionamento(t *testing.T) {
	r := Relacionamento{
		ID:               "rel-1",
		SourcePersonID:   "person-1",
		TargetPersonID:   "person-2",
		Quality:          QualityNourishing,
		Strength:         0.8,
		Label:            "friend",
		ReciprocityIndex: 0.7,
	}

	if r.Quality != QualityNourishing {
		t.Errorf("Relacionamento.Quality = %q, want %q", r.Quality, QualityNourishing)
	}
	if r.Label != "friend" {
		t.Errorf("Relacionamento.Label = %q, want %q", r.Label, "friend")
	}
}

func TestSocialRoleConstants(t *testing.T) {
	roles := []SocialRole{
		RoleBridge, RoleMentor, RoleAnchor, RoleCatalyst,
		RoleObserver, RoleDrain, RoleUnknown,
	}
	unique := make(map[SocialRole]bool)
	for _, r := range roles {
		if unique[r] {
			t.Errorf("duplicate SocialRole: %q", r)
		}
		unique[r] = true
	}
	if len(unique) != 7 {
		t.Errorf("expected 7 unique SocialRole constants, got %d", len(unique))
	}
}

func TestQualityConstants(t *testing.T) {
	qualities := []Quality{QualityNourishing, QualityNeutral, QualityDraining, QualityConflicted, QualityUnknown}
	unique := make(map[Quality]bool)
	for _, q := range qualities {
		if unique[q] {
			t.Errorf("duplicate Quality: %q", q)
		}
		unique[q] = true
	}
	if len(unique) != 5 {
		t.Errorf("expected 5 unique Quality constants, got %d", len(unique))
	}
}
