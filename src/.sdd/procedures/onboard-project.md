# Documentar proyecto existente

## Objetivo

Construir contexto estable y verificable de un repositorio existente para reducir exploraciones repetidas.

## Precondiciones

- El contrato SDD está inicializado.
- Se dispone del código del proyecto o de la mayor parte accesible de este.
- Las fuentes faltantes, contradictorias o todavía desconocidas pueden declararse explícitamente.

## Resultado

Actualizar los siguientes documentos bajo `.sdd/context/`:

- `project.md`: propósito, alcance, límites, comandos y referencias.
- `architecture/software-architecture.md`: arquitectura general, sistemas conectados, autenticación, plataforma y despliegue.
- `architecture/project-architecture.md`: estructura concreta del repositorio, capas, convenciones y patrones implementados.
- `architecture/data-modeling.md`: entidades, relaciones, persistencia, validaciones y migraciones conocidas.
- `architecture/data-and-process-flow.md`: flujos de datos, procesos principales y diagramas relevantes.

Priorizar el codebase y las fuentes verificables. No inventar contexto ausente: marcar como pendiente o desconocido lo que no pueda confirmarse. No completar `domain-language.md`; ese documento se construye progresivamente durante el trabajo con el proyecto.
