# SDD Workflow — Motor de Spec-Driven Development

`sdd-cli` es el **motor determinista** del framework de Spec-Driven Development (SDD):
gobierna el proceso de trabajo (fases, gates humanos, transiciones y trazabilidad)
sobre el filesystem, de forma agent-agnostic y sin depender de que un LLM recuerde el
estado. Es un binario nativo autocontenido escrito en Go, sin runtime externo.

> Estado: **BETA**.

## Instalacion rapida

**macOS / Linux (Ubuntu):**

```bash
curl -fsSL https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/scripts/install.sh | sh
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/scripts/install.ps1 | iex
```

Los binarios se publican en [GitHub Releases](https://github.com/matiasezequieldimuro/ssd-workflow/releases).
Para instalacion manual, build desde fuente o troubleshooting, consulta la
[**Guia de instalacion**](./docs/10-guia-instalacion.md).

## Primeros pasos

```bash
sdd-cli version                       # verificar instalacion
sdd-cli init                          # crear la estructura .sdd/ en tu proyecto
sdd-cli adapters install claude-code  # (opcional) adapter de Claude Code
sdd-cli --help                        # ver todos los comandos
```

## Documentacion

- [Guia de instalacion](./docs/10-guia-instalacion.md)
- [Guia tecnica de la CLI](./docs/9-guia-tecnica-cli.md) — comandos, workflows y modelo de estados
- [Contrato del framework](./docs/3-sdd-framework-spec-doc.md) — escenarios, fases y artefactos
- [`docs/`](./docs) — diseño, implementacion y auditoria completos

## Desarrollo

Requiere [Go 1.22+](https://go.dev/dl/).

```bash
cd src/cli
go generate ./...   # sincroniza recursos embebidos (.sdd y adapters)
go build -o sdd-cli .
go test ./...
```

## CI/CD

- **CI** (`.github/workflows/ci.yml`): valida formato, `go vet`, tests y build en cada
  push/PR hacia `main` y `development`.
- **Release** (`.github/workflows/release.yml`): al publicar un tag `v*` compila los
  binarios multiplataforma, genera `SHA256SUMS` y crea el GitHub Release.

Publicar una version:

```bash
git tag v0.1.0-beta
git push origin v0.1.0-beta
```
