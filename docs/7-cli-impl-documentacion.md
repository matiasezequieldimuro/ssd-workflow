# CLI del Motor SDD — Documentación Técnica

---

## 1. Implementación: ¿Qué se construyó?

La CLI es el **motor determinista del framework SDD**: recibe comandos, valida transiciones de estado y actualiza el manifiesto (`manifest.yaml`) y la bitácora (`events.jsonl`). El LLM propone; la CLI decide si la transición es válida.

### Ubicación en el repositorio

```text
src/
├── .sdd/                    ← Contrato del motor (schemas, workflows, templates)
└── cli/                     ← Código fuente de la CLI en Go
    ├── main.go              ← Punto de entrada
    ├── go.mod / go.sum      ← Módulo y lockfile de dependencias
    ├── cmd/                 ← Capa de comandos (interfaz terminal)
    ├── embeds/              ← Archivos .sdd por defecto embebidos en el binario
    └── internal/
        ├── domain/          ← Modelos y errores del negocio
        ├── ports/           ← Capacidades externas requeridas por aplicación
        ├── infra/           ← Implementación en disco (YAML/JSONL)
        └── usecases/        ← Lógica de negocio orquestada
```

### Arquitectura en Capas (Clean Architecture)

```mermaid
graph TD
    Terminal["👤 Terminal / Agente IA"]
    CMD["cmd/ — Cobra CLI\nParsea flags y args"]
    UC["internal/usecases/\nOrquesta la lógica"]
    DOMAIN["internal/domain/\nModelos y reglas"]
    PORTS["internal/ports/\nRepositorios, artifacts, reloj e IDs"]
    INFRA["internal/infra/\nCommit transaccional en filesystem"]

    Terminal --> CMD
    CMD --> UC
    UC --> DOMAIN
    UC --> PORTS
    INFRA -->|implementa| PORTS

    style DOMAIN fill:#f0f4ff,stroke:#4a6cf7
    style PORTS fill:#fff8e6,stroke:#f5a623
    style INFRA fill:#f0fff4,stroke:#27ae60
```

### Flujo de datos: ejemplo de `sdd approve`

```mermaid
sequenceDiagram
    participant User
    participant CMD as cmd/approve.go
    participant UC as usecases/approve_uc.go
    participant FS as infra/fs_repository.go
    participant Disk as .sdd/work-items/

    User->>CMD: sdd approve feat-001 --phase prd --by matias
    CMD->>UC: ApproveUseCase.Execute(...)
    UC->>FS: GetWorkItem("feat-001")
    FS->>Disk: Lee manifest.yaml
    Disk-->>FS: WorkItem struct
    FS-->>UC: WorkItem
    UC->>UC: Dominio valida status == awaiting_approval
    UC->>UC: Dominio exige actor human
    UC->>UC: Transición prd → approved
    UC->>UC: Dominio desbloquea dependencias → ready
    UC->>FS: CommitWorkItem(manifest, artifacts, events)
    FS->>FS: Lock + valida revision
    FS->>Disk: Prepara snapshot temporal
    FS->>Disk: Sync + publica snapshot
    UC-->>CMD: WorkItem actualizado
    CMD-->>User: JSON / Texto con resultado
```

### Componentes por capa

| Capa | Archivos | Responsabilidad |
| :--- | :--- | :--- |
| **Domain** | `work_item.go`, `workflow.go`, `event.go`, `errors.go` | Entidades, tipos e invariantes de transición |
| **Ports** | `repository.go` | Lectura, existencia, idempotencia, commit conjunto, artifacts, inicialización, reloj e IDs |
| **Infra** | `fs_repository.go`, `artifact_manager.go`, `project_initializer.go`, `runtime.go` | Filesystem transaccional, templates locales, inicialización, reloj de sistema e IDs aleatorios |
| **Use Cases** | `init_uc.go`, `start_uc.go`, `begin_phase_uc.go`, `deliver_phase_uc.go`, `approve_uc.go`, `complete_uc.go`, `next_uc.go`, `record_event_uc.go` | Orquestación de operaciones y persistencia |
| **CMD** | `composition.go`, `root.go` y comandos | Composición centralizada, parseo con Cobra `RunE` y contrato uniforme de salida |
| **Embeds** | `embeds.go`, `default_sdd/` | Templates de `.sdd/` embebidos en el binario para `sdd init` |

---

## 2. Cómo probar el proyecto

### A) Tests automatizados (sin compilar)

