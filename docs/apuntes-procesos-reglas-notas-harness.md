# Flujo SDD

## Escenarios Típicos de Trabajo

- Desarrollar una nueva feature desde cero.
- Realizar un cambio, ajuste, change requirement en una feature existente.
- Corregir un bug o un issue de una feature existente.
- Realizar una investigación o análisis de viabilidad de cierta tecnología sin desarrollo, nada, solo exploración, y evaluar con codebase actual también.
- Realizar consultas técnicas o funcionales del proyecto o bien de la feature existente.
- Realizar un pequeño onboarding en un proyecto nuevo, donde debo entender el código, patrones, estructuras, etc.

## Paso a paso de cada escenario - Proceso, Reglas, Etapas

Nueva Feature
- Se parte de una necesidad puntual de negocio (en un párrafo) o bien de algún documento (artefacto) funcional (PRD, especificación o plan técnico).
- Toda feature, idealmente, debería contar con un PRD, Especificación, Plan y Reporte de Implementación, en ese orden. Si se provee, inicialmente, otro que no sea el primero, el flujo continuaría, en es mismo orden, desde ese mismo.
- Así, se podrían distinguir las siguientes etapas (considerando que se parte en el orden indicado arriba): problema de negocio (user) --> PRD (AI) —> Ve rificación (user) —> Especificación (AI, con alguna posible sugerencia del user) —> Verificación (user) —> Plan (AI) —> Verificación (user) —> Implementación (AI) -> Validación (AI) —> Reporte Implementación (AI) —> Verificación Mínima (user) —> Archivar (AI).
- Cable aclarar que el user podría proveer la PRD o la especificación, y continuar desde ese punto.

Cambio en Feature Existente
- En el caso de un cambio en los requerimientos o en la especificación, se debe generar un nuevo docuemento de tipo CR (Change Request) - similar a un PRD pero explicando los cambios respecto a la documentación original.
- El flujo, etapas y documents (a excepción del PRD) son los mismos que la Nueva feature.

Ajuste puntual y especifico Feature Existente (FAST - Saltea fases)
- Sea de acuerdo a un PRD o un CR, se parte de una serie de cambios (presentados por el usuario) a realizar. El usuario sabe lo que quiere, en lineas generales por lo menos (de QUE hacer principalmente, pudiéndose complementar de un COMO hacerlo). 
- Orientado para cambios o fixes que no requieran mucha exploración o investigación .
- No requiere realizar una gran especificación, debería alcanzar con los inputs del user.
- En base a ellos, se elabora un Plan detallado para su posterior Implementación.
- Desde el Plan (inclusive) en adelante, el proceso sigue el orden de los dos escenarios previos.

Debugging de feature existente (FAST - Saltea fases)
- Se narra un issue nuevo detectado en el sistema. Pertenece a una feature en particular.
- El user sabe que paso, donde y -por lo menos a grandes rasgos- como podría solucionarse (provee referencias y contexto necesarios para que la IA se saltee la exploración profunda).
- Se documenta brevemente el issue + la exploración, y se genera un PLAN detallado de acción. El user lo verifica. Si lo aprueba, se implementa y se sigue el mismo flujo y etapas comentados en las secciones anteriores (implementación -> validación -> reporte -> verificación -> archive) 
- Que diferencia hay con sección “Ajuste puntual (FAST)”? Principalmente el tipo de cambio: debe quedar registrado y documentado el tipo de cambio: fue una modificación simple de código? Fue un CR? Fue un issue que encontré yo y yo mismo detecté como soluciónar? Lo delegue a la IA para debuguear?
- En este caso tendríamos 2 documentos nuevos: issue (el user lo provee, la IA lo reescribe) + exploración de cambios (es el reporte del user de sus ideas sobre el causante raiz del issue y los cambios sugeridos para el fix). 

Debugging + Ajuste puntual de feature existente.
- Idem punto anterior, solo que con una fase mas: debugging.
- El debugging consiste en exploración profunda (sea revisar codebase, librerías o probar en vivo la app/servicio) del issue reportado. 
- Primero el user reporta el issue, la AI lo reescribe. Posteriormente, comienza a debuguear. Cuando encuentra la causa raiz, genera un documento Exploración a modo reporte y, en base a el, genera un Plan.
- La diferencia respecto al anterior es que el debugging y Exploraciones performado 

