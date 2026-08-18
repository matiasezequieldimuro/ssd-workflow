# Auditoría técnica del codebase de la CLI SDD

> Estado: evaluación técnica del prototipo actual.  
> Alcance: `src/cli/`, su integración con el contrato `.sdd/` y la documentación vigente.  
> Objetivo: determinar si la implementación actual constituye una base suficientemente robusta para continuar desarrollando el motor determinista.

---

## 1. Contexto y objetivo de la auditoría

La CLI SDD tiene la responsabilidad de actuar como autoridad determinista del harness. Los agentes pueden analizar, redactar artefactos o solicitar acciones, pero la CLI debe decidir qué transición es válida, persistirla de forma consistente y dejar evidencia auditable.

Por esa razón, esta evaluación no se limita a comprobar que el código compile, esté separado en carpetas o produzca resultados en los escenarios felices. Se revisaron especialmente:

- cumplimiento del contrato documentado;
- integridad de manifiestos, artefactos y eventos;
- seguridad de las operaciones sobre el filesystem;
- implementación real de la máquina de estados;
- aplicación de SOLID y Clean Architecture;
- claridad e idiomaticidad del código Go;
- calidad y alcance de los tests;
- capacidad de evolución hacia nuevos comandos y workflows.

La pregunta principal de la auditoría es:

> ¿La implementación actual es una base confiable sobre la cual continuar agregando comandos como `validate`, `reject` y `archive`?

## 2. Conclusión ejecutiva

La base presenta una **buena intención arquitectónica y una modularización inicial razonable**, pero todavía no es suficientemente robusta para actuar como autoridad determinista del workflow.

La estructura `domain / ports / usecases / infra / cmd` facilita la lectura y muestra una dirección correcta. También son buenas decisiones el uso de workflows declarativos, recursos embebidos, manifiestos YAML, eventos JSONL y dependencias inyectadas en varios casos de uso.

Sin embargo, existen problemas que afectan propiedades centrales del motor:

- una operación fallida puede dejar estado persistido parcialmente;
- algunos errores críticos se ignoran y se devuelve éxito;
- IDs no validados permiten escribir fuera del directorio esperado;
- `next`, definido como consulta, modifica el manifiesto;
- la máquina de estados no implementa el ciclo completo de una fase;
- el motor interpreta workflows declarativos mediante supuestos lineales y paths hardcodeados;
- los schemas existen, pero no se ejecutan;
- las reglas importantes todavía no están protegidas dentro del dominio;
- los tests cubren principalmente caminos felices y no detectan estas inconsistencias.

Por lo tanto:

> **No se recomienda continuar directamente con los comandos pendientes. Primero debe estabilizarse el núcleo de dominio, validación y persistencia.**

La implementación actual puede conservarse y refactorizarse; no es necesario descartarla. El problema principal no es la estructura general, sino que las garantías prometidas por el contrato aún no están implementadas.

## 3. Estado actual de la implementación

### 3.1. Funcionalidades construidas

Actualmente se encuentran implementados:

- `sdd init`;
- `sdd start`;
- `sdd status`;
- `sdd approve`;
- `sdd next`;
- `sdd record-event`;
- carga de workflows YAML;
- persistencia de manifests YAML;
- registro append-only de eventos JSONL;
- creación inicial de artefactos desde templates embebidos;
- inicio desde un artefacto externo;
- cinco workflows iniciales;
- modelos Go para workflows, work items, fases, aprobaciones y eventos;
- tests de integración para un recorrido parcial y un inicio con bypass.

### 3.2. Funcionalidades pendientes documentadas

La documentación ya identifica como pendientes:

- `sdd validate`;
- integración con el agente orquestador;
- `sdd reject`;
- `sdd archive`;
- validación de event types;
- capacidades posteriores de memoria y observabilidad.

La auditoría detectó que, antes de esos puntos, también faltan capacidades internas esenciales:

- finalizar o entregar el resultado de una fase;
- pasar una fase a `awaiting_approval`;
- completar una fase sin aprobación;
- completar el work item;
- validar entry points e inputs externos;
- validar el grafo del workflow;
- persistir una operación de forma atómica;
- controlar actualizaciones concurrentes;
- garantizar que toda transición produzca eventos consistentes.

## 4. Aspectos positivos de la solución

### 4.1. Organización comprensible

La distribución por capas permite identificar rápidamente:

- modelos y errores en `internal/domain`;
- interfaces en `internal/ports`;
- lógica de aplicación en `internal/usecases`;
- filesystem y artefactos en `internal/infra`;
- adaptación a Cobra en `cmd`.

Para un primer proyecto en Go, la estructura es clara y evita concentrar toda la lógica en comandos Cobra.

### 4.2. Contrato declarativo

Los workflows, fases, dependencias, artefactos y approvals están definidos en YAML. Esta decisión es coherente con el objetivo agent-agnostic y con el principio de que el motor debe poder extenderse por datos.

El problema no está en el contrato declarativo, sino en que la implementación todavía no lo interpreta por completo.

### 4.3. Tecnologías adecuadas para el alcance

Las elecciones técnicas son razonables:

- Go para distribuir un binario portable;
- Cobra para la interfaz CLI;
- `go:embed` para entregar un template autocontenido;
- YAML para configuración legible;
- JSONL para eventos append-only;
- `internal/` para evitar exponer paquetes internos.

No se observa una necesidad actual de incorporar una base de datos o un framework de mayor complejidad.

### 4.4. Manejo de errores parcialmente correcto

Cuando los errores son propagados, normalmente se agrega contexto mediante `%w`. Esto conserva la causa original y permite utilizar `errors.Is` posteriormente.

El problema principal no es la forma de envolver errores, sino que varias operaciones importantes los descartan.

## 5. Resumen de puntos a corregir