```bash
cd /Users/matiasdimuro/Documents/Webdev/sdd-harness/src/cli
/opt/homebrew/bin/go test -v ./...
```

Incluye:
- `TestContractFixtures` — valida estructural y semánticamente los fixtures JSON del contrato
- `TestEveryWorkflowCompletesItsMandatoryLifecycle` — descubre todos los `*.workflow.yaml` y completa sus fases obligatorias con artifacts, eventos y manifest válidos
- `TestCLICompletesFastChangeLifecycle` — compila el binario y recorre init → start → next/status → begin/deliver/approve → complete → record-event, validando stdout, stderr y exit codes
- `TestFullWorkItemLifecycle` — ciclo de integración detallado del workflow `feature-standard`
- `TestBypassModeStart` — inicio desde artefacto externo, incluyendo el gate requerido
- `state_machine_test.go` — transiciones table-driven, gates humanos, rework y fases opcionales
- `persistence_test.go` — rollback completo, locks, revisión optimista y recuperación tras interrupciones
- `contract_integration_test.go` — defaults, templates locales, artefactos externos, seguridad de paths y validación contractual
- `idempotency_test.go` — reintentos idempotentes mediante `operation_id`
- `dependency_injection_test.go` — use cases en memoria, tiempo/IDs deterministas y propagación table-driven de fallos de ports
- `cmd/root_test.go` — errores de argumentos y flags requeridos dentro del envelope JSON

### B) Compilar el binario

```bash
cd /Users/matiasdimuro/Documents/Webdev/sdd-harness/src/cli
/opt/homebrew/bin/go build -o sdd main.go
```

Genera el ejecutable `sdd` en la misma carpeta. **No commitear este archivo.**

### C) Probar manualmente en un proyecto nuevo

```bash
# 1. Compilar
cd /Users/matiasdimuro/Documents/Webdev/sdd-harness/src/cli
/opt/homebrew/bin/go build -o sdd main.go

# 2. Crear carpeta de prueba
mkdir ~/Desktop/mi-proyecto-test
cd ~/Desktop/mi-proyecto-test

# 3. Ejecutar flujo completo
SDD=/Users/matiasdimuro/Documents/Webdev/sdd-harness/src/cli/sdd
$SDD init
$SDD start feat-001 --title "Feature de prueba" --summary "Probar el motor"
$SDD status feat-001
$SDD next feat-001
$SDD deliver feat-001 --phase prd --actor-id copilot
$SDD approve feat-001 --phase prd --by matias --comment "OK"
$SDD next feat-001 --json
$SDD begin feat-001 --phase specification --actor-id copilot
```

### D) Instalar globalmente (más cómodo)

```bash
cd /Users/matiasdimuro/Documents/Webdev/sdd-harness/src/cli
/opt/homebrew/bin/go install .
# El binario queda en ~/go/bin/sdd → usable como "sdd" desde cualquier carpeta
```

---

## 3. Comandos disponibles

Todos los comandos comparten flags globales:

| Flag | Descripción | Default |
| :--- | :--- | :--- |
| `--json` | Salida en formato JSON estructurado (para agentes IA) | `false` |
| `--dir` | Directorio raíz del proyecto destino | `.` (directorio actual) |

Los comandos mutantes (`start`, `begin`, `deliver`, `approve`, `complete` y `record-event`) aceptan `--operation-id`. El agente debe reutilizar el mismo valor al reintentar una invocación incierta; una operación ya confirmada devuelve el estado existente sin duplicar eventos ni aumentar la revisión.

La v0.1 admite múltiples lectores, pero sólo un escritor simultáneo por work item. Cada mutación confirma conjuntamente manifest, artifacts y eventos; una revisión obsoleta o un lock ocupado producen error sin sobrescribir estado.

Todos los comandos usan Cobra `RunE`: los errores se propagan hasta el root command y `os.Exit` se ejecuta únicamente en `main`. Con `--json`, tanto los errores de negocio como los de argumentos o flags respetan el mismo envelope:

```json
{
  "success": false,
  "error": {
    "code": "invalid_arguments",
    "message": "accepts 1 arg(s), received 0"
  }
}
```

Los códigos iniciales son `invalid_arguments`, `invalid_input`, `not_found`, `already_exists`, `invalid_transition`, `concurrent_modification`, `work_item_locked` e `internal_error`.

---

### `sdd init`
Inicializa la estructura `.sdd/` en el proyecto actual desde las plantillas embebidas en el binario.

