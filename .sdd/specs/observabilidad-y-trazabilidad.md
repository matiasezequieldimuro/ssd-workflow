# Spec: Observabilidad y trazabilidad

## Propósito

Mantener un historial auditable e inmutable de cada transición y decisión sobre un work item, permitir la idempotencia de operaciones de agentes y exponer observabilidad de uso de tokens.

## Comportamiento observable

### Historial de eventos (`events.jsonl`)

- Cada mutación del work item appends exactamente un evento a `events.jsonl`.
- El archivo es append-only; nunca se sobrescribe ni edita.
- Cada evento incluye: ID único, tipo, work_item_id, revisión del item, actor, timestamp y payload específico del tipo.
- El archivo vive junto al `manifest.yaml` en `.sdd/work-items/active/<id>/events.jsonl`.
- Al archivar, el `events.jsonl` se mueve junto al expediente y queda inmutable.

### Eventos custom (`sdd-cli record-event`)

- Permite registrar eventos semánticos sin alterar el estado de las fases.
- Casos de uso: resultados de validación, notas de observabilidad, decisiones no mapeadas a transiciones estándar.
- Incrementa la revisión del work item.
- Participa del commit transaccional (atomicidad garantizada).
- Valida actor y schema del evento.
- Acepta `--operation-id` para idempotencia.

### Idempotencia (`--operation-id`)

- Cualquier comando mutador puede recibir un `--operation-id` (1–128 chars, `[a-zA-Z0-9.:_-]`).
- Antes de ejecutar la transición, la CLI verifica si ese `operation-id` ya fue aplicado.
- Si ya fue aplicado: retorna el estado persistido sin mutar nada (sin error).
- Si no fue aplicado: ejecuta la transición y persiste el `operation-id` junto al evento.
- Garantiza que un agente que pierda la respuesta de la CLI pueda reintentar con seguridad.

### Revisión optimista

- `manifest.yaml` incluye un campo `revision` (entero incremental).
- Cada escritura verifica que la revisión actual coincida con la esperada.
- Si otra escritura concurrente incrementó la revisión: error `concurrent_modification`.
- Previene que dos agentes simultáneos corrompan el estado.

### Token usage (observabilidad opcional)

- El campo `observability.token_usage` en el manifest puede registrar el uso de tokens del agente.
- Estados: `not_reported`, `partial`, `recorded`.
- Campos opcionales: `input_tokens`, `output_tokens`, `cache_read_tokens`, `cache_write_tokens`, `source`.
- Activación: `observability.token_usage: optional` en `config.yaml`.

### Trazabilidad entre work items

- `traceability.related_work_items`: lista de IDs de work items relacionados.
- `traceability.baseline_specs`: lista de rutas a specs que sirven de línea base para el work item.

## Entradas

| Entrada | Descripción |
| --- | --- |
| `--operation-id` | Clave de idempotencia para reintentar sin duplicar |
| `--type` | Tipo de evento custom (en `record-event`) |
| `--message` | Mensaje del evento custom |
| `--actor-kind`, `--actor-id` | Actor que registra el evento |

## Salidas

- Evento appended a `events.jsonl`.
- `manifest.yaml` con revisión incrementada.
- Confirmación en stdout (texto o JSON).

## Reglas de negocio

- `events.jsonl` es inmutable: solo se puede leer o appendear.
- `operation-id` debe cumplir el patrón `[a-zA-Z0-9.:_-]{1,128}`.
- Un `operation-id` ya aplicado hace la operación un no-op sin error.
- La revisión optimista aplica a todas las escrituras, incluyendo `record-event`.

## Dependencias

- `sdd-cli` binario.
- `events.jsonl` del work item.
- `manifest.yaml` del work item.
- `.sdd/schemas/event.schema.json` para validación de eventos.
