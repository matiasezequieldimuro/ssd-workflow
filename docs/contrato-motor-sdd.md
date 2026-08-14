# Contrato del motor SDD

> Estado: propuesta de contrato v0.1.  
> Alcance: estructura, datos, estados y reglas del motor. No define todavía CLI, agentes, skills, comandos, MCPs ni adaptadores ejecutables.

## 1. Decisiones de diseño

El framework es un **motor de proceso SDD** y no un conjunto de prompts. OpenSpec es una referencia para organizar cambios en artefactos versionados; este contrato agrega el control determinista de transiciones, gates humanos y evidencia operativa que el harness necesita.

Principios no negociables:

1. **Portable por defecto.** El contrato vive en `.sdd/` y no contiene sintaxis de Codex, Claude Code, Copilot u otro agente. Un adaptador sólo transforma la interacción de su plataforma en operaciones sobre este contrato.
2. **Git y Markdown son la fuente de verdad.** Los documentos legibles y los manifiestos se versionan con el código. Una memoria vectorial/episódica futura (por ejemplo, Engram) sólo indexa, resume o enlaza esta información: nunca decide el estado del proceso.
3. **El motor, no el prompt, gobierna.** El modelo puede proponer contenido y solicitar una acción; una futura CLI validará precondiciones y actualizará el manifiesto. No se confía en que el modelo recuerde gates desde el chat.
4. **Un cambio, un directorio.** Todo el contexto específico y la evidencia de una feature, CR o bug se conserva junto bajo un `work item` identificable.
5. **Humano en los compromisos.** Requerimientos, especificación y plan generados requieren aprobación explícita. Implementar, archivar o efectuar efectos externos queda bloqueado hasta que el manifiesto habilite la operación y exista la autorización requerida.
6. **Pequeño núcleo, extensible por datos.** Workflows, artefactos y validaciones se declaran en YAML/Markdown; no deben quedar codificados en agentes ni prompts.
7. **Evidencia antes que telemetría.** La auditoría mínima funciona sin proveedor, modelo, MCP ni estimación de tokens. Esos datos son campos opcionales de eventos, no dependencias del flujo.

## 2. Vocabulario del contrato

| Término | Definición |
| --- | --- |
| **Workflow** | Plantilla declarativa y versionada que define fases, dependencias, gates y artefactos de una clase de trabajo. |
| **Work item** | Instancia concreta de un workflow: una feature, CR, cambio FAST o bug. Tiene estado propio. |
| **Fase** | Unidad gobernada del workflow. Puede producir evidencia, requerir aprobación o habilitar una acción. |
| **Artefacto** | Archivo Markdown o dato estructurado que prueba el resultado de una fase. |
| **Gate** | Decisión explícita que habilita la transición posterior. Los gates humanos los resuelve una persona identificada. |
| **Transición** | Cambio válido de estado de una fase o work item. Sólo el motor la realiza. |
| **Evento** | Registro inmutable, append-only, de una acción observada o una transición. |
| **Adaptador** | Integración específica de un agente; nunca es autoridad sobre estado ni duplica un workflow. |
| **Capacidad** | Operación reutilizable que un adaptador puede ofrecer. El contrato la anuncia mediante un registro neutral; no usa “skill” como concepto del núcleo. |

## 3. Estructura canónica de un proyecto