| ID | Prioridad | Área | Hallazgo |
| --- | --- | --- | --- |
| CLI-01 | Crítica | Seguridad | Los IDs permiten path traversal y escritura fuera de `.sdd/work-items/active`. |
| CLI-02 | Crítica | Integridad | Una operación fallida puede dejar manifests y directorios parcialmente creados. |
| CLI-03 | Crítica | Errores | Se ignoran errores de workflows, artefactos y eventos, devolviendo éxitos falsos. |
| CLI-04 | Alta | Semántica | `next` modifica el estado aunque el contrato lo define como consulta. |
| CLI-05 | Alta | Dominio | La máquina de estados de fases está incompleta. |
| CLI-06 | Alta | Seguridad del proceso | La aprobación humana no está protegida como una invariante del dominio. |
| CLI-07 | Alta | Workflows | La implementación usa parcialmente el grafo, pero inicio, bypass y selección siguen dependiendo del orden del array. |
| CLI-08 | Alta | Workflows | Los paths de artefactos se hardcodean por phase ID y contradicen el workflow. |
| CLI-09 | Alta | Validación | Los JSON Schemas y reglas semánticas no se aplican. |
| CLI-10 | Alta | Persistencia | No hay escrituras atómicas ni una política definida frente a escrituras concurrentes. |
| CLI-11 | Media | Arquitectura | Los use cases dependen directamente de infraestructura concreta. |
| CLI-12 | Media | Diseño de dominio | El dominio es principalmente anémico y utiliza strings libres. |
| CLI-13 | Media | Auditoría | Los eventos no garantizan unicidad, orden lógico ni registro de todas las transiciones. |
| CLI-14 | Media | CLI | El manejo de errores y output no ofrece un contrato JSON uniforme. |
| CLI-15 | Alta | Tests | La suite no cubre casos inválidos, atomicidad, CLI ni workflows completos. |
| CLI-16 | Baja | Go y repositorio | Hay archivos sin `gofmt` y un binario de plataforma versionado. |
| CLI-17 | Media | Contrato | Los templates generados no cumplen todo el front matter documentado. |
| CLI-18 | Alta | Contrato local | La generación de artefactos ignora los templates locales de `.sdd/` y usa siempre los embebidos. |
| CLI-19 | Media | Configuración | El workflow default de `config.yaml` no se utiliza; Cobra hardcodea `feature-standard`. |

### 5.1. Revalidación de los hallazgos

Los puntos anteriores no representan todos el mismo tipo de problema. Para evitar tratar recomendaciones arquitectónicas como si fueran bugs reproducidos, se realizó una segunda revisión con ejecución manual, inspección de código y contraste contra el contrato.

| ID | Naturaleza | Resultado de la revalidación |
| --- | --- | --- |
| CLI-01 | Bug reproducible | Confirmado: un ID con `../` creó el manifest fuera de `.sdd/work-items/active`. |
| CLI-02 | Bug reproducible | Confirmado: un `start` fallido dejó un manifest persistido y consultable. |
| CLI-03 | Bug reproducible | Confirmado: `approve` devolvió éxito sin workflow y también cuando no pudo escribir el evento. |
| CLI-04 | Bug reproducible | Confirmado: `next` persistió `in_progress` pero devolvió `status: ready`. |
| CLI-05 | Bug reproducible y modelo incompleto | Confirmado: se aprobó una fase sin artifact y también una fase declarada con `approval: none`. |
| CLI-06 | Invariante ausente | Confirmado por código. La CLI Cobra fuerza `human`, pero el use case no lo valida; el riesgo aparece al reutilizarlo desde otro adaptador. |
| CLI-07 | Limitación de implementación | Confirmado con matiz: `unblockDependentPhases` sí consulta `requires`; el problema es que inicio, bypass y `next` todavía dependen del orden físico. |
| CLI-08 | Bug reproducible | Confirmado: el manifest referenció `implementation.md` mientras se creó `implementation-report.md`. |
| CLI-09 | Funcionalidad ausente | Confirmado por código y tests: los schemas no son ejecutados y los fixtures inválidos no se prueban. |
| CLI-10 | Riesgo de persistencia | Confirmada la escritura directa y la ausencia de control concurrente. No se simuló un crash ni un lost update; esas consecuencias son riesgos técnicos, no incidentes observados. |
| CLI-11 | Deuda arquitectónica | Confirmado estructuralmente. No implica por sí solo un fallo funcional actual. |
| CLI-12 | Deuda de diseño | Confirmado estructuralmente. Es una recomendación para centralizar invariantes, no un bug aislado. |
| CLI-13 | Bug reproducible | Confirmado: dos eventos consecutivos recibieron el mismo ID y el bypass quedó antes de `work_item.created`. |
| CLI-14 | Bug reproducible y deuda de diseño | Confirmado: con `--json`, un error de argumentos produjo usage y texto plano, no el envelope JSON. |
| CLI-15 | Brecha de tests | Confirmado mediante inspección de la suite y cobertura por paquetes. |
| CLI-16 | Higiene | Confirmado mediante `gofmt`, Git y tipo del binario. |
| CLI-17 | Drift documental | Confirmado. Debe aclararse que hoy no existe un artifact schema que lo detecte automáticamente. |
| CLI-18 | Bug reproducible | Confirmado: modificar `.sdd/templates/prd.md` no cambió el artifact generado. |
| CLI-19 | Funcionalidad ausente | Confirmado: `start` usa un default hardcodeado y no carga `config.yaml`. |

La revalidación no invalidó ninguno de los hallazgos principales. Sí corrigió dos posibles sobregeneralizaciones:

1. El grafo no se ignora por completo: las dependencias se consultan al desbloquear fases, aunque otras operaciones sigan siendo order-dependent.
2. La falta de locking es una carencia de diseño verificable, pero un conflicto concurrente concreto no fue reproducido durante esta auditoría.

## 6. Hallazgos detallados

### CLI-01 — Path traversal mediante IDs y paths no validados

#### Situación actual

Los repositorios construyen las rutas concatenando directamente valores externos:

