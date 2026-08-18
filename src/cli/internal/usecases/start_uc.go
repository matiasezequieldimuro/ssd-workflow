package usecases

import (
	"crypto/sha256"
	"fmt"
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
	OperationID  string
}

type StartWorkItemUseCase struct {
	workItemRepo ports.WorkItemRepository
	workflowRepo ports.WorkflowRepository
	configRepo   ports.ConfigRepository
}

func NewStartWorkItemUseCase(
	wiRepo ports.WorkItemRepository,
	wfRepo ports.WorkflowRepository,
	configRepo ports.ConfigRepository,
) *StartWorkItemUseCase {
	return &StartWorkItemUseCase{
		workItemRepo: wiRepo,
		workflowRepo: wfRepo,
		configRepo:   configRepo,
	}
}

func (uc *StartWorkItemUseCase) Execute(baseDir string, in StartWorkItemInput) (*domain.WorkItem, error) {
	if err := domain.ValidateIdentifier("work item id", in.ID); err != nil {
		return nil, err
	}
	if err := domain.ValidateActor(in.Actor); err != nil {
		return nil, err
	}
	if err := domain.ValidateOperationID(in.OperationID); err != nil {
		return nil, err
	}
	if in.Title == "" {
		return nil, fmt.Errorf("%w: title cannot be empty", domain.ErrInvalidWorkItem)
	}
	if (in.FromArtifact == "") != (in.Phase == "") {
		return nil, fmt.Errorf("%w: --from-artifact and --phase must be used together", domain.ErrInvalidExternalArtifact)
	}
	exists, err := uc.workItemRepo.WorkItemExists(baseDir, in.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		if in.OperationID != "" {
			applied, err := operationApplied(baseDir, in.ID, in.OperationID, uc.workItemRepo)
			if err != nil {
				return nil, err
			}
			if applied {
				return uc.workItemRepo.GetWorkItem(baseDir, in.ID)
			}
		}
		return nil, domain.ErrWorkItemAlreadyExists
	}

	workflowID := in.WorkflowID
	if workflowID == "" {
		config, err := uc.configRepo.GetConfig(baseDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load default workflow: %w", err)
		}
		workflowID = config.Defaults.Workflow
	}
	wf, err := uc.workflowRepo.GetWorkflow(baseDir, workflowID)
	if err != nil {
		return nil, fmt.Errorf("invalid workflow: %w", err)
	}

	var externalPath, externalHash, externalArtifactID string
	entryPhase, err := wf.EntryPhaseFor("user_prompt")
	if in.FromArtifact != "" {
		entryPhase = in.Phase
		externalArtifactID, err = wf.ExternalArtifactForEntry(entryPhase)
		if err != nil {
			return nil, err
		}
		externalPath, externalHash, err = validateExternalArtifact(in.FromArtifact)
	}
	if err != nil {
		return nil, err
	}

	phasesState := make(map[string]domain.PhaseState)

	if in.FromArtifact != "" {
		ancestors, err := wf.Ancestors(entryPhase)
		if err != nil {
			return nil, err
		}
		for _, ph := range wf.Phases {
			artPath := wf.ArtifactPathForPhase(ph.ID)
			if _, isAncestor := ancestors[ph.ID]; isAncestor {
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
			ExternalArtifact: externalArtifactReference(
				externalArtifactID,
				externalPath,
				externalHash,
			),
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

	artifactMgr := infra.NewArtifactManager()
	templateVars := map[string]string{
		"title":      in.Title,
		"id":         in.ID,
		"created_at": item.CreatedAt,
		"type":       item.Type,
	}
	artifacts, err := artifactMgr.PrepareArtifactsForPhase(baseDir, wf, entryPhase, in.ID, templateVars)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare artifacts for entry phase: %w", err)
	}

	if in.FromArtifact != "" {
		artifacts, err = artifactMgr.ImportExternalArtifact(
			wf,
			entryPhase,
			externalArtifactID,
			externalPath,
			artifacts,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to import external artifact: %w", err)
		}
	}

	events := []domain.Event{
		newOperationEvent(in.ID, "work_item.created", in.Actor, map[string]interface{}{
			"workflow": wf.ID,
			"title":    in.Title,
		}, in.OperationID),
	}
	if in.FromArtifact != "" {
		events = append(events, newOperationEvent(in.ID, "phase.bypassed_by_external_input", in.Actor, map[string]interface{}{
			"phase":             in.Phase,
			"external_artifact": externalPath,
			"sha256":            externalHash,
		}, in.OperationID))
	}

	persisted, err := commitWorkItem(baseDir, uc.workItemRepo, item, artifacts, events, in.OperationID)
	if err != nil {
		return nil, fmt.Errorf("failed to commit work item: %w", err)
	}
	return persisted, nil
}

func validateExternalArtifact(path string) (string, string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("%w: resolve path: %v", domain.ErrInvalidExternalArtifact, err)
	}
	info, err := os.Stat(absolutePath)
	if err != nil {
		return "", "", fmt.Errorf("%w: inspect %s: %v", domain.ErrInvalidExternalArtifact, absolutePath, err)
	}
	if !info.Mode().IsRegular() {
		return "", "", fmt.Errorf("%w: %s is not a regular file", domain.ErrInvalidExternalArtifact, absolutePath)
	}
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return "", "", fmt.Errorf("%w: read %s: %v", domain.ErrInvalidExternalArtifact, absolutePath, err)
	}
	hash := sha256.Sum256(content)
	return absolutePath, fmt.Sprintf("%x", hash), nil
}

func externalArtifactReference(artifactID, path, hash string) *domain.ExternalArtifactReference {
	if artifactID == "" {
		return nil
	}
	return &domain.ExternalArtifactReference{
		Artifact: artifactID,
		Path:     path,
		SHA256:   hash,
	}
}
