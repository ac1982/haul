<!-- Translation of README.md; maintenance: docs/TRANSLATING.md -->

<p align="center">
  <img src="docs/assets/brand.svg" alt="haul" width="1200">
</p>

<!-- languages:start -->
<p align="center">
  <a href="README.md"><bdi>English</bdi></a> ·
  <a href="README.zh-CN.md"><bdi>简体中文</bdi></a> ·
  <a href="README.zh-TW.md"><bdi>繁體中文</bdi></a> ·
  <a href="README.ja.md"><bdi>日本語</bdi></a> ·
  <a href="README.hi.md"><bdi>हिन्दी</bdi></a> ·
  <bdi><strong>Español</strong></bdi> ·
  <a href="README.ar.md"><bdi>العربية</bdi></a> ·
  <a href="README.fr.md"><bdi>Français</bdi></a><br>
  <a href="README.bn.md"><bdi>বাংলা</bdi></a> ·
  <a href="README.pt.md"><bdi>Português</bdi></a> ·
  <a href="README.id.md"><bdi>Bahasa Indonesia</bdi></a> ·
  <a href="README.ur.md"><bdi>اردو</bdi></a> ·
  <a href="README.ru.md"><bdi>Русский</bdi></a> ·
  <a href="README.de.md"><bdi>Deutsch</bdi></a> ·
  <a href="README.pcm.md"><bdi>Naijá</bdi></a> ·
  <a href="README.arz.md"><bdi>مصري</bdi></a>
</p>
<!-- languages:end -->

<h1 align="center">Un enlace. Tu contenido.</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="Estado de la compilación"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="Última versión"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 o posterior">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS, Linux y Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="Licencia MIT"></a>
</p>

<p align="center">
  <strong>Un descargador de vídeo y audio por línea de comandos, pensado tanto para personas como para agentes de IA.</strong><br>
  Dale un enlace. Elige tus pistas. Obtén vídeo, audio, subtítulos, capítulos y portada en un solo archivo.
</p>

<p align="center">
  <a href="#install">Instalación</a> ·
  <a href="#usage">Inicio rápido</a> ·
  <a href="#sites">Sitios compatibles</a> ·
  <a href="#for-ai-agents-and-scripts">Guía para agentes</a> ·
  <a href="#options">Opciones</a> ·
  <a href="../../releases">Versiones</a>
</p>

---

<a id="small-command-complete-download"></a>

## Un comando pequeño. Una descarga completa.

| ↓ Tu contenido, a tu manera | ⌘ Un binario, cualquier escritorio | { } Listo para automatizar |
| :--- | :--- | :--- |
| Elige calidad, códecs, páginas y nombres de archivo con el mismo vocabulario en cinco sitios | Un único binario Go para macOS, Linux y Windows, con descargas paralelas por rangos que se reanudan | Un documento JSON en stdout, registros en stderr, códigos de salida con significado y ninguna pregunta sin terminal |

| 01 / INSPECCIONAR | 02 / ELEGIR | 03 / DESCARGAR | 04 / COMBINAR |
| :--- | :--- | :--- | :--- |
| **Todas las pistas a la vista** | **Define tus prioridades** | **Llévatelo a casa** | **Todo junto** |
| Páginas, códecs y tamaños | Calidad, audio y páginas | Por rangos o por segmentos | Pistas, subtítulos y portada |
| `haul info "<url>"` | `-q 720p -c avc,m4a` | `haul "<url>"` | `ffmpeg / MP4Box` |

<details>
<summary><strong>Echa un vistazo al terminal</strong></summary>

```
$ haul info -q 720p "https://youtu.be/DdCEmlAydcw"
haul v1.0.0

  Inside Anthropic's molecular biology lab
  channel Anthropic  ·  2026-09-24  ·  00:01:15

  Video
  ▶  0  720p   1190×720   AVC  24fps   755 kbps    6.8 MB
     1  720p   1190×720   VP9  24fps   528 kbps    4.7 MB
     2  720p   1190×720   AV1  24fps   388 kbps    3.5 MB
     3  1080p  2048×1240  VP9  24fps  5827 kbps   52.3 MB
     …
    19  144p   238×144    AVC  24fps    50 kbps  459.0 KB
  Audio
  ▶ 0  OPUS  131 kbps    1.2 MB
    1  M4A   130 kbps    1.2 MB
    …
```