```go
filepath.Join(baseDir, ".sdd", "work-items", "active", id, "manifest.yaml")
```

El mismo patrón se utiliza para eventos y artefactos.

Aunque `work-item.schema.json` declara un patrón kebab-case para el ID, el schema no se ejecuta antes de usar ese valor como parte de una ruta.

#### Evidencia

Una ejecución equivalente a:

```bash
sdd start ../../../escaped-item --title test
```

permitió crear el manifest fuera de `.sdd/work-items/active`, dentro de:

```text
<proyecto>/escaped-item/manifest.yaml
```

#### Por qué es importante

No es solamente un problema de validación funcional. Un argumento controlado por un usuario, script o agente puede hacer que la CLI escriba en ubicaciones distintas de las autorizadas.

El mismo riesgo puede aparecer en:

- work item IDs;
- workflow IDs;
- artifact paths declarados en YAML;
- IDs alterados manualmente dentro de un manifest.

#### Corrección esperada

- Validar IDs antes de cualquier operación sobre disco.
- Rechazar separadores, segmentos `..`, valores vacíos y formatos no permitidos.
- Resolver la ruta final y comprobar que permanezca dentro del directorio raíz autorizado.
- Aplicar la misma política a los paths declarados en workflows.
- No depender únicamente de JSON Schema para seguridad de rutas.

### CLI-02 — Operaciones fallidas dejan estado parcial

#### Situación actual

`StartWorkItemUseCase` realiza aproximadamente estas operaciones:

1. construye el work item;
2. guarda el manifest;
3. crea artefactos;
4. copia el artefacto externo, si corresponde;
5. registra eventos.

Si falla cualquiera de los pasos posteriores al guardado, el manifest ya existe.

#### Evidencia

Al iniciar con una fase externa inexistente:

```bash
sdd start feat-partial \
  --title test \
  --from-artifact input.md \
  --phase nonexistent
```

el comando devolvió error, pero dejó un work item persistido y consultable.

Esto contradice la regla del contrato:

> Una operación fallida no deja el manifiesto a medio actualizar.

#### Impacto

- El exit code indica fallo, pero el estado del proyecto fue modificado.
- Un reintento encuentra el ID ya utilizado.
- El usuario o agente no puede confiar en que un error signifique “no ocurrió nada”.
- Manifests, artefactos y eventos pueden quedar desincronizados.

#### Corrección esperada

La operación debe prepararse y validarse antes del commit:

1. validar todos los inputs;
2. cargar y validar el workflow;
3. calcular el estado resultante;
4. preparar artifacts y eventos;
5. persistir el conjunto de forma controlada;
6. publicar el resultado sólo cuando toda la operación haya finalizado.

Para el filesystem puede utilizarse un directorio temporal o archivos temporales más rename atómico. No es necesario incorporar una base de datos para resolver esta etapa.

### CLI-03 — Errores críticos ignorados

#### Situación actual

Existen varias operaciones cuyo error se descarta explícitamente:

```go
_ = uc.workItemRepo.AppendEvent(...)
_ = artifactMgr.CreateArtifactsForPhase(...)
```

En `ApproveUseCase`, si el workflow no puede cargarse, la aprobación continúa igualmente.

#### Evidencia

Se reprodujeron dos casos:

1. Al eliminar el workflow antes de aprobar, el comando respondió con éxito y aprobó la fase, pero no desbloqueó su dependencia.
2. Al impedir la escritura en `events.jsonl`, la aprobación respondió con éxito aunque no quedó registrada en la bitácora.

#### Impacto

El motor puede devolver una respuesta con forma de éxito mientras:

- no aplicó toda la transición;
- no creó el artefacto esperado;
- no dejó trazabilidad;
- dejó estados incompatibles entre sí.

Esto es especialmente grave porque la CLI está diseñada para ser la autoridad de estado del harness.

#### Corrección esperada

- No ignorar ningún error que forme parte del resultado contractual.
- Clasificar explícitamente qué operaciones son obligatorias y cuáles son best-effort.
- En el núcleo actual, workflow, manifest, artifacts requeridos y eventos de transición deben considerarse obligatorios.
- Hacer que el comando falle si la operación completa no puede garantizarse.

### CLI-04 — `next` modifica el manifiesto

#### Situación actual

`NextUseCase` busca la próxima fase y, si está en `ready`, la cambia a `in_progress` y guarda el work item.

#### Problema contractual

La documentación define `status`, `next` y `validate` como operaciones de consulta que no modifican archivos.

Además, la respuesta utiliza el estado anterior para algunos campos. Se observó una respuesta con:

```json
{
  "status": "ready",
  "message": "Next active phase is 'specification' (in_progress)."
}
```

mientras el manifest ya contenía `in_progress`.

#### Impacto

- Consultar dos veces produce resultados diferentes.
- Un agente puede iniciar trabajo accidentalmente sólo por inspeccionar el estado.
- La respuesta no representa de manera consistente el estado persistido.
- Se dificulta razonar sobre idempotencia.

#### Corrección esperada

`next` debe limitarse a calcular y devolver la siguiente acción permitida.

El cambio `ready → in_progress` debe pertenecer a una operación mutante explícita, por ejemplo:

```bash
sdd begin-phase <id> --phase specification
```

El nombre definitivo debe definirse al cerrar la máquina de estados.

### CLI-05 — Máquina de estados incompleta

#### Situación actual

El contrato describe estados como:

```text
blocked → ready → in_progress → awaiting_approval → approved
```

y contempla fases sin approval que deben llegar a `completed`.

La CLI actual sólo ofrece transiciones parciales:

- `start` establece la primera fase en `in_progress`;
- `next` cambia `ready` a `in_progress`;
- `approve` acepta `in_progress` o `awaiting_approval` y cambia a `approved`.

No existe una operación para:

- declarar terminado un artefacto;
- solicitar aprobación;
- completar una fase sin gate;
- completar el work item;
- distinguir generación, entrega, aprobación y aplicación de efectos.

