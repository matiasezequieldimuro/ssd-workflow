package usecases

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/infra"
	"sdd-cli/internal/ports"
)

type StartWorkItemInput struct {
	ID           string
	WorkflowID   string
	Title        string
	Summary      string
	FromArtifact string
	Phase        string
	Actor        domain.Actor
}

type StartWorkItemUseCase struct {
	workItemRepo ports.WorkItemRepository
	workflowRepo ports.WorkflowRepository
}

func NewStartWorkItemUseCase(wiRepo ports.WorkItemRepository, wfRepo ports.WorkflowRepository) *StartWorkItemUseCase {
	return &StartWorkItemUseCase{
		workItemRepo: wiRepo,
		workflowRepo: wfRepo,
	}
}

func (uc *StartWorkItemUseCase) Execute(baseDir string, in StartWorkItemInput) (*domain.WorkItem, error) {
	if uc.workItemRepo.WorkItemExists(baseDir, in.ID) {
		return nil, domain.ErrWorkItemAlreadyExists
	}

	wf, err := uc.workflowRepo.GetWorkflow(baseDir, in.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("invalid workflow: %w", err)
	}

	entryPhase := wf.Phases[0].ID
	if in.FromArtifact != "" && in.Phase != "" {
		entryPhase = in.Phase
	}

	phasesState := make(map[string]domain.PhaseState)

	if in.FromArtifact != "" {
		foundEntry := false
		for _, ph := range wf.Phases {
			artPath := wf.ArtifactPathForPhase(ph.ID)
			if ph.ID == in.Phase {
				foundEntry = true
				phasesState[ph.ID] = domain.PhaseState{
					Status:   domain.PhaseBlocked,
					Artifact: artPath,
				}
				continue
			}

			if !foundEntry {
				phasesState[ph.ID] = domain.PhaseState{
					Status:   domain.PhaseNotApplicable,
					Artifact: artPath,
				}
			} else {
				phasesState[ph.ID] = domain.PhaseState{
					Status:   domain.PhaseBlocked,
					Artifact: artPath,
				}
			}
		}
	} else {
		for _, ph := range wf.Phases {
			phasesState[ph.ID] = domain.PhaseState{
				Status:   domain.PhaseBlocked,
				Artifact: wf.ArtifactPathForPhase(ph.ID),
			}
		}
		entryState := phasesState[entryPhase]
		entryState.Status = domain.PhaseReady
		phasesState[entryPhase] = entryState
	}

	inputSource := "user_prompt"
	if in.FromArtifact != "" {
		inputSource = "external_artifact"
	}

	item := &domain.WorkItem{
		SchemaVersion: "0.1",
		Kind:          "work-item",
		ID:            in.ID,
		Title:         in.Title,
		Type:          wf.WorkItemType,
		Status:        domain.WorkItemActive,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		CreatedBy:     &in.Actor,
		Workflow: domain.WorkItemWorkflow{
			ID:         wf.ID,
			Version:    wf.SchemaVersion,
			EntryPhase: entryPhase,
		},
		Input: domain.WorkItemInput{
			Source:  inputSource,
			Summary: in.Summary,
		},
		Phases: phasesState,
		Traceability: domain.Traceability{
			Events: "events.jsonl",
		},
	}

	if in.FromArtifact == "" {
		if _, err := item.BeginPhase(wf, entryPhase); err != nil {
			return nil, fmt.Errorf("failed to begin entry phase: %w", err)
		}
	} else {
		if _, err := item.AcceptExternalPhase(wf, entryPhase); err != nil {
			return nil, fmt.Errorf("failed to accept external entry phase: %w", err)
		}
	}

	if err := uc.workItemRepo.SaveWorkItem(baseDir, item); err != nil {
		return nil, fmt.Errorf("failed to save work item: %w", err)
	}

	// Create artifacts for the entry phase
	artifactMgr := infra.NewArtifactManager()
	templateVars := map[string]string{
		"title":      in.Title,
		"id":         in.ID,
		"created_at": item.CreatedAt,
		"type":       item.Type,
	}
	if err := artifactMgr.CreateArtifactsForPhase(baseDir, wf, entryPhase, in.ID, templateVars); err != nil {
		return nil, fmt.Errorf("failed to create artifacts for entry phase: %w", err)
	}

	// Copy external artifact if provided
	if in.FromArtifact != "" {
		targetArtifactPath := filepath.Join(baseDir, ".sdd", "work-items", "active", in.ID, wf.ArtifactPathForPhase(in.Phase))
		if err := copyFile(in.FromArtifact, targetArtifactPath); err != nil {
			return nil, fmt.Errorf("failed to copy external artifact: %w", err)
		}

		bypassEvent := domain.NewEvent(in.ID, "phase_bypassed_by_external_input", in.Actor, map[string]interface{}{
			"phase":             in.Phase,
			"external_artifact": in.FromArtifact,
		})
		_ = uc.workItemRepo.AppendEvent(baseDir, in.ID, bypassEvent)
	}

	createEvent := domain.NewEvent(in.ID, "work_item.created", in.Actor, map[string]interface{}{
		"workflow": wf.ID,
		"title":    in.Title,
	})
	_ = uc.workItemRepo.AppendEvent(baseDir, in.ID, createEvent)

	return item, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
