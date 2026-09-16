# Guia de instalacion de sdd-cli

> Estado: BETA.
> Audiencia: usuarios que quieren instalar y usar el motor SDD.
> Alcance: instalacion del binario `sdd-cli` en Windows, macOS y Linux (Ubuntu/WSL).

---

## 1. Que voy a instalar

`sdd-cli` es **un unico binario nativo** escrito en Go. No necesita Node.js, Python
ni ningun runtime: se descarga un solo ejecutable, se pone en el `PATH` y queda listo.

- No hay instalador con dependencias ni servicios en segundo plano.
- "Instalar" = copiar el ejecutable a una carpeta que este en tu `PATH`.
- "Actualizar" = reemplazar ese ejecutable por uno mas nuevo.
- "Desinstalar" = borrar ese ejecutable.

Los binarios se publican en **GitHub Releases**:
<https://github.com/matiasezequieldimuro/ssd-workflow/releases>

---

## 2. Que archivo me corresponde

Elegi el archivo segun tu sistema operativo y arquitectura:

| Sistema operativo | Arquitectura | Archivo a descargar |
| --- | --- | --- |
| Linux (Ubuntu, WSL) | x86_64 / amd64 | `sdd-cli_<version>_linux_amd64.tar.gz` |
| macOS Intel | x86_64 / amd64 | `sdd-cli_<version>_darwin_amd64.tar.gz` |
| macOS Apple Silicon (M1/M2/M3) | arm64 | `sdd-cli_<version>_darwin_arm64.tar.gz` |
| Windows | x86_64 / amd64 | `sdd-cli_<version>_windows_amd64.zip` |

`<version>` es la etiqueta del release, por ejemplo `v0.1.0-beta`.
Cada release incluye `SHA256SUMS` para verificar la integridad.

Como saber tu arquitectura:

```bash
# macOS / Linux
uname -m        # x86_64 -> amd64 ; arm64 o aarch64 -> arm64
```

```powershell
# Windows
$env:PROCESSOR_ARCHITECTURE   # AMD64 -> amd64
```

---

## 3. Metodos de instalacion

Hay tres caminos. Elegi **uno**:

- **3.1 Rapida (recomendada):** un comando, resuelve todo automaticamente.
- **3.2 Manual desde Releases:** descargas el archivo vos mismo.
- **3.3 Compilar desde fuente:** requiere Go, util para desarrollo.

### 3.1. Instalacion rapida (recomendada)

Detecta tu plataforma, descarga el binario correcto del ultimo release, verifica el
checksum y lo instala en el `PATH`.

**macOS y Linux (Ubuntu/WSL):**

```bash
curl -fsSL https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/install.sh | sh
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/install.ps1 | iex
```

Opciones (variables de entorno):

| Variable | Efecto |
| --- | --- |
| `SDD_VERSION` | Instala una version especifica (ej. `v0.1.0-beta`). Default: ultimo release. |
| `SDD_INSTALL_DIR` | Carpeta destino. Default: si ya hay un `sdd-cli` en el `PATH`, lo sobrescribe en su lugar; si no, `/usr/local/bin` (o `~/.local/bin`) en Unix y `%LOCALAPPDATA%\Programs\sdd-cli` en Windows. |

```bash
# Ejemplo: version fija y carpeta propia
SDD_VERSION=v0.1.0-beta SDD_INSTALL_DIR="$HOME/.local/bin" \
  curl -fsSL https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/install.sh | sh
```

### 3.2. Instalacion manual desde Releases

Usa este metodo si preferis no ejecutar el script. Es el mismo para instalar por
primera vez o para actualizar (sobrescribe el binario).

