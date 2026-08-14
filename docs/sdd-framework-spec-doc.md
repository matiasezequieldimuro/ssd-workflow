# CONTRATO: ESCENARIOS, FASES, ARTEFACTOS, COMPONENTES.

## WORKFLOW-BASED SCENARIOS

Los escenarios descriptos a continuación pueden ser desencadenados al proveer distintos inputs; es decir, el usuario puede proveer el input primario inicial de un proceso (como un requerimiento) o bien un artefacto cualquiera del mismo (como el Plan). Este comportamiento tiene como fin flexibilizar la interacción con el framework SDD, ya que el usuario podría contar de antemano con un artefacto, evitando la ejecución de fases innecesarias. 

La situación mencionada previamente es el único caso donde el usuario puede saltear fases del flujo; es decir, el proceso inicia desde el input/artefacto provisto por el usuario. Tanto en esos escenarios (para las fases posteriores) como en cualquier otro, toda etapa debe ser ejecutada sin excepción.

Típicamente, cada fase será documentada por medio de un artefacto; un archivo markdown que sirve como fuente de verdad/conocimiento de toda feature o issue reportado en el proyecto.

Idealmente, en todo proceso de varias fases, las tareas deben ser delegadas desde un agente orquestador a subagentes con objetivos específicos.

En las fases especificadas como “user”, el mismo debe revisar y aprobar sin excepción (Human In The Loop).

(Nombrar logs y engram?)

### Nueva Feature

El Flujo típico es:

1. User: Necesidad puntual de negocio (párrafo).
2. IA: Genera PRD.
3. User: Revisa y aprueba PRD.
4. IA: Genera Especificación.
5. User: Revisa y aprueba Especificación.
6. IA: Elabora Plan.
7. User: Revisa y aprueba Plan.
8. IA: Implementacion (coding).
9. IA: Verificación (unit testing, smoke tests, UI-based end-to-end tests, comparacion con especificaciones).
10. User: Revisa código.
11. IA: Archiva (modifica changelog, commitea, pushea, genera PR).

El user podría proveer la PRD, Especificación o Plan como máximo. 

Los artefactos de esta etapa son PRD, especificación, plan, reporte implementación, pruebas. Toda feature debería contar con estos documentos.

En Engram se debe persistir como memoria un resumen de la implementación de la nueva feature (a grandes rasgos, cual es el requerimiento, la especificación, el plan y como se implementó). Una buena idea es dejar referencias de archivos relevantes.

Quienes participan del flujo?
1. Agente Funcional: encargado de la redacción de requerimientos (PRD) y Especificación.
2. Agente Planificador: encargado de explorar, investigar y analizar detallada y cuidadosamente como implementar la feature. Es, probablemente, el mas critico y relevante, de el dependerá, en gran medida, el éxito del proceso. Debe documentar el Plan lo mas claro y completo posible como codificar la solución.
3. Agente Desarrollador: encargado de implementar la feature siguiendo el plan generado. Al final, crea un Reporte de Implementación explicando los cambios realizados en el codebase. Adicionalmente, si así se comunica, debería escribir tests.
4. Agente verificador: encargado del test de la feature (ejecutar tests unitarios, de integración, smoke tests, probar la UI en vivo para ver si cambios visuales (en el frontend) funciónan correctamente).
5. Agente Archivador: encargado de commitear, pushear al repo y generar PRs.
6. Agente Orquestador: encargado de coordinar el flujo y es quien interactúa con el user.

Que skills podriamos definir? (SDD only, NO tecnologías o librerías de código x el momento).
1. Generar PRD
2. Generar Especificación 
3. Generar Plan
4. Generar Reporte de Implementación
5. Archivar
    1. Modificar Changelog
    2. Generar commit.
    3. Generar ADO/GitHub PR.

Otras que podrían ir, mas adelante, son:
1. Desarrollo CAP.
2. Desarrollo UI5.
3. Desarrollo SOLID y Clean Code.
4. UI Testing.

Que MCPs podrían incluirse?
- Engram.
- Context7
- Azure DevOps (y/o uno para github)
- CAP, UI5, Fiori (mas adelante)
- Codegraph (mas adelante)

Que deberíamos guardar en los logs? Principalmente, toda interacción de un agente: Timestamp, Agente, Modelo, Tokens Input, Tokens Salida, Tokens Cache, Descripción, etc.

### Actualizar Feature

El flujo es el mismo que “Nueva Feature” con la distinción del Change Request (CR) en vez de una PRD. El usuario, podría ingresar un párrafo con un cambio de requerimientos, o bien un archivo markdown CR.md.

Se sumaria una skill “Generar CR”, en vez de “Generar PRD”.

Todas las demás fases, artefactos y herramientas son las mismas.

### Ajuste Rápido Feature

