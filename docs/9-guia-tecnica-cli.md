# Guia tecnica de la CLI del motor SDD

> Estado: documentacion de la implementacion actual.  
> Audiencia: desarrolladores que se incorporan al proyecto, con experiencia en JavaScript/TypeScript o Java y sin experiencia previa en Go.  
> Alcance: contrato instalado en `src/.sdd/` y motor implementado en `src/cli/`.  
> Version documentada: contrato y CLI v0.1, estado al 2026-08-19.

---

## 1. Objetivo de este documento

Este documento explica la CLI como un sistema completo:

- que problema resuelve dentro del harness;
- que funcionalidades soporta hoy;
- que workflows gobierna;
- por que se eligio Go y sus librerias principales;
- como esta organizada la arquitectura;
- como interactuan las capas y componentes;
- como fluyen los datos desde un comando hasta el filesystem;
- que garantias de seguridad, consistencia e idempotencia ofrece;
- que partes siguen fuera del alcance actual.

La informacion se reconstruyo cruzando:

1. el contrato vigente de `src/.sdd/`;
2. el codigo de produccion de `src/cli/`;
3. la suite contractual y end-to-end;
4. los documentos de diseño, implementacion y auditoria previos.

Cuando existe una diferencia entre una idea historica y el comportamiento actual, este documento prioriza el codigo y el contrato ejecutable.

---

## 2. Resumen ejecutivo

La CLI `sdd` es el **motor determinista** del framework de Spec-Driven Development.

Su responsabilidad no es redactar documentos, programar una feature ni decidir que modelo de IA usar. Su responsabilidad es gobernar el proceso:

- crear instancias de trabajo;
- conocer el estado de cada fase;
- validar transiciones;
- impedir saltos no permitidos;
- exigir gates humanos;
- preparar los artifacts definidos por el workflow;
- registrar eventos auditables;
- persistir el cambio completo de forma consistente.

La separacion conceptual mas importante es:

```text
Agente o persona: realiza trabajo cognitivo y propone acciones.
CLI: valida reglas, cambia estado y registra evidencia.
```

Ejemplo:

```text
El agente redacta el plan.
El agente ejecuta `sdd deliver ... --phase plan`.
La CLI verifica que `plan` estaba en progreso.
La CLI lo deja esperando aprobacion.
Una persona ejecuta `sdd approve ... --phase plan`.
La CLI registra la aprobacion y desbloquea `implementation`.
```

Esto evita depender de que el LLM recuerde correctamente el estado del proceso a partir del chat.

---

## 3. Que es y que no es la CLI

### 3.1. Que es

La CLI es:

- una interfaz de linea de comandos;
- una maquina de estados para work items y fases;
- un interprete de workflows declarativos;
- un validador estructural y semantico;
- un coordinador de persistencia transaccional sobre filesystem;
- una fuente de respuestas estructuradas para agentes;
- un productor de trazabilidad mediante eventos JSONL.

### 3.2. Que no es

La CLI no:

- genera por si sola el contenido intelectual de un PRD, plan o especificacion;
- modifica codigo como parte de una fase;
- ejecuta commits, pushes o pull requests;
- elige agentes, modelos, skills o MCPs;
- reemplaza al orquestador;
- reemplaza Git como fuente de verdad;
- mantiene una base de datos oculta;
- archiva fisicamente work items todavia;
- todavia no expone el comando `validate`.

Los campos `procedure` y `effects` del workflow describen que debe hacerse, pero el motor no ejecuta esas herramientas. Esa responsabilidad pertenece al agente o adaptador.

---

## 4. Conceptos de dominio

| Concepto | Significado | Ejemplo |
| --- | --- | --- |
| **Workflow** | Plantilla declarativa de un proceso. Define fases, dependencias, artifacts y gates. | `feature-standard` |
| **Work item** | Instancia concreta de un workflow. | `feat-add-coupons` |
| **Fase** | Unidad de trabajo gobernada por estados. | `plan`, `implementation` |
| **Artifact** | Documento Markdown producido como evidencia de una fase. | `artifacts/plan.md` |
| **Manifest** | Snapshot YAML del estado actual del work item. | `manifest.yaml` |
| **Evento** | Registro inmutable de una accion o transicion. | `phase.transitioned` |
| **Gate** | Decision requerida antes de continuar. | Aprobacion humana del plan |
| **Entry point** | Fase desde la cual el workflow permite comenzar. | `prd`, `specification`, `plan` |
| **Procedure** | Instruccion portable para producir el resultado de una fase. | `procedures/create-plan.md` |
| **Operation ID** | Clave estable usada para reintentar una mutacion sin duplicarla. | `run:plan:deliver:001` |
| **Actor** | Identidad que ejecuta una operacion. | `{kind: agent, id: copilot}` |

### 4.1. Las tres fuentes de informacion de un work item

```mermaid
flowchart LR
    Manifest["manifest.yaml<br/>Estado actual"]
    Artifacts["artifacts/*.md<br/>Evidencia legible"]
    Events["events.jsonl<br/>Historial inmutable"]

    Manifest --- WorkItem["Work item"]
    Artifacts --- WorkItem
    Events --- WorkItem
```

- El **manifest** responde: "¿en que estado esta ahora?".
- Los **artifacts** responden: "¿que produjo cada fase?".
- Los **eventos** responden: "¿como llego hasta este estado?".

El estado no se deduce leyendo el chat ni buscando checkboxes dentro de Markdown.

---

## 5. Por que se utiliza Go

Go es una buena eleccion para este motor porque combina distribucion simple, tipos estaticos, tooling estandar y un modelo de concurrencia adecuado para una CLI portable.

### 5.1. Justificacion tecnica

| Necesidad del motor | Aporte de Go |
| --- | --- |
| Distribuir la herramienta entre proyectos | Genera un binario autocontenido por plataforma |
| Ser agent-agnostic | No requiere runtime de Node.js, JVM ni un agente especifico |
| Mantener reglas deterministas | Tipos estaticos y compilacion previa |
| Operar sobre filesystem | Libreria estandar madura y multiplataforma |
| Probar dependencias externas | Interfaces implicitas y composicion simple |
| Incluir recursos base | `go:embed` integra `.sdd/` dentro del binario |
| Evitar configuracion compleja | `go test`, `go build`, `go vet` y `gofmt` vienen con el toolchain |
| Controlar concurrencia local | Locks de archivo y manejo explicito de errores |

### 5.2. Dependencias externas

| Libreria | Funcion |
| --- | --- |
| `github.com/spf13/cobra` | Parseo de comandos, argumentos y flags |
| `gopkg.in/yaml.v3` | Lectura y escritura de YAML |
| `github.com/santhosh-tekuri/jsonschema/v5` | Validacion contra JSON Schema |
| `github.com/gofrs/flock` | Lock exclusivo por work item |

No se usa un framework de aplicacion ni una base de datos.

---

## 6. Introduccion a Go aplicada a este codebase

### 6.1. Equivalencias utiles

