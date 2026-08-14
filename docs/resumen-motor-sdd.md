# Resumen de investigación: motor SDD custom

## Decisión

El harness se construirá alrededor de un **motor SDD propio**, portable entre agentes y adaptable a las necesidades del proyecto. OpenSpec será una referencia útil para organizar artefactos de especificación, pero no será el núcleo del sistema.

La razón principal es que el harness necesita gobernar un proceso, no sólo documentarlo: debe conocer el estado de cada trabajo, aplicar reglas, requerir aprobaciones humanas y dejar evidencia de lo que ocurrió.

## Qué ofrece OpenSpec y por qué no alcanza

OpenSpec es un framework de Spec-Driven Development. Convierte una intención en artefactos versionados antes de modificar código: propuesta, especificación, diseño, tareas, implementación y archivo. Su propósito principal es documentar cambios mediante artefactos asociados a fases preestablecidas.

Es una buena capa de especificación porque promueve documentación clara, versionada y revisable. Sin embargo, no cubre por completo las necesidades del harness:

- No modela todos los escenarios y artefactos requeridos, como PRD, Change Request (CR), issue, exploración o reporte de implementación.
- No gestiona presupuesto ni consumo de tokens.
- No está pensado como un motor de flujo estricto con gates Human in the Loop (HITL).
- No aporta por sí mismo observabilidad operativa ni una estrategia de memoria.

La diferencia central es la siguiente:

```text
OpenSpec: documenta el cambio.
Motor SDD custom: documenta el cambio y controla qué puede ocurrir después.
```

Por ejemplo, si una feature está esperando aprobación del plan, el motor debe impedir la implementación. No alcanza con pedirle al LLM que respete esa regla: el sistema debe validarla de forma determinista.

## Modelo de conocimiento y evidencia

El harness separa tres tipos de información:

```text
Markdown/Git = evidencia y fuente de verdad
Engram        = recuperación semántica y memoria episódica complementaria
Logs/eventos  = evidencia operativa
```

Los artefactos Markdown representan documentos de fases concretas, por ejemplo PRD, CR, especificación y plan. Se versionan, commitean y pushean junto al código.

Engram complementa los documentos: permite recuperar decisiones, bugs encontrados y conocimiento episódico. No reemplaza a Git ni a los artefactos como fuente de verdad.

Los eventos registran el comportamiento operativo del flujo: qué agente actuó, cuándo, sobre qué work item, qué transición realizó y, más adelante, costos o tokens. Se pueden guardar, por ejemplo, en un archivo JSON de logs.

## Interacción y gates humanos

Todo trabajo comienza con input humano. El usuario puede proporcionar:

- una descripción de requerimiento o cambio;
- un PRD;
- un CR;
- una especificación;
- un plan.

El flujo empieza desde el artefacto que el usuario aporta. Esto evita duplicar trabajo cuando otro miembro del equipo ya preparó un documento.

Los artefactos producidos por IA requieren aprobación humana:

```text
PRD → aprobación → especificación → aprobación → plan → aprobación
→ implementación → validación → revisión humana de código → archive opcional
```

El archive es opcional; las validaciones y aprobaciones anteriores no lo son.

## Flujos y escenarios

No todos los escenarios deben modelarse con el mismo nivel de formalidad.

### Flujos principales de cambio

Estos sí requieren workflow, work item, estado y gates:

- Feature nueva.
- Change Request (CR) sobre una feature existente.
- Cambio FAST: ajuste puntual que comienza directamente desde el plan.
- Bug con causa o solución conocida.
- Bug que requiere debugging e investigación de causa raíz.

Un workflow define el proceso reutilizable. Un work item es una instancia concreta de ese workflow para una feature, cambio o bug particular.

### Escenarios auxiliares

Onboarding, investigación y consultas técnicas o funcionales son escenarios auxiliares. Pueden delegar trabajo, pero no deben forzarse dentro del pipeline principal `PRD → especificación → plan → implementación`.

En general, una skill con un template Markdown es suficiente para estos casos. Por ejemplo, una investigación puede generar un `research.md` con fuentes, hallazgos, riesgos y recomendación. Un workflow sólo vale la pena cuando aparecen fases, gates, bifurcaciones o reglas operativas relevantes.

## Componentes técnicos

La estructura propuesta separa el contrato portable del harness de las integraciones específicas de cada agente:

```text
sdd/
  workflows/              # Procesos reutilizables.
  templates/              # PRD, spec, plan, research, etc.
  work-items/             # Instancias reales de workflows.
  knowledge/              # Dominio, arquitectura, decisiones e investigaciones.
  procedures/             # Instrucciones portables para generar y validar.
  schema/                 # Formatos de manifest y eventos.

adapters/
  codex/
    prompts/
    skills/
    hooks/
    install-or-sync.md
  claude-code/
    commands/
    skills/
    hooks/
  copilot/
    prompts/
    instructions/

bin/
  sdd                     # CLI portable del motor.
```

