package domain

type WorkItemWorkflow struct {
	ID         string `json:"id" yaml:"id"`
	Version    string `json:"version" yaml:"version"`
	EntryPhase string `json:"entry_phase" yaml:"entry_phase"`
}

type WorkItemInput struct {
	Source     string   `json:"source" yaml:"source"` // user_prompt, external_artifact, imported_artifact
	Summary    string   `json:"summary" yaml:"summary"`
	References []string `json:"references,omitempty" yaml:"references,omitempty"`
}

type PhaseState struct {
	Status   string `json:"status" yaml:"status"` // not_applicable, blocked, ready, in_progress, awaiting_approval, approved, completed, rejected, accepted, superseded
	Artifact string `json:"artifact,omitempty" yaml:"artifact,omitempty"`
}

type Approval struct {
	Phase   string `json:"phase" yaml:"phase"`
	Status  string `json:"status" yaml:"status"` // pending, approved, rejected, superseded
	By      *Actor `json:"by,omitempty" yaml:"by,omitempty"`
	At      string `json:"at,omitempty" yaml:"at,omitempty"`
	Comment string `json:"comment,omitempty" yaml:"comment,omitempty"`
}

type Traceability struct {
	Events           string   `json:"events" yaml:"events"`
	RelatedWorkItems []string `json:"related_work_items,omitempty" yaml:"related_work_items,omitempty"`
	BaselineSpecs    []string `json:"baseline_specs,omitempty" yaml:"baseline_specs,omitempty"`
}

type TokenUsage struct {
	Status           string  `json:"status" yaml:"status"` // not_reported, partial, recorded
	Source           *string `json:"source,omitempty" yaml:"source,omitempty"`
	InputTokens      *int    `json:"input_tokens,omitempty" yaml:"input_tokens,omitempty"`
	OutputTokens     *int    `json:"output_tokens,omitempty" yaml:"output_tokens,omitempty"`
	CacheReadTokens  *int    `json:"cache_read_tokens,omitempty" yaml:"cache_read_tokens,omitempty"`
	CacheWriteTokens *int    `json:"cache_write_tokens,omitempty" yaml:"cache_write_tokens,omitempty"`
}

type Observability struct {
	TokenUsage *TokenUsage `json:"token_usage,omitempty" yaml:"token_usage,omitempty"`
}

type WorkItem struct {
	SchemaVersion string                `json:"schema_version" yaml:"schema_version"`
	Kind          string                `json:"kind" yaml:"kind"`
	ID            string                `json:"id" yaml:"id"`
	Title         string                `json:"title" yaml:"title"`
	Type          string                `json:"type" yaml:"type"`
	Status        string                `json:"status" yaml:"status"` // active, completed, archived, cancelled
	CreatedAt     string                `json:"created_at" yaml:"created_at"`
	CreatedBy     *Actor                `json:"created_by,omitempty" yaml:"created_by,omitempty"`
	Workflow      WorkItemWorkflow      `json:"workflow" yaml:"workflow"`
	Input         WorkItemInput         `json:"input" yaml:"input"`
	Phases        map[string]PhaseState `json:"phases" yaml:"phases"`
	Approvals     []Approval            `json:"approvals,omitempty" yaml:"approvals,omitempty"`
	Traceability  Traceability          `json:"traceability" yaml:"traceability"`
	Observability *Observability        `json:"observability,omitempty" yaml:"observability,omitempty"`
}