| Go | JavaScript/TypeScript | Java |
| --- | --- | --- |
| `package` | Modulo/carpeta | Package |
| `struct` | Objeto tipado o `interface` de datos | Clase de datos/record |
| Metodo con receiver | Metodo de objeto | Metodo de instancia |
| `interface` | Interface estructural de TypeScript | Interface, pero implementada implicitamente |
| Funcion constructora `NewX` | Factory function | Constructor/factory |
| Composicion de interfaces | `extends` de interfaces | `extends` de interfaces |
| `error` como retorno | `throw`/`Promise.reject` | Excepcion checked conceptualmente |
| Exportado con mayuscula | `export` | `public` |
| Minuscula inicial | Privado al package | Package-private/private aproximado |

### 6.2. No hay clases ni herencia

El dominio se modela con structs y metodos:

```go
type WorkItem struct {
    ID     string
    Status WorkItemStatus
    Phases map[string]PhaseState
}

func (item *WorkItem) BeginPhase(
    workflow *Workflow,
    phaseID string,
) (PhaseMutation, error) {
    // Regla de negocio.
}
```

`(item *WorkItem)` es el **receiver**. Es comparable a un metodo de instancia:

```ts
class WorkItem {
  beginPhase(workflow: Workflow, phaseID: string): PhaseMutation {}
}
```

El puntero `*WorkItem` permite modificar la misma instancia.

### 6.3. Las interfaces se implementan implicitamente

```go
type WorkItemReader interface {
    GetWorkItem(baseDir string, id string) (*domain.WorkItem, error)
}
```

`FSWorkItemRepository` implementa esa interfaz porque posee un metodo con la misma firma. No existe `implements WorkItemReader`.

Esto permite que un use case reciba:

- el repositorio real de filesystem en produccion;
- un doble en memoria durante un test.

### 6.4. Manejo de errores

Go representa los fallos como valores:

```go
item, err := repo.GetWorkItem(baseDir, id)
if err != nil {
    return nil, fmt.Errorf("failed to get work item: %w", err)
}
```

`%w` agrega contexto sin perder la causa original. Luego `errors.Is` permite clasificar el error en la capa CLI.

No hay excepciones que atraviesen la aplicacion implicitamente: cada funcion decide si maneja o propaga el error.

### 6.5. `internal/`

Go reserva semanticamente el directorio `internal/`: sus packages no pueden importarse desde fuera del arbol permitido.

En esta CLI ayuda a indicar que `domain`, `ports`, `usecases` e `infra` son detalles internos del ejecutable y no una libreria publica estable.

---

## 7. Tecnologia y formatos de persistencia

| Informacion | Formato | Motivo |
| --- | --- | --- |
| Configuracion | YAML | Legible y editable por proyecto |
| Workflows | YAML | Permite declarar el proceso como datos |
| Manifest | YAML | Facil de revisar y versionar |
| Artifacts | Markdown con front matter YAML | Legibles por humanos y procesables por maquinas |
| Eventos | JSON Lines | Append-only y procesable linea por linea |
| Contratos | JSON Schema | Validacion interoperable y determinista |

La CLI no persiste estado fuera de `.sdd/`.

---

## 8. Estructura del proyecto

### 8.1. Contrato fuente

`src/.sdd/` contiene la definicion portable del motor:

```text
src/.sdd/
├── config.yaml
├── schemas/
├── workflows/
├── templates/
├── procedures/
├── registry/
├── context/
├── specs/
├── research/
├── tests/
└── work-items/
```

Este directorio es la fuente de verdad para workflows, schemas, templates y procedimientos incluidos en el framework.

### 8.2. Codigo Go

```text
src/cli/
├── main.go
├── generate.go
├── go.mod
├── cmd/
├── embeds/
├── internal/
│   ├── domain/
│   ├── ports/
│   ├── usecases/
│   └── infra/
└── tools/
    └── syncsdd/
```

### 8.3. Estructura instalada en un proyecto usuario

Luego de `sdd init`:

```text
mi-proyecto/
└── .sdd/
    ├── config.yaml
    ├── schemas/
    ├── workflows/
    ├── templates/
    ├── procedures/
    ├── registry/
    ├── context/
    ├── specs/
    ├── research/
    └── work-items/
        ├── active/
        └── archive/
```

Los fixtures contractuales de `tests/` no se instalan en el proyecto usuario.

### 8.4. Estructura de un work item activo

```text
.sdd/work-items/active/<id>/
├── manifest.yaml
├── events.jsonl
├── artifacts/
│   ├── plan.md
│   └── ...
└── evidence/
```

`evidence/` esta preparado para outputs crudos, screenshots o reportes, aunque la CLI actual no posee un comando especifico para administrarlos.

---

## 9. Recursos embebidos e inicializacion

El binario debe poder ejecutar `sdd init` sin depender de la ubicacion del repositorio original. Para eso utiliza `go:embed`.

```mermaid
flowchart LR
    Contract["src/.sdd/<br/>Fuente versionada"]
    Generate["go generate<br/>tools/syncsdd"]
    Generated["src/cli/embeds/default_sdd/<br/>Generado, no versionado"]
    Binary["Binario sdd<br/>go:embed"]
    Project["Proyecto destino<br/>.sdd/"]

    Contract --> Generate --> Generated --> Binary
    Binary -->|"sdd init"| Project
```

Puntos importantes:

- `src/.sdd/` se versiona.
- `src/cli/embeds/default_sdd/` es generado y esta en `.gitignore`.
- `generate.go` declara `//go:generate go run tools/syncsdd/main.go`.
- `embeds/embeds.go` incluye el directorio generado con `//go:embed`.
- `sdd init` publica la copia embebida en el proyecto destino.

En un checkout limpio debe sincronizarse el embed antes de compilar:

```bash
cd src/cli
go generate ./...
go build
```

---

## 10. Arquitectura de alto nivel

La implementacion sigue una variante pragmatica de Clean Architecture con ports and adapters.

```mermaid
flowchart TB
    User["Persona o agente"]
    Main["main.go"]
    Cmd["cmd/<br/>Cobra y presentacion"]
    UseCases["internal/usecases/<br/>Orquestacion"]
    Domain["internal/domain/<br/>Entidades e invariantes"]
    Ports["internal/ports/<br/>Capacidades requeridas"]
    Infra["internal/infra/<br/>Filesystem, schemas, templates"]
    Contract[".sdd/<br/>Contrato local"]

    User --> Main --> Cmd --> UseCases
    UseCases --> Domain
    UseCases --> Ports
    Infra --> Ports
    Infra --> Domain
    Cmd --> Infra
    Infra <--> Contract

    classDef core fill:#eaf2ff,stroke:#2864dc
    classDef outer fill:#f5f5f5,stroke:#666
    class Domain,Ports,UseCases core
    class User,Main,Cmd,Infra,Contract outer
```

### 10.1. Regla de dependencias

La direccion importante es:

```text
cmd -> usecases -> domain
                 -> ports
infra ---------> domain + ports
```

Los use cases no importan `infra`. Conocen interfaces pequenas definidas en `ports`.

`cmd/composition.go` es la excepcion intencional donde se conectan las implementaciones reales. Ese archivo es el **composition root**.

### 10.2. Responsabilidades por capa