```bash
sdd init
sdd init --dir /ruta/a/mi-proyecto
```

> ⚠️ Falla si `.sdd/` ya existe. No inicializa Git.

---

### `sdd start <id>`
Crea un nuevo work item activo con su manifiesto y bitácora de eventos.

```bash
# Inicio estándar desde prompt
sdd start feat-023 --title "Cupones en checkout" --summary "Agregar campo de cupón"

# Inicio desde artefacto existente (bypass de fase prd)
sdd start feat-023 \
  --title "Cupones en checkout" \
  --from-artifact ./mi-prd.md \
  --phase prd
```

| Flag | Obligatorio | Default | Descripción |
| :--- | :--- | :--- | :--- |
| `--title` / `-t` | ✅ | — | Título del work item |
| `--workflow` / `-w` | ❌ | `defaults.workflow` de `.sdd/config.yaml` | ID del workflow a usar |
| `--summary` / `-s` | ❌ | — | Resumen del input inicial |
| `--from-artifact` | ❌ | — | Ruta a un artefacto externo existente |
| `--phase` | ❌* | — | Fase de entrada al usar `--from-artifact` |
| `--actor-kind` | ❌ | `human` | Tipo de actor creador |
| `--actor-id` | ❌ | `user` | ID del actor creador |
| `--operation-id` | ❌ | — | Clave estable para reintentos idempotentes |

> *Requerido si se usa `--from-artifact`

---

### `sdd status <id>`
Muestra el estado actual de todas las fases del work item.

```bash
sdd status feat-023
sdd status feat-023 --json
```

**Salida texto:**
```text
Work Item: feat-023 [active]
Title: Cupones en checkout
Workflow: feature-standard
------------------------------------------------------
PHASE                STATUS               ARTIFACT
------------------------------------------------------
prd                  approved             artifacts/prd.md
specification        ready                artifacts/specification.md
plan                 blocked              -
```

---

### `sdd approve <id>`
Registra la aprobación humana de una fase, desbloqueando automáticamente las fases dependientes.

```bash
sdd approve feat-023 --phase prd --by matias
sdd approve feat-023 --phase specification --by matias --comment "Revisado y OK"
```

| Flag | Obligatorio | Default | Descripción |
| :--- | :--- | :--- | :--- |
| `--phase` / `-p` | ✅ | — | ID de la fase a aprobar |
| `--by` / `-b` | ❌ | `human` | ID del aprobador humano |
| `--comment` / `-c` | ❌ | — | Comentario opcional |
| `--operation-id` | ❌ | — | Clave estable para reintentos idempotentes |

> Sólo aprueba fases en estado `awaiting_approval`, con política `required` u `optional`. La invariante de actor humano vive en el dominio, no sólo en Cobra.

---

### `sdd begin <id>`
Comienza explícitamente una fase `ready`, `rejected` o `superseded`.

```bash
sdd begin feat-023 --phase specification --actor-kind agent --actor-id copilot
sdd begin feat-023 --phase specification --operation-id run:specification:begin:001
```

`next` nunca realiza esta transición: sólo informa qué fase corresponde.

---

### `sdd deliver <id>`
Entrega la evidencia de una fase `in_progress`.

```bash
sdd deliver feat-023 --phase specification --actor-id copilot
sdd deliver feat-023 --phase archive --request-approval
sdd deliver feat-023 --phase specification --operation-id run:specification:deliver:001
```

- `approval: required` produce `awaiting_approval`.
- `approval: optional` produce `completed`, salvo que se use `--request-approval`.
- `approval: none` produce `completed` y rechaza `--request-approval`.

---

### `sdd complete <id>`
Completa una fase `approved`/`accepted`, o el work item si se omite `--phase`.

```bash
sdd complete feat-023 --phase prd
sdd complete feat-023
sdd complete feat-023 --operation-id run:work-item:complete:001
```

El work item sólo puede completarse cuando todas las fases obligatorias están satisfechas. Una fase opcional puede omitirse, pero si comenzó también debe terminar.

---

### `sdd next <id>`
Informa cuál es la próxima fase activa, qué procedimiento seguir y si requiere aprobación humana. Es una consulta pura: no escribe el manifest ni inicia la fase.

```bash
sdd next feat-023
sdd next feat-023 --json
```

**Salida JSON:**
```json
{
  "success": true,
  "data": {
    "phase_id": "specification",
    "status": "ready",
    "procedure": "generate-specification",
    "artifact": "artifacts/specification.md",
    "needs_approval": true,
    "optional": false,
    "message": "Next active phase is 'specification' (ready). Follow procedure 'generate-specification'."
  }
}
```