```text
mi-proyecto/
├── AGENTS.md                         # Reglas generales, redactadas de forma neutral
├── .sdd/
│   ├── config.yaml                    # Identidad y defaults del proyecto
│   ├── schemas/                       # JSON Schema del contrato, para validación determinista
│   │   ├── workflow.schema.json
│   │   ├── work-item.schema.json
│   │   └── event.schema.json
│   ├── workflows/                     # Workflows declarativos reutilizables
│   │   ├── feature-standard.workflow.yaml
│   │   ├── change-request.workflow.yaml
│   │   ├── fast-change.workflow.yaml
│   │   ├── bug-known-cause.workflow.yaml
│   │   └── bug-investigation.workflow.yaml
│   ├── templates/                     # Plantillas Markdown de artefactos
│   │   ├── prd.md
│   │   ├── change-request.md
│   │   ├── specification.md
│   │   ├── plan.md
│   │   ├── issue.md
│   │   ├── exploration.md
│   │   ├── implementation-report.md
│   │   └── verification-report.md
│   ├── procedures/                    # Cómo crear/revisar artefactos (instrucciones portables)
│   │   ├── generate-prd.md
│   │   ├── generate-specification.md
│   │   ├── create-plan.md
│   │   ├── implement.md
│   │   ├── verify.md
│   │   └── archive.md
│   ├── registry/
│   │   └── capabilities.yaml          # Capacidades neutras y sus adaptadores disponibles
│   ├── context/                       # Contexto estable, curado y de bajo volumen
│   │   ├── project.md
│   │   ├── architecture/
│   │   │   ├── software-architecture.md
│   │   │   ├── project-architecture.md
│   │   │   ├── data-modeling.md
│   │   │   └── data-and-process-flow.md
│   │   └── domain-language.md
│   ├── specs/                         # Comportamiento vigente por dominio (baseline durable)
│   │   └── <domain>/spec.md
│   ├── work-items/
│   │   ├── active/<id>/
│   │   │   ├── manifest.yaml          # Estado y trazabilidad de la instancia
│   │   │   ├── artifacts/             # Documentos de fase: PRD, spec, plan, reportes, etc.
│   │   │   ├── evidence/              # Pruebas crudas: outputs, screenshots, reportes de herramientas
│   │   │   └── events.jsonl           # Bitácora append-only local al work item
│   │   └── archive/YYYY-MM-DD-<id>/   # Work item cerrado; nunca se sobrescribe
│   └── research/                      # Investigaciones auxiliares versionadas
│       └── <topic>.md
```

`.sdd/` debe versionarse. Se excluyen de Git sólo caches, índices semánticos, secretos y telemetría centralizada. La ubicación y formato de los adaptadores quedan fuera de v0.1: se definirán al diseñar agentes, prompts, comandos y permisos de cada plataforma.

### Formato elegido

| Información | Formato | Motivo |
| --- | --- | --- |
| Configuración, workflows y manifiestos | YAML | Legible para personas y sencillo de editar/revisar. |
| Validación de datos | JSON Schema | Estándar interoperable para la futura CLI, IDEs y cualquier runtime. |
| Artefactos, procedimientos y conocimiento | Markdown con front matter YAML | Revisión natural en Git, enlaces, Mermaid y metadatos estructurados. |
| Eventos | JSON Lines (`.jsonl`) | Append-only, streaming y análisis sin reescritura de historial. |

No se usan bases de datos, estados ocultos ni formatos propietarios en v0.1.

### `specs/` vs. `work-items/<id>/artifacts/`

Ambas carpetas contienen Markdown, pero su horizonte y finalidad son diferentes:

| Ubicación | Contiene | Vida útil | Pregunta que responde |
| --- | --- | --- | --- |
| `.sdd/specs/<domain>/spec.md` | El **baseline** de comportamiento actualmente aceptado del producto, organizado por dominio. | Persistente; evoluciona al archivar cambios. | “¿Qué hace hoy el sistema?” |
| `.sdd/work-items/.../artifacts/` | El expediente de un cambio: intención, análisis, plan, implementación y verificación. Puede incluir una **specificación delta**. | Nace y se archiva con el work item; no se reescribe como conocimiento vigente. | “¿Por qué y cómo cambió el sistema?” |

Por ejemplo, `artifacts/specification.md` de un CR puede decir que agrega o modifica tres requisitos del dominio `checkout`. Al archivar el cambio, una persona o procedimiento consolida esos deltas en `.sdd/specs/checkout/spec.md`. No se copia un documento entero: el baseline refleja el estado final, mientras el work item conserva el razonamiento y la trazabilidad.

### `artifacts/` vs. `evidence/`

El **reporte de verificación sí es un artefacto** y queda en `artifacts/verification-report.md`: es el documento legible que interpreta los resultados, cubre criterios de aceptación y declara limitaciones. `evidence/` conserva los insumos verificables que respaldan ese reporte y que no conviene mezclar con la documentación narrativa: salida JUnit/JSON, logs de una ejecución, screenshots, videos, reportes de cobertura, resultados de smoke tests o links a una corrida CI.

