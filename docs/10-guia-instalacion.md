# Guia de instalacion de sdd-cli

> Estado: BETA.
> Audiencia: usuarios que quieren instalar y usar el motor SDD.
> Alcance: instalacion del binario `sdd-cli` en Windows, macOS y Linux (Ubuntu).

`sdd-cli` es un binario nativo autocontenido escrito en Go. No requiere Node.js,
Python ni ningun runtime: se descarga un unico ejecutable y queda listo para usar.

Los binarios se publican en **GitHub Releases**:
<https://github.com/matiasezequieldimuro/ssd-workflow/releases>

Plataformas publicadas:

| Sistema operativo | Arquitectura | Archivo |
| --- | --- | --- |
| Linux (Ubuntu, WSL) | x86_64 (amd64) | `sdd-cli_<version>_linux_amd64.tar.gz` |
| macOS (Intel) | x86_64 (amd64) | `sdd-cli_<version>_darwin_amd64.tar.gz` |
| macOS (Apple Silicon) | arm64 (M1/M2/M3) | `sdd-cli_<version>_darwin_arm64.tar.gz` |
| Windows | x86_64 (amd64) | `sdd-cli_<version>_windows_amd64.zip` |

Cada release incluye un archivo `SHA256SUMS` para verificar la integridad de las descargas.

---

## 1. Instalacion rapida (recomendada)

### 1.1. macOS y Linux (Ubuntu)

Ejecuta en una terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/install.sh | sh
```

El script:

1. detecta tu sistema operativo y arquitectura;
2. resuelve la ultima version publicada (incluye pre-releases beta);
3. descarga el binario correcto y verifica su checksum;
4. lo instala en `/usr/local/bin` (o en `~/.local/bin` si el primero no es escribible).

Opciones mediante variables de entorno:

```bash
# Instalar una version especifica
SDD_VERSION=v0.1.0-beta curl -fsSL https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/install.sh | sh

# Elegir el directorio de instalacion
SDD_INSTALL_DIR="$HOME/.local/bin" curl -fsSL https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/install.sh | sh
```

Si el instalador avisa que el directorio no esta en tu `PATH`, agrega esta linea a
tu `~/.bashrc`, `~/.zshrc` o `~/.profile`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### 1.2. Windows (PowerShell)

Abri **PowerShell** y ejecuta:

```powershell
irm https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/install.ps1 | iex
```

El script descarga el binario, verifica el checksum, lo instala en
`%LOCALAPPDATA%\Programs\sdd-cli` y agrega ese directorio a tu `PATH` de usuario.
Cerra y volve a abrir la terminal para que el `PATH` tome efecto.

Opciones mediante variables de entorno:

```powershell
# Version especifica
$env:SDD_VERSION = "v0.1.0-beta"; irm https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/install.ps1 | iex
```

> Si PowerShell bloquea la ejecucion por la politica de scripts, podes habilitarla
> para la sesion actual con:
> `Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass`

---

## 2. Instalacion manual desde Releases

Usa este metodo si preferis no ejecutar el script de instalacion.

### 2.1. Linux (Ubuntu / WSL)

```bash
# 1. Descargar (reemplaza <version> por la etiqueta del release, ej. v0.1.0-beta)
curl -fsSLO https://github.com/matiasezequieldimuro/ssd-workflow/releases/download/<version>/sdd-cli_<version>_linux_amd64.tar.gz

# 2. (Opcional) Verificar checksum
curl -fsSLO https://github.com/matiasezequieldimuro/ssd-workflow/releases/download/<version>/SHA256SUMS
sha256sum -c SHA256SUMS --ignore-missing

# 3. Extraer e instalar
tar -xzf sdd-cli_<version>_linux_amd64.tar.gz
sudo mv sdd-cli /usr/local/bin/
sudo chmod +x /usr/local/bin/sdd-cli
```

### 2.2. macOS

```bash
# Intel: usa darwin_amd64. Apple Silicon (M1/M2/M3): usa darwin_arm64.
curl -fsSLO https://github.com/matiasezequieldimuro/ssd-workflow/releases/download/<version>/sdd-cli_<version>_darwin_arm64.tar.gz

