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
		// Bypass mode
		foundEntry := false
		for _, ph := range wf.Phases {
			if ph.ID == in.Phase {
				foundEntry = true
				artPath := fmt.Sprintf("artifacts/%s.md", ph.ID)
				phasesState[ph.ID] = domain.PhaseState{
					Status:   "accepted",
					Artifact: artPath,
				}
				continue
			}

			if !foundEntry {
				phasesState[ph.ID] = domain.PhaseState{
					Status: "not_applicable",
				}
			} else {
				// The phase immediately following entry phase becomes ready
				if len(phasesState) == countPhasesUntil(wf.Phases, in.Phase)+1 {
					phasesState[ph.ID] = domain.PhaseState{
						Status: "ready",
					}
				} else {
					phasesState[ph.ID] = domain.PhaseState{
						Status: "blocked",
					}
				}
			}
		}
	} else {
		// Standard start mode
		for i, ph := range wf.Phases {
			artPath := fmt.Sprintf("artifacts/%s.md", ph.ID)
			if i == 0 {
				phasesState[ph.ID] = domain.PhaseState{
					Status:   "in_progress",
					Artifact: artPath,
				}
			} else {
				phasesState[ph.ID] = domain.PhaseState{
					Status: "blocked",
				}
			}
		}
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
		Status:        "active",
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
		targetArtifactPath := filepath.Join(baseDir, ".sdd", "work-items", "active", in.ID, "artifacts", fmt.Sprintf("%s.md", in.Phase))
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

func countPhasesUntil(phases []domain.WorkflowPhase, phaseID string) int {
	for i, ph := range phases {
		if ph.ID == phaseID {
			return i
		}
	}
	return 0
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
