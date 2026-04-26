package domain

import "fmt"

type NotFoundError struct {
	Entity string
	ID     string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s %s not found", e.Entity, e.ID)
}

type AlreadyExistsError struct {
	Entity string
	ID     string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s %s already exists", e.Entity, e.ID)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func ValidateRelacionamento(r *Relacionamento) error {
	if r.SourcePersonID == "" {
		return &ValidationError{Field: "source_person_id", Message: "required"}
	}
	if r.TargetPersonID == "" {
		return &ValidationError{Field: "target_person_id", Message: "required"}
	}
	if r.SourcePersonID == r.TargetPersonID {
		return &ValidationError{Field: "target_person_id", Message: "cannot be same as source"}
	}
	if r.Strength < 0 || r.Strength > 1 {
		return &ValidationError{Field: "strength", Message: "must be between 0.0 and 1.0"}
	}
	if r.ReciprocityIndex < 0 || r.ReciprocityIndex > 1 {
		return &ValidationError{Field: "reciprocity_index", Message: "must be between 0.0 and 1.0"}
	}
	return nil
}

func ValidateInteracao(i *Interacao) error {
	if i.UserID == "" {
		return &ValidationError{Field: "user_id", Message: "required"}
	}
	if i.RawText == "" {
		return &ValidationError{Field: "raw_text", Message: "required"}
	}
	if len(i.RawText) > 10000 {
		return &ValidationError{Field: "raw_text", Message: "max 10000 characters"}
	}
	return nil
}