tar -xzf sdd-cli_<version>_darwin_arm64.tar.gz
sudo mv sdd-cli /usr/local/bin/
sudo chmod +x /usr/local/bin/sdd-cli
```

> **Gatekeeper (macOS):** como el binario no esta firmado/notarizado por Apple,
> la primera ejecucion puede bloquearse. Autorizalo con:
>
> ```bash
> xattr -d com.apple.quarantine /usr/local/bin/sdd-cli
> ```

### 2.3. Windows

1. Descarga `sdd-cli_<version>_windows_amd64.zip` desde la
   [pagina de Releases](https://github.com/matiasezequieldimuro/ssd-workflow/releases).
2. Extrae `sdd-cli.exe` (por ejemplo a `C:\Users\<tu-usuario>\sdd-cli\`).
3. Agrega esa carpeta al `PATH` de usuario:
   - Buscar "Editar las variables de entorno del sistema" -> **Variables de entorno**.
   - En **Path** (usuario) -> **Nuevo** -> pega la ruta de la carpeta.
4. Abri una nueva terminal para que el cambio tome efecto.

Alternativa por PowerShell:

```powershell
$dir = "$env:USERPROFILE\sdd-cli"
Expand-Archive -Path .\sdd-cli_<version>_windows_amd64.zip -DestinationPath $dir -Force
[Environment]::SetEnvironmentVariable('Path', "$([Environment]::GetEnvironmentVariable('Path','User'));$dir", 'User')
```

---

## 3. Compilar desde fuente

Requiere [Go 1.22 o superior](https://go.dev/dl/).

```bash
git clone https://github.com/matiasezequieldimuro/ssd-workflow.git
cd ssd-workflow/src/cli

# Sincroniza los recursos embebidos (.sdd y adapters) antes de compilar
go generate ./...

# Compila el binario
go build -o sdd-cli .

# (Opcional) Instalar en el PATH
sudo mv sdd-cli /usr/local/bin/
```

Para incrustar la informacion de version al compilar:

```bash
go build -ldflags "-X sdd-cli/cmd.version=v0.1.0-dev" -o sdd-cli .
```

---

## 4. Verificar la instalacion

```bash
sdd-cli version
sdd-cli --help
```

`sdd-cli version` muestra la version, el commit y la fecha de compilacion.

---

## 5. Primeros pasos

Dentro de tu proyecto:

```bash
# 1. Inicializar la estructura .sdd/
sdd-cli init

# 2. (Opcional) Instalar el adapter de Claude Code
sdd-cli adapters install claude-code

# 3. Ver los comandos disponibles
sdd-cli --help
```

Consulta la [Guia tecnica de la CLI](./9-guia-tecnica-cli.md) para el detalle de
comandos, workflows y el modelo de estados.

---

## 6. Actualizar

Volve a ejecutar el instalador rapido; sobrescribe el binario con la ultima version:

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/install.sh | sh
```

```powershell
# Windows
irm https://raw.githubusercontent.com/matiasezequieldimuro/ssd-workflow/main/install.ps1 | iex
```

---

## 7. Desinstalar

```bash
# macOS / Linux (segun donde se instalo)
sudo rm /usr/local/bin/sdd-cli
# o
rm ~/.local/bin/sdd-cli
```

```powershell
# Windows
Remove-Item "$env:LOCALAPPDATA\Programs\sdd-cli\sdd-cli.exe"
```

---

## 8. Solucion de problemas

| Sintoma | Causa probable | Solucion |
| --- | --- | --- |
| `sdd-cli: command not found` | El directorio de instalacion no esta en el `PATH` | Agregalo al `PATH` (ver seccion 1) y reabri la terminal |
| macOS bloquea el binario | Gatekeeper / quarantine | `xattr -d com.apple.quarantine <ruta>/sdd-cli` |
| PowerShell no ejecuta el script | Politica de ejecucion | `Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass` |
| `could not determine the latest release tag` | Aun no hay releases o problema de red | Especifica `SDD_VERSION` manualmente |
| Checksum mismatch | Descarga corrupta | Reintenta la descarga |
