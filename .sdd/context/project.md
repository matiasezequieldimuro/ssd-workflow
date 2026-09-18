# Proyecto

## Propósito

SDD Workflow es un framework de **Spec-Driven Development** que hace trazable y gobernada la colaboración entre personas y agentes de IA. En lugar de confiar en que el modelo "recuerde" el estado del proceso, modela cada entrega como fases explícitas con gates humanos, artefactos y un historial auditable.

El producto principal es `sdd-cli`: un binario nativo Go que actúa como motor determinista del proceso. Los agentes realizan el trabajo cognitivo (redactar, planificar, implementar); la CLI gobierna transiciones, valida reglas y persiste evidencia.

## Alcance y límites

- Incluye:
  - CLI (`sdd-cli`) multiplataforma (Linux, macOS, Windows) sin runtime externo.
  - Contrato declarativo (`.sdd/`) con workflows, schemas, templates y procedures.
  - Adapter para Claude Code (`sdd-cli adapters install claude-code`).
  - Cinco workflows predefinidos: `feature-standard`, `change-request`, `bug-investigation`, `bug-known-cause`, `fast-change`.
  - Scripts de instalación (`scripts/install.sh`, `scripts/install.ps1`).
  - CI/CD: validación en push/PR y release multiplataforma via GitHub Actions.

- No incluye:
  - Redacción de contenido (PRDs, planes, código): responsabilidad de los agentes.
  - Ejecución de commits, PRs o deploys: las fases con `effects` declaran la intención; el agente o la persona los ejecuta.
  - UI gráfica.
  - Soporte para LLMs distintos de Claude Code en la BETA.

- Límites conocidos:
  - BETA `v0.1.0-beta`: contrato `schema_version: "0.1"`, sin garantía de compatibilidad hacia atrás.
  - Sin `--force` en `adapters install`: no sobrescribe archivos existentes.

## Componentes principales

| Componente | Ubicación | Responsabilidad |
| --- | --- | --- |
| CLI (`sdd-cli`) | `src/cli/` | Motor determinista: gestiona estado, valida transiciones, persiste evidencia |
| Contrato SDD | `.sdd/` (embebido en el binario) | Workflows, schemas, templates, procedures |
| Adapter Claude Code | Embebido en el binario | Instala `CLAUDE.md`, `.mcp.json`, agents, skills, commands, hooks |
| Scripts de instalación | `scripts/` | Descargar e instalar el binario en cada OS |

## Repositorios y dependencias relacionadas

- Repositorio principal: `github.com/matiasezequieldimuro/ssd-workflow`
- Sin dependencias de runtime externas (binario Go estático).
- Dependencias de build: Go 1.22+.

## Desarrollo local

- Requisitos: Go 1.22+.
- Build:
  ```bash
  cd src/cli
  go generate ./...   # sincroniza .sdd/ y adapters embebidos
  go build -o sdd-cli .
  ```
- Pruebas: `go test ./...` desde `src/cli/`.
- Lint/formato: `gofmt`, `go vet` (ejecutados en CI).

## Documentación y referencias

- `README.md` — resumen y guía de inicio rápido.
- `docs/CLI.md` — referencia de comandos, flags y contrato de salida.
- `docs/SDD_WORKFLOW.md` — modelo de fases, gates, artefactos y fuentes de verdad.
- `docs/GUIA_INSTALACION.md` — instalación, actualización y desinstalación.
- `CHANGELOG.md` — scope funcional de la BETA.
- `docs/internal/` — notas de diseño no publicadas.

## Estado y pendientes

- Estado actual: BETA (`v0.1.0-beta`), publicada en GitHub Releases.
- Rama de desarrollo activa: `development`.
- Pendientes o preguntas abiertas:
  - Soporte para múltiples adapters simultáneos.
  - Estrategia de upgrade del contrato (schema versioning).
  - Adapter para otros agentes (Copilot, Cursor).