| Capa | Responsabilidad | No deberia hacer |
| --- | --- | --- |
| `cmd` | Parsear input, ejecutar use cases y presentar output | Decidir transiciones |
| `usecases` | Orquestar una operacion completa | Conocer detalles de archivos o Cobra |
| `domain` | Modelar estados, reglas e invariantes | Leer disco o imprimir output |
| `ports` | Declarar capacidades externas requeridas | Implementar filesystem |
| `infra` | Implementar persistencia, schemas, templates, tiempo e IDs | Decidir reglas de negocio |
| `.sdd` | Declarar workflows, templates y contratos | Ejecutar operaciones |

---

## 11. Componentes principales del codigo

### 11.1. `main.go`

Es el punto de entrada minimo:

```go
func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

Es el unico lugar que convierte un error en exit code `1`.

### 11.2. `cmd/root.go`

El root command:

- registra flags globales;
- agrega todos los subcomandos;
- centraliza success y error output;
- convierte errores de dominio en codigos JSON estables;
- separa stdout y stderr.

### 11.3. `cmd/composition.go`

Construye las dependencias productivas una sola vez:

```text
FS repositories
ArtifactManager
SystemClock
CryptoIDGenerator
        |
        v
Use cases
        |
        v
Comandos Cobra
```

Esto evita que cada comando cree su propio repositorio o que un use case conozca implementaciones concretas.

### 11.4. `internal/domain`

Contiene:

- `WorkItem`;
- `Workflow`;
- estados de fase y work item;
- approvals;
- actors y events;
- validacion semantica;
- construccion inicial del agregado;
- maquina de estados.

Es el nucleo con mayor valor de negocio.

### 11.5. `internal/ports`

Divide capacidades segun el consumidor:

| Port | Capacidad |
| --- | --- |
| `WorkItemReader` | Leer un work item |
| `WorkItemExistenceChecker` | Detectar colisiones de ID |
| `OperationTracker` | Consultar si un operation ID ya fue aplicado |
| `WorkItemCommitter` | Confirmar manifest, artifacts y eventos |
| `WorkflowRepository` | Cargar un workflow |
| `ConfigRepository` | Cargar defaults del proyecto |
| `ArtifactPreparer` | Preparar artifacts desde templates |
| `ExternalArtifactImporter` | Resolver e importar input externo |
| `ProjectInitializer` | Inicializar `.sdd/` |
| `Clock` | Proveer timestamps |
| `IDGenerator` | Generar IDs de eventos |

Las interfaces compuestas:

```go
type WorkItemMutationRepository interface {
    WorkItemReader
    OperationTracker
    WorkItemCommitter
}
```

son comparables a:

```ts
interface WorkItemMutationRepository
  extends WorkItemReader, OperationTracker, WorkItemCommitter {}
```

### 11.6. `internal/usecases`

Cada operacion publica tiene un caso de uso:

- `InitUseCase`;
- `StartWorkItemUseCase`;
- `StatusUseCase`;
- `NextUseCase`;
- `BeginPhaseUseCase`;
- `DeliverPhaseUseCase`;
- `ApproveUseCase`;
- `CompleteUseCase`;
- `RecordEventUseCase`.

`transition_helpers.go` concentra pasos compartidos:

- cargar item y workflow;
- consultar idempotencia;
- preparar artifacts desbloqueados;
- generar eventos;
- confirmar el commit.

### 11.7. `internal/infra`

Implementa:

- repositorios YAML/filesystem;
- persistencia por snapshots;
- locks y revision optimista;
- recuperacion de transacciones interrumpidas;
- validacion JSON Schema;
- seguridad de paths y symlinks;
- renderizado/importacion de artifacts;
- inicializacion atomica;
- reloj real;
- IDs criptograficos.

---

## 12. Funcionalidades soportadas

### 12.1. Flags globales

| Flag | Default | Funcion |
| --- | --- | --- |
| `--dir` | `.` | Define el proyecto que contiene `.sdd/` |
| `--json` | `false` | Devuelve un envelope JSON para integraciones |

### 12.2. Matriz de comandos

| Comando | Tipo | Funcion | Modifica estado | `--operation-id` |
| --- | --- | --- | --- | --- |
| `sdd init` | Inicializacion | Instala `.sdd/` | Si | No |
| `sdd start` | Creacion | Crea un work item | Si | Si |
| `sdd status` | Consulta | Muestra manifest y fases | No | No |
| `sdd next` | Consulta | Informa la proxima accion | No | No |
| `sdd begin` | Transicion | Inicia una fase habilitada | Si | Si |
| `sdd deliver` | Transicion | Entrega el resultado de una fase | Si | Si |
| `sdd approve` | Gate humano | Aprueba una fase | Si | Si |
| `sdd reject` | Gate humano | Rechaza una fase para retrabajo | Si | Si |
| `sdd complete` | Transicion | Completa una fase o work item | Si | Si |
| `sdd record-event` | Observabilidad | Agrega un evento custom | Si | Si |

### 12.3. `sdd init`

```bash
sdd init
sdd init --dir /ruta/al/proyecto
```

Comportamiento:

1. verifica que `.sdd/` no exista;
2. crea un staging temporal;
3. copia los recursos embebidos;
4. excluye fixtures contractuales y work items del repositorio template;
5. crea `work-items/active` y `work-items/archive`;
6. publica el directorio mediante rename.

No inicializa Git ni sobrescribe una instalacion existente.

### 12.4. `sdd start <id>`

Inicio normal:

```bash
sdd start feat-add-coupons \
  --title "Agregar cupones" \
  --summary "Permitir descuentos en checkout"
```

Inicio con workflow explicito:

```bash
sdd start bug-payment-timeout \
  --workflow bug-investigation \
  --title "Timeout de pago"
```

Inicio desde artifact existente:

```bash
sdd start feat-add-coupons \
  --title "Agregar cupones" \
  --from-artifact ./plan-aprobado.md \
  --phase plan
```

Flags:

| Flag | Requerido | Default |
| --- | --- | --- |
| `--title`, `-t` | Si | - |
| `--workflow`, `-w` | No | `.sdd/config.yaml` |
| `--summary`, `-s` | No | Vacio |
| `--from-artifact` | Junto con `--phase` | Vacio |
| `--phase` | Junto con `--from-artifact` | Vacio |
| `--actor-kind` | No | `human` |
| `--actor-id` | No | `user` |
| `--operation-id` | No | Vacio |

La entrada externa:

- debe corresponder a un entry point declarado;
- se resuelve a path absoluto;
- se hashea con SHA-256;
- se importa al artifact canonico;
- marca ancestros como `not_applicable`;
- deja la fase en `accepted` o `awaiting_approval`, segun su gate.

No existe un skip arbitrario de fases.

En un inicio normal, la fase de entrada aceptada para `user_prompt` se crea y comienza automaticamente en `in_progress`. No es necesario ejecutar `begin` para esa primera fase.

### 12.5. `sdd status <id>`

```bash
sdd status feat-add-coupons
sdd status feat-add-coupons --json
```

Devuelve:

- estado general;
- workflow;
- revision;
- datos de input;
- approvals;
- estado y artifact de cada fase.

La salida de texto presenta las fases en orden topologico. La salida JSON conserva la forma contractual del manifest, donde `phases` es un mapa y no promete orden de claves.

Es una consulta pura.

### 12.6. `sdd next <id>`

```bash
sdd next feat-add-coupons
```

Prioriza:

1. una fase `awaiting_approval`;
2. una fase `in_progress`;
3. una fase `ready`.

Devuelve:

- phase ID;
- estado;
- procedure;
- artifact;
- necesidad de aprobacion;
- opcionalidad;
- mensaje orientativo.

`next` **no inicia la fase** y no aumenta la revision.

### 12.7. `sdd begin <id>`

```bash
sdd begin feat-add-coupons \
  --phase specification \
  --actor-kind agent \
  --actor-id copilot
