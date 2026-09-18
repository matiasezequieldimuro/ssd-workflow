# Spec: Consultas de estado del work item

## Propósito

Permitir que agentes y personas obtengan el estado actual del proceso sin mutarlo, para orientar la próxima acción, diagnosticar problemas y validar consistencia.

## Comportamiento observable

### `sdd-cli status <id>`

- Devuelve el estado general del work item: ID, tipo, status (`active`/`completed`/`archived`), workflow, revisión, actor creador, timestamp.
- Devuelve el estado de cada fase: status, ruta del artefacto.
- En modo texto: fases ordenadas topológicamente.
- En modo JSON (`--json`): fases como mapa (sin orden garantizado).
- Resuelve la ubicación del work item: busca en `active/` primero, luego en `archive/`.
- Incluye el historial de aprobaciones.

### `sdd-cli next <id>`

- Informa la próxima acción a tomar, priorizando:
  1. Fase con `awaiting_approval`.
  2. Fase con `in_progress`.
  3. Fase con `ready`.
- Devuelve: phase ID, status, procedure de referencia, ruta del artefacto, si requiere aprobación.
- Si no hay próxima acción: devuelve vacío (work item terminado o sin fases habilitadas).
- No comienza la fase ni incrementa la revisión.

### `sdd-cli validate [id]`

**Sin ID (proyecto completo):**
- Valida: config, schemas, workflows, templates, procedures y todos los work items activos.

**Con ID (work item específico):**
- Valida: manifest, workflow referenciado, coherencia de aprobaciones, artefactos referenciados existentes, eventos y referencias cruzadas.

**Salida de cada check:**
- `status`: `passed` | `warning` | `failed`.
- `category`: categoría del check.
- `code`: código de error específico.
- `target`: elemento validado.
- `message`: descripción legible.

**Exit codes:**
- `0`: sin failures (puede tener warnings).
- `1`: al menos un check `failed`.

Las advertencias no cambian el exit code. `validate` es una consulta pura: no crea locks, no registra eventos, no incrementa revisión.

## Entradas

| Comando | Entrada | Descripción |
| --- | --- | --- |
| `status` | `<id>` | ID del work item |
| `next` | `<id>` | ID del work item |
| `validate` | `[id]` | ID opcional; sin ID valida todo el proyecto |
| Todos | `--json` | Formato de salida JSON |
| Todos | `--dir` | Directorio del proyecto |

## Salidas

- `status`: estado completo del work item y sus fases.
- `next`: próxima fase a atender con su procedure y artefacto.
- `validate`: lista de checks con status, category, code, target y message.

## Reglas de negocio

- Ninguna consulta muta estado, registra eventos ni incrementa revisión.
- `status` y `validate` resuelven work items archivados (buscan en `archive/`).
- `validate` con failures devuelve exit code `1`; `archive` requiere `validate <id>` sin failures como precondición.

## Dependencias

- `sdd-cli` binario.
- `manifest.yaml` del work item.
- `.sdd/workflows/` para resolución de topología de fases.
- `.sdd/schemas/` para validación de schemas.
- Sistema de archivos local.
