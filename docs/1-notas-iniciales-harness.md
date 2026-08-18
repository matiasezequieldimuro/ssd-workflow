# SDD Harness — Notas & Diseño

## Sobre el flujo SDD

Los puntos críticos que constituyen un buen harness son la definición de **procesos**, **estructura** y mecanismos de **control**. Esto es aplicar **ingeniería**.

---

## Interacción humana dentro del flujo SDD

El desarrollador piensa con criterio y plasma los requerimientos e ideas de implementación. Con ayuda de la IA, se **formaliza**.

---

## Componentes del flujo SDD

- Skill de **orquestación de flujo**.
- Un agente/skill **orquestador** se encarga de enrutar el trabajo (routeo).
- Decisión de mecanismo de **delegación**: inline, vanilla fresh-context, SDD.
- **Artefactos** (PRDs, por ejemplo) y **memoria** como fuentes de conocimiento.
- **Etapas** obligatorias bien definidas y ordenadas, con manejo de dependencias.
- Sistema de **monitoreo** y **logging** → qué se hizo, cuándo, cómo, cuánto consumió.

---

## Ideas para el flujo SDD

- Comando/Skill de **onboarding** → convierte un repo desconocido en entorno operativo.
- **Engram** como memoria → recordar decisiones, bugs, lineamientos → ahorrar tiempo y tokens.
- Fase de **verificación** obligatoria → Tests = Ingeniería.
- **Skill Registry** → cada agente carga las cosas de distinta manera; esto lo soluciona → el orquestador digiere el registry y pasa un resumen a los subagentes.
- **Subagentes** → Contexto Limpio.
- **TDD** → funciona bien solo si hay requerimientos claros y bien definidos (bottom-up).
- **Open-Spec** para guardar artefactos (documentación) vs **Engram** para memoria (decisiones, bugs).
- Implementación de **doble memoria** (por ejemplo, archivos `.md` + Engram).

---

## Notas del flujo SDD

- **CLAUDE.md** → contexto del proyecto y normas generales → en qué consiste el proyecto, cómo trabajamos.
- **Skills** → manual de cómo hacer algo (si lo escribí 2 veces o más, es reutilizable) → forma estandarizada de hacer cosas, como un manual para compañeros o empleados.
- **Subagentes** → contratar a una persona nueva: darle una identidad, explicarle brevemente qué espero de ella (QUÉ), mas no los detalles profundos (CÓMO — que consulte la skill).

---

## Dudas del flujo SDD

- ¿Conviene un agente definido específicamente o construir, en su lugar, una skill que spawnee un agente default?
- ¿Convienen comandos o skills para prompts reutilizables por mí mismo?

---

## Beneficios del Harness

- Framework/Sistema/Flujo que se puede usar en cualquier agente (a pesar de que cada uno tenga sus particularidades y caprichos).

---

## Cómo pensar mi flujo SDD

- ¿Qué **escenarios** de trabajo puedo automatizar? (aunque sean las tareas más simples del día a día, en desarrollo, testing, documentación, etc.) ¿Cómo lo modelo en un **proceso** con reglas y **etapas**?
- ¿Necesito **subagentes** para mi proceso? ¿Necesito **skills**? ¿Necesito **MCPs**?
- ¿Cómo controlar y **monitorear** mis procesos (cuánto se gastó, qué se hizo)? Podría ser memoria (Engram), documentación (artifacts) o mensajes de subagentes (archivo de logs).
- ¿Cómo genero mi **base de conocimientos** compartida (tipo Drive, SharePoint)?
- ¿Cómo **restringir** accesos y permisos (`settings.json`)? ¿Cómo establecer **límites** y scope?
- ¿Cómo garantizar que cada etapa se ejecuta correctamente y en el orden especificado?

---

## Ideas puntuales para el flujo SDD

### PR — Peer Review

- No más de **400 líneas de código** por PR; si supera, separar.

### PR Reviewer / Review PR