```

Permite:

```text
ready | rejected | superseded -> in_progress
```

Una fase `blocked` no puede comenzar.

### 12.8. `sdd deliver <id>`

```bash
sdd deliver feat-add-coupons \
  --phase specification \
  --actor-id copilot
```

El destino depende de la politica:

| Politica | Resultado |
| --- | --- |
| `required` | `awaiting_approval` |
| `optional` | `completed` |
| `optional` + `--request-approval` | `awaiting_approval` |
| `none` | `completed` |
| `none` + `--request-approval` | Error |

Al satisfacerse dependencias, la CLI desbloquea fases y prepara sus templates dentro del mismo commit.

### 12.9. `sdd approve <id>`

```bash
sdd approve feat-add-coupons \
  --phase plan \
  --by matias \
  --comment "Aprobado"
```

Reglas:

- la fase debe estar `awaiting_approval`;
- la politica no puede ser `none`;
- el actor debe ser humano;
- se resuelve el approval pendiente;
- la fase pasa a `approved`;
- se desbloquean dependencias satisfechas.

La exigencia de actor humano vive en dominio, no solamente en Cobra.

### 12.10. `sdd reject <id>`

```bash
sdd reject feat-add-coupons \
  --phase plan \
  --by matias \
  --comment "Falta detallar el rollback"
```

Reglas:

- la fase debe estar `awaiting_approval`;
- la politica no puede ser `none`;
- el actor debe ser humano;
- se resuelve el approval pendiente como `rejected`;
- la fase pasa a `rejected`;
- no se desbloquean dependencias.

El retrabajo comienza explícitamente con:

```bash
sdd begin feat-add-coupons --phase plan
```

La nueva entrega crea otra iteracion de approval sin borrar el rechazo anterior.

### 12.11. `sdd complete <id>`

Completar fase:

```bash
sdd complete feat-add-coupons --phase plan
```

Permite:

```text
approved | accepted -> completed
```

Completar work item:

```bash
sdd complete feat-add-coupons
```

El work item solo pasa a `completed` cuando:

- todas las fases obligatorias estan satisfechas;
- las fases opcionales no iniciadas pueden omitirse;
- toda fase opcional iniciada tambien termino.

Los estados `approved` y `accepted` ya satisfacen dependencias y el cierre del work item. Por eso `complete --phase` no es obligatorio para avanzar: se utiliza cuando se quiere explicitar que una fase aprobada o aceptada ya fue aplicada y debe quedar en `completed`.

Una fase opcional como `archive` puede ejecutarse despues de completar el work item, pero el comando actual no mueve el directorio a `work-items/archive/`.

### 12.12. `sdd record-event <id>`

```bash
sdd record-event feat-add-coupons \
  --type validation.completed \
  --message "Suite verde" \
  --actor-kind agent \
  --actor-id verifier
```

Agrega un evento custom sin alterar fases. Aun asi:

- aumenta la revision;
- participa del mismo commit transaccional;
- valida actor y schema;
- puede ser idempotente mediante `--operation-id`.

### 12.13. Funcionalidades modeladas pero no expuestas

Tambien existen estados `archived` y `cancelled`, pero no hay comandos publicos que los apliquen.

---

## 13. Workflows disponibles

| Workflow | Tipo | Entry points | Flujo principal |
| --- | --- | --- | --- |
| `feature-standard` | `feature` | PRD, specification, plan | PRD -> spec -> plan -> implementacion -> verificacion -> review -> archive |
| `change-request` | `change-request` | CR, specification, plan | CR -> spec delta -> plan -> implementacion -> verificacion -> review -> archive |
| `fast-change` | `fast-change` | plan | Plan -> implementacion -> verificacion -> review -> archive |
| `bug-known-cause` | `bug` | issue, plan | Issue -> exploracion -> plan -> implementacion -> verificacion -> review -> archive |
| `bug-investigation` | `bug` | issue, plan | Issue -> debugging -> plan -> implementacion -> verificacion -> review -> archive |

`archive` es opcional en los cinco workflows.

### 13.1. Gates por workflow

| Fase | Feature | CR | Fast | Bug conocido | Bug investigado |
| --- | --- | --- | --- | --- | --- |
| PRD | Requerido | - | - | - | - |
| Change Request | - | Requerido | - | - | - |
| Specification | Requerido | Requerido | - | - | - |
| Plan | Requerido | Requerido | Requerido | Requerido | Requerido |
| Implementation | Ninguno | Ninguno | Ninguno | Ninguno | Ninguno |
| Verification | Ninguno | Ninguno | Ninguno | Ninguno | Ninguno |
| Human code review | Requerido | Requerido | Requerido | Requerido | Requerido |
| Archive | Ninguno/opcional | Ninguno/opcional | Ninguno/opcional | Ninguno/opcional | Ninguno/opcional |

### 13.2. Grafo general de una feature

```mermaid
flowchart LR
    PRD["PRD<br/>approval required"]
    Spec["Specification<br/>approval required"]
    Plan["Plan<br/>approval required"]
    Impl["Implementation<br/>approval none"]
    Verify["Verification<br/>approval none"]
    Review["Human code review<br/>approval required"]
    Archive["Archive<br/>optional"]

    PRD --> Spec --> Plan --> Impl --> Verify --> Review --> Archive
```

Aunque los workflows actuales son lineales, el motor no depende del orden fisico del YAML. Usa `requires` y ordenamiento topologico, por lo que el contrato puede evolucionar hacia grafos con ramas.

---

## 14. Maquina de estados

### 14.1. Estados de fase

```mermaid
stateDiagram-v2
    [*] --> blocked
    blocked --> ready: dependencias satisfechas
    ready --> in_progress: begin
    rejected --> in_progress: begin
    superseded --> in_progress: begin

    in_progress --> awaiting_approval: deliver / gate requerido
    in_progress --> completed: deliver / sin gate

    awaiting_approval --> approved: approve humano
    awaiting_approval --> rejected: reject humano

    approved --> completed: complete phase
    accepted --> completed: complete phase

    blocked --> not_applicable: bypass por entry point posterior
```

### 14.2. Estados de work item

```mermaid
stateDiagram-v2
    [*] --> active: start
    active --> completed: complete
    completed --> archived: pendiente de comando archive
    active --> cancelled: pendiente de comando cancel