El directorio `sdd/` contiene el contrato común. Los adaptadores no duplican workflows ni reglas: sólo traducen la interacción propia de cada agente hacia ese contrato.

## Workflows, manifests y CLI

Un workflow es un template declarativo de proceso. Por ejemplo, `feature-standard` puede definir que una feature produce PRD, especificación y plan, y que cada artefacto necesita aprobación antes de continuar.

Un manifest es la instancia de ese template para un trabajo concreto. Registra, como mínimo:

- identificador y tipo de work item;
- workflow seleccionado;
- fase y estado actuales;
- artefactos y aprobaciones;
- próxima acción permitida.

Ejemplo simplificado de contrato de manifest:

```yaml
schema_version: "1.0"
id: "FEAT-023-add-coupons"
title: "Aplicar cupones en checkout"
type: feature
status: awaiting_prd_approval

input:
  source: user_prompt
  summary: "Permitir aplicar un cupón de descuento en checkout."

workflow:
  profile: feature-standard
  entry_phase: prd
  current_phase: prd

phases:
  prd:
    artifact: prd.md
    status: generated
    approval: required
  specification:
    artifact: specification.md
    status: blocked
    requires:
      - prd.approved
    approval: required
  plan:
    artifact: plan.md
    status: blocked
    requires:
      - specification.approved
    approval: required
  implementation:
    status: blocked
    requires:
      - plan.approved

traceability:
  events: events.jsonl
```

El contrato puede empezar con menos campos. La idea esencial es que el manifest permita saber, sin depender del contexto del chat, qué trabajo es, en qué estado está, qué evidencia existe y cuál es la próxima acción válida.

La CLI es la pieza que valida determinísticamente las transiciones. El LLM hace el trabajo cognitivo —analiza, escribe artifacts, implementa y valida—, mientras que la CLI aplica reglas mecánicas:

```text
LLM: propone generar o aprobar una fase.
CLI: verifica si la transición es válida y actualiza el estado.
```

La CLI debe devolver resultados estructurados, preferentemente JSON, para que cualquier agente pueda interpretar el estado sin ambigüedad. Ejemplos futuros:

```bash
sdd start --type feature --title "Cupones en checkout"
sdd status FEAT-023
sdd approve FEAT-023 --phase prd
sdd next FEAT-023 --json
```

No es necesario construir una CLI compleja desde el primer día. Una versión inicial puede limitarse a crear work items, consultar estado y registrar aprobaciones.

## Delegación

Delegar a subagentes no es una regla automática. Sólo conviene cuando existe trabajo independiente, riesgo o complejidad suficiente para compensar el costo de rehidratar contexto.

Las consultas pequeñas pueden resolverse inline. Investigación, planificación, implementación o revisión pueden delegarse cuando el contexto fresco y la separación de responsabilidades aporten valor.

## Instalación en proyectos nuevos

Este repositorio funciona como template del harness. Al iniciar un proyecto nuevo, se instala una instancia mínima del contrato y sólo los adaptadores de los agentes que se usarán:

```text
mi-proyecto/
  src/
  tests/
  .sdd/
    workflows/
    templates/
    procedures/
    knowledge/
    work-items/
  .codex/                  # Sólo si se usa Codex.
  AGENTS.md                # Reglas generales neutrales.
```

La carpeta `.sdd/` se versiona dentro del proyecto porque contiene conocimiento, artefactos y procesos de ese proyecto. Los adaptadores específicos se instalan según necesidad; no hace falta copiar configuraciones de agentes que no se usarán.

## Próximos pasos

La implementación se realizará por capas, validando una base utilizable antes de sumar complejidad:

1. **Definir el contrato.** Reunir notas y documentar en detalle los documentos, fases y componentes del flujo. El objetivo es dejar claras las etapas, reglas y artefactos antes de implementar el motor.

2. **Definir los componentes esenciales del flujo.** Con el contrato definido, diseñar los agentes, skills esenciales, settings de agentes (permisos y accesos) y MCPs básicos. El alcance inicial es el funcionamiento del SDD, no skills de dominio o tecnología específicos —por ejemplo, desarrollo CAP SAP BTP—. Se busca una base funcional para iterar, no una solución perfecta.

3. **Construir y probar una CLI pequeña.** Con contrato y componentes definidos, crear una CLI capaz de administrar los artefactos SDD. Primero se probará directamente por línea de comandos para verificar que crea, consulta y actualiza correctamente los work items y sus artifacts.

4. **Integrar el agente principal.** Instruir al agente principal para usar la CLI y realizar pruebas reales del flujo. La CLI sigue siendo la autoridad de estado; el agente la utiliza para operar el proceso.

5. **Agregar capacidades avanzadas.** Cuando el motor y sus pruebas reales funcionen correctamente, incorporar Engram, CodeGraph y skills/MCPs adicionales.

La meta no es automatizar todo desde el inicio, sino construir un núcleo confiable que pueda evolucionar sin quedar acoplado a un agente o herramienta particular.

---

Created on: 2026-08-10  
Last modified: 2026-08-10
---