#### Consecuencias

- Una fase puede aprobarse sin haber sido formalmente entregada.
- Las fases `approval: none` no tienen una transición correcta.
- La única manera práctica de avanzar es abusar de `approve`.
- `awaiting_approval` prácticamente no se utiliza.
- El estado global del work item nunca se completa.

Durante la revalidación se eliminó físicamente `artifacts/prd.md` antes de ejecutar `approve`. La aprobación igualmente finalizó con éxito. También se avanzó hasta `implementation`, fase declarada con `approval: none`, y `approve` la dejó en estado `approved`.

Por lo tanto, el comentario del código que indica que se permite aprobar `in_progress` “si el artifact existe” no coincide con la implementación: no se comprueba la existencia ni el contenido del artifact.

#### Corrección esperada

Antes de escribir más comandos debe definirse una tabla explícita de transiciones. Por ejemplo:

| Estado actual | Operación | Condición | Estado resultante |
| --- | --- | --- | --- |
| `blocked` | desbloqueo interno | dependencias satisfechas | `ready` |
| `ready` | begin | operación autorizada | `in_progress` |
| `in_progress` | submit | evidencia válida y approval requerido | `awaiting_approval` |
| `in_progress` | complete | evidencia válida y sin approval | `completed` |
| `awaiting_approval` | approve | actor humano válido | `approved` |
| `awaiting_approval` | reject | actor humano y motivo | `rejected` |

La tabla final debe resolver también qué significa `approved → completed` y cuándo se actualiza el status global.

### CLI-06 — Aprobación humana protegida sólo en Cobra

#### Situación actual

El comando `approve` construye un actor con `Kind: "human"`, pero `ApproveUseCase` no valida esa condición.

La regla está protegida en la interfaz terminal, no en la lógica del negocio.

#### Impacto

Una futura integración podría invocar el caso de uso con:

```go
domain.Actor{Kind: "agent", ID: "planner"}
```

y el dominio aceptaría la aprobación.

También falta validar:

- ID de aprobador no vacío;
- que la fase realmente declare `approval: required` u `optional`;
- que el artefacto requerido exista y sea válido;
- que la fase esté exactamente en el estado desde el cual puede aprobarse.

#### Corrección esperada

Las invariantes de aprobación deben residir en el dominio o en un servicio de dominio invocado por cualquier adaptador. Cobra sólo debe parsear inputs.

### CLI-07 — Uso parcial del grafo y dependencia del orden

#### Situación actual

La implementación:

- toma `wf.Phases[0]` como entrada default;
- recorre el slice para encontrar la fase externa;
- marca las anteriores como `not_applicable`;
- habilita la fase inmediatamente posterior;
- utiliza el orden físico del YAML para decidir progresión.

No toda la lógica ignora el grafo. `unblockDependentPhases` sí recorre `requires` y verifica que las dependencias estén satisfechas. El hallazgo preciso es que esa interpretación no se aplica de manera uniforme en todas las operaciones.

#### Problema

Un workflow declara un grafo mediante IDs y `requires`. El orden del array puede ayudar a presentarlo, pero no debería reemplazar las dependencias.

Tampoco se valida:

- que la fase sea un entry point;
- que el tipo de input esté permitido por `accepts`;
- que todos los IDs sean únicos;
- que las referencias existan;
- que el grafo sea acíclico;
- que una fase sea alcanzable.

Además, `--from-artifact` no exige realmente `--phase`. Al omitirlo, la ejecución:

- persistió `source: external_artifact`;
- conservó `entry_phase: prd`;
- creó el template normal `artifacts/prd.md`;
- copió el input externo a un archivo adicional llamado `artifacts/.md`.

El comando finalizó con éxito a pesar de producir un estado ambiguo.

#### Impacto

El comportamiento es confiable principalmente para workflows lineales, correctamente ordenados y similares a los cinco ejemplos actuales. Parte de un DAG puede desbloquearse mediante `requires`, pero inicio, bypass y selección no implementan todavía una interpretación uniforme del grafo.

Esto limita el principio de extensión por datos y hace que un workflow sintácticamente válido pueda producir estados incorrectos.

#### Corrección esperada

Implementar un validador semántico de workflows, adicional a JSON Schema:

- unicidad de fases;
- referencias válidas;
- DAG sin ciclos;
- entry points válidos;
- artifacts producidos existentes;
- dependencias alcanzables;
- paths seguros;
- approvals y effects válidos.

La próxima fase debe calcularse desde el grafo y el estado, no desde la posición en el slice.

### CLI-08 — Paths de artifacts hardcodeados

#### Situación actual

Los estados de fase calculan el artifact así:

```go
fmt.Sprintf("artifacts/%s.md", phaseID)
```

El workflow ya declara el artifact real mediante `produces` y `artifacts`.

#### Ejemplos de inconsistencia

| Fase | Path hardcodeado en manifest | Path declarado por workflow |
| --- | --- | --- |
| `implementation` | `artifacts/implementation.md` | `artifacts/implementation-report.md` |
| `verification` | `artifacts/verification.md` | `artifacts/verification-report.md` |
| `debugging` | `artifacts/debugging.md` | `artifacts/exploration.md` |

`ArtifactManager` sí utiliza el path del workflow, por lo que puede crear un archivo distinto del que referencia el manifest.

#### Impacto

- El manifest deja de ser confiable.
- El agente puede buscar un archivo inexistente.
- Una futura validación semántica de artifacts debería detectar inconsistencias generadas por la propia CLI.
- Agregar nuevos artifacts requiere coincidir accidentalmente con el phase ID.

#### Corrección esperada

Toda referencia debe derivarse desde:

```text
phase.produces → workflow.artifacts[artifactID].path
```

Debe definirse también cómo representar fases con:

- ningún artefacto;
- un artefacto;
- múltiples artefactos.