Consulta Técnica / Funcional
- El user puede tener dudas técnicas o funcionales del proyecto.
- Se debe contar con distintas fuentes de conocimiento: codebase (sean archivos o grafo indexado de los mismos), artefactos de features (sean PRDs, CRs, issues), memoria (Engram por ejemplo), archivos de logging y monitoreo (a revisar este punto, tal vez pertenezca a toda feature) o bien documentación general del mismo (como README.md, un archivo con lenguaje común de dominio, ARCHITECTURE.md, etc).
- Se busca info en las fuentes q sean necesarias y se devuelve la respuesta al usuario.

Investigación / Análisis de Viabilidad
- El usuario reporta una necesidad de investigar cierta tecnología/framework.
- Se debe investigar en documentación oficial.
- Luego, adicionalmente, si el usuario lo solicita, se evalúa la viabilidad de integrarlo en el proyecto, compatibilidad, etc.
- A su vez, si el usuario lo solicita, se documenta en un apartado de investigaciones.

Onboarding
- Se clona un proyecto nuevo y el usuario quiere integrarse rápida, fácil y cómodamente. Al mismo tiempo, se debe hacer un Setup del framework SDD para que pueda ser utilizado correctamente.
- Primero, se genera/copia el template del framework SDD (estructura básica, AGENTS.md, .mcp.json y otros archivos relevantes) - algunos archivos-documentos (como el lenguaje común de dominio, utilizado para que la IA y el user hablen un mismo idioma sobre los distintos elementos del dominio) comenzaran vacios, y el user tendrá que ir modificándolos con el tiempo. 
- Luego, se corre algun comando para que la IA analice el proyecto y genere la documentación necesaria (resumen general del proyecto, diagrama de arquitectura, estructura del proyecto -carpetas, patrones de diseño utilizados, etc), diagrama de flujo de datos, diagrama de secuencia, etc).


## Componentes - Subagentes, Skills, Comandos, Hooks, MCPs.
TODO

## Base de Conocimientos - Artefactos, Memoria, Logs, Costos.
TODO


## Reglas Generales

- No se debería realizar cambios masivos de una misma de manera simultánea. Es importante ir paso a paso y revisar cada uno. 
- Toda interacción (sea una investigación, implementación o un ajuste), debe ser documentado.
- El usuario podría querer saltear el archive por ejemplo, es lo único que podría no realizar. Todas las demás etapas deben cumplirse obligatoriamente, muchas de ellas con documentación.


## Dudas consultadas

* ¿Conviene un agente definido específicamente o construir, en su lugar, una skill que spawnee un agente default? -> Los agentes permiten definir una identidad/contrato/tools/permisos específicos, que no es posible mediante una skill.
* ¿Convienen comandos o skills para prompts reutilizables por mí mismo? -> Las skills son tareas repetitivas que puede ejecutar tanto la IA como el user. El comando, es un atajo tipo “prompt” reutilzable. Se suelen combinar para que el Comando mande en un prompt determinado ciertos parámetros y active/desencadene la Skill necesaria; que explicara el paso a paso de como realizar algo, usando esos parámetros.

## Otras Dudas
- Que es OpenSpec y como me puede ayudar a desarrollar mi framework harness SDD? Es la mejor forma? Hay otros?
- Que puedo configurar en settings.json? Es lo mismo para todos los agentes o cambia? Como establezco limites y scope? Principio Least Privilege.
- Revisar: OPEN SPEC?CODEGRAPH?TDD?

## Ideas a modo recordatorio para implementar
- Interaccion INLINE vs DELEGACION: para tareas con mas de 1 etapa, conviene delegar a subagentes (contexto fresco principalmente). Para consultas puntuales (por ejemplo funciónales, algo técnico puntual) conviene INLINE. En la DELEGACION se ve con mayor presencia el poder del ORQUESTADOR. En INLINE, el orquestador responde x si solo.
- Doble memoria (Engram y md)
- Distintos modos: Modo Team Lead (yo delego), Modo Junior (me explican, aprendo, yo codifico). El modo Team Lead esta orientado a trabajo diario profesional, el Junior para aprender mientras trabajo, forzarme a ir mas despacio, codificar alguna q otra cosa yo.
- Los agentes deben logear que hacen y mismo se deberían guardar los para saber el estado de un proyecto en un momento dado y, aparte, logs sobre que hizo cada agente.

---

Created on: 2026-08-06  
Last modified: 2026-08-08