Así, `artifacts/` es lo que una persona revisa para decidir; `evidence/` es lo que permite comprobar o reproducir esa afirmación. Para una verificación pequeña que sólo ejecuta dos comandos, `evidence/` puede quedar vacío y el reporte documenta los comandos y resultados. La carpeta no es obligatoria por sí misma.

## 4. Configuración global

`config.yaml` identifica el contrato y fija sólo los defaults que el usuario puede configurar por proyecto. Las reglas de seguridad y flujo no son switches: pertenecen al contrato y a cada workflow.

```yaml
schema_version: "0.1"
project:
  id: "checkout-service"
  name: "Checkout Service"

defaults:
  workflow: feature-standard
  artifact_language: es
  archive_policy: optional

interaction:
  mode: team-lead                 # team-lead | junior

observability:
  token_usage: optional           # Se registra cuando proveedor/adaptador lo informa
```

Los modos no cambian el workflow ni relajan gates. Sólo son una **política de interacción** del adaptador:

- `team-lead`: el adaptador puede delegar trabajo independiente y devolver resultados ejecutivos, evidencia y decisiones necesarias.
- `junior`: el adaptador prioriza explicación, pasos revisables y participación del usuario en la codificación.

El modo es global (y puede ser sobrescrito por una instrucción explícita en la conversación), no se ata al work item ni altera su estado. Toda ejecución conserva en sus eventos el adaptador/estrategia realmente utilizada cuando esa trazabilidad esté disponible.

## 5. Contrato de workflow

Un workflow declara el grafo válido. Las dependencias se expresan por IDs de fase, no por nombres de archivos ni prompts. Los artefactos obligatorios son outputs de una fase y los procedimientos explican cómo producirlos.

```yaml
schema_version: "0.1"
kind: workflow
id: feature-standard
title: "Nueva feature"
work_item_type: feature
description: "Desde necesidad de negocio hasta verificación y archive opcional."

entry_points:
  - phase: prd
    accepts: [user_prompt, prd]
  - phase: specification
    accepts: [specification]
  - phase: plan
    accepts: [plan]

phases:
  - id: prd
    produces: [prd]
    procedure: generate-prd
    approval: required
  - id: specification
    requires: [prd]
    produces: [specification]
    procedure: generate-specification
    approval: required
  - id: plan
    requires: [specification]
    produces: [plan]
    procedure: create-plan
    approval: required
  - id: implementation
    requires: [plan]
    produces: [implementation-report]
    procedure: implement
    approval: none
    effects: [repository_write]
  - id: verification
    requires: [implementation]
    produces: [verification-report]
    procedure: verify
    approval: none
  - id: human-code-review
    requires: [verification]
    produces: [code-review-record]
    approval: required
  - id: archive
    requires: [human-code-review]
    produces: [archive-record]
    procedure: archive
    approval: optional
    effects: [changelog_write, git_commit, git_push, pull_request]

artifacts:
  prd: { path: artifacts/prd.md, template: prd }
  specification: { path: artifacts/specification.md, template: specification }
  plan: { path: artifacts/plan.md, template: plan }
  implementation-report: { path: artifacts/implementation-report.md, template: implementation-report }
  verification-report: { path: artifacts/verification-report.md, template: verification-report }
  code-review-record: { path: artifacts/human-code-review.md, template: human-code-review }
  archive-record: { path: artifacts/archive.md, template: archive }
```

Reglas del grafo:

- Un workflow es acíclico y cada `id` es único, kebab-case.
- `requires` significa que la fase requerida está en `approved`, `completed` o `accepted` cuando el artefacto fue aportado por el usuario.
- `approval: required` convierte el estado de una fase generada en `awaiting_approval`; no basta que el archivo exista.
- `effects` expresa el potencial de mutación. El motor no ejecuta herramientas: una futura capa de políticas/adaptador decide permiso y registra la autorización.
- `archive` puede ser `optional`, pero si se inicia debe respetar sus dependencias y dejar evidencia. Nunca se infiere por el agente.
- El workflow no prescribe agente, modelo ni subagente. Esa elección es una estrategia de ejecución del adaptador y se registra como evento.

