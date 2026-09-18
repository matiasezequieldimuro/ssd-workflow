# Spec: Workflows predefinidos

## Propósito

Proveer templates de proceso reutilizables para los casos de entrega más comunes: nueva feature, cambio de requerimientos, bugs con y sin causa conocida, y ajustes rápidos. Cada workflow define las fases, dependencias, gates y artefactos de ese tipo de trabajo.

## Comportamiento observable

Todos los workflows comparten la misma estructura terminal (`implementation → verification → human-code-review → archive?`) y se diferencian en las fases iniciales.

### `feature-standard` — Nueva feature completa

Fases: `prd → specification → plan → implementation → verification → human-code-review → archive?`

- Inicia desde: prompt de usuario, PRD existente, spec existente o plan existente.
- Gates humanos obligatorios: `prd`, `specification`, `plan`, `human-code-review`.
- `archive` es opcional.
- Uso: funcionalidad nueva que requiere validar necesidad de negocio, especificación técnica y plan antes de implementar.

### `change-request` — Cambio de requerimiento

Fases: `change-request → specification → plan → implementation → verification → human-code-review → archive?`

- Inicia desde: prompt, change-request existente, spec o plan.
- Gates humanos obligatorios: `change-request`, `specification`, `plan`, `human-code-review`.
- Diferencia con `feature-standard`: la primera fase es un Change Request (delta sobre comportamiento existente) en lugar de un PRD.

### `bug-investigation` — Bug con causa desconocida

Fases: `issue → debugging → plan → implementation → verification → human-code-review → archive?`

- Inicia desde: prompt, issue existente o plan.
- Gates humanos obligatorios: `plan`, `human-code-review`.
- Las fases `issue` y `debugging` no requieren aprobación; permiten al agente documentar y explorar libremente antes de proponer un plan.

### `bug-known-cause` — Bug con causa conocida

Fases: `issue → exploration → plan → implementation → verification → human-code-review → archive?`

- Idéntico a `bug-investigation` en estructura.
- La distinción semántica es que el reporte ya incluye una causa probable; la fase `exploration` confirma antes de planificar.

### `fast-change` — Ajuste rápido

Fases: `plan → implementation → verification → human-code-review → archive?`

- Inicia desde: prompt o plan existente.
- Gates humanos obligatorios: `plan`, `human-code-review`.
- Uso: cambio acotado con input claro que no requiere PRD ni spec elaborada.

## Estructura común de fases terminales

| Fase | Aprobación | Efectos declarados |
| --- | --- | --- |
| `implementation` | `none` | `repository_write` |
| `verification` | `none` | — |
| `human-code-review` | `required` | — |
| `archive` | `none` (opcional) | `changelog_write`, `git_commit`, `git_push`, `pull_request` |

## Entradas

- Un workflow se selecciona al crear un work item con `--workflow <id>` o desde el default en `config.yaml`.
- Los entry points de cada workflow definen qué fases permiten inicio con `--from-artifact` y qué tipos de artefacto aceptan.

## Salidas

- Work item creado con el conjunto de fases del workflow seleccionado.
- Cada fase genera su artefacto correspondiente (plantilla Markdown) al desbloquearse.

## Reglas de negocio

- Un work item está ligado a un workflow al crearse; no se puede cambiar.
- Los workflows son DAGs (grafos sin ciclos); la CLI los valida en `validate`.
- Los `effects` son declarativos: la CLI los registra pero no los ejecuta (responsabilidad del agente o persona).

## Dependencias

- `.sdd/workflows/*.workflow.yaml` (embebidos en el binario).
- `.sdd/templates/` para los artefactos de cada fase.
- `.sdd/procedures/` para las guías de procedure por fase.
