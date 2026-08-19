package usecases

import (
	"sort"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type ValidationScope string

const (
	ValidationScopeProject  ValidationScope = "project"
	ValidationScopeWorkItem ValidationScope = "work_item"
)

type ValidationSummary struct {
	Total    int `json:"total"`
	Passed   int `json:"passed"`
	Warnings int `json:"warnings"`
	Failed   int `json:"failed"`
}

type ValidationReport struct {
	Scope   ValidationScope          `json:"scope"`
	Target  string                   `json:"target"`
	Valid   bool                     `json:"valid"`
	Summary ValidationSummary        `json:"summary"`
	Checks  []domain.ValidationCheck `json:"checks"`
}

type ValidateUseCase struct {
	inspector ports.ValidationInspector
}

func NewValidateUseCase(inspector ports.ValidationInspector) *ValidateUseCase {
	return &ValidateUseCase{inspector: inspector}
}

func (uc *ValidateUseCase) Execute(baseDir, workItemID string) (*ValidationReport, error) {
	var (
		scope  ValidationScope
		target string
		checks []domain.ValidationCheck
		err    error
	)
	if workItemID == "" {
		scope = ValidationScopeProject
		target = "."
		checks, err = uc.inspector.InspectProject(baseDir)
	} else {
		if err := domain.ValidateIdentifier("work item id", workItemID); err != nil {
			return nil, err
		}
		scope = ValidationScopeWorkItem
		target = workItemID
		checks, err = uc.inspector.InspectWorkItem(baseDir, workItemID)
	}
	if err != nil {
		return nil, err
	}

	ordered := append([]domain.ValidationCheck(nil), checks...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Target != ordered[j].Target {
			return ordered[i].Target < ordered[j].Target
		}
		if ordered[i].Category != ordered[j].Category {
			return ordered[i].Category < ordered[j].Category
		}
		if ordered[i].Code != ordered[j].Code {
			return ordered[i].Code < ordered[j].Code
		}
		if ordered[i].Status != ordered[j].Status {
			return ordered[i].Status < ordered[j].Status
		}
		return ordered[i].Message < ordered[j].Message
	})

	summary := ValidationSummary{Total: len(ordered)}
	for _, check := range ordered {
		switch check.Status {
		case domain.CheckPassed:
			summary.Passed++
		case domain.CheckWarning:
			summary.Warnings++
		case domain.CheckFailed:
			summary.Failed++
		}
	}

	return &ValidationReport{
		Scope:   scope,
		Target:  target,
		Valid:   summary.Failed == 0,
		Summary: summary,
		Checks:  ordered,
	}, nil
}
