# Spec: Ciclo de vida de un work item

## Propósito

Permitir crear, avanzar, aprobar y cerrar una unidad de trabajo (work item) de forma gobernada, trazable y sin requerir memoria del chat. El motor determinista (`sdd-cli`) impone las reglas; el agente o persona solo puede mover el proceso a través de las transiciones válidas.

## Comportamiento observable

### Creación (`sdd-cli start`)

- Se crea un `manifest.yaml` bajo `.sdd/work-items/active/<id>/`.
- La primera fase del workflow (o la de entrada declarada con `--phase`) arranca automáticamente en `in_progress`.
- Si se usa `--from-artifact`, las fases anteriores a la entrada se marcan `not_applicable` y el artefacto externo se importa (SHA256 verificado).
- El work item nace con `status: active` y `revision: 1`.

### Comienzo de fase (`sdd-cli begin`)

- Transición `ready | rejected | superseded → in_progress`.
- Una fase `blocked` no puede comenzar.

### Entrega de fase (`sdd-cli deliver`)

- Transición `in_progress →`:
  - `awaiting_approval` si `approval: required`.
  - `awaiting_approval` si `approval: optional` y se usa `--request-approval`.
  - `completed` si `approval: optional` sin flag.
  - `completed` si `approval: none`.
  - Error si `approval: none` y se usa `--request-approval`.
- Al completar una fase, las dependientes cuyos requisitos quedan satisfechos pasan de `blocked → ready`.
- Los artefactos de las fases desbloqueadas se preparan a partir de sus templates.

### Aprobación humana (`sdd-cli approve`)

- Solo válida si la fase está `awaiting_approval` y su política no es `none`.
- El actor debe ser `kind: human`.
- Transición: `awaiting_approval → approved`.
- Desbloquea fases dependientes satisfechas.
- Se registra quién, cuándo y el comentario en `approvals[]`.

### Rechazo humano (`sdd-cli reject`)

- Mismas precondiciones que `approve`.
- Transición: `awaiting_approval → rejected`.
- No desbloquea dependientes.
- El retrabajo se reabre con `begin`; la nueva entrega crea otra entrada en `approvals[]` sin borrar el rechazo previo.

### Completar fase (`sdd-cli complete --phase`)

- Normaliza `approved | accepted → completed`.
- No es necesario para desbloquear dependencias (ya las satisfacen en `approved/accepted`); sirve para hacer explícito que la fase fue aplicada.

### Cerrar work item (`sdd-cli complete`)

- Sin `--phase`: cierra el work item si todas las fases obligatorias están satisfechas.
- Las fases opcionales no iniciadas pueden omitirse; las iniciadas deben haber terminado.
- Transición: `WorkItem.status: active → completed`.

### Archivar (`sdd-cli archive`)

- Precondiciones: work item `completed`, fase `archive` satisfecha (si el workflow la declara), `validate <id>` sin failures, `artifacts/archive.md` válido, destino ausente.
- Resultado: renombra `active/<id>/` → `archive/YYYY-MM-DD-<id>/` (operación atómica).
- El expediente archivado queda inmutable; `status` y `validate` resuelven su ubicación.

## Entradas

| Entrada | Descripción |
| --- | --- |
| `id` | Identificador del work item (`^[a-z0-9]+(?:-[a-z0-9]+)*$`) |
| `--title` | Título legible (requerido en `start`) |
| `--workflow` | Workflow a usar (default: `config.yaml`) |
| `--summary` | Resumen del input |
| `--from-artifact` + `--phase` | Importar artefacto externo como punto de entrada |
| `--phase` | Fase objetivo de la transición |
| `--by` | Actor humano (en `approve`/`reject`) |
| `--comment` | Comentario de aprobación/rechazo |
| `--request-approval` | Solicitar aprobación en fase `optional` |
| `--actor-kind`, `--actor-id` | Actor que ejecuta la operación |
| `--operation-id` | Clave de idempotencia |

## Salidas

- `manifest.yaml` actualizado con la nueva revisión, estado y aprobaciones.
- Evento appended a `events.jsonl`.
- Artefactos de fases desbloqueadas creados desde template.
- Salida en stdout (texto o JSON envelope según `--json`).

## Reglas de negocio

- Solo un actor `human` puede aprobar o rechazar.
- Una fase `blocked` no puede comenzar.
- La política `none` no permite `--request-approval`.
- El work item debe estar `active` para mutar fases (excepción: fases opcionales con status `completed`).
- La revisión optimista previene escrituras concurrentes (`concurrent_modification`).
- `--operation-id` hace cualquier operación idempotente en reintentos.

## Dependencias

- `sdd-cli` (binario).
- `.sdd/config.yaml` para defaults de workflow.
- `.sdd/workflows/<id>.workflow.yaml` para definición de fases.
- `.sdd/templates/` para generación de artefactos.
- Sistema de archivos local (`.sdd/work-items/`).
