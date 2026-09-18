# Spec: Inicialización del contrato SDD

## Propósito

Instalar la estructura `.sdd/` en un proyecto existente y, opcionalmente, desplegar el adapter de Claude Code, dejando el proyecto listo para operar el motor SDD sin configuración manual.

## Comportamiento observable

### `sdd-cli init`

- Verifica que `.sdd/` no exista en el directorio destino; falla si ya existe (no sobrescribe).
- Copia los recursos embebidos en el binario: workflows, schemas, templates, procedures, config, registry.
- Crea los directorios `work-items/active/` y `work-items/archive/`.
- No inicializa Git ni modifica el proyecto más allá de `.sdd/`.
- Resultado: directorio `.sdd/` con el contrato completo, listo para `sdd-cli start`.

### `sdd-cli adapters list`

- Lista los adapters embebidos en el binario con ID, título y descripción.
- En la BETA: un único adapter (`claude-code`).
- No muta estado.

### `sdd-cli adapters install <id>`

- Precondición: `.sdd/` inicializado (`init` previo).
- Comprueba todas las colisiones de archivos antes de escribir; si algún destino existe, falla sin escribir nada (no hay `--force` en la BETA).
- Instala en el proyecto:
  - `CLAUDE.md` — instrucciones del framework SDD.
  - `.mcp.json` — configuración MCP del adapter.
  - `.claude/agents/sdd-*.md` — agentes especializados.
  - `.claude/skills/<skill>/` — skills invocables.
  - `.claude/commands/sdd/` — comandos slash.
  - `.claude/hooks/` — hooks de protección.
  - `.claude/settings.local.json.example` — plantilla de settings privados (el real queda ignorado por `.gitignore`).

## Entradas

| Entrada | Descripción |
| --- | --- |
| `--dir` | Directorio del proyecto destino (default: `.`) |
| `<id>` | ID del adapter a instalar (`claude-code`) |

## Salidas

- `sdd-cli init`: directorio `.sdd/` con el contrato completo.
- `sdd-cli adapters install`: archivos del adapter copiados al proyecto.
- Salida en stdout: lista de archivos instalados (texto o JSON).

## Reglas de negocio

- `init` es idempotente en el sentido negativo: falla si `.sdd/` ya existe.
- `adapters install` falla atómicamente si hay cualquier colisión: no deja instalación parcial.
- El contrato embebido en el binario es la fuente de verdad de la versión instalada.

## Dependencias

- `sdd-cli` binario (contiene el contrato y adapters embebidos).
- Sistema de archivos local con permisos de escritura en el directorio destino.
