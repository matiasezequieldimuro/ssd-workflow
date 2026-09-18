# Arquitectura del proyecto

## Estructura del repositorio

```text
ssd-workflow/
├── .claude/                    # Adapter Claude Code (instalado por adapters install)
│   ├── agents/                 # Definiciones de agentes especializados (sdd-*)
│   ├── skills/                 # Skills invocables (onboard-project, implement, etc.)
│   ├── commands/sdd/           # Comandos slash del adapter
│   ├── hooks/                  # Hooks de protección (protect-sdd-paths, protect-secrets)
│   └── settings.json           # Permisos y configuración del adapter
├── .sdd/                       # Contrato SDD del propio proyecto
│   ├── config.yaml             # Configuración del proyecto (workflow, idioma, modo)
│   ├── workflows/              # Workflows predefinidos (YAML)
│   ├── schemas/                # JSON Schemas de validación
│   ├── templates/              # Plantillas Markdown de artefactos
│   ├── procedures/             # Guías de procedimiento por fase
│   ├── context/                # Contexto estable del proyecto (este directorio)
│   ├── specs/                  # Specs de funcionalidades existentes
│   ├── research/               # Evidencia exploratoria fechada
│   ├── registry/capabilities.yaml  # Registro de capacidades SDD
│   └── work-items/
│       ├── active/             # Work items en curso
│       └── archive/            # Work items completados y archivados
├── src/cli/                    # Código fuente de sdd-cli (Go)
│   ├── main.go                 # Punto de entrada
│   ├── cmd/                    # Comandos Cobra (uno por subcomando)
│   ├── internal/
│   │   ├── domain/             # Entidades, reglas y máquina de estados
│   │   ├── ports/              # Interfaces (contratos de infraestructura)
│   │   ├── usecases/           # Casos de uso (orquestación)
│   │   └── infra/              # Implementaciones de puertos (FS, schemas, etc.)
│   ├── embeds/                 # Recursos embebidos (contrato .sdd, adapters)
│   ├── tools/                  # Herramientas de generación (syncsdd, syncadapters)
│   └── generate.go             # Directivas go:generate
├── docs/                       # Documentación pública
│   ├── CLI.md                  # Referencia de la CLI
│   ├── SDD_WORKFLOW.md         # Modelo de workflow (fases, gates, artefactos)
│   ├── GUIA_INSTALACION.md     # Instalación por OS
│   └── internal/               # Notas de diseño no publicadas
├── scripts/                    # Scripts de instalación (sh, ps1)
├── .github/workflows/          # CI (ci.yml) y Release (release.yml)
├── CLAUDE.md                   # Instrucciones del adapter (instalado por adapters install)
├── AGENTS.md                   # Descripción de agentes para AGENTS.md raíz
├── README.md                   # Presentación pública del proyecto
└── CHANGELOG.md                # Historial de cambios
```

## Organización por capas o módulos

| Capa / Módulo | Responsabilidad | Dependencias principales |
| --- | --- | --- |
| `domain` | Entidades (`WorkItem`, `Workflow`), estados, máquina de transiciones, validaciones | Ninguna (puro Go) |
| `ports` | Contratos de interfaz (repositorios, servicios, reloj, generador de IDs) | `domain` |
| `usecases` | Orquestación de casos de uso: `start`, `begin`, `deliver`, `approve`, `reject`, `complete`, `archive`, `validate`, `next`, `status`, `init`, `adapters` | `domain`, `ports` |
| `infra` | Implementaciones concretas: `fs_repository` (FS), `schema_validator`, `artifact_manager`, `config_repository`, `project_initializer`, `path_security` | `ports`, `domain` |
| `cmd` | Comandos Cobra: parseo de flags, llamada al use case, formateo de salida | `usecases`, `ports` |
| `embeds` | Recursos estáticos embebidos: contrato `.sdd/` y adapters | — |
| `tools` | Sincronizadores de recursos embebidos (`syncsdd`, `syncadapters`) | — |

## Puntos de entrada

- CLI: `src/cli/main.go` → Cobra root command (`cmd/root.go`).
- Subcomandos: `cmd/start.go`, `cmd/begin.go`, `cmd/deliver.go`, `cmd/approve.go`, `cmd/reject.go`, `cmd/complete.go`, `cmd/archive.go`, `cmd/status.go`, `cmd/next.go`, `cmd/validate.go`, `cmd/init.go`, `cmd/adapters.go`, `cmd/record_event.go`, `cmd/version.go`.
- `go generate ./...`: ejecuta `tools/syncsdd` y `tools/syncadapters` para actualizar recursos embebidos antes del build.

## Patrones implementados

- **Hexagonal (Ports & Adapters)**: `domain` y `usecases` son independientes de infraestructura; `ports` define las interfaces; `infra` las implementa.
- **Máquina de estados finita**: en `domain/work_item.go`; todas las transiciones de fase pasan por métodos que validan la transición antes de aplicarla.
- **Persistencia transaccional**: `CommitWorkItem` en `infra/fs_repository.go` escribe manifest + artefactos + eventos en una sola operación atómica (rename de archivo temporal).
- **Idempotencia**: `OperationTracker` persiste `operation_id` aplicados; los use cases verifican antes de ejecutar para evitar dobles efectos.
- **Locking optimista**: revisión incremental (`Revision`) en `manifest.yaml`; una escritura concurrente con revisión obsoleta falla con `concurrent_modification`.
- **Embed de recursos**: `go:embed` en `embeds/embeds.go` empaqueta el contrato SDD y los adapters en el binario.

## Convenciones

- Nomenclatura: `snake_case` en YAML/JSON; `PascalCase` en Go para exports; `camelCase` en Go privados.
- Organización: un archivo por comando Cobra; un archivo por use case.
- Manejo de errores: errores centinela en `domain/errors.go` (`ErrInvalidTransition`, `ErrPhaseNotFound`, etc.); wrapping con `fmt.Errorf("%w: ...")`.
- Configuración: `.sdd/config.yaml` + flags globales `--dir`, `--json`, `--actor-*`, `--operation-id`.
- Pruebas: unitarias en `domain/` e `infra/`; integración en `usecases/contract_integration_test.go` y `cmd/cli_e2e_test.go`.

## Dependencias internas

- `usecases` depende de `ports` (interfaces); nunca de `infra` directamente.
- `cmd` instancia `infra` y las inyecta en `usecases` vía `cmd/composition.go`.
- `domain` es la capa más interna; no importa ningún otro paquete interno.

## Pendientes y fuentes

- Pendientes: verificar si `fs_archive_repository.go` implementa locking propio o hereda el de `fs_repository`.
- Fuentes verificadas: `src/cli/` (todos los archivos Go), `README.md`, `docs/CLI.md`.