- Para no quemarse analizando PRs del equipo.
- Detectar **RIESGOS** (seguridad, componentes sensibles) y **LEGIBILIDAD** (estructura, separación de capas, SRP —Single Responsibility Principle—, etc.).

---

## Otras notas del flujo SDD

### 4 Tipos de Memoria

Todo sistema agéntico debe contar con los siguientes tipos de memoria:

| Tipo | Descripción | Implementación |
|---|---|---|
| **Corto plazo** | Información temporal de las tareas realizadas en el momento | Context window |
| **Semántica** | Conocimiento general, ideas clave, referencias, guías | Archivos Markdown |
| **Procedural** | Reglas y how-to de tareas específicas (potencialmente repetitivas) | Skills |
| **Episódica** | Decisiones tomadas, issues/bugs encontrados | Engram, docs Markdown, logs JSON |

> Fuente: [SWE Best Practices](https://docs.google.com/document/d/1CZEkcbBHnKSoBJ1kZ3rlc2YtqyzjugkBj37zywXiNKY/edit?tab=t.tjfpgxcrs2lr#heading=h.8m4197r8ds2s)

---

### ¿Por qué fallan los agentes?

#### Loop Infinito

- **Falta de definición de terminación**: usar `maxRetries` y/o instrucciones que definan cuándo y en qué casos terminar la tarea.
- **Falta de trackeo de actividad**: monitorear qué hizo el agente y tener capacidad de registrar todas las operaciones y decisiones tomadas.

#### Alucinación

- **Mala definición de tools**: indicar correcta y explícitamente cuándo y cómo usar cada herramienta.
- **Falta de validación**: implementar Human In The Loop o un agente verificador.
- **Límites pobres**: indicar qué puede hacer el agente y qué no; prohibir asumir cosas sin información clara.

#### Uso indebido de herramientas

- **Tools over-privileged**: aplicar principio de *least privilege* y separar herramientas en lectura/escritura.
- **No approval workflow**: falta de confirmación del usuario en decisiones críticas.

> Fuente: [SWE Best Practices](https://docs.google.com/document/d/1CZEkcbBHnKSoBJ1kZ3rlc2YtqyzjugkBj37zywXiNKY/edit?tab=t.g2i8j5isoexs)

---

### Los fundamentos importan más que nunca

- *"Bad code is the most expensive it's ever been."*
- La calidad de las respuestas de un LLM es, en gran parte, proporcional a la calidad de la codebase sobre la que trabaja.
- Mantener un **documento del dominio** (glosario o guía funcional) con conceptos, reglas de negocio y terminología.
- *"Always take small, deliberate steps. The rate of feedback is your speed limit."*
- *"Think and design the interface, delegate the implementation."*

> Fuente: [SWE Best Practices](https://docs.google.com/document/d/1CZEkcbBHnKSoBJ1kZ3rlc2YtqyzjugkBj37zywXiNKY/edit?tab=t.52xqoq1yv5pc)

---

### Estructura de un buen prompt

1. **Rol y tarea**
2. **Contexto**
3. **Instrucciones**
4. **Ejemplos** *(opcional)*
5. **Reglas Críticas**

> Fuente: [SWE Best Practices](https://docs.google.com/document/d/1CZEkcbBHnKSoBJ1kZ3rlc2YtqyzjugkBj37zywXiNKY/edit?tab=t.0#heading=h.qwrpddcqgfnr)

---

### Token Poker

Será una habilidad valiosa poder estimar la cantidad de tokens que puede demandar una feature o proyecto antes de ejecutarla.

> Fuente: [Predicciones & Notas & Reflexiones sobre AI](https://docs.google.com/document/d/1BPHMPOJUrTQfb8s_OcUGKgbeQyE3PIIIjadSRJnhfCA/edit?tab=t.y7ueb0a4gaye#heading=h.mhbkc9ggeozo)

---

Created on: 2026-08-03 
Last modified: 2026-08-04
