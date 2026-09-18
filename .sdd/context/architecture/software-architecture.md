# Arquitectura de software

## Resumen

SDD Workflow es una herramienta de línea de comandos de responsabilidad única: **gobernar el proceso de desarrollo**, no producir el contenido. La arquitectura sigue Hexagonal/Ports & Adapters con tres capas: dominio puro (reglas), puertos (contratos de interfaz) y adaptadores de infraestructura.

El binario embebe el contrato SDD (`.sdd/`) y los adapters (Claude Code) como recursos estáticos (`go:embed`). Esto garantiza que cualquier instalación sea autónoma y reproducible.

## Sistemas y servicios conectados

| Sistema o servicio | Propósito | Dirección | Datos intercambiados |
| --- | --- | --- | --- |
| Sistema de archivos local | Persistencia de work items, eventos y artefactos | CLI → FS | YAML (manifests), JSONL (eventos), Markdown (artefactos) |
| GitHub Releases | Distribución del binario | Release pipeline → GitHub | Binarios multiplataforma, SHA256SUMS |
| GitHub Actions | CI y release automation | Push/PR → Actions | Resultado de build, test y vet |
| Claude Code (adapter) | Agente que opera la CLI | Adapter instalado → CLI | Comandos `sdd-cli`, artefactos Markdown |

## Plataforma y ejecución

- Plataforma: Linux (incluyendo WSL2), macOS, Windows.
- Runtime: ninguno — binario Go nativo estático.
- Go version requerida para build: 1.22+.
- Infraestructura: repositorio GitHub, CI en GitHub Actions.
- Ambientes: local (desarrollo y uso); GitHub (CI y distribución de releases).
- Despliegue: `git tag v*` dispara el workflow `release.yml` que compila y publica el GitHub Release.

## Autenticación y autorización

- No requiere autenticación en runtime; opera sobre el sistema de archivos local.
- Los actors (`human`, `agent`, `cli`) se registran en eventos pero no hay verificación criptográfica de identidad.
- La integridad de artefactos externos se verifica por SHA256 al importar (`--from-artifact`).
- Los hooks de Claude Code (`protect-sdd-paths.py`, `protect-secrets.py`) previenen escrituras directas en `.sdd/` sin pasar por la CLI.

## Integraciones y comunicación

- La CLI opera sobre el FS local; no tiene API REST ni mensajería.
- Modo agente (`--json`): produce un único JSON envelope por stdout para integración con agentes.
- Modo humano (sin `--json`): texto legible por humanos.
- Exit code `0` en éxito, `1` en error; estandarizado para pipelines.

## Observabilidad y operación

- Logs: errores por stderr (modo humano) o JSON envelope (modo agente).
- Historial de eventos: `events.jsonl` en cada work item (inmutable, append-only).
- Trazabilidad: `manifest.yaml` de cada work item registra revisión, actor y timestamp de cada transición.
- Idempotencia: `--operation-id` permite reintentar sin duplicar estado.
- No hay métricas, trazas distribuidas ni alertas en la BETA.

## Seguridad y restricciones

- Path security (`src/cli/internal/infra/path_security.go`): bloquea path traversal al leer/escribir work items.
- Hooks de Claude Code: protegen `.sdd/` de ediciones manuales y evitan que se persistan secretos.
- Sin credenciales en el binario ni en el contrato.

## Pendientes y fuentes

- Pendientes: estrategia de upgrade del schema, métricas de uso, soporte multi-adapter.
- Fuentes verificadas: `src/cli/`, `.sdd/workflows/`, `CLAUDE.md`, `docs/CLI.md`, `docs/SDD_WORKFLOW.md`.