El modelo actual `PhaseState.Artifact string` sólo representa correctamente el segundo caso.

### CLI-09 — Schemas sin ejecución y validación semántica ausente

#### Situación actual

Los tres JSON Schemas están presentes, pero ninguna operación de lectura o escritura los ejecuta.

El test de fixtures válidos sólo hace `json.Unmarshal` y comprueba tres campos básicos. Los fixtures inválidos no se utilizan.

#### Impacto

El motor acepta:

- estados no permitidos;
- actors inválidos;
- IDs inseguros;
- manifests incompletos;
- workflows inconsistentes;
- event types fuera del patrón.

Además, JSON Schema no será suficiente para reglas de grafo o filesystem.

#### Corrección esperada

Separar dos niveles:

1. **Validación estructural:** JSON Schema.
2. **Validación semántica:** reglas del dominio, grafo y seguridad de paths.

Toda operación mutante debe validar:

- estado actual cargado;
- workflow relacionado;
- comando solicitado;
- estado resultante antes de escribir.

### CLI-10 — Persistencia no atómica y política de concurrencia ausente

#### Situación actual

El manifest se escribe directamente con `os.WriteFile`. No existen locks ni control de versión.

Los eventos se escriben por separado del manifest y los artifacts.

#### Hechos confirmados

- El manifest se trunca y reescribe directamente.
- Manifest, artifacts y eventos se persisten mediante operaciones separadas.
- No existe lock, revision number ni compare-and-swap.
- El código no define si dos mutaciones simultáneas están permitidas.

#### Riesgos derivados

- Un proceso podría leer un manifest durante una escritura incompleta.
- Dos agentes podrían cargar la misma versión y sobrescribir mutuamente sus cambios.
- Un fallo entre manifest y evento puede dejar información contradictoria.
- Una interrupción puede dejar una transición sin auditoría.

Estos escenarios se derivan de las garantías del filesystem y de la ausencia de coordinación, pero no se reprodujo durante la auditoría un crash real ni un lost update concurrente.

La atomicidad de una única escritura es necesaria desde ahora. El locking puede implementarse inmediatamente o posponerse si se declara explícitamente que v0.1 sólo admite un escritor; no debe quedar como comportamiento implícito.

#### Corrección esperada

Como mínimo:

- escritura a archivo temporal;
- sincronización y rename atómico;
- lock por work item si se permiten mutaciones concurrentes;
- versión o revisión del manifest para detectar lost updates;
- orden contractual entre manifest y eventos;
- estrategia explícita de recuperación si falla el commit.

### CLI-11 — Dependencias hacia infraestructura dentro de use cases

#### Situación actual

`StartWorkItemUseCase` y `ApproveUseCase` crean directamente:

```go
infra.NewArtifactManager()
```

`InitUseCase` depende directamente de `os`, `io/fs` y del package de embeds.

#### Evaluación

La inversión de dependencias se aplicó a repositorios, pero no a artifacts ni inicialización.

Esto produce una Clean Architecture parcial:

```text
usecases → ports        correcto
infra    → ports        correcto
usecases → infra        incorrecto para la frontera propuesta
```

#### Impacto

- Los casos de uso no son independientes del filesystem.
- Los tests necesitan recursos reales y directorios temporales.
- Se dificulta simular fallos concretos.
- Cambiar la estrategia de artifacts exige modificar aplicación.

#### Corrección esperada

Definir contratos según las necesidades del consumidor y la frontera de consistencia. Por ejemplo:

- un repositorio para cargar el agregado;
- un committer o repositorio capaz de persistir conjuntamente el manifest, los artifacts y los eventos de una mutación;
- `WorkflowRepository`;
- `TemplateSource`;
- `Clock`;
- `IDGenerator`.

Separar ciegamente `WorkItemStore`, `ArtifactStore` y `EventStore` podría dificultar la atomicidad. La frontera correcta no depende sólo del tipo de archivo, sino de qué elementos deben confirmarse como una única operación.

No es necesario crear una interfaz para cada función. Sólo deben abstraerse dependencias externas o variables que aporten testabilidad, aislamiento o una garantía de consistencia.

### CLI-12 — Dominio anémico y strings libres

#### Situación actual

Los structs de `domain` representan correctamente los datos, pero no protegen sus invariantes. Los estados se expresan como strings:

```go
Status string
Kind   string
```

La lógica de transición está dispersa entre `Start`, `Approve` y `Next`.

#### Riesgos

- Typos aceptados por compilación.
- Transiciones diferentes pueden interpretar estados de forma distinta.
- Es posible construir entidades inválidas.
- Cada nuevo comando debe volver a implementar reglas.

#### Corrección esperada

Introducir tipos y constantes:

```go
type PhaseStatus string
type ActorKind string
type ApprovalPolicy string
```

El work item debería exponer comportamiento, por ejemplo:

```go
func (w *WorkItem) BeginPhase(...)
func (w *WorkItem) SubmitPhase(...)
func (w *WorkItem) ApprovePhase(...)
func (w *WorkItem) CompletePhase(...)
```

Estos métodos deben validar la transición y producir el estado resultante. No hace falta implementar DDD pesado; alcanza con concentrar las invariantes donde viven los datos.

### CLI-13 — Debilidades del registro de eventos

#### Situación actual

El ID se genera con timestamp a precisión de segundos:

```go
"evt_" + time.Now().Format("20060102150405")
```

Dos eventos del mismo work item creados en el mismo segundo pueden compartir ID.

También se observan otros problemas:

- eventos obligatorios ignorados si fallan;
- eventos de bypass escritos antes de `work_item.created`;
- ausencia de eventos para muchas transiciones;
- `CorrelationID` definido pero no utilizado;
- ausencia de un generador inyectable para tests.

La colisión fue reproducida con un inicio desde artifact externo. Los eventos `phase_bypassed_by_external_input` y `work_item.created` obtuvieron exactamente el mismo ID y timestamp.