```

Actualmente la superficie publica implementa `active -> completed`. Los otros estados existen en el modelo, pero no tienen comando.

### 14.3. Diferencia entre `approved` y `completed`

`approved` significa:

> Una persona acepto el contenido entregado.

`completed` significa:

> La fase termino y sus dependientes pueden considerarla satisfecha.

Por eso una fase con gate sigue:

```text
in_progress -> awaiting_approval -> approved -> completed
```

Una fase sin gate sigue:

```text
in_progress -> completed
```

---

## 15. Modelo de datos conceptual

```mermaid
classDiagram
    class Workflow {
        string id
        string workItemType
        EntryPoint[] entryPoints
        WorkflowPhase[] phases
        map artifacts
    }

    class WorkflowPhase {
        string id
        string[] requires
        string[] produces
        string procedure
        ApprovalPolicy approval
        bool optional
        string[] effects
    }

    class WorkItem {
        string id
        int revision
        WorkItemStatus status
        WorkItemWorkflow workflow
        map phases
        Approval[] approvals
        Traceability traceability
    }

    class PhaseState {
        PhaseStatus status
        string artifact
    }

    class Approval {
        string phase
        ApprovalStatus status
        Actor by
        string at
        string comment
    }

    class Event {
        string id
        string at
        string workItem
        string type
        Actor actor
        map data
        string correlationId
    }

    Workflow "1" *-- "*" WorkflowPhase
    WorkItem --> Workflow : references
    WorkItem "1" *-- "*" PhaseState
    WorkItem "1" *-- "*" Approval
    WorkItem "1" --> "*" Event : traceability
```

El `WorkItem` funciona como **aggregate root**: las mutaciones de fases y approvals se realizan a traves de sus metodos para preservar invariantes.

---

## 16. Flujo de datos general

```mermaid
flowchart LR
    Input["Args + flags"]
    Cobra["cmd<br/>parseo"]
    UseCase["use case<br/>orquestacion"]
    Load["ports de lectura"]
    Domain["dominio<br/>reglas"]
    Artifact["artifact service"]
    Event["clock + ID<br/>eventos"]
    Commit["WorkItemCommit"]
    Repo["FS repository"]
    Schemas["JSON Schemas"]
    Disk["manifest + artifacts + events"]
    Output["texto o JSON"]

    Input --> Cobra --> UseCase
    UseCase --> Load --> Disk
    UseCase --> Domain
    UseCase --> Artifact
    UseCase --> Event
    Domain --> Commit
    Artifact --> Commit
    Event --> Commit
    Commit --> Repo
    Schemas --> Repo
    Repo --> Disk
    Repo --> UseCase --> Cobra --> Output
```

La unidad de persistencia no es un archivo aislado. Es:

```go
type WorkItemCommit struct {
    Item        *domain.WorkItem
    Artifacts   []ArtifactWrite
    Events      []domain.Event
    OperationID string
}
```

Esto mantiene juntos estado, documentos y auditoria.

---

## 17. Recorrido de codigo: arranque y composition root

```mermaid
sequenceDiagram
    participant OS as Sistema operativo
    participant Main as main.go
    participant Root as cmd/root.go
    participant Composition as cmd/composition.go
    participant Cobra

    OS->>Main: ejecuta binario
    Main->>Root: Execute()
    Root->>Composition: NewProductionApplication()
    Composition-->>Root: use cases con adapters reales
    Root->>Cobra: NewRootCommand(application)
    Cobra->>Cobra: parsea comando y flags
    Cobra-->>Main: nil o error
    Main-->>OS: exit 0 o 1
```

Este diseño mantiene `main.go` libre de logica y hace visible el wiring productivo en un solo archivo.

---

## 18. Recorrido de codigo: `sdd start`

`start` es uno de los recorridos mas completos porque crea el agregado, resuelve configuracion, prepara artifacts y emite varios eventos.

### 18.1. Secuencia

```mermaid
sequenceDiagram
    actor Caller as Persona/Agente
    participant Cmd as cmd/start.go
    participant UC as StartWorkItemUseCase
    participant Config as ConfigRepository
    participant WFRepo as WorkflowRepository
    participant Artifacts as ArtifactService
    participant Domain as domain.NewWorkItem
    participant Repo as WorkItemRepository
    participant Disk as .sdd/work-items

    Caller->>Cmd: sdd start feat-x --title ...
    Cmd->>UC: Execute(baseDir, input)
    UC->>UC: valida ID, actor y operation ID
    UC->>Repo: WorkItemExists(id)

    alt workflow no informado
        UC->>Config: GetConfig()
        Config-->>UC: defaults.workflow
    end

    UC->>WFRepo: GetWorkflow(workflowID)
    WFRepo->>WFRepo: schema + semantica + templates
    WFRepo-->>UC: Workflow

    opt input externo
        UC->>Artifacts: ResolveExternalArtifact(path)
        Artifacts-->>UC: path absoluto + SHA-256 + contenido
    end

    UC->>Domain: NewWorkItem(workflow, params)
    Domain->>Domain: crea fases y aplica entrada
    Domain-->>UC: WorkItem + PhaseMutation

    UC->>Artifacts: PrepareArtifactsForPhase()
    opt input externo
        UC->>Artifacts: ImportExternalArtifact()
    end

    UC->>UC: genera eventos
    UC->>Repo: CommitWorkItem(item, artifacts, events)
    Repo->>Disk: publica snapshot atomico
    Repo-->>UC: commit confirmado
    UC-->>Cmd: WorkItem
    Cmd-->>Caller: texto o JSON
```

### 18.2. Que decide cada capa

| Decision | Capa |
| --- | --- |
| `--title` es requerido | `cmd` |
| ID y actor son validos | `domain` |
| Workflow default | `usecase` + `ConfigRepository` |
| Workflow es un DAG valido | `domain`/`infra` |
| Estados iniciales | `domain.NewWorkItem` |
| Template local a utilizar | `infra.ArtifactManager` |
| Eventos de creacion | `usecase` |
| Revision y atomicidad | `infra.FSWorkItemRepository` |

---

## 19. Recorrido de codigo: ciclo de una fase

El ciclo normal usa cuatro comandos:

```text
begin -> deliver -> approve -> complete
```

`approve` se omite para una fase sin gate. `complete phase` es una normalizacion explicita de `approved` o `accepted` a `completed`; no es necesaria para desbloquear dependencias ni para completar el work item.

Si el humano rechaza la entrega, la fase vuelve al circuito de trabajo:

```text
deliver -> reject -> begin -> deliver -> approve | reject
```

```mermaid
sequenceDiagram
    actor Agent
    actor Human
    participant CLI
    participant UC as Use cases
    participant WI as WorkItem domain
    participant Repo

    Agent->>CLI: begin --phase plan
    CLI->>UC: BeginPhase
    UC->>WI: BeginPhase()
    WI-->>UC: ready -> in_progress
    UC->>Repo: commit + transition event

    Agent->>CLI: deliver --phase plan
    CLI->>UC: DeliverPhase
    UC->>WI: DeliverPhase()
    WI-->>UC: in_progress -> awaiting_approval
    UC->>Repo: commit + approval.requested

    alt Aprobacion
        Human->>CLI: approve --phase plan
        CLI->>UC: Approve
        UC->>WI: ApprovePhase()
        WI-->>UC: awaiting_approval -> approved + unlock
        UC->>Repo: commit + approval.recorded + artifacts ready
    else Rechazo y retrabajo
        Human->>CLI: reject --phase plan
        CLI->>UC: Reject
        UC->>WI: RejectPhase()
        WI-->>UC: awaiting_approval -> rejected
        UC->>Repo: commit + approval.recorded
        Agent->>CLI: begin --phase plan
        CLI->>UC: BeginPhase
        UC->>WI: BeginPhase()
        WI-->>UC: rejected -> in_progress
        UC->>Repo: commit + transition event
    end

    opt Normalizacion explicita
    Agent->>CLI: complete --phase plan
    CLI->>UC: CompletePhase
    UC->>WI: CompletePhase()
    WI-->>UC: approved -> completed
    UC->>Repo: commit de la transicion
    end