## 6. Workflows iniciales

| ID | Tipo | Entrada mínima | Secuencia gobernada |
| --- | --- | --- | --- |
| `feature-standard` | Feature nueva | necesidad, PRD, spec o plan | PRD → spec → plan → implementación → verificación → revisión humana → archive opcional. |
| `change-request` | Cambio de feature | cambio de requisitos o CR | CR → spec delta → plan → implementación → verificación → revisión humana → archive opcional. |
| `fast-change` | Ajuste acotado | cambio y referencia a PRD/CR | plan → implementación → verificación → revisión humana → archive opcional. |
| `bug-known-cause` | Bug con diagnóstico aportado | issue y causa/posible fix | issue → exploración breve → plan → implementación → verificación → revisión humana → archive opcional. |
| `bug-investigation` | Bug sin causa conocida | issue reportado | issue → debugging/exploración → plan → implementación → verificación → revisión humana → archive opcional. |

La `specification` de un CR debe contener deltas de comportamiento (`added`, `modified`, `removed`) y referencias a `.sdd/specs/<domain>/`. Durante archive, un procedimiento posterior podrá consolidar esos deltas en el baseline; la v0.1 sólo registra que dicha consolidación fue requerida o realizada, sin intentar automatizar merges semánticos.

Onboarding, setup, investigación y consulta son **operaciones auxiliares**, no work items de cambio por defecto. Producen artefactos bajo `context/` o `research/`; sólo se convierten en workflow cuando surjan gates/dependencias operativas reales.

## 7. Contrato del manifiesto de work item

Cada work item tiene un único `manifest.yaml`. Es la fuente de estado actual; los eventos son su historial y los archivos son la evidencia. Un proceso no debe deducir el estado leyendo títulos, checks de Markdown o historial del chat.

```yaml
schema_version: "0.1"
kind: work-item
id: "feat-023-add-coupons"
title: "Aplicar cupones en checkout"
type: feature
status: active                       # active | completed | archived | cancelled
created_at: "2026-08-13T14:30:00Z"
created_by: { kind: human, id: "matias" }

workflow:
  id: feature-standard
  version: "0.1"
  entry_phase: prd

input:
  source: user_prompt                # user_prompt | external_artifact | imported_artifact
  summary: "Permitir aplicar cupones de descuento en checkout."
  references: []

phases:
  prd: { status: awaiting_approval, artifact: artifacts/prd.md }
  specification: { status: blocked, artifact: artifacts/specification.md }
  plan: { status: blocked, artifact: artifacts/plan.md }
  implementation: { status: blocked, artifact: artifacts/implementation-report.md }
  verification: { status: blocked, artifact: artifacts/verification-report.md }
  human-code-review: { status: blocked, artifact: artifacts/human-code-review.md }
  archive: { status: blocked, artifact: artifacts/archive.md }

approvals:
  - phase: prd
    status: pending                  # pending | approved | rejected | superseded

traceability:
  events: events.jsonl
  related_work_items: []
  baseline_specs: []

observability:
  token_usage:                    # Opcional; se agrega sólo si el proveedor lo informa
    status: not_reported           # not_reported | partial | recorded
    source: null                   # p. ej. provider_usage | adapter_estimate
    input_tokens: null
    output_tokens: null
    cache_read_tokens: null
    cache_write_tokens: null
```

Estados permitidos para una fase:

```text
blocked → ready → in_progress → awaiting_approval → approved → completed
                                    │                  │
                                    └── rejected ──────┘

Entrada humana: accepted → completed
                         \→ awaiting_approval (si el workflow exige revisión)
```

`completed` es para una fase sin gate o para una fase aprobada cuyo efecto ya fue aplicado. `approved` conserva la distinción esencial entre “el humano aceptó el plan” y “la implementación terminó”. Una corrección sustancial de un artefacto aprobado lo marca `superseded`, invalida dependientes configurados por el motor y requiere una nueva aprobación; no se permite editar silenciosamente un plan aprobado.