El orden también quedó confirmado: el evento que describe el bypass se escribió antes del evento que declara la creación del work item. Esto no viola actualmente una regla de schema, pero dificulta interpretar cronológicamente la bitácora.

#### Impacto

La bitácora no garantiza identificación única ni reconstrucción completa de lo ocurrido.

#### Corrección esperada

- Utilizar ULID, UUID u otro ID con unicidad suficiente.
- Inyectar reloj y generador.
- Definir qué evento corresponde a cada transición.
- Mantener un orden lógico estable.
- Registrar `from`, `to`, fase y causa de toda transición.
- Tratar los eventos contractuales como parte de la operación.

### CLI-14 — Contrato de output y errores de la CLI

#### Situación actual

`outputResult` centraliza parte del formato, pero:

- llama `os.Exit` desde funciones internas;
- ignora errores de `json.MarshalIndent`;
- los comandos usan `Run` en lugar de `RunE`;
- Cobra puede devolver errores antes de llegar a `outputResult`;
- los formatters de texto están distribuidos en comandos;
- `status` recorre un map y produce orden no determinista.

#### Impacto

- No todos los errores respetan el envelope JSON.
- La CLI es difícil de probar sin ejecutar procesos.
- Los defers pueden no ejecutarse al llamar `os.Exit`.
- El output para humanos puede variar entre ejecuciones.

La inconsistencia JSON fue reproducida ejecutando `status` sin su argumento obligatorio junto con `--json`. Cobra imprimió usage y errores en texto plano, incluso duplicando el mensaje de error, en lugar de devolver `JSONResponse`.

#### Corrección esperada

- Construir el root command mediante una función.
- Utilizar `RunE`.
- Propagar errores hasta `main`.
- Ejecutar `os.Exit` únicamente en el punto de entrada.
- Centralizar serialización JSON y errores.
- Ordenar las fases según el workflow para output de texto.
- Definir códigos de error estables para consumo de agentes.

### CLI-15 — Tests insuficientes para las garantías requeridas

#### Situación actual

Los tests pasan, pero su alcance es limitado:

- `TestFullWorkItemLifecycle` no completa el workflow;
- los use cases utilizan filesystem real en lugar de doubles;
- sólo se prueban fixtures válidos;
- deserializar JSON no equivale a validarlo contra schema;
- no existen tests de `cmd`;
- no hay assertions profundas sobre eventos;
- no se prueban fallos entre pasos;
- no se prueban IDs o paths inválidos;
- no se prueban workflows no lineales;
- no se prueban actualizaciones concurrentes.

La cobertura está concentrada en `usecases`; `cmd`, `domain` e `infra` no cuentan con cobertura directa significativa.

#### Corrección esperada

Incorporar:

1. tests unitarios de la máquina de estados;
2. tests table-driven por transición válida e inválida;
3. tests del validador de workflows;
4. tests de repositorios y atomicidad;
5. tests de CLI con stdout, stderr y exit codes;
6. tests de todos los fixtures inválidos;
7. tests de path containment;
8. tests de fallos inyectados;
9. tests de concurrencia;
10. tests end-to-end de cada workflow.

Los tests deben demostrar las propiedades del motor, no sólo que el camino feliz produce un archivo.

### CLI-16 — Higiene de Go y repositorio

#### Hallazgos

Los checks principales de compilación son correctos:

- `go build` finaliza;
- `go test ./...` pasa;
- `go test -race ./...` pasa para los casos actuales;
- `go vet ./...` no reporta problemas.

Sin embargo:

- `cmd/record_event.go` no está formateado con `gofmt`;
- `internal/usecases/approve_uc.go` no está formateado;
- `embeds/embeds.go` no está formateado;
- `src/cli/sdd` continúa versionado aunque `.gitignore` lo excluye;
- el binario versionado es Mach-O ARM64 y no es portable a Linux;
- el módulo `sdd-cli` es suficiente localmente, pero debería revisarse si la CLI se distribuirá mediante el repositorio.

#### Corrección esperada

- Ejecutar `gofmt` sobre todos los archivos Go.
- Retirar el binario del índice de Git.
- Compilar binarios mediante releases, CI o comandos locales.
- Definir una estrategia de distribución antes de estabilizar el module path.

### CLI-17 — Drift entre contrato y templates

#### Situación actual

El contrato mínimo de artefactos documenta front matter con:

- `schema_version`;
- `kind`;
- `id`;
- `work_item`;
- `phase`;
- `status`;
- `created_at`;
- `created_by`;
- `sources`.

Los templates actuales contienen principalmente:

```yaml
kind: artifact
id: prd
phase: prd
status: draft
```

`StartWorkItemUseCase` prepara variables como `id`, `created_at` y `type`, pero los templates no las utilizan.

#### Impacto

- Los artefactos generados no cumplen el contrato documentado.
- Se pierde trazabilidad entre archivo, work item y actor.
- Una futura validación semántica de artifacts podría marcar como inválidos archivos creados por la propia CLI.

Actualmente sólo existen schemas para workflow, work item y event. No existe un artifact schema, por lo que `validate` no detectaría automáticamente este drift salvo que se agregue esa validación.

#### Corrección esperada

- Alinear templates y contrato antes de implementar validación.
- Definir qué campos son realmente obligatorios en v0.1.
- Evitar documentar metadatos que el motor no puede producir consistentemente.
- Agregar tests de renderizado de cada template.

### CLI-18 — Los templates locales del proyecto son ignorados

#### Situación actual

`sdd init` copia los templates embebidos a:

```text
<proyecto>/.sdd/templates/
```

Esa carpeta forma parte del contrato local, versionado y personalizable del proyecto. Sin embargo, `ArtifactManager` no lee esos archivos. Siempre obtiene el contenido desde:

```go
embeds.DefaultSDDResources
```

#### Evidencia

Después de ejecutar `init`, se reemplazó `.sdd/templates/prd.md` por un template local identificable. Al ejecutar `start`, el artifact generado continuó utilizando el template embebido original.

