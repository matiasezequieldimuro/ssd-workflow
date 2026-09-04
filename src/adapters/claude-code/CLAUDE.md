# Claude Code - SDD adapter

Este proyecto usa el motor SDD como autoridad determinista del proceso.

## Inicio de una tarea

1. Leer `.sdd/registry/capabilities.yaml` para descubrir capacidades disponibles.
2. Consultar `sdd-cli status`, `sdd-cli next --json` o `sdd-cli validate` segun corresponda.
3. Cargar solamente el procedure requerido por la capability elegida.

## Limites

- No editar manualmente manifests, estados ni eventos.
- No saltar fases ni autoaprobar gates humanos.
- No duplicar el contenido de `.sdd/procedures/` en skills, agents o commands.
- Delegar solo cuando la separacion de contexto compense el costo.

<!-- TODO: completar instrucciones especificas del proyecto. -->