---

### `sdd record-event <id>`
Inyecta un evento personalizado de forma inmutable en `events.jsonl`.

```bash
sdd record-event feat-023 --type validation.completed --message "Tests pasaron OK"
sdd record-event feat-023 \
  --type custom.checkpoint \
  --message "Revisión intermedia" \
  --actor-kind agent \
  --actor-id planner
```

| Flag | Obligatorio | Default | Descripción |
| :--- | :--- | :--- | :--- |
| `--type` / `-t` | ✅ | — | Tipo de evento (ej: `validation.completed`) |
| `--message` / `-m` | ❌ | — | Descripción del evento |
| `--actor-kind` | ❌ | `agent` | Tipo de actor (`human`, `agent`, `cli`, `system`) |
| `--actor-id` | ❌ | `agent` | ID del actor |
| `--operation-id` | ❌ | — | Clave estable para reintentos idempotentes |

---

## 4. Fronteras arquitectónicas y runtime determinista

La Fase 4 corrige la dependencia parcial hacia infraestructura que todavía existía después de estabilizar la persistencia. El objetivo no fue crear una interfaz por archivo, sino separar únicamente capacidades externas o variables y conservar `CommitWorkItem` como frontera transaccional.

### 4.1. Ports según el consumidor

El contrato de aplicación se divide por capacidad:

| Port | Responsabilidad | Consumidores |
| :--- | :--- | :--- |
| `WorkItemReader` | Cargar el agregado | `status`, `next` y operaciones mutantes |
| `WorkItemExistenceChecker` | Detectar colisiones al crear | `start` |
| `OperationTracker` | Comprobar `operation_id` aplicado | Operaciones mutantes |
| `WorkItemCommitter` | Confirmar manifest, artifacts y eventos juntos | Operaciones mutantes |
| `ArtifactPreparer` | Renderizar y validar artifacts sin persistirlos | `start`, `deliver`, `approve`, `complete` |
| `ExternalArtifactImporter` | Resolver, hashear e importar evidencia externa | `start` |
| `ProjectInitializer` | Publicar `.sdd/` de forma atómica | `init` |
| `Clock` | Proveer timestamps controlables | Creación, approvals, artifacts y eventos |
| `IDGenerator` | Generar IDs únicos de eventos | Todas las mutaciones |

`WorkItemMutationRepository` y `WorkItemCreationRepository` componen las capacidades mínimas requeridas por cada tipo de operación. Esta segregación no separa artificialmente manifest, artifacts y eventos en stores independientes: los tres siguen confirmándose mediante un único `WorkItemCommit`.

### 4.2. Dependencias de use cases

Los packages de `internal/usecases` no importan `internal/infra`. La infraestructura productiva implementa los ports, mientras los tests pueden reemplazarla por dobles en memoria:

```text
usecases ──> domain
usecases ──> ports
infra    ──> domain + ports
cmd      ──> usecases
```

`InitUseCase` ya no conoce `os`, `io/fs` ni los embeds. Delega en `ProjectInitializer`. La preparación de artifacts tampoco se instancia dentro de los casos de uso: se recibe mediante `ArtifactPreparer` o `ArtifactService`.

### 4.3. Creación del agregado

La construcción inicial de estados salió de `StartWorkItemUseCase` y se concentra en:

```go
domain.NewWorkItem(workflow, params)
```

El dominio decide:

- qué fases comienzan `blocked`;
- cuál pasa de `ready` a `in_progress` en un inicio normal;
- qué ancestros quedan `not_applicable` al ingresar desde evidencia externa;
- si la fase externa queda `accepted` o `awaiting_approval`;
- cómo se construyen workflow, input, traceability y estado inicial.

El use case queda limitado a validar input de aplicación, cargar configuración, resolver evidencia externa, solicitar artifacts, crear eventos y pedir el commit.

### 4.4. Reloj, IDs y secuencia de eventos

`SystemClock` entrega tiempo UTC y `CryptoIDGenerator` produce IDs `evt_` con 128 bits aleatorios. El ID ya no deriva del timestamp, por lo que dos eventos creados en el mismo instante no colisionan.

Cada transición emite `phase.transitioned` con:

```json
{
  "phase": "specification",
  "from": "blocked",
  "to": "ready",
  "cause": "dependencies_satisfied"
}
```

