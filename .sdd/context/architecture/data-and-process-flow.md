# Flujos de datos y procesos

## Flujo principal: ciclo de vida de un work item

```mermaid
sequenceDiagram
    actor H as Humano
    participant CLI as sdd-cli
    participant FS as Sistema de archivos (.sdd/)
    participant A as Agente (Claude Code)

    H->>CLI: start <id> --title "..." --summary "..."
    CLI->>FS: Crear manifest.yaml + events.jsonl<br/>(primera fase en in_progress)
    CLI-->>H: Work item creado, fase en progreso

    loop Por cada fase del workflow
        A->>A: Produce artefacto (PRD, spec, plan, etc.)
        A->>CLI: deliver <id> --phase <fase>
        CLI->>FS: Actualizar manifest.yaml<br/>Append a events.jsonl
        alt Aprobación requerida
            CLI-->>A: Fase en awaiting_approval
            H->>CLI: approve <id> --phase <fase> --by <actor>
            CLI->>FS: Registrar aprobación<br/>Desbloquear siguiente fase
            CLI-->>H: Fase aprobada
            H->>CLI: complete <id> --phase <fase>
            CLI->>FS: Fase → completed
        else Sin aprobación
            CLI->>FS: Fase → completed<br/>Desbloquear siguiente fase
        end
        A->>CLI: begin <id> --phase <siguiente>
        CLI->>FS: Siguiente fase → in_progress
    end

    H->>CLI: complete <id>
    CLI->>FS: WorkItem → completed
    H->>CLI: archive <id>
    CLI->>FS: Mover active/<id> → archive/YYYY-MM-DD-<id>
```

## Flujo de inicio desde artefacto externo (`--from-artifact`)

Permite saltar fases anteriores si ya existe un artefacto aprobado fuera del work item.

```mermaid
flowchart LR
    A["--from-artifact ./plan.md<br/>--phase plan"] --> B["Calcular fases ancestras"]
    B --> C["Marcar ancestras como not_applicable"]
    C --> D["Aceptar artefacto externo<br/>(SHA256 verificado)"]
    D --> E["plan → accepted o awaiting_approval"]
```

## Flujo de transiciones de fase (máquina de estados)

```mermaid
stateDiagram-v2
    [*] --> blocked: Work item creado
    blocked --> ready: Dependencias satisfechas
    ready --> in_progress: begin
    rejected --> in_progress: begin (retrabajo)
    superseded --> in_progress: begin (tras rechazo múltiple)
    in_progress --> awaiting_approval: deliver (approval=required o optional+flag)
    in_progress --> completed: deliver (approval=none o optional sin flag)
    awaiting_approval --> approved: approve (human)
    awaiting_approval --> rejected: reject (human)
    approved --> completed: complete --phase
    not_applicable --> [*]: (fases previas a entrada externa)
    accepted --> completed: complete --phase
```

## Flujo de persistencia transaccional

Cada mutación de `CommitWorkItem`:

1. Adquirir lock de archivo (`.lock`).
2. Verificar revisión actual del manifest (control optimista).
3. Escribir artefactos nuevos en paths temporales.
4. Serializar manifest actualizado a path temporal.
5. Append de eventos a `events.jsonl`.
6. Rename atómico de temporales a definitivos.
7. Liberar lock.

En caso de error en cualquier paso: rollback sin efectos visibles.

## Flujo de idempotencia

```mermaid
flowchart TD
    A["Use case recibe --operation-id"] --> B{"OperationApplied?"}
    B -- Sí --> C["Retornar estado persistido\n(sin mutar)"]
    B -- No --> D["Ejecutar transición"]
    D --> E["Persistir + registrar operation-id en eventos"]
```

## Flujo de inicialización de proyecto

```mermaid
flowchart LR
    A["sdd-cli init --dir /proyecto"] --> B{"Existe .sdd/?"}
    B -- Sí --> C["Error: ya inicializado"]
    B -- No --> D["Copiar recursos embebidos\n(.sdd/ sin fixtures de test)"]
    D --> E["Crear work-items/active/ y archive/"]

    F["sdd-cli adapters install claude-code"] --> G{"Existe .sdd/?"}
    G -- No --> H["Error: init primero"]
    G -- Sí --> I{"Colisión de archivos?"}
    I -- Sí --> J["Error: sin --force en BETA"]
    I -- No --> K["Instalar CLAUDE.md, .mcp.json,\n.claude/ (agents, skills, hooks, commands)"]
```

## Flujo de validación

```mermaid
flowchart TD
    A["sdd-cli validate"] --> B["InspectProject: config, schemas,\nworkflows, templates, procedures, work items"]
    A2["sdd-cli validate <id>"] --> C["InspectWorkItem: manifest, workflow,\napprovals, artefactos, eventos, referencias"]
    B & C --> D["Checks: passed | warning | failed"]
    D -- "failed" --> E["Exit code 1"]
    D -- "solo warnings" --> F["Exit code 0"]
```

## Flujo del adapter Claude Code

El adapter instala en el proyecto destino:

- `CLAUDE.md` — instrucciones del framework SDD para el agente.
- `.claude/agents/sdd-*.md` — agentes especializados (orchestrator, developer, planner, etc.).
- `.claude/skills/` — skills invocables (implement, create-plan, verify, etc.).
- `.claude/commands/sdd/` — comandos slash del adapter.
- `.claude/hooks/` — `protect-sdd-paths.py` y `protect-secrets.py`.
- `.mcp.json` — configuración de MCP para el adapter.

Una vez instalado, el agente (Claude Code) opera la CLI como motor del proceso:

```text
Agente invoca skill → Skill lee procedimiento .sdd/procedures/ → Agente produce artefacto → Agente invoca sdd-cli deliver → CLI valida y persiste
```