Al iniciar desde un artefacto que aporta el usuario, sólo son elegibles los `entry_points` declarados por el workflow. La fase queda `accepted` y se registra su procedencia y hash; las fases anteriores se marcan `not_applicable` con un evento `phase_bypassed_by_external_input`. No existe un comando genérico de “skip phase”.

## 8. Contrato mínimo de artefacto

Todo Markdown de un work item comienza con front matter. El cuerpo sigue la plantilla asociada, pero el front matter permite trazabilidad automática.

```markdown
---
schema_version: "0.1"
kind: artifact
id: plan
work_item: feat-023-add-coupons
phase: plan
status: draft
created_at: "2026-08-13T15:10:00Z"
created_by: { kind: agent, id: planner }
sources:
  - artifacts/specification.md
---

# Plan de implementación

<!-- Contenido definido por templates/plan.md -->
```

Las plantillas deben pedir lo mínimo útil para tomar decisiones y verificar resultados, no replicar información que ya existe. Como línea base:

- PRD/CR: problema, objetivo, alcance/no alcance, reglas, criterios de aceptación y riesgos.
- Especificación: comportamiento verificable, escenarios y, para CR, deltas respecto al baseline.
- Plan: pasos pequeños y ordenados, archivos/áreas afectadas, estrategia de pruebas, riesgos y rollback cuando aplique.
- Issue/exploración: síntoma, impacto, evidencia, causa raíz conocida o hipótesis y límites de la investigación.
- Reporte de implementación: cambios realmente realizados, desviaciones aprobadas y referencias a código/tests.
- Verificación: comandos/entorno, resultados, cobertura de criterios y limitaciones conocidas.

## 9. Eventos y observabilidad

Cada línea de `events.jsonl` es un JSON válido. El formato es estable, extensible y no exige métricas que un proveedor no exponga.

```json
{"schema_version":"0.1","id":"evt_01J...","at":"2026-08-13T15:20:10Z","work_item":"feat-023-add-coupons","type":"phase.transitioned","actor":{"kind":"cli","id":"sdd"},"data":{"phase":"prd","from":"in_progress","to":"awaiting_approval"},"correlation_id":"run_01J..."}
```

Tipos iniciales: `work_item.created`, `artifact.created`, `artifact.updated`, `phase.transitioned`, `approval.requested`, `approval.recorded`, `validation.completed`, `authorization.recorded`, `archive.completed` y `failure.recorded`.

Campos opcionales de `actor` o `data`: adaptador, agente/rol, modelo, proveedor, duración, tokens de entrada/salida/cache, comandos ejecutados, hashes de artefactos, resultado de validación y referencias externas. Si el proveedor expone consumo, cada evento puede registrar el detalle y el bloque `observability.token_usage` del manifiesto conserva un agregado opcional. Los secretos, prompts completos con datos sensibles y credenciales no se guardan en el evento.

## 10. Procedimientos, skills y registro de capacidades

Un **procedimiento** no es una skill: es el manual portable y canónico de una operación (`.sdd/procedures/create-plan.md`). Define objetivo, precondiciones, entradas, pasos, outputs, controles y criterio de terminación. Cualquier persona o agente puede seguirlo aunque no haya adaptador instalado.

Una **skill** es la materialización de ese procedimiento en una plataforma concreta. Puede incluir el prompt, metadata, permisos y forma de invocación que Codex, Claude Code u otra herramienta requiera. Se diseñará después, dentro del adaptador; debe referenciar y no reescribir el procedimiento.

El **registro de capacidades** es el índice pequeño que conecta nombres estables con procedimientos, entradas y salidas. Resuelve “¿qué operaciones están disponibles para este harness?” y permite a un orquestador cargar sólo la guía necesaria. No contiene el cuerpo de prompts, no es una skill y no ejecuta nada. En la práctica será la versión neutral de tu idea de *skill registry*.

El contrato puede iniciar sin adaptadores; el registro sólo declara la capacidad portable:

```yaml
schema_version: "0.1"
capabilities:
  - id: sdd.generate-plan
    title: "Generar plan de implementación"
    procedure: procedures/create-plan.md
    inputs: [work_item_id, specification]
    outputs: [plan]
  - id: sdd.status
    title: "Consultar estado del work item"
    inputs: [work_item_id]
    outputs: [manifest, next_actions]
```

Cuando se diseñen adaptadores, se añadirá un mapeo opcional bajo cada capacidad (por ejemplo, la skill o comando de Codex y Claude Code). Mientras tanto, el procedimiento es la especificación de comportamiento de la capacidad. El registro no es un cargador de código ni ejecuta acciones; sólo resuelve nombres, inputs, outputs y documentación.

## 11. Límites de responsabilidad

| Componente | Es responsable de | No es responsable de |
| --- | --- | --- |
| Motor/CLI futura | Validar esquema, calcular siguiente fase, aplicar transiciones, registrar aprobaciones y emitir JSON. | Redactar documentos, modificar código o elegir modelo. |
| Workflow | Declarar proceso, dependencias, gates, artefactos y efectos potenciales. | Ejecutar prompts o llamar herramientas. |
| Procedimiento/template | Guiar la producción y revisión de un artefacto. | Mantener estado. |
| Adaptador | Ofrecer comandos, skills, hooks e instrucciones propias del agente. | Ser fuente de verdad, duplicar reglas o saltar gates. |
| Agente/orquestador | Reunir contexto, elegir ejecución inline/delegada, pedir la operación al motor y presentar resultados. | Autoaprobar al humano o mutar el manifiesto directamente. |
| Memoria/índice futuro | Recuperar conocimiento y sugerir referencias. | Sustituir Git, aprobar o alterar estados. |

## 12. Reglas operativas para la futura CLI

Estas reglas definen el comportamiento, aunque la CLI se implemente después:

1. Todas las operaciones mutantes validan el JSON Schema y el grafo del workflow antes de escribir.
2. Una operación fallida no deja el manifiesto a medio actualizar; el evento de fallo se agrega sólo si puede hacerse sin romper la integridad.
3. Las operaciones de consulta (`status`, `next`, `validate`) no modifican archivos y pueden devolver JSON para cualquier agente.
4. Las aprobaciones incluyen fase, identidad humana, fecha, decisión y comentario opcional. Una aprobación no puede ser emitida por un actor `agent`.
5. Acciones de efecto externo (commit, push, PR, tickets, despliegue) exigen una autorización registrada separada del gate de contenido; se delegan a la política/permisos del adaptador.
6. La verificación registra resultados reales. No se marca `completed` si no hay reporte de evidencia, aun cuando alguna prueba esté explícitamente marcada como no aplicable.
7. El archivo de un work item conserva todos sus artefactos y eventos; sólo entonces puede actualizarse el baseline de especificaciones y/o changelog.

## 13. Camino de implementación propuesto

1. Crear las carpetas vacías, `config.yaml`, cinco workflows y sus plantillas mínimas.
2. Publicar los tres JSON Schemas y tests de fixtures válidos/inválidos; esto fija el contrato antes de escribir la CLI.
3. Implementar una CLI mínima: `init`, `start`, `status --json`, `next --json`, `approve`, `validate` y `record-event`.
4. Ejercitar manualmente cada workflow con work items de ejemplo, incluidos rechazo, input externo y un intento inválido de implementar sin plan aprobado.
5. Recién entonces crear el adaptador inicial y sus capacidades; posteriormente, memoria semántica, CodeGraph, métricas avanzadas y automatización de archive.

## 14. Preguntas deliberadamente pospuestas

- El proveedor de memoria episódica/semántica y su política de sincronización.
- Autenticación, presupuestos, límites de tokens y agregación centralizada de observabilidad.
- Cómo se ejecutan pruebas UI/E2E y qué MCPs se autorizan por proyecto.
- Convenciones de commits, proveedor de pull requests e integración con Azure DevOps/GitHub.
- El catálogo de agentes, skills, comandos y permisos específicos de cada plataforma.

Posponerlas evita que el contrato portable quede contaminado por decisiones de implementación que aún no fueron validadas en uso real.
