package domain

import "errors"

var (
	ErrWorkItemNotFound         = errors.New("work item not found")
	ErrWorkItemAlreadyExists    = errors.New("work item already exists")
	ErrWorkflowNotFound         = errors.New("workflow not found")
	ErrPhaseNotFound            = errors.New("phase not found in workflow or work item")
	ErrPhaseNotAwaitingApproval = errors.New("phase is not awaiting approval")
	ErrPhaseBlocked             = errors.New("phase is blocked and cannot be started")
	ErrInvalidTransition        = errors.New("invalid phase transition")
	ErrHumanActorRequired       = errors.New("a human actor is required")
	ErrApprovalNotAllowed       = errors.New("phase does not allow approval")
	ErrWorkItemCannotComplete   = errors.New("work item cannot be completed")
)