```

La preparacion del artifact de la siguiente fase sucede al aprobar o completar la dependencia que la desbloquea, no cuando se ejecuta `begin`.

---

## 20. Recorrido de codigo: consulta `next`

`next` demuestra la separacion entre consulta y comando.

```mermaid
sequenceDiagram
    participant Caller
    participant Cmd as cmd/next.go
    participant UC as NextUseCase
    participant Repo
    participant Domain as WorkItem.NextPhase

    Caller->>Cmd: sdd next feat-x
    Cmd->>UC: Execute()
    UC->>Repo: GetWorkItem()
    UC->>Repo: GetWorkflow()
    UC->>Domain: NextPhase(workflow)
    Domain-->>UC: fase activa o nil
    UC-->>Cmd: NextAction
    Cmd-->>Caller: procedure + artifact + gate
```

No existe llamada a `CommitWorkItem`, por lo que:

- no cambia fases;
- no escribe eventos;
- no aumenta `revision`;
- puede ejecutarse repetidamente sin efectos.

---

## 21. Persistencia transaccional

El filesystem no ofrece una transaccion multiarchivo nativa. La infraestructura construye una mediante snapshots y renames.

### 21.1. Algoritmo

```mermaid
flowchart TD
    Start["CommitWorkItem"]
    Lock["Adquirir lock exclusivo"]
    Recover["Recuperar transaccion previa"]
    Idempotency["Verificar operation ID"]
    Revision["Comparar revision"]
    Validate["Validar item, artifacts y eventos"]
    Stage["Crear staging"]
    Copy["Copiar snapshot actual"]
    Write["Escribir artifacts, events y manifest"]
    Sync["fsync de archivos y directorios"]
    Backup["Renombrar actual a backup"]
    Publish["Renombrar staging a activo"]
    DirSync["Sincronizar directorio padre"]
    Success["Revision confirmada"]
    Rollback["Restaurar snapshot anterior"]

    Start --> Lock --> Recover --> Idempotency --> Revision --> Validate
    Validate --> Stage --> Copy --> Write --> Sync --> Backup --> Publish --> DirSync --> Success
    Backup -. fallo .-> Rollback
    Publish -. fallo .-> Rollback
    DirSync -. fallo .-> Rollback
```

### 21.2. Garantias

- Un fallo antes de publicar no cambia el estado visible.
- Un fallo durante la publicacion intenta restaurar el snapshot anterior.
- Una operacion no puede confirmar manifest sin sus eventos o artifacts asociados.
- La siguiente lectura recupera un backup si detecta una interrupcion entre renames.
- Los symlinks y archivos no regulares no se copian dentro del snapshot.

### 21.3. Revision optimista

Cada manifest posee:

```yaml
revision: 7
```

El repositorio compara la revision leida con la revision persistida:

```text
esperada != actual -> concurrent_modification
```

En un commit exitoso:

```text
revision persistida = revision esperada + 1
```

---

## 22. Concurrencia

La politica v0.1 es:

```text
Multiples lectores.
Un solo escritor simultaneo por work item.
```

Cada mutacion intenta adquirir:

```text
.sdd/work-items/.locks/<work-item-id>.lock
```

Si otro proceso ya escribe el mismo item, la operacion falla con `work_item_locked`.

El lock evita escritores simultaneos y `revision` evita sobrescribir una version nueva usando un objeto obsoleto.

Dos work items diferentes pueden mutarse independientemente porque el lock es por ID, no global.

---

## 23. Idempotencia y `operation_id`

Un agente puede perder la respuesta de una operacion aunque la CLI haya confirmado el cambio. Repetir ciegamente el comando podria duplicar eventos o intentar una transicion ya aplicada.

El `operation_id` resuelve ese caso.

```mermaid
sequenceDiagram
    participant Agent
    participant CLI
    participant Events as events.jsonl

    Agent->>CLI: deliver --operation-id op-123
    CLI->>Events: existe correlation_id op-123?
    Events-->>CLI: no
    CLI->>CLI: aplica mutacion y commit
    CLI-->>Agent: respuesta perdida

    Agent->>CLI: mismo comando, mismo op-123
    CLI->>Events: existe correlation_id op-123?
    Events-->>CLI: si
    CLI-->>Agent: devuelve estado persistido
