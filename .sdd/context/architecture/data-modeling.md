# Modelado de datos

## Entidades principales

### WorkItem (`manifest.yaml`)

El expediente de una unidad de trabajo. Persiste en `.sdd/work-items/active/<id>/manifest.yaml`.

| Campo | Tipo | Descripción |
| --- | --- | --- |
| `schema_version` | string | Versión del schema (`"0.1"`) |
| `kind` | string | Siempre `"work-item"` |
| `id` | string | Identificador kebab-case único (`^[a-z0-9]+(?:-[a-z0-9]+)*$`) |
| `revision` | int | Contador incremental; incrementa en cada mutación |
| `title` | string | Título legible |
| `type` | string | Tipo derivado del workflow (`feature`, `bug`, `change-request`, `fast-change`) |
| `status` | WorkItemStatus | `active` \| `completed` \| `archived` \| `cancelled` |
| `created_at` | string (ISO 8601) | Timestamp de creación |
| `created_by` | Actor | Actor que creó el item |
| `workflow` | WorkItemWorkflow | ID de workflow, versión y fase de entrada |
| `input` | WorkItemInput | Fuente (`user_prompt`, `external_artifact`), resumen, referencias |
| `phases` | map[string]PhaseState | Estado de cada fase del workflow |
| `approvals` | []Approval | Historial de aprobaciones y rechazos |
| `traceability` | Traceability | Referencia a `events.jsonl`, items relacionados, specs baseline |
| `observability` | Observability | Token usage (opcional) |

### PhaseState

| Campo | Tipo | Descripción |
| --- | --- | --- |
| `status` | PhaseStatus | Ver estados abajo |
| `artifact` | string | Ruta relativa al artefacto de la fase |

**Estados de fase:**

```text
not_applicable → (fases previas a la entrada cuando se usa --from-artifact)
blocked        → (dependencias no satisfechas)
ready          → (dependencias satisfechas, pendiente begin)
in_progress    → (comenzada)
awaiting_approval → (entregada, esperando aprobación humana)
approved       → (aprobada, puede completarse)
accepted       → (importada externamente, sin gate de aprobación requerido)
completed      → (finalizada)
rejected       → (rechazada, debe recomenzarse con begin)
superseded     → (reemplazada por un re-inicio tras rechazo múltiple)
```

**Satisfacción de dependencias:** `approved | completed | accepted`.
**Satisfacción de completion:** `approved | completed | accepted | not_applicable`.

### Approval

| Campo | Tipo | Descripción |
| --- | --- | --- |
| `phase` | string | ID de la fase aprobada/rechazada |
| `status` | ApprovalStatus | `pending` \| `approved` \| `rejected` \| `superseded` |
| `by` | Actor | Actor humano que tomó la decisión |
| `at` | string (ISO 8601) | Timestamp |
| `comment` | string | Comentario opcional |

### Actor

| Campo | Tipo | Descripción |
| --- | --- | --- |
| `kind` | ActorKind | `human` \| `agent` \| `cli` |
| `id` | string | Identificador del actor |

Las aprobaciones y rechazos requieren actor `human`.

### Workflow (`*.workflow.yaml`)

Almacenado en `.sdd/workflows/`. Embebido en el binario.

| Campo | Tipo | Descripción |
| --- | --- | --- |
| `id` | string | Identificador del workflow |
| `work_item_type` | string | Tipo de work item que crea |
| `entry_points` | []EntryPoint | Fases de entrada válidas y tipos de artefacto que aceptan |
| `phases` | []WorkflowPhase | Lista ordenada de fases |
| `artifacts` | map[string]WorkflowArtifact | Mapa artifactID → path + template |

### WorkflowPhase

| Campo | Tipo | Descripción |
| --- | --- | --- |
| `id` | string | Identificador de la fase |
| `requires` | []string | IDs de fases que deben satisfacerse antes |
| `produces` | []string | IDs de artefactos que genera |
| `procedure` | string | Procedimiento de referencia para el agente |
| `approval` | ApprovalPolicy | `none` \| `required` \| `optional` |
| `optional` | bool | Si `true`, puede omitirse sin bloquear completion |
| `effects` | []string | Efectos declarados (`repository_write`, `git_commit`, `pull_request`, etc.) |

### Event (`events.jsonl`)

Registro inmutable, append-only. Cada línea es un JSON.

| Campo | Tipo | Descripción |
| --- | --- | --- |
| `schema_version` | string | `"0.1"` |
| `kind` | string | `"event"` |
| `id` | string | UUID del evento |
| `type` | string | Tipo de evento (ej. `phase.begun`, `phase.delivered`, `approval.approved`) |
| `work_item_id` | string | ID del work item |
| `revision` | int | Revisión del work item al momento del evento |
| `actor` | Actor | Quién lo generó |
| `at` | string (ISO 8601) | Timestamp |
| `payload` | object | Datos específicos del tipo de evento |

## Persistencia

```text
.sdd/work-items/
├── active/
│   └── <id>/
│       ├── manifest.yaml     ← WorkItem serializado
│       ├── events.jsonl      ← Historial de eventos (append-only)
│       └── artifacts/
│           ├── prd.md
│           ├── specification.md
│           └── ...
└── archive/
    └── YYYY-MM-DD-<id>/      ← Directorio renombrado al archivar (inmutable)
        ├── manifest.yaml
        ├── events.jsonl
        └── artifacts/
```

## Validaciones conocidas

- ID de work item: `^[a-z0-9]+(?:-[a-z0-9]+)*$`.
- ID de actor: no vacío.
- Actor de aprobación/rechazo: debe ser `human`.
- Transiciones de fase: validadas por la máquina de estados antes de persistir.
- Revisión optimista: `Revision` en el manifest previene escrituras concurrentes.
- Integridad de artefactos externos: SHA256 verificado al importar.
- Operation ID: 1–128 caracteres, `[a-zA-Z0-9.:_-]`.

## Pendientes y fuentes

- Pendientes: verificar campos exactos de `Config` (`domain/config.go`) para completar modelo.
- Fuentes verificadas: `domain/work_item.go`, `domain/workflow.go`, `domain/event.go`, `ports/repository.go`, `docs/CLI.md`.