#### Impacto

- Las personalizaciones por proyecto no tienen efecto.
- Un workflow local no puede utilizar un template nuevo que no exista en el binario.
- `.sdd/templates/` aparenta ser fuente de verdad, pero no lo es en runtime.
- Para modificar un template sería necesario recompilar la CLI.
- Se rompe el objetivo de que el contrato sea portable y configurable por proyecto.

#### Corrección esperada

- Utilizar los recursos embebidos únicamente durante `sdd init`.
- Después de la inicialización, leer templates desde `<baseDir>/.sdd/templates`.
- Validar que el template referenciado por el workflow exista localmente.
- Proteger el path del template con las mismas reglas de containment.
- Agregar un test que modifique un template local y verifique su uso.

### CLI-19 — El workflow default de `config.yaml` es ignorado

#### Situación actual

El contrato declara:

```yaml
defaults:
  workflow: feature-standard
```

pero el flag Cobra se registra con:

```go
StringVarP(&startWorkflow, "workflow", "w", "feature-standard", "Workflow ID")
```

No existe un repositorio o loader para `config.yaml`.

#### Impacto

- Cambiar `defaults.workflow` no modifica el comportamiento.
- El usuario recibe una configuración que aparenta ser efectiva, pero no lo es.
- El contrato local y la interfaz CLI pueden divergir.

#### Corrección esperada

- Distinguir entre “flag no informado” y “flag informado”.
- Cargar el default desde `.sdd/config.yaml`.
- Permitir que `--workflow` lo sobrescriba explícitamente.
- Validar la configuración antes de iniciar el work item.

## 7. Evaluación de SOLID

### 7.1. Single Responsibility Principle

**Aplicación parcial.**

La separación por archivos y paquetes es positiva, pero algunos use cases acumulan:

- validación;
- cálculo de transición;
- persistencia;
- creación y copia de archivos;
- renderizado de templates;
- emisión de eventos.

La solución no requiere dividir cada función en servicios mínimos, pero sí separar reglas de dominio, preparación de artifacts y commit de persistencia.

### 7.2. Open/Closed Principle

**Aplicación débil en el comportamiento.**

Los workflows son declarativos, lo cual favorece extensión. Sin embargo, los supuestos lineales y los paths hardcodeados hacen que nuevos workflows requieran adaptar el código o respetar convenciones implícitas.

El motor será realmente extensible cuando interprete `requires`, `produces`, `entry_points` y `artifacts` sin asumir coincidencia entre IDs ni orden lineal.

### 7.3. Liskov Substitution Principle

**No hay suficiente variedad de implementaciones para evaluarlo en profundidad.**

Las interfaces actuales sólo tienen implementaciones de filesystem. El riesgo principal futuro será que los contratos no documenten atomicidad, errores o comportamiento esperado.

### 7.4. Interface Segregation Principle

**Mejorable.**

`WorkItemRepository` combina:

- lectura;
- escritura;
- existencia;
- registro de eventos.

Un caso de uso de consulta depende de métodos que no utiliza. En Go suele ser preferible definir interfaces pequeñas cerca del consumidor.

### 7.5. Dependency Inversion Principle

**Aplicación parcial.**

Los use cases reciben repositorios por interfaces, pero también crean directamente implementaciones de artifacts y utilizan filesystem.

La dirección de dependencias debe corregirse para que aplicación y dominio no importen infraestructura concreta.

## 8. Evaluación de Clean Code y prácticas Go

### 8.1. Fortalezas

- Funciones mayormente cortas y fáciles de seguir.
- Nombres generales comprensibles.
- Estructura de archivos predecible.
- Uso correcto de wrapping de errores cuando se propagan.
- Punto de entrada pequeño.
- Uso apropiado de `internal`.
- Pocas dependencias externas.

### 8.2. Debilidades

- strings mágicos para estados y tipos;
- comentarios que describen un comportamiento no implementado;
- errores descartados;
- duplicación de `templateVars`;
- orden no determinista de maps;
- dependencias concretas creadas dentro de use cases;
- helpers de copia y persistencia sin garantías atómicas;
- nombres de tests que prometen más cobertura de la que realizan;
- mezcla de responsabilidades entre dominio, aplicación e infraestructura.

## 9. Riesgo de continuar sin refactorizar

Agregar nuevos comandos sobre el diseño actual amplificaría la deuda:

- `reject` necesitaría duplicar reglas dispersas de transición;
- `archive` operaría sobre work items cuyo completion no está definido;
- `validate` encontraría manifests y artifacts inválidos creados por la propia CLI;
- el agente orquestador recibiría éxitos falsos;
- la concurrencia aumentaría la probabilidad de lost updates;
- cada workflow nuevo dependería de supuestos no documentados.

El mayor riesgo no es que la CLI falle visiblemente. Es que responda con éxito mientras el estado queda incompleto o inconsistente.

> Un motor que aparenta determinismo pero no garantiza atomicidad, validación y trazabilidad genera una confianza incorrecta en todo el harness.

## 10. Arquitectura objetivo sugerida

La solución puede mantenerse pequeña. No se recomienda introducir patrones complejos sin necesidad.

Una dirección posible es:

```text
cmd
  └── parsea inputs, presenta resultados y traduce errores

application/usecases
  └── carga datos, invoca dominio y coordina un commit

domain
  ├── WorkItem
  ├── Workflow
  ├── tipos de estado
  ├── validación semántica
  └── reglas de transición

ports
  ├── WorkItemRepository / MutationCommitter
  ├── WorkflowRepository
  ├── TemplateSource
  ├── Clock
  └── IDGenerator

infra
  ├── filesystem
  ├── atomic writer / locking
  ├── JSON Schema validator
  ├── embedded templates
  └── ULID/UUID generator
```

Un caso de uso mutante debería seguir un flujo similar:

