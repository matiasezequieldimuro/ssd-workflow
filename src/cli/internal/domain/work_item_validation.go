package domain

import "fmt"

func (item *WorkItem) ValidateAgainst(workflow *Workflow) error {
	if item.Workflow.ID != workflow.ID {
		return fmt.Errorf("%w: manifest workflow %s does not match %s", ErrInvalidWorkItem, item.Workflow.ID, workflow.ID)
	}
	if item.Workflow.Version != workflow.SchemaVersion {
		return fmt.Errorf("%w: workflow version %s does not match %s", ErrInvalidWorkItem, item.Workflow.Version, workflow.SchemaVersion)
	}
	if item.Type != workflow.WorkItemType {
		return fmt.Errorf("%w: work item type %s does not match workflow type %s", ErrInvalidWorkItem, item.Type, workflow.WorkItemType)
	}
	if !workflow.hasEntryPhase(item.Workflow.EntryPhase) {
		return fmt.Errorf("%w: %s is not a declared entry point", ErrInvalidEntryPoint, item.Workflow.EntryPhase)
	}
	if item.Input.Source == "external_artifact" {
		if item.Input.ExternalArtifact == nil {
			return fmt.Errorf("%w: external artifact metadata is required", ErrInvalidExternalArtifact)
		}
		acceptedArtifact, err := workflow.ExternalArtifactForEntry(item.Workflow.EntryPhase)
		if err != nil {
			return err
		}
		if item.Input.ExternalArtifact.Artifact != acceptedArtifact {
			return fmt.Errorf(
				"%w: entry phase %s accepts %s, got %s",
				ErrInvalidExternalArtifact,
				item.Workflow.EntryPhase,
				acceptedArtifact,
				item.Input.ExternalArtifact.Artifact,
			)
		}
	} else if item.Input.ExternalArtifact != nil {
		return fmt.Errorf("%w: external artifact metadata is only valid for external input", ErrInvalidExternalArtifact)
	}
	if len(item.Phases) != len(workflow.Phases) {
		return fmt.Errorf("%w: manifest has %d phases, workflow requires %d", ErrInvalidWorkItem, len(item.Phases), len(workflow.Phases))
	}

	ancestors, err := workflow.Ancestors(item.Workflow.EntryPhase)
	if err != nil {
		return err
	}
	for _, phase := range workflow.Phases {
		state, exists := item.Phases[phase.ID]
		if !exists {
			return fmt.Errorf("%w: phase %s has no state", ErrInvalidWorkItem, phase.ID)
		}
		if state.Artifact != "" && state.Artifact != workflow.ArtifactPathForPhase(phase.ID) {
			return fmt.Errorf("%w: phase %s artifact path %q does not match workflow", ErrInvalidWorkItem, phase.ID, state.Artifact)
		}

		if state.Status == PhaseNotApplicable {
			if _, allowed := ancestors[phase.ID]; !allowed {
				return fmt.Errorf("%w: phase %s is not an ancestor of entry phase %s", ErrInvalidWorkItem, phase.ID, item.Workflow.EntryPhase)
			}
			continue
		}

		dependenciesSatisfied := item.dependenciesSatisfied(phase)
		isEntry := phase.ID == item.Workflow.EntryPhase
		if state.Status == PhaseBlocked {
			if dependenciesSatisfied && !isEntry && !phase.Optional {
				return fmt.Errorf("%w: phase %s is blocked despite satisfied dependencies", ErrInvalidWorkItem, phase.ID)
			}
			continue
		}
		if !isEntry && !dependenciesSatisfied {
			return fmt.Errorf("%w: phase %s is %s before its dependencies are satisfied", ErrInvalidWorkItem, phase.ID, state.Status)
		}
		if (state.Status == PhaseAwaitingApproval || state.Status == PhaseApproved || state.Status == PhaseRejected) &&
			phase.Approval == ApprovalNone {
			return fmt.Errorf("%w: phase %s uses approval state with policy none", ErrInvalidWorkItem, phase.ID)
		}
		if state.Status == PhaseAccepted && (isEntry == false || item.Input.Source == "user_prompt") {
			return fmt.Errorf("%w: phase %s cannot be accepted for input source %s", ErrInvalidWorkItem, phase.ID, item.Input.Source)
		}
	}
	for phaseID := range item.Phases {
		if _, exists := workflow.Phase(phaseID); !exists {
			return fmt.Errorf("%w: manifest contains unknown phase %s", ErrInvalidWorkItem, phaseID)
		}
	}

	if item.Status == WorkItemCompleted {
		if err := item.validateCompletion(workflow); err != nil {
			return err
		}
	}

	return nil
}

func (item *WorkItem) validateCompletion(workflow *Workflow) error {
	for _, phase := range workflow.Phases {
		state := item.Phases[phase.ID]
		if phase.Optional && (state.Status == PhaseBlocked || state.Status == PhaseReady || state.Status == PhaseNotApplicable) {
			continue
		}
		if !state.Status.satisfiesCompletion() {
			return fmt.Errorf("%w: completed work item has unsatisfied phase %s in %s", ErrInvalidWorkItem, phase.ID, state.Status)
		}
	}
	return nil
}

func (workflow *Workflow) hasEntryPhase(phaseID string) bool {
	for _, entry := range workflow.EntryPoints {
		if entry.Phase == phaseID {
			return true
		}
	}
	return false
}
