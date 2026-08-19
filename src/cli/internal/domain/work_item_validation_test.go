package domain_test

import (
	"errors"
	"testing"

	"sdd-cli/internal/domain"
)

func TestWorkItemValidationAccumulatesViolations(t *testing.T) {
	workflow := testWorkflow(
		[]domain.WorkflowPhase{
			{ID: "first", Produces: []string{"first"}, Approval: domain.ApprovalRequired},
			{ID: "second", Requires: []string{"first"}, Produces: []string{"second"}, Approval: domain.ApprovalNone},
		},
		[]domain.EntryPoint{{Phase: "first", Accepts: []string{"user_prompt", "first"}}},
	)
	item := &domain.WorkItem{
		Workflow: domain.WorkItemWorkflow{ID: "other", Version: "9", EntryPhase: "missing"},
		Type:     "other",
		Status:   domain.WorkItemCompleted,
		Input:    domain.WorkItemInput{Source: "user_prompt"},
		Phases: map[string]domain.PhaseState{
			"first":   {Status: domain.PhaseAwaitingApproval, Artifact: "artifacts/wrong.md"},
			"unknown": {Status: domain.PhaseReady},
		},
	}

	violations := item.ViolationsAgainst(workflow)
	if len(violations) < 5 {
		t.Fatalf("ViolationsAgainst() = %#v", violations)
	}
	if err := item.ValidateAgainst(workflow); !errors.Is(err, domain.ErrInvalidWorkItem) {
		t.Fatalf("ValidateAgainst() error = %v, want %v", err, domain.ErrInvalidWorkItem)
	}
}

func TestWorkItemValidationAcceptsRejectedApprovalDuringRework(t *testing.T) {
	workflow := testWorkflow(
		[]domain.WorkflowPhase{{ID: "first", Produces: []string{"first"}, Approval: domain.ApprovalRequired}},
		[]domain.EntryPoint{{Phase: "first", Accepts: []string{"user_prompt", "first"}}},
	)
	actor := domain.Actor{Kind: domain.ActorHuman, ID: "reviewer"}
	item := &domain.WorkItem{
		Workflow: domain.WorkItemWorkflow{ID: workflow.ID, Version: workflow.SchemaVersion, EntryPhase: "first"},
		Type:     workflow.WorkItemType,
		Status:   domain.WorkItemActive,
		Input:    domain.WorkItemInput{Source: "user_prompt"},
		Phases: map[string]domain.PhaseState{
			"first": {Status: domain.PhaseInProgress, Artifact: "artifacts/first.md"},
		},
		Approvals: []domain.Approval{{
			Phase: "first", Status: domain.ApprovalRejected, By: &actor, At: "2026-08-19T00:00:00Z",
		}},
	}
	if err := item.ValidateAgainst(workflow); err != nil {
		t.Fatalf("ValidateAgainst() error = %v", err)
	}
}