Partiendo de un PRD o CR como referencia o adjunto, el usuario propone realizar una serie de cambios, intuyendo QUE es lo que quiere principalmente, pudiéndose complementar de COMO hacerlo. 

Estos cambios no requieren de mucha burocracia dado que, como su nombre lo indica, es un cambio pequeño y se desea ser rápido.

Con los inputs del usuario, se debe generar un PLAN. Luego, el flujo es idéntico al “Nueva Feature”.

### Debug Rápido

El Flujo típico es:

1. User: Issue detectado en el sistema, breve exploración donde ha identificado posibles causas y/o código a modificar para solucionarlo, pertenece a una feature en particular.
2. AI: Documenta brevemente el Issue + Exploración.
3. AI: Genera Plan detallado de acción.
4. User: Revisa y aprueba Plan.
5. …

Las demás fases son idénticas a las de “Nueva Feature”

1. User: Necesidad puntual de negocio (párrafo).
2. IA: Genera PRD.
3. User: Revisa y aprueba PRD.
4. IA: Genera Especificación.
5. User: Revisa y aprueba Especificación.
6. IA: Elabora Plan.
7. User: Revisa y aprueba Plan.
8. IA: Implementacion (coding).
9. IA: Verificación (unit testing, smoke tests, UI-based end-to-end tests, comparacion con especificaciones).
10. User: Revisa código.
11. IA: Archiva (modifica changelog, commitea, pushea, genera PR).

El user podría proveer la PRD, Especificación o Plan como máximo. 

Los artefactos de esta etapa son Issue, Exploración, plan, reporte implementación, pruebas. Toda feature debería contar con estos documentos.

En Engram se debe persistir como memoria un resumen de la implementación del fin (a grandes rasgos, cual es el issue, la causa raiz, el plan y como se implementó). Una buena idea es dejar referencias de archivos relevantes.

Quienes participan del flujo? Los mismos, hasta ahora, que en los demás flujos. 
Se agrega alguna skill nueva? Sí, “Generar Issue”, “Documentar Exploración Issue”.

### Debug Profundo

Este flujo es idéntico al anterior, solo que presenta una fase mas: debugging; dado que el user no ha identificado ni explorado la causa raiz del issue.

El user reporta, en un párrafo, el issue, la IA lo reescribe y, posteriormente, comienza a debuguear (la fase de exploración). Luego, el flujo sigue como antes.

Aparece un nuevo agente “debugger”


## ESCENARIOS AUXILIARES

Los escenarios descriptos a continuación no constan de numerosas fases o etapas con aprobación mandatoria; son mas simples. 

Delegar a subagentes no es una regla automática. Sólo conviene cuando existe trabajo independiente, riesgo o complejidad suficiente para compensar el costo de rehidratar contexto. Las consultas pequeñas pueden resolverse inline. Investigación, planificación, implementación o revisión pueden delegarse cuando el contexto fresco y la separación de responsabilidades aporten valor.

A pesar de no incluir fases, toda interacción podría ser documentada si así lo desea el usuario.

### Setup

Considerando se esta iniciando un nuevo proyecto.
Generar AGENTS.md (o equivalente como CLAUDE.md, copilot-instructions.md, etc), .mcp.json, skills, comandos, carpeta “.sdd”, skill registry, etc.

Idealmente, debería ser un comando de la CLI, al que el user podría acceder mediante comando, y la IA ejecutar mediante skill. Debe dejar el proyecto con todo configurado para poder usar SDD.

Estaria bueno reporte cuantos tokens gasta, y mismo quede en los logs y Engram.

### Onboarding

Considerando se utiliza en proyecto existente.

Deben hacer un “setup” del proyecto (skill anterior).

Ademas, debe explorar el repositorio para comprender de que trata el proyecto y como esta construida la solución. Para ello, deberá leer archivos de documentación -como README.md, ARCHITECTURE.md, TESTING.md u otro- y codebase para generar:

- SOFTWARE_ARCHITECTURE.md
- PROJECT_ARCHITECTURE.md
- DATA_MODELING.md
- DATA_AND_PROCESS_FLOW.md
- DOMAIN_LANGUAGE.md

Estos artefactos están documentados en la sección “OTROS ARTEFACTOS”.

Correspondería a una skill. Idealmente, debería reportar, al final de la interacción, cuantos tokens han sido consumidos.

### Consulta Técnica o Funcional.

El user puede tener dudas técnicas o funcionales del proyecto. Teniendo disponibles distintas fuentes de conocimiento (codebase, artefactos, memoria, logging, lenguaje común de dominio, etc), se debe responder al usuario exactamente aquello que ha preguntado.

Correspondería, mas bien, a una skill. Idealmente, debería reportar, al final de la interacción, cuantos tokens han sido consumidos 

## OTROS ARTEFACTOS

Hay artefactos que no corresponden a una fase específica ni a una feature en particular. Es el caso de la documentación general del proyecto como:

### Arquitectura de software

Documento orientado a la arquitectura general de software. Describe los distintos servicios/aplicaciones del sistema (IdP, API Gateway, BD, Seguridad, servicios externos, backend, frontend, etc) y la plataforma/infraestructura de despliegue. Es de alto nivel, permite conocer y entender cuales son todos los componentes del sistema a nivel general.

Idealmente, debería mostrar mediante gráficos Mermaid el diagrama, explicarlo brevemente con texto, mencionando cada uno de los componentes y luego esplashandose un poco mas sobre el/los aplicativo/s del repositorio actual (si por ejemplo, el repositorio contiene al frontend y backend, dar un poco mas de detalle de estos 2 y explicar como interactúan / que rol cumplen en el sistema).

El objetivo es comprender a grandes rasgos cuales son los componentes a nivel plataforma del sistema y como se vinculan unos a otros.

Comunmente, en los repositorios, esta info se encuentra en README.md o ARCHITECTURE.md. En el caso de no contar con la info. necesaria NO asumir, el codebase no refleja siempre la arquitectura general.

### Arquitectura del proyecto

Documento orientado a la arquitectura de un proyecto/modulo específico del repositorio (si se tiene 2 UI frontend y 1 backend, serian 3 documentos). 

A modo exploración, sera necesario leer documentación existente y codebase para luego cruzarlas (siempre prioriza el codebase, es la fuente de verdad).

Tiene como objetivo explicar:
- Stack Tecnológico utilizado.
- Como es la estructura de carpetas y archivos.
- Patrones de diseño utilizados.
- Nomenclaturas, convenciones.

El propósitósito es entender la tecnología, arquitectura de proyecto (hexagonal, clean, mcv, etc) y reglas/patrones aplicados.

### Modelado de datos

Documento orientado a explicar las entidades definidas (cuando mediante código se crean las tablas/vistas) o consumidas (se obtienen datos de un sistema externo) del sistema.

Para la exploración, revisar en los archivos típicos markdown como README.md o ARQUITECTURE.md si se documenta el flujo de datos y entidades. En cualquier caso, también revisar el codebase para obtener dicho conocimiento.

Se debe generar en formato Mermaid:

- DER -Diagrama entidad Relacion- o Modelo Lógico -> entender que entidades existen, sus campos, como se conectan.

Adicionalmente, explicar brevemente cada entidad y campo.

### Diagrama de Secuencia y Flujo

Documento orientado a explicar como fluyen los datos en todo proceso / caso de uso , punta a punta, desde que el usuario -por ejemplo- clickea un botón hasta que la UI los renderiza.

Se debe revisar documentación -como README.md o ARCHITECTURE.md- y codebase para entender que componentes (clases, funciones, DTOs, etc) manejan los datos y como fluye la información.

Generar en formato Mermaid:
- Diagrama de secuencia.
- DFD - Diagrama de flujo de datos -.

### Lenguaje Común Dominio

Documento orientado a generar un lenguaje que tanto usuario/desarrollador como IA entiendan sobre palabras/frases/conceptos a nivel funcional que se utilizan para referirse a ciertos procesos/componentes de la solución. Por ejemplo, si mi aplicación se encarga de automatizar el cierre contable y se integra en un sistema financiero externo que nombra a los cierres como “listas”, los empleados funcionales (no técnicos) pueden no saber a que nos referimos los desarrolladores (técnicos) cuando hablamos de “listas”. Lo mismo podría ocurrir con nombres de campos técnicos u otros ejemplos. Este documento es un manual donde registramos sinónimos o términos/expresiones equivalentes. Normalmente, el usuario lo va contemplando progresivamente. SDD solo genera el template.

## OTRAS NOTAS
- Skill Registry, archivo markdown con referencias a las skills disponibles para que el orquestador las digiera y luego entregue al subagente las que este necesite, y ahorrar tokens.
- Para cada fase/escenario, un comando que sirva de interfaz para el user.
- Cada flujo debe ser capaz de documentar la cantidad de tokens que han sido empleados.
- Quiero que todo quede organizado en features, dentro habrá artefactos.
- Recordar que Engram, Codegraph y otros MCPs/Skills se quedan fuera hasta tener un SDD robusto.
- Hay artefactos que ire generando hasta llegar a uno que realmente me guste -> en ese caso, debo guardarlo como EJEMPLO que hoy no tengo para mostrarle al LLM

## NEXT STEPS
1. Definir contrato YAML de workflows, estructura de .sdd, etc.
2. Generar skills, agentes, comandos, settings.json y MCPs basicos, lo mas simple posible para luego iterar.
3. CLI (evolutivo)
4. Probar CLI manual, luego instruir al Orquestador.
5. Agregar Engram Codegraph y el resto.

---

Created on: 2026-08-13  
Last modified: 2026-08-13
---