```text
1. Validar comando e identidad.
2. Cargar workflow y work item.
3. Validar schema y semántica.
4. Pedir al dominio la transición.
5. Obtener nuevo estado y eventos.
6. Preparar artifacts requeridos.
7. Persistir todo mediante una operación controlada.
8. Devolver el estado confirmado.
```

## 11. Próximos pasos sugeridos

### Fase 0 — Congelar nuevas capacidades

No implementar todavía:

- `reject`;
- `archive`;
- integración definitiva con el orquestador.

Aplicar además la higiene mínima del repositorio: `gofmt` y eliminación del binario de plataforma versionado.

El objetivo es evitar construir nuevas operaciones sobre invariantes incompletas.

**Hallazgos cerrados al completar la fase:** CLI-16.

### Fase 1 — Cerrar el modelo de estados

1. Documentar todas las transiciones posibles.
2. Definir operaciones explícitas para comenzar, entregar, completar, aprobar y rechazar fases.
3. Definir cuándo una fase queda `approved` y cuándo `completed`.
4. Definir cómo se completa el work item.
5. Resolver el comportamiento de approvals `required`, `optional` y `none`.
6. Modelar esas reglas mediante tipos y métodos de dominio.

**Criterio de finalización:** ninguna transición válida o inválida depende de lógica dispersa en Cobra o filesystem.

**Hallazgos cerrados al completar la fase:** CLI-04, CLI-06 y CLI-12.

**Hallazgos encaminados pero todavía no cerrados:** CLI-05, porque `reject` seguirá sin exponerse como comando hasta la Fase 6; CLI-07, porque la selección ya no muta estado pero la validación completa del grafo pertenece a la Fase 2; y CLI-15, porque se agregan tests de dominio pero la suite contractual completa pertenece a la Fase 5.

**Estado al 2026-08-18:** completada. El dominio concentra las transiciones mediante estados y políticas tipadas; `next` es una consulta pura; `begin`, `deliver` y `complete` son operaciones explícitas; `approve` exige `awaiting_approval` y actor humano; el rechazo está modelado en dominio pero permanece sin comando público, respetando la congelación de la Fase 0.

### Fase 2 — Seguridad y validación

1. Validar IDs y paths.
2. Implementar path containment.
3. Integrar JSON Schema.
4. Implementar validación semántica del workflow.
5. Validar entry points e inputs externos.
6. Alinear templates con el contrato.
7. Leer templates desde el contrato local.
8. Aplicar los defaults de `config.yaml`.

**Criterio de finalización:** la CLI no puede cargar ni escribir un estado estructural o semánticamente inválido.

**Hallazgos cerrados al completar la fase:** CLI-01, CLI-07, CLI-08, CLI-09, CLI-17, CLI-18 y CLI-19.

### Fase 3 — Integridad de persistencia

1. Eliminar errores ignorados.
2. Implementar escrituras atómicas.
3. Definir si v0.1 admite múltiples escritores y, si los admite, implementar locks o control de revisiones.
4. Garantizar consistencia entre manifest, artifacts y eventos.
5. Hacer idempotentes las operaciones que corresponda.
6. Asegurar que una operación fallida no deje estado parcial.

**Criterio de finalización:** todo éxito representa una transición completa y todo fallo deja el estado anterior intacto.

**Hallazgos cerrados al completar la fase:** CLI-02, CLI-03 y CLI-10.

**Hallazgo encaminado pero todavía no cerrado:** CLI-13, porque aquí se garantiza la escritura consistente de eventos, mientras que su unicidad y generación determinista se terminan en la Fase 4.

### Fase 4 — Corregir fronteras arquitectónicas

1. Extraer la creación de artifacts detrás de un port.
2. Separar interfaces según las necesidades de los consumidores.
3. Inyectar reloj y generador de IDs.
4. Mover reglas desde use cases hacia dominio.
5. Centralizar el wiring de dependencias en la composición de la CLI.
6. Convertir comandos a `RunE` y unificar errores/output.

**Criterio de finalización:** los use cases pueden probarse sin filesystem real y no importan infraestructura concreta.

**Hallazgos cerrados al completar la fase:** CLI-11, CLI-13 y CLI-14.

### Fase 5 — Construir una suite de tests contractual

1. Tests table-driven de estados.
2. Tests de schemas y fixtures inválidos.
3. Tests de grafos inválidos.
4. Tests de path traversal.
5. Tests de fallos entre pasos.
6. Tests de eventos y unicidad.
7. Tests de concurrencia.
8. Tests de CLI y JSON.
9. Ciclo end-to-end completo para cada workflow.
10. Tests de templates locales y defaults de configuración.

**Criterio de finalización:** cada regla no negociable del contrato tiene al menos un test que falla si se viola.

**Hallazgo cerrado al completar la fase:** CLI-15.

### Fase 6 — Continuar el roadmap

Con el núcleo estabilizado:

1. implementar `sdd validate`;
2. implementar `sdd reject`;
3. implementar `sdd archive`;
4. integrar el agente orquestador;
5. realizar pruebas reales del harness;
6. evaluar observabilidad, Engram y CodeGraph.

**Hallazgo cerrado al completar la fase:** CLI-05, al exponer el rechazo pendiente y completar la superficie pública del lifecycle.

## 12. Decisión recomendada

La estructura actual debe utilizarse como punto de partida, pero no como diseño cerrado.

La prioridad inmediata no es agregar más comandos. Es convertir la CLI en un motor cuyas respuestas tengan garantías claras:

- inputs seguros;
- transiciones válidas;
- consultas sin efectos;
- escrituras atómicas;
- errores visibles;
- eventos confiables;
- workflows realmente declarativos;
- tests que demuestren esas propiedades.

Una vez alcanzadas esas garantías, los comandos restantes serán más sencillos de implementar y no necesitarán duplicar ni compensar reglas incompletas.

---

Created on: 2026-08-18  
Last modified: 2026-08-18