```

Propiedades:

- el mismo ID identifica el mismo intento logico;
- no duplica eventos;
- no vuelve a incrementar revision;
- no ejecuta dos veces la transicion;
- diferentes operaciones deben usar IDs diferentes.

Formato permitido:

```text
1 a 128 caracteres:
letras, numeros, punto, dos puntos, guion bajo y guion
```

Ejemplos:

```text
run:feat-x:plan:deliver:001
agent-session-42.approve.plan
```

---

## 24. Eventos y trazabilidad

Cada linea de `events.jsonl` es un JSON independiente.

```json
{
  "schema_version": "0.1",
  "id": "evt_4d3c...",
  "at": "2026-08-19T12:00:00Z",
  "work_item": "feat-add-coupons",
  "type": "phase.transitioned",
  "actor": {
    "kind": "agent",
    "id": "copilot"
  },
  "data": {
    "phase": "plan",
    "from": "in_progress",
    "to": "awaiting_approval",
    "cause": "phase_delivered"
  },
  "correlation_id": "run:plan:deliver:001"
}
```

### 24.1. Eventos generados por el motor

| Evento | Momento |
| --- | --- |
| `work_item.created` | Creacion del item |
| `phase.bypassed_by_external_input` | Inicio desde artifact externo |
| `phase.transitioned` | Toda transicion principal o derivada |
| `approval.requested` | Entrega que necesita gate |
| `approval.recorded` | Aprobacion o rechazo humano |
| `work_item.completed` | Cierre logico del item |
| Tipo custom | `record-event` |

### 24.2. Orden dentro de una operacion

Para una transicion normal:

1. evento principal de la operacion, si aplica;
2. transicion principal;
3. transiciones derivadas por desbloqueo.

Para input externo:

1. `work_item.created`;
2. `phase.bypassed_by_external_input`;
3. `phase.transitioned`.

### 24.3. IDs y tiempo

- `SystemClock` entrega UTC.
- `CryptoIDGenerator` genera 128 bits aleatorios.
- Los IDs comienzan con `evt_`.
- El timestamp no se usa como ID.

Reloj e IDs son ports inyectables para que los tests sean deterministas.

---

## 25. Validacion estructural y semantica

La CLI aplica dos niveles de validacion.

### 25.1. JSON Schema

| Schema | Protege |
| --- | --- |
| `workflow.schema.json` | Forma de workflows |
| `work-item.schema.json` | Forma de manifests |
| `artifact.schema.json` | Front matter de artifacts |
| `event.schema.json` | Forma de eventos |

### 25.2. Reglas semanticas

El schema no puede expresar todas las invariantes. `Workflow.ValidateSemantics` verifica ademas:

- IDs unicos y kebab-case;
- al menos una fase y un entry point;
- politicas de approval conocidas;
- exactamente un artifact por fase en v0.1;
- dependencias existentes;
- ausencia de auto-dependencias;
- ausencia de ciclos;
- artifacts producidos una unica vez;
- entry points no ambiguos;
- inputs externos compatibles;
- fases alcanzables;
- paths normalizados dentro de `artifacts/`;
- extension Markdown;
- templates existentes y validos.

`WorkItem.ValidateAgainst` verifica:

- workflow y version correctos;
- tipo compatible;
- entry point declarado;
- metadata de artifact externo;
- mismas fases que el workflow;
- paths de artifacts;
- uso valido de `not_applicable`;
- dependencias satisfechas;
- approvals compatibles;
- completitud real de un item `completed`.

---

## 26. Seguridad de paths

La CLI opera sobre paths controlados por IDs y YAML, por lo que protege:

- traversal mediante `../`;
- paths absolutos en ubicaciones internas donde no estan permitidos;
- paths no normalizados;
- artifacts fuera de `artifacts/`;
- symlinks que escapan del root;
- archivos no regulares dentro del snapshot.

La excepcion intencional es `--from-artifact`: puede apuntar a un archivo externo al proyecto. La CLI resuelve y registra su path absoluto, pero nunca lo utiliza como destino de escritura.

Los IDs usan:

```text
^[a-z0-9]+(?:-[a-z0-9]+)*$
```

Ejemplos validos:

```text
feat-add-coupons
bug-123
fast-change
```

Ejemplos invalidos:

```text
../outside
Feature_A
feat/add
```

`containedPath` valida tanto el path textual como los componentes existentes resueltos mediante symlinks.

---

## 27. Artifacts y templates locales

Los recursos embebidos se usan solamente para instalar `.sdd/`.

Durante la ejecucion normal, `ArtifactManager` lee:

```text
<proyecto>/.sdd/templates/<template>.md
```

Esto permite personalizar templates por proyecto sin recompilar la CLI.

### 27.1. Preparacion

Para cada artifact:

1. obtiene la configuracion desde el workflow;
2. lee el template local;
3. reemplaza placeholders;
4. verifica que no queden `{{...}}`;
5. extrae front matter;
6. valida `artifact.schema.json`;
7. verifica `id`, `phase` y `work_item`;
8. devuelve un `ArtifactWrite` aun no persistido.

La escritura real ocurre luego dentro del commit transaccional.

### 27.2. Input externo

La CLI conserva el front matter canonico generado por el motor y reemplaza el cuerpo con el contenido importado.

Ademas registra en el manifest:

```yaml
input:
  source: external_artifact
  external_artifact:
    artifact: plan
    path: /ruta/absoluta/plan.md
    sha256: ...
```

---

## 28. Contrato de salida

### 28.1. Modo humano

Sin `--json`, los exitos se imprimen en stdout y los errores en stderr.

Ejemplo:

```text
Phase 'plan' delivered for work item 'feat-add-coupons'.
```

### 28.2. Modo agente

Con `--json`, stdout contiene un unico envelope:

```json
{
  "success": true,
  "data": {}
}
```

Error:

```json
{
  "success": false,
  "error": {
    "code": "invalid_transition",
    "message": "invalid phase transition: ..."
  }
}
```

### 28.3. Codigos de error

| Codigo | Significado |
| --- | --- |
| `invalid_arguments` | Args, flags o comando incorrectos |
| `invalid_input` | ID, actor, path, schema o contrato invalido |
| `not_found` | Work item, workflow o fase inexistente |
| `already_exists` | Colision al crear |
| `invalid_transition` | Operacion no permitida por el estado |
| `concurrent_modification` | Revision obsoleta |
| `work_item_locked` | Otro escritor posee el lock |
| `internal_error` | Error no clasificado |

El proceso devuelve:

- exit code `0` en exito;
- exit code `1` en error.

---

## 29. DFD del motor

```mermaid
flowchart TB
    Human["Entidad externa:<br/>Persona"]
    Agent["Entidad externa:<br/>Agente IA"]

    P1["P1: Interpretar comando"]
    P2["P2: Ejecutar caso de uso"]
    P3["P3: Aplicar reglas de dominio"]
    P4["P4: Validar contrato"]
    P5["P5: Confirmar snapshot"]
    P6["P6: Formatear respuesta"]

    D1[(".sdd/config.yaml")]
    D2[(".sdd/workflows")]
    D3[(".sdd/templates")]
    D4[("manifest.yaml")]
    D5[("artifacts/*.md")]
    D6[("events.jsonl")]
    D7[(".sdd/schemas")]

    Human -->|args y flags| P1
    Agent -->|args y flags JSON| P1
    P1 -->|input tipado| P2

    D1 --> P2
    D2 --> P2
    D3 --> P2
    D4 --> P2

    P2 --> P3
    P3 -->|nuevo estado| P2
    P2 --> P4
    D7 --> P4

    P4 -->|commit validado| P5
    P5 --> D4
    P5 --> D5
    P5 --> D6
    P5 -->|resultado confirmado| P6

    P6 -->|texto| Human
    P6 -->|envelope JSON| Agent
```

---

## 30. Suite de tests y garantias

La suite esta organizada por propiedades contractuales, no solo por funciones.

| Grupo | Que demuestra |
| --- | --- |
| Dominio table-driven | Estados validos e invalidos |
| Workflow validation | DAG, entry points, dependencies y paths |
| Schema fixtures | Casos validos fallan si dejan de ser validos e invalidos son rechazados |
| Persistencia | Atomicidad, rollback, recuperacion, locks y revision |
| Idempotencia | Reintentos no duplican operaciones |
| Dependency injection | Fallos de ports se propagan sin commits falsos |
| Contract integration | Templates locales, config, artifacts externos y seguridad |
| Workflow lifecycle | Todos los workflows completan su recorrido obligatorio |
| CLI E2E | Binario real, stdout, stderr, JSON y exit codes |

Tests destacados:

- `TestEveryWorkflowCompletesItsMandatoryLifecycle`;
- `TestCLICompletesFastChangeLifecycle`;
- `TestFullWorkItemLifecycle`;
- `TestBypassModeStart`;
- matrices de `BeginPhase`, `DeliverPhase` y `CompletePhase`;
- fallos antes y durante publicacion;
- writers concurrentes y revisiones obsoletas.

La prueba de lifecycle descubre automaticamente todos los `*.workflow.yaml`. Agregar un workflow nuevo lo incorpora a la obligacion contractual sin mantener una lista duplicada.

---

## 31. Como desarrollar y ejecutar la CLI

### 31.1. Preparar recursos y compilar

```bash
cd src/cli
go generate ./...
go build -o sdd .
```

No commitear:

```text
src/cli/sdd
src/cli/embeds/default_sdd/
```

### 31.2. Ejecutar tests

```bash
cd src/cli
go test ./...
go test -race ./...
go vet ./...
```

### 31.3. Formato

```bash
gofmt -w .
```

`gofmt` es el formateador canonico de Go. No se discute estilo de espacios o llaves como en otros ecosistemas: el toolchain impone el formato.

### 31.4. Probar contra otro proyecto

```bash
./sdd init --dir /tmp/my-project