</details>

> Para uso personal, de investigación y otros fines no comerciales. Eres responsable de respetar los derechos de autor y las condiciones de cada sitio.

<a id="install"></a>

## Instalación

haul funciona en **macOS, Linux y Windows** (amd64 y arm64). Necesita [ffmpeg](https://ffmpeg.org) para combinar pistas; YouTube y X también necesitan [yt-dlp](https://github.com/yt-dlp/yt-dlp), y los desafíos del reproductor de YouTube requieren [deno](https://deno.com):

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg pipx unzip       # Linux (Debian / Ubuntu)
pipx install "yt-dlp[default]" && pipx ensurepath
curl -fsSL https://deno.land/install.sh | sh -s -- -y
winget install Gyan.FFmpeg               # Windows, un paquete cada vez
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

En Linux, instala yt-dlp con pipx o con el binario `yt-dlp_linux` de sus [versiones](https://github.com/yt-dlp/yt-dlp/releases/latest) (`yt-dlp_linux_aarch64` en arm64), no desde tu distribución: YouTube cambia a menudo y los paquetes se quedan atrás. Después de instalar, abre una terminal nueva para que las herramientas estén en tu `PATH`.

<a id="get-the-binary"></a><a id="binary"></a>

### Obtener el binario

Descarga el archivo comprimido para tu sistema desde [Releases](../../releases) —`haul-<version>-<os>-<arch>.tar.gz` (`.zip` en Windows)— y coloca `haul` en cualquier directorio de tu `PATH`:

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
```

El flujo actual firma las versiones de macOS y obtiene su notarización de Apple. El instalador `haul-<version>-darwin-<arch>.pkg` incluye el comprobante e instala `haul` en `/usr/local/bin`. El archivo comprimido contiene el mismo ejecutable firmado, pero la primera verificación puede necesitar internet. Las versiones antiguas pueden no estar firmadas.

<details>
<summary><strong>Compilar desde el código fuente</strong> · Go 1.26+</summary>

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

o bien, desde un clon del repositorio:

```sh
go build -o haul ./cmd/haul
```

</details>

<a id="usage"></a>

## Inicio rápido

Empieza con un enlace. haul elige por defecto las mejores pistas disponibles.

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

Consulta la información antes de descargar, o elige una calidad y un códec:

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # H.264 + AAC para QuickTime
```

<details>
<summary><strong>Referencia de comandos</strong></summary>

| Comando | Significado |
| --- | --- |
| `haul <url> [options]` | Descargar; equivale a `haul download <url>` |
| `haul info <url> [--urls]` | Mostrar el elemento, sus páginas y pistas sin descargar nada |
| `haul login bilibili` | Iniciar sesión en bilibili para obtener calidades superiores |
| `haul templates` | Variables de las plantillas de nombres de archivo |
| `haul --help` | Sitios, uso por agentes, códigos de salida y ejemplos |
| `haul download --help` | Todas las opciones |

</details>

<a id="sites"></a>

### Sitios compatibles

| Sitio | Enlaces | Requiere |
|---|---|---|
| **YouTube** | `youtube.com/watch?v=…`, `youtu.be/…`, `/shorts/…`, `/embed/…`, `/live/…` | yt-dlp, deno |
| **X** | `x.com/<user>/status/<id>`, `twitter.com/…`; `/video/<n>` selecciona un vídeo de la publicación | yt-dlp |
| **bilibili** | vídeos, bangumi, cursos, colecciones, series, favoritos, espacios de usuario, `b23.tv` e identificadores sueltos `BV…` `av…` `ep…` `ss…` `md…` | – |
| **Xiaoyuzhou** | `xiaoyuzhoufm.com/episode/<id>`, `/podcast/<id>` | – |
| **Apple Podcasts** | `podcasts.apple.com/<cc>/podcast/<name>/id<show>`, con `?i=<episode>` para un solo episodio | – |

<a id="make-it-yours"></a><a id="examples"></a>

### A tu manera

**Solo el audio**

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # AAC en lugar de Opus
```

**Algunas partes, una temporada completa o el último episodio**

```sh
# Partes seleccionadas, guardadas en tu carpeta Movies
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# Todos los episodios de una temporada
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# El episodio de pódcast más reciente
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# Todos los vídeos de una publicación de X
haul -p ALL "https://x.com/<user>/status/<id>"
```

**Archivos organizados, subtítulos o una elección interactiva**

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # solo estos idiomas de subtítulos
haul -i "BV1qt4y1X7TW"                                 # elige las pistas con las teclas de flecha
```

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## Para agentes de IA y scripts

`haul --help` incluye todo lo que necesita un agente; [llms.txt](llms.txt) es la misma guía en forma de archivo. En resumen:

```sh
haul info --json "<url>"                                # 1. consultar
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. descargar exactamente esas pistas
haul --json -q 720p -c avc,m4a "<url>"                  #    o dejar que decidan las prioridades
```

- Con `--json`, **stdout contiene solo un documento JSON**; el progreso y los registros van a stderr. Su campo `files` enumera los archivos de salida, nuevos o ya existentes.
- **Sin terminal no hay nada interactivo.** `-i` sin terminal falla con el código de salida 2 e indica las opciones que deben usarse en su lugar.
- Los índices de pista de `info` siguen el orden en que elige haul; pasa los mismos `-q` / `-c` a `info` y a la descarga.
- En una lista de reproducción, temporada, programa o publicación con varios vídeos, `info` enumera sus páginas; `-p <n>` añade las pistas y los subtítulos de esa página.
- Las páginas que ya están en disco se indican como `"status": "skipped", "reason": "exists"`, así que repetir la ejecución es seguro.

<details>
<summary><strong>Ejemplo de respuesta JSON</strong> · una descarga correcta</summary>

Una descarga imprime:

```json
{
  "ok": true,
  "command": "download",
  "site": "youtube",
  "input": "https://youtu.be/DdCEmlAydcw",
  "title": "Inside Anthropic's molecular biology lab",
  "uploader": "Anthropic",
  "published": "2026-09-23T18:01:32Z",
  "description": "…",
  "pageCount": 1,
  "files": ["/Users/me/Movies/Inside Anthropic's molecular biology lab.mp4"],
  "pages": [
    {
      "index": 1,
      "id": "DdCEmlAydcw",
      "title": "Inside Anthropic's molecular biology lab",
      "durationSeconds": 75,
      "published": "2026-09-23T18:01:32Z",
      "selected": true,
      "status": "downloaded",
      "file": "/Users/me/Movies/Inside Anthropic's molecular biology lab.mp4",
      "sizeBytes": 3434260,
      "selectedVideo": 0,
      "selectedAudio": 0,
      "video": [{ "index": 0, "quality": "360p", "resolution": "594x360", "codec": "AVC", "fps": 24, "bitrateKbps": 214, "sizeBytes": 2014782 }],
      "audio": [{ "index": 0, "codec": "M4A", "bitrateKbps": 130, "sizeBytes": 1220994 }],
      "subtitles": [{ "lang": "en" }]
    }
  ]
}
```

Es la salida de `haul --json -q 360p -c avc,m4a …`. Aquí las claves siguen el orden de lectura y las listas de pistas se recortan a las elegidas; el documento real ordena sus claves alfabéticamente e incluye todas las pistas.

</details>

Un fallo imprime `"ok": false` con `"error": {"kind", "message", "exitCode"}` y conserva las páginas completadas antes del error.

| Código de salida | Significado | `error.kind` |
|---|---|---|
| 0 | completado | – |
| 1 | la descarga o la extracción falló | `failed` |
| 2 | enlace no compatible o valor de opción incorrecto | `input` |
| 3 | falta una herramienta necesaria (ffmpeg, yt-dlp) | `dependency` |
| 4 | inicio de sesión necesario o caducado | `auth` |
| 64 | línea de comandos incorrecta (opción desconocida, argumento ausente) | – |
| 130 | cancelado | `cancelled` |

<a id="options"></a>

## Opciones

`haul download --help` las enumera todas. Las principales:

| Grupo | Opciones |
|---|---|
| General | `--json`, `--config <file>`, `--debug` |
| Pistas | `-q, --quality <list>`, `-c, --codec <list>`, `--video-stream <n>`, `--audio-stream <n>`, `-i, --interactive`, `--video-ascending`, `--audio-ascending` |
| Páginas | `-p, --pages <spec>`, `--show-all`, `--hide-streams` |
| Contenido | `--audio-only`, `--video-only`, `--subtitle-only`, `--cover-only`, `--skip-subtitle`, `--sub-lang <list>`, `--auto-subtitles`, `--skip-cover`, `--skip-mux` |
| Salida | `-o, --output <template>`, `--multi-output <template>`, `-w, --work-dir <dir>`, `--lang <code>`, `--no-tags`, `--archive`, `--delay <seconds>` |
| bilibili | `--api web\|tv\|app\|intl`, `--danmaku`, `--danmaku-only`, `--danmaku-format xml,ass`, `--cookie`, `--token`; más opciones con `--help-hidden` |
| Herramientas | `--ffmpeg`, `--yt-dlp`, `--use-mp4box`, `--mp4box`, `--use-aria2c`, `--aria2c`, `--aria2c-args`, `--single-connection` |

<details>
<summary><strong>Prioridades de calidad y códecs</strong></summary>

**Calidades y códecs.** `-q` acepta las etiquetas que muestra la tabla de pistas: `1080p`, `720p60` en YouTube; `8K`, `Dolby Vision`, `HDR`, `4K`, `1080P60`, `1080P+`, `1080P`, `720P` en bilibili. `-c` acepta `av1 vp9 hevc avc` para vídeo y `m4a opus flac eac3 mp3` para audio. Son prioridades, no filtros: lo que no se indica va después, de mejor a peor. `--video-ascending` / `--audio-ascending` invierten el orden para obtener la descarga más pequeña.

</details>

<details>
<summary><strong>Selección de páginas</strong></summary>

**Páginas.** `8`, `1,2`, `3-5`, `1-3,10`, `ALL`, `LAST` (la última página, que es el episodio más reciente de un programa; también vale `LATEST`). Un enlace a un episodio, o `?p=N`, selecciona esa página por sí solo. Las páginas son las partes de un vídeo de bilibili, los episodios de una temporada, programa o lista, y los vídeos de una publicación de X.

</details>

<details>
<summary><strong>Nombres de archivo y variables de plantilla</strong></summary>

**Nombres de archivo.** Un elemento con una sola página: `<title>`; con varias (aunque `-p` seleccione solo una): `<title>/[P<pageNumberWithZero>]<pageTitle>`, con el número rellenado con ceros según el total de páginas. Variables: `<title>` `<pageNumber>` `<pageNumberWithZero>` `<pageTitle>` `<id>` `<site>` `<uploader>` `<uploaderId>` `<quality>` `<resolution>` `<fps>` `<videoCodec>` `<videoBitrate>` `<audioCodec>` `<audioBitrate>` `<publishDate>` `<pageDate>`, además de las de bilibili: `<bvid>` `<aid>` `<cid>` `<api>`. Las fechas admiten un formato: `<publishDate:yyyy-MM-dd>`. La extensión se añade automáticamente.

</details>

<a id="site-notes"></a><a id="notes"></a>

## Notas por sitio

<details>
<summary><strong>YouTube</strong> · extracción, subtítulos y listas de reproducción</summary>

YouTube protege las URL de sus pistas con desafíos del reproductor que necesitan un entorno de ejecución de JavaScript, por lo que la extracción se delega en `yt-dlp -J` (que los ejecuta en deno); todo lo que viene después lo hace haul. Los subtítulos subidos se incorporan al archivo; los generados automáticamente, solo con `--auto-subtitles`. googlevideo solo sirve rangos de bytes limitados, así que las pistas se descargan por rangos de 10 MB, uno tras otro. `watch?v=…&list=…` descarga solo el vídeo. Mantén yt-dlp actualizado: ahí llegan las correcciones para los cambios de YouTube.

</details>

<details>
<summary><strong>X</strong> · publicaciones públicas y audio</summary>

X no requiere iniciar sesión para las publicaciones públicas. Sus vídeos son archivos MP4 únicos con el audio incluido, por lo que la tabla solo muestra vídeo; `--audio-only` extrae el audio.

</details>

<details>
<summary><strong>bilibili</strong> · inicio de sesión, calidades superiores y danmaku</summary>

bilibili se lee a través de sus propias API web, TV, APP (gRPC) e internacional. Sin sesión solo ofrece calidades bajas (normalmente hasta 480P); inicia sesión para obtener 1080P, 4K, HDR, Dolby Vision y audio Hi-Res:

```sh
haul login bilibili                # escanea un código QR con la app de bilibili
haul login bilibili --from-edge    # macOS: reutiliza la sesión de Microsoft Edge (o --from-chrome; --profile "Profile 1")
haul login bilibili --tv           # token de acceso de TV, para --api tv / --api app
```

Leer la sesión de un navegador solo funciona en macOS por ahora: requiere acceso total al disco para el terminal, y macOS pide una vez acceso al elemento “Safe Storage” del llavero. `--danmaku` guarda los comentarios superpuestos (danmaku) en XML y ASS; `--danmaku-format ass` conserva solo uno de los dos.

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · episodios y metadatos</summary>

Xiaoyuzhou y Apple Podcasts no necesitan yt-dlp: las páginas de Xiaoyuzhou incluyen el enlace de audio, y los enlaces de Apple pasan por la API pública de iTunes y el feed RSS del programa. Los episodios conservan su formato (`.mp3` o `.m4a`), con la portada incrustada, el programa como álbum y el presentador como artista. El enlace de un programa enumera sus episodios más recientes del más antiguo al más nuevo, así que `-p LAST` es el más reciente.

</details>

<a id="config"></a>

## Configuración

`~/.config/haul/` (o `$HAUL_HOME`) contiene:

| Archivo | Uso |
|---|---|
| `config.json` | valores predeterminados para cualquier opción |
| `cookie.txt`, `tv-token.txt`, `app-token.txt` | sesión de bilibili |
| `archives.txt` | páginas ya descargadas (`--archive`) |

`config.json` usa los nombres de las opciones en camelCase, con las de bilibili bajo `"bilibili"`, y solo necesita lo que quieras cambiar; una lista puede ser un array o una cadena separada por comas. Las opciones de la línea de comandos tienen prioridad sobre él:

```json
{
  "workDir": "~/Movies/haul",
  "codec": "avc,hevc,m4a",
  "quality": "1080p,1080P",
  "output": "<uploader>/<title>",
  "archive": true,
  "bilibili": { "danmaku": true, "danmakuFormat": ["ass"] }
}
```

<a id="development"></a>

## Desarrollo

```sh
go build ./cmd/haul
go test ./...                                  # sin red, unos segundos
HAUL_LIVE=1 go test -run Live ./internal/...   # también accede a bilibili, YouTube, X y Apple Podcasts
```

Las pruebas sin conexión nunca acceden a la red: hablan con un `http.RoundTripper` simulado que responde con respuestas de API grabadas y desde CDN simuladas (servidores que solo sirven rangos, conexiones cortadas, servidores que ignoran los rangos). Las pruebas de extremo a extremo combinan pistas con un ffmpeg real y comprueban el resultado con ffprobe; se omiten si ffmpeg no está instalado. Consulta la arquitectura y las convenciones en [AGENTS.md](AGENTS.md).

Al publicar una etiqueta `v*`, GitHub Actions ejecuta las pruebas, compila para cada plataforma y publica una versión para todas ellas.

<a id="acknowledgements"></a>

## Agradecimientos

[yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://ffmpeg.org), [GPAC](https://gpac.io), [aria2](https://aria2.github.io), [pflag](https://github.com/spf13/pflag), [x/term](https://pkg.go.dev/golang.org/x/term), [rsc.io/qr](https://pkg.go.dev/rsc.io/qr) y las notas sobre las API de [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) y [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api).

<a id="license"></a>

## Licencia

[MIT](LICENSE)

---

<p align="center">
  <strong>Un enlace. Tu contenido.</strong><br>
  <a href="#install">Consigue haul</a> · <a href="llms.txt">Referencia para agentes</a> · <a href="AGENTS.md">Contribuir</a>
</p>