El orden se preserva dentro del mismo commit:

1. evento que describe la intención principal;
2. transición principal;
3. transiciones derivadas por desbloqueo de dependencias.

Para un inicio desde artifact externo se registra `work_item.created` antes de `phase.bypassed_by_external_input`, seguido de la transición de entrada. El mismo `operation_id` se conserva como `correlation_id` en todos los eventos de la operación.

### 4.5. Composition root y salida de CLI

`cmd/composition.go` es el único lugar que construye implementaciones productivas y las conecta con los casos de uso. Los archivos de comandos sólo:

1. definen argumentos y flags;
2. construyen el input tipado;
3. ejecutan el use case;
4. presentan el resultado.

Todos los handlers usan `RunE`. Los errores llegan al root command, que decide entre salida de texto y `JSONResponse`; ningún helper llama `os.Exit`. El punto de entrada `main.go` es el único responsable de convertir un error en exit code `1`.

Los tests se nombran por comportamiento:

- `contract_integration_test.go`;
- `idempotency_test.go`;
- `dependency_injection_test.go`;
- `cmd/root_test.go`.

No se usan nombres asociados a fases del roadmap porque perderían significado cuando el backlog evolucione.

### 4.6. Suite contractual

La Fase 5 organiza la verificación por garantía y no por archivo de implementación:

| Frontera | Propiedad demostrada |
| :--- | :--- |
| Dominio | Las matrices de `begin`, `deliver` y `complete` aceptan sólo estados contractuales; approvals, rework y fases opcionales mantienen sus invariantes |
| Schemas y semántica | Todos los fixtures válidos deben pasar y todos los inválidos deben fallar; una carpeta vacía de fixtures también falla la suite |
| Workflows | Ciclos, dependencias inexistentes, entry points y paths inválidos son rechazados |
| Persistencia | Los fallos antes o durante publicación no dejan estado parcial; locks y revisiones obsoletas impiden escritores conflictivos |
| Ports | Errores de repositorio, config, workflows, artifacts e IDs se propagan sin producir un commit exitoso |
| Lifecycle | Cada archivo `*.workflow.yaml` inicializado se recorre hasta completar todas sus fases obligatorias y validar artifacts, manifest y eventos |
| CLI | El binario compilado conserva consultas puras, orden de texto, envelopes JSON, canales stdout/stderr y exit codes |

El test de workflows no mantiene una lista duplicada: descubre los archivos instalados en `.sdd/workflows/`. Agregar un workflow nuevo lo incorpora automáticamente al ciclo contractual. El test de CLI se ejecuta contra el composition root productivo y el filesystem real; los dobles quedan reservados para aislar fallos de ports y controlar reloj e IDs.

## 5. Próximos pasos

```mermaid
graph LR
    CLI["✅ CLI Fases 1-5\nImplementadas"]
    VALIDATE["🔴 sdd validate\nValidar manifest vs JSON Schema"]
    AGENT["🔴 Integración Agente\nOrquestador usa la CLI"]
    REJECT["🟡 sdd reject\nRechazar fase con motivo"]
    ARCHIVE["🟡 sdd archive\nArchivar work item"]
    ENGRAM["🟢 Engram + CodeGraph\nMemoria semántica"]

    CLI --> VALIDATE
    CLI --> AGENT
    VALIDATE --> AGENT
    AGENT --> REJECT
    AGENT --> ARCHIVE
    ARCHIVE --> ENGRAM
```

| Prioridad | Paso | Descripción |
| :--- | :--- | :--- |
| 🔴 Alta | **`sdd validate`** | Exponer una validación explícita bajo demanda; las operaciones actuales ya validan schemas y semántica al cargar/persistir |
| 🔴 Alta | **Integración del agente orquestador** | Instruir al agente (via `AGENTS.md` o skill) para que invoque la CLI como autoridad de estado |
| 🟡 Media | **`sdd reject`** | Rechazar una fase con motivo, bloqueando el flujo hasta nueva generación |
| 🟡 Media | **`sdd archive <id>`** | Mover un work item completado de `active/` a `archive/YYYY-MM-DD-<id>/` |
| 🟢 Baja | **Validación de Event Types** | Validar que el tipo en `record-event` respete el patrón del schema |
| 🟢 Baja | **Engram + CodeGraph** | Integración con memoria semántica (fuera del scope de Fase 1) |

---

*Motor SDD CLI v0.1 — Actualizado el 2026-08-18*
