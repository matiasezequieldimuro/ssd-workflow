package domain

type EntryPoint struct {
	Phase   string   `json:"phase" yaml:"phase"`
	Accepts []string `json:"accepts" yaml:"accepts"`
}

type WorkflowPhase struct {
	ID        string   `json:"id" yaml:"id"`
	Requires  []string `json:"requires,omitempty" yaml:"requires,omitempty"`
	Produces  []string `json:"produces,omitempty" yaml:"produces,omitempty"`
	Procedure string   `json:"procedure,omitempty" yaml:"procedure,omitempty"`
	Approval  string   `json:"approval" yaml:"approval"` // none, required, optional
	Effects   []string `json:"effects,omitempty" yaml:"effects,omitempty"`
}

type WorkflowArtifact struct {
	Path     string `json:"path" yaml:"path"`
	Template string `json:"template" yaml:"template"`
}

type Workflow struct {
	SchemaVersion string                      `json:"schema_version" yaml:"schema_version"`
	Kind          string                      `json:"kind" yaml:"kind"`
	ID            string                      `json:"id" yaml:"id"`
	Title         string                      `json:"title" yaml:"title"`
	WorkItemType  string                      `json:"work_item_type" yaml:"work_item_type"`
	Description   string                      `json:"description,omitempty" yaml:"description,omitempty"`
	EntryPoints   []EntryPoint                `json:"entry_points,omitempty" yaml:"entry_points,omitempty"`
	Phases        []WorkflowPhase             `json:"phases" yaml:"phases"`
	Artifacts     map[string]WorkflowArtifact `json:"artifacts" yaml:"artifacts"`
}