> **Tip:** definimos `VERSION` una sola vez al principio y el resto de los comandos
> se copian y pegan tal cual. Cambia solo esa linea por la version que quieras
> (mira la ultima en la [pagina de Releases](https://github.com/matiasezequieldimuro/ssd-workflow/releases)).

**Linux (Ubuntu / WSL) — amd64:**

```bash
VERSION="v0.1.0-beta"     # <-- unica linea a ajustar

cd /tmp
curl -fsSLO "https://github.com/matiasezequieldimuro/ssd-workflow/releases/download/${VERSION}/sdd-cli_${VERSION}_linux_amd64.tar.gz"
curl -fsSLO "https://github.com/matiasezequieldimuro/ssd-workflow/releases/download/${VERSION}/SHA256SUMS"

# Verificar integridad (debe imprimir: ...linux_amd64.tar.gz: OK)
sha256sum -c SHA256SUMS --ignore-missing

# Instalar (sobrescribe cualquier version previa en esa ruta)
tar -xzf "sdd-cli_${VERSION}_linux_amd64.tar.gz"
sudo install -m 0755 sdd-cli /usr/local/bin/sdd-cli

hash -r                   # limpia el cache de rutas del shell
sdd-cli version
```

**macOS — Intel (amd64) o Apple Silicon (arm64):**

```bash
VERSION="v0.1.0-beta"
ARCH="arm64"              # Intel: usa "amd64"

cd /tmp
curl -fsSLO "https://github.com/matiasezequieldimuro/ssd-workflow/releases/download/${VERSION}/sdd-cli_${VERSION}_darwin_${ARCH}.tar.gz"
curl -fsSLO "https://github.com/matiasezequieldimuro/ssd-workflow/releases/download/${VERSION}/SHA256SUMS"

shasum -a 256 -c SHA256SUMS --ignore-missing

tar -xzf "sdd-cli_${VERSION}_darwin_${ARCH}.tar.gz"
sudo install -m 0755 sdd-cli /usr/local/bin/sdd-cli

# macOS: el binario no esta firmado por Apple; autorizalo si Gatekeeper lo bloquea
xattr -d com.apple.quarantine /usr/local/bin/sdd-cli 2>/dev/null || true

hash -r
sdd-cli version
```

**Windows (PowerShell) — amd64:**

```powershell
$Version = "v0.1.0-beta"
$Dir = "$env:LOCALAPPDATA\Programs\sdd-cli"

$zip = "$env:TEMP\sdd-cli.zip"
Invoke-WebRequest -Uri "https://github.com/matiasezequieldimuro/ssd-workflow/releases/download/$Version/sdd-cli_${Version}_windows_amd64.zip" -OutFile $zip -UseBasicParsing

New-Item -ItemType Directory -Path $Dir -Force | Out-Null
Expand-Archive -Path $zip -DestinationPath $Dir -Force

# Agregar al PATH de usuario (una sola vez)
$userPath = [Environment]::GetEnvironmentVariable('Path','User')
if ($userPath -notlike "*$Dir*") {
    [Environment]::SetEnvironmentVariable('Path', "$userPath;$Dir", 'User')
}
# Abri una NUEVA terminal y luego:
sdd-cli version
```

### 3.3. Compilar desde fuente

Requiere [Go 1.22 o superior](https://go.dev/dl/).

```bash
git clone https://github.com/matiasezequieldimuro/ssd-workflow.git
cd ssd-workflow/src/cli

go generate ./...    # sincroniza recursos embebidos (.sdd y adapters)
go build -ldflags "-X sdd-cli/cmd.version=v0.1.0-dev" -o sdd-cli .

sudo install -m 0755 sdd-cli /usr/local/bin/sdd-cli   # opcional: dejarlo en el PATH
sdd-cli version
```

> **Cuidado al mezclar `go install` con la instalacion por curl/release.** `go install`
> deja el binario en `~/go/bin/sdd-cli` (tu `GOBIN`/`GOPATH/bin`), mientras que el
> instalador rapido o la descarga manual lo dejan en `/usr/local/bin` o `~/.local/bin`.
> Si tenes ambos, **el que gana es el que aparece primero en tu `PATH`**, no el que
> instalaste ultimo: por eso podes "instalar la version nueva" y seguir ejecutando una
> vieja. Ademas, un binario hecho con `go install`/`go build` sin `-ldflags` reporta
> `version = dev`, asi que no siempre distinguis cual estas corriendo. Recomendaciones:
> para desarrollo usa un solo camino (por ejemplo `go run .` o `go build -o /tmp/sdd .`
> e invocalo por ruta), o mantene dev y release en carpetas separadas y controla el
> orden del `PATH`. Ante la duda, corre `type -a sdd-cli` (macOS/Linux) o
> `Get-Command sdd-cli -All` (Windows) para ver todas las copias, y `sdd-cli version`
> para confirmar cual se ejecuta. El instalador rapido, cuando ya existe un `sdd-cli`
> en el `PATH`, sobrescribe esa copia activa y avisa si detecta otras.

---

## 4. Verificar la instalacion

```bash
sdd-cli version      # muestra version, commit y fecha de compilacion
sdd-cli --help       # lista los comandos disponibles
```

Salida esperada de `sdd-cli version`:

```text
sdd-cli v0.1.0-beta (commit 91e5b13, built 2026-09-16T15:03:14Z)
```

Para confirmar que ejecutable se esta usando y detectar copias viejas:

```bash
type -a sdd-cli      # macOS/Linux: lista TODAS las copias en el PATH, en orden
```

```powershell
Get-Command sdd-cli -All   # Windows
```

---

## 5. Reinstalar / actualizar

Actualizar es simplemente **reemplazar el binario** por uno mas nuevo.

- **Si instalaste con el metodo rapido (3.1):** volve a ejecutar el mismo comando;
  sobrescribe el binario con el ultimo release.

  ```bash
  # macOS / Linux
  curl -fsSL https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/install.sh | sh
  ```

  ```powershell
  # Windows
  irm https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/install.ps1 | iex
  ```

- **Si instalaste manual (3.2):** repeti los pasos con la nueva `VERSION`. El
  comando `install -m 0755 ...` sobrescribe el binario existente en su lugar.

> **Importante:** despues de actualizar corre `hash -r` (bash/zsh) y verifica con
> `type -a sdd-cli` que no haya quedado **otra copia vieja antes en el PATH**. Si la
> hay, borrala (ver seccion 6). Es la causa mas comun de "ejecute la version nueva
> pero sigue comportandose como la vieja".

---

## 6. Desinstalar

1. Encontra donde esta(n) instalado(s):

   ```bash
   type -a sdd-cli          # macOS/Linux
   ```
   ```powershell
   Get-Command sdd-cli -All  # Windows
   ```

2. Borra el/los ejecutable(s):

   ```bash
   # macOS / Linux (una linea por cada ruta que haya reportado type -a)
   sudo rm /usr/local/bin/sdd-cli
   rm ~/.local/bin/sdd-cli 2>/dev/null || true
   ```

   ```powershell
   # Windows
   Remove-Item "$env:LOCALAPPDATA\Programs\sdd-cli\sdd-cli.exe"
   ```

3. (Opcional) Quita la carpeta del `PATH` si la habias agregado a mano
   (Windows: "Editar las variables de entorno del sistema"; Unix: la linea
   `export PATH=...` en tu `~/.bashrc` / `~/.zshrc`).

> Desinstalar el binario **no** borra las carpetas `.sdd/` de tus proyectos; esas
> son datos tuyos y se eliminan aparte si asi lo queres.

---

## 7. Primeros pasos

Dentro de tu proyecto:

```bash
sdd-cli init                          # crea la estructura .sdd/
sdd-cli adapters install claude-code  # (opcional) adapter de Claude Code
sdd-cli --help                        # ver todos los comandos
```

Consulta la [Guia tecnica de la CLI](./9-guia-tecnica-cli.md) para el detalle de
comandos, workflows y el modelo de estados.

---

## 8. Solucion de problemas

| Sintoma | Causa probable | Solucion |
| --- | --- | --- |
| `sdd-cli: command not found` | La carpeta de instalacion no esta en el `PATH` | Agregala al `PATH` (ver 3.1/3.2) y reabri la terminal |
| `unknown command "version"` o comportamiento viejo tras actualizar | Hay un binario **viejo antes en el PATH** | `type -a sdd-cli` para ubicarlo, borralo (seccion 6) y corre `hash -r` |
| macOS bloquea el binario | Gatekeeper / quarantine (no esta firmado) | `xattr -d com.apple.quarantine <ruta>/sdd-cli` |
| PowerShell no ejecuta el script | Politica de ejecucion | `Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass` |
| `404 Not Found` al descargar | `VERSION` incorrecta o release inexistente | Copia la etiqueta exacta desde la pagina de Releases |
| `could not determine the latest release tag` | Aun no hay releases o problema de red | Especifica `SDD_VERSION` manualmente |
| Checksum mismatch | Descarga corrupta | Reintenta la descarga |