./sdd start feat-example \
  --dir /tmp/my-project \
  --title "Ejemplo"

./sdd status feat-example \
  --dir /tmp/my-project
```

---

## 32. Ejemplo completo: `fast-change`

```bash
SDD=./sdd
PROJECT=/tmp/example-project

$SDD init --dir "$PROJECT"

$SDD start fast-update-copy \
  --dir "$PROJECT" \
  --workflow fast-change \
  --title "Actualizar copy de checkout" \
  --summary "Corregir mensaje de validacion" \
  --operation-id run:start:001

# `plan` comienza automaticamente in_progress al crear el item.
$SDD deliver fast-update-copy \
  --dir "$PROJECT" \
  --phase plan \
  --operation-id run:plan:deliver:001

$SDD approve fast-update-copy \
  --dir "$PROJECT" \
  --phase plan \
  --by matias \
  --operation-id run:plan:approve:001

$SDD complete fast-update-copy \
  --dir "$PROJECT" \
  --phase plan \
  --operation-id run:plan:complete:001

$SDD begin fast-update-copy \
  --dir "$PROJECT" \
  --phase implementation \
  --actor-id copilot \
  --operation-id run:implementation:begin:001

$SDD deliver fast-update-copy \
  --dir "$PROJECT" \
  --phase implementation \
  --actor-id copilot \
  --operation-id run:implementation:deliver:001

# Repetir begin/deliver para verification.
# Human code review requiere deliver -> approve.
# complete --phase es opcional para normalizar approved -> completed.

$SDD complete fast-update-copy \
  --dir "$PROJECT" \
  --operation-id run:work-item:complete:001
```

La CLI gobierna estados y evidencia. El contenido del plan, implementacion y verificacion debe producirlo el agente siguiendo el `procedure` indicado.

---

## 33. Decisiones arquitectonicas relevantes

### 33.1. Workflows como datos

Agregar o modificar un workflow debe hacerse en YAML. La maquina no contiene un `switch` por tipo de feature o bug.

### 33.2. Dominio rico

Las invariantes viven en metodos de `WorkItem` y `Workflow`, no dispersas entre comandos.

### 33.3. Ports pequenos

No existe una interfaz gigante por conveniencia. Cada use case depende de las capacidades que necesita.

### 33.4. Commit conjunto

Manifest, artifacts y eventos no son stores independientes desde la perspectiva de una mutacion. Forman una unica unidad de consistencia.

### 33.5. Consultas puras

`status` y `next` no mutan estado. La accion explicita se realiza con `begin`, `deliver`, `approve`, `reject` o `complete`.

### 33.6. Contrato local

Luego de `init`, el proyecto controla sus workflows y templates locales. El binario no impone silenciosamente la copia original.

---

## 34. Limites actuales y roadmap

### 34.1. Pendiente de alta prioridad

| Capacidad | Estado |
| --- | --- |
| `sdd validate` | No expuesto |
| Integracion con agente orquestador | No implementada |
| Pruebas reales del harness completo | Pendientes |

### 34.2. Pendiente de prioridad media

| Capacidad | Estado |
| --- | --- |
| `sdd archive` | Fase declarativa existente, movimiento fisico ausente |
| Consolidacion de specs baseline | Procedimiento futuro |
| Autorizacion de efectos externos | Fuera del motor actual |

### 34.3. Pendiente posterior

- Engram;
- CodeGraph;
- observabilidad avanzada de tokens;
- adapters especificos por agente;
- catalogo de skills y permisos;
- integraciones GitHub/Azure DevOps.

### 34.4. Distincion importante sobre archive

Hoy pueden coexistir estas dos ideas:

1. una fase `archive` dentro del workflow;
2. el directorio `.sdd/work-items/archive/`.

La primera esta implementada como fase opcional y artifact. La segunda todavia no recibe el work item mediante un comando.

Completar `archive` hoy **no mueve**:

```text
work-items/active/<id>
```

hacia:

```text
work-items/archive/<fecha>-<id>
```

---

## 35. Guia de lectura para incorporarse al proyecto

Orden recomendado:

1. Este documento para comprender el sistema completo.
2. `src/.sdd/workflows/` para ver los procesos reales.
3. `internal/domain/work_item.go` para la maquina de estados.
4. `internal/domain/workflow_validation.go` para las reglas del grafo.
5. `internal/ports/repository.go` para las fronteras.
6. `internal/usecases/start_uc.go` y un use case de transicion.
7. `cmd/composition.go` para ver el wiring.
8. `internal/infra/fs_repository.go` para persistencia.
9. `cmd/cli_e2e_test.go` para observar el comportamiento externo.
10. `internal/usecases/workflow_lifecycle_test.go` para entender la garantia global.

No conviene comenzar por `fs_repository.go`: es el archivo mas largo y contiene detalles mecanicos que se entienden mejor despues de conocer dominio y ports.

---

## 36. Checklist para agregar una capacidad

Antes de implementar un comando nuevo:

1. Definir la regla en el contrato.
2. Verificar si pertenece al dominio o solo a un adapter.
3. Modelar la invariante dentro de `domain`.
4. Definir el port minimo si requiere una capacidad externa.
5. Implementar el use case sin importar `infra`.
6. Conectar dependencias en `cmd/composition.go`.
7. Crear el comando Cobra solo como adapter.
8. Emitir eventos contractuales.
9. Confirmar mediante `WorkItemCommit`.
10. Agregar tests de exito, estados invalidos, fallos de ports y CLI.
11. Actualizar este documento si cambia la superficie publica.

---

## 37. Glosario rapido para desarrolladores

| Termino Go/proyecto | Lectura practica |
| --- | --- |
| Receiver | Objeto sobre el que se ejecuta un metodo |
| Pointer `*T` | Referencia mutable a una instancia |
| Port | Interface requerida por la aplicacion |
| Adapter | Implementacion concreta de un port |
| Composition root | Lugar donde se conectan implementaciones |
| Aggregate root | Entidad que protege las invariantes del conjunto |
| Table-driven test | Matriz de casos recorrida por un test |
| Snapshot | Copia completa y consistente del work item |
| Optimistic concurrency | Rechazar escritura si la version cambio |
| Idempotencia | Repetir la misma operacion logica sin duplicar efectos |
| JSONL | Un objeto JSON por linea |
| Front matter | Metadata YAML al inicio de un Markdown |
| DAG | Grafo dirigido sin ciclos |

---

## 38. Conclusiones

La CLI actual ya no es solamente un prototipo de comandos. Es un nucleo determinista con:

- workflows declarativos;
- maquina de estados centralizada;
- gates humanos;
- entrada desde artifacts externos;
- validacion por schema y semantica;
- paths contenidos;
- artifacts locales personalizables;
- eventos ordenados y correlacionados;
- commits multiarchivo controlados;
- locks, revision e idempotencia;
- output estable para personas y agentes;
- una suite contractual que recorre todos los workflows.

Su siguiente desafio no es endurecer nuevamente el mismo nucleo, sino completar la superficie publica pendiente e integrarla con el orquestador sin romper la frontera fundamental:

```text
El agente realiza el trabajo.
La CLI gobierna el proceso.
El contrato define las reglas.
Git conserva la evidencia.
```
