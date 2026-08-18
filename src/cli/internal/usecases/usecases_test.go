package usecases_test

import (
	"os"
	"path/filepath"
	"testing"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/infra"
	"sdd-cli/internal/usecases"
)

func setupTestEnv(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "sdd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	initUC := usecases.NewInitUseCase()
	if err := initUC.Execute(tmpDir); err != nil {
		t.Fatalf("InitUseCase failed: %v", err)
	}

	return tmpDir
}

func TestFullWorkItemLifecycle(t *testing.T) {
	tmpDir := setupTestEnv(t)
	defer os.RemoveAll(tmpDir)

	wiRepo := infra.NewFSWorkItemRepository()
	wfRepo := infra.NewFSWorkflowRepository()

	actor := domain.Actor{Kind: "human", ID: "matias"}

	// 1. Start Work Item
	startUC := usecases.NewStartWorkItemUseCase(wiRepo, wfRepo)
	item, err := startUC.Execute(tmpDir, usecases.StartWorkItemInput{
		ID:         "feat-test-lifecycle",
		WorkflowID: "feature-standard",
		Title:      "Test Lifecycle Feature",
		Summary:    "Testing the end to end engine CLI lifecycle",
		Actor:      actor,
	})
	if err != nil {
		t.Fatalf("StartWorkItemUseCase failed: %v", err)
	}

	if item.ID != "feat-test-lifecycle" {
		t.Errorf("Expected ID 'feat-test-lifecycle', got '%s'", item.ID)
	}
	if item.Phases["prd"].Status != "in_progress" {
		t.Errorf("Expected prd status 'in_progress', got '%s'", item.Phases["prd"].Status)
	}

	// 2. Status Check
	statusUC := usecases.NewStatusUseCase(wiRepo)
	statusItem, err := statusUC.Execute(tmpDir, "feat-test-lifecycle")
	if err != nil {
		t.Fatalf("StatusUseCase failed: %v", err)
	}
	if statusItem.Title != "Test Lifecycle Feature" {
		t.Errorf("Expected title 'Test Lifecycle Feature', got '%s'", statusItem.Title)
	}

	// 3. Approve PRD phase
	approveUC := usecases.NewApproveUseCase(wiRepo, wfRepo)
	approvedItem, err := approveUC.Execute(tmpDir, usecases.ApproveInput{
		WorkItemID: "feat-test-lifecycle",
		PhaseID:    "prd",
		ApprovedBy: actor,
		Comment:    "PRD looks great!",
	})
	if err != nil {
		t.Fatalf("ApproveUseCase failed: %v", err)
	}
	if approvedItem.Phases["prd"].Status != "approved" {
		t.Errorf("Expected prd status 'approved', got '%s'", approvedItem.Phases["prd"].Status)
	}
	if approvedItem.Phases["specification"].Status != "ready" {
		t.Errorf("Expected specification status 'ready', got '%s'", approvedItem.Phases["specification"].Status)
	}

	// 4. Next Action Check
	nextUC := usecases.NewNextUseCase(wiRepo, wfRepo)
	nextAction, err := nextUC.Execute(tmpDir, "feat-test-lifecycle")
	if err != nil {
		t.Fatalf("NextUseCase failed: %v", err)
	}
	if nextAction.PhaseID != "specification" {
		t.Errorf("Expected next phase 'specification', got '%s'", nextAction.PhaseID)
	}

	// 5. Record Event
	recordEventUC := usecases.NewRecordEventUseCase(wiRepo)
	err = recordEventUC.Execute(tmpDir, usecases.RecordEventInput{
		WorkItemID: "feat-test-lifecycle",
		EventType:  "test.completed",
		Message:    "Unit test verification passed",
		Actor:      actor,
	})
	if err != nil {
		t.Fatalf("RecordEventUseCase failed: %v", err)
	}
}

func TestBypassModeStart(t *testing.T) {
	tmpDir := setupTestEnv(t)
	defer os.RemoveAll(tmpDir)

	wiRepo := infra.NewFSWorkItemRepository()
	wfRepo := infra.NewFSWorkflowRepository()

	// Create dummy external artifact
	extArtPath := filepath.Join(tmpDir, "external-prd.md")
	if err := os.WriteFile(extArtPath, []byte("# My External PRD"), 0644); err != nil {
		t.Fatalf("Failed to create external artifact: %v", err)
	}

	actor := domain.Actor{Kind: "human", ID: "matias"}

	startUC := usecases.NewStartWorkItemUseCase(wiRepo, wfRepo)
	item, err := startUC.Execute(tmpDir, usecases.StartWorkItemInput{
		ID:           "feat-bypass",
		WorkflowID:   "feature-standard",
		Title:        "Bypass Test Feature",
		Summary:      "Starting from existing PRD",
		FromArtifact: extArtPath,
		Phase:        "prd",
		Actor:        actor,
	})
	if err != nil {
		t.Fatalf("StartWorkItemUseCase bypass mode failed: %v", err)
	}

	if item.Phases["prd"].Status != "accepted" {
		t.Errorf("Expected prd status 'accepted', got '%s'", item.Phases["prd"].Status)
	}
	if item.Phases["specification"].Status != "ready" {
		t.Errorf("Expected specification status 'ready', got '%s'", item.Phases["specification"].Status)
	}

	// Verify external artifact file was copied
	copiedArtPath := filepath.Join(tmpDir, ".sdd", "work-items", "active", "feat-bypass", "artifacts", "prd.md")
	if _, err := os.Stat(copiedArtPath); err != nil {
		t.Errorf("Expected copied artifact at %s, but file not found: %v", copiedArtPath, err)
	}
}
