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
  <a href="README.es.md"><bdi>Español</bdi></a> ·
  <a href="README.ar.md"><bdi>العربية</bdi></a> ·
  <bdi><strong>Français</strong></bdi><br>
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

<h1 align="center">Un lien. Vos médias.</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="État de la compilation"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="Dernière version"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 ou plus récent">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS, Linux et Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="Licence MIT"></a>
</p>

<p align="center">
  <strong>Un outil en ligne de commande pour télécharger vidéo et audio, conçu pour les personnes comme pour les agents IA.</strong><br>
  Donnez-lui un lien. Choisissez vos flux. Obtenez vidéo, audio, sous-titres, chapitres et couverture dans un seul fichier.
</p>

<p align="center">
  <a href="#install">Installation</a> ·
  <a href="#usage">Démarrage rapide</a> ·
  <a href="#sites">Sites pris en charge</a> ·
  <a href="#for-ai-agents-and-scripts">Guide pour les agents</a> ·
  <a href="#options">Options</a> ·
  <a href="../../releases">Versions</a>
</p>

---

<a id="small-command-complete-download"></a>

## Une petite commande. Un téléchargement complet.

| ↓ Vos médias, à votre façon | ⌘ Un binaire, tous les ordinateurs | { } Prêt pour l’automatisation |
| :--- | :--- | :--- |
| Choisissez qualité, codecs, pages et noms de fichiers avec le même vocabulaire sur cinq sites | Un seul binaire Go pour macOS, Linux et Windows, avec des téléchargements parallèles par plages qui reprennent là où ils se sont arrêtés | Un document JSON sur stdout, les journaux sur stderr, des codes de sortie explicites et aucune question sans terminal |

| 01 / EXAMINER | 02 / CHOISIR | 03 / TÉLÉCHARGER | 04 / ASSEMBLER |
| :--- | :--- | :--- | :--- |
| **Voir chaque flux** | **Fixer vos priorités** | **Tout rapatrier** | **Tout réunir** |
| Pages, codecs et tailles | Qualité, audio et pages | Par plages ou par segments | Pistes, sous-titres et couverture |
| `haul info "<url>"` | `-q 720p -c avc,m4a` | `haul "<url>"` | `ffmpeg / MP4Box` |

<details>
<summary><strong>Un coup d’œil dans le terminal</strong></summary>

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

> Pour un usage personnel, de recherche et autres usages non commerciaux. Il vous appartient de respecter les droits d’auteur et les conditions de chaque site.

<a id="install"></a>

## Installation

haul fonctionne sous **macOS, Linux et Windows** (amd64 et arm64). Il nécessite [ffmpeg](https://ffmpeg.org) pour assembler les pistes ; YouTube et X nécessitent aussi [yt-dlp](https://github.com/yt-dlp/yt-dlp), et les défis du lecteur YouTube nécessitent [deno](https://deno.com) :

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg pipx unzip       # Linux (Debian / Ubuntu)
pipx install "yt-dlp[default]" && pipx ensurepath
curl -fsSL https://deno.land/install.sh | sh -s -- -y
winget install Gyan.FFmpeg               # Windows, un paquet à la fois
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

Sous Linux, installez yt-dlp avec pipx ou avec le binaire `yt-dlp_linux` de ses [versions](https://github.com/yt-dlp/yt-dlp/releases/latest) (`yt-dlp_linux_aarch64` sur arm64), et non depuis votre distribution : YouTube change souvent, et les paquets prennent du retard. Après l’installation, ouvrez un nouveau terminal pour que les outils soient dans votre `PATH`.

<a id="get-the-binary"></a><a id="binary"></a>

### Obtenir le binaire

Téléchargez l’archive correspondant à votre système depuis [Releases](../../releases) — `haul-<version>-<os>-<arch>.tar.gz` (`.zip` sous Windows) — et placez `haul` dans n’importe quel dossier de votre `PATH` :

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
```

Le processus actuel signe les versions macOS et les fait notarier par Apple. Choisissez `haul-<version>-darwin-<arch>.pkg` pour installer `haul` dans `/usr/local/bin` avec le ticket de notarisation joint. L’archive contient le même exécutable signé, mais sa première vérification peut nécessiter Internet. Les anciennes versions peuvent ne pas être signées.

<details>
<summary><strong>Compiler depuis les sources</strong> · Go 1.26+</summary>

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

ou, depuis un clone du dépôt :

```sh
go build -o haul ./cmd/haul
```

</details>

<a id="usage"></a>

## Démarrage rapide

Commencez par un lien. Par défaut, haul choisit les meilleurs flux disponibles.

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

Examinez avant de télécharger, ou choisissez une qualité et un codec :

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # H.264 + AAC pour QuickTime
```

<details>
<summary><strong>Référence des commandes</strong></summary>

| Commande | Signification |
| --- | --- |
| `haul <url> [options]` | Télécharger ; équivaut à `haul download <url>` |
| `haul info <url> [--urls]` | Afficher l’élément, ses pages et ses flux, sans rien télécharger |
| `haul login bilibili` | Se connecter à bilibili pour accéder aux qualités supérieures |
| `haul templates` | Variables des modèles de noms de fichiers |
| `haul --help` | Sites, usage par les agents, codes de sortie et exemples |
| `haul download --help` | Toutes les options |

</details>

<a id="sites"></a>

### Sites pris en charge

| Site | Liens | Nécessite |
|---|---|---|
| **YouTube** | `youtube.com/watch?v=…`, `youtu.be/…`, `/shorts/…`, `/embed/…`, `/live/…` | yt-dlp, deno |
| **X** | `x.com/<user>/status/<id>`, `twitter.com/…` ; `/video/<n>` choisit une vidéo de la publication | yt-dlp |
| **bilibili** | vidéos, bangumi, cours, collections, séries, favoris, espaces utilisateurs, `b23.tv` et identifiants seuls `BV…` `av…` `ep…` `ss…` `md…` | – |
| **Xiaoyuzhou** | `xiaoyuzhoufm.com/episode/<id>`, `/podcast/<id>` | – |
| **Apple Podcasts** | `podcasts.apple.com/<cc>/podcast/<name>/id<show>`, avec `?i=<episode>` pour un seul épisode | – |

<a id="make-it-yours"></a><a id="examples"></a>

### À votre façon

**L’audio seulement**

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # AAC plutôt qu’Opus
```

**Quelques parties, une saison entière ou le dernier épisode**

```sh
# Parties choisies, enregistrées dans votre dossier Movies
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# Tous les épisodes d’une saison
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# Le dernier épisode d’un podcast
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# Toutes les vidéos d’une publication X
haul -p ALL "https://x.com/<user>/status/<id>"
```

**Fichiers organisés, sous-titres ou choix interactif**

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # uniquement ces langues de sous-titres
haul -i "BV1qt4y1X7TW"                                 # choisir les flux avec les touches fléchées
```

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## Pour les agents IA et les scripts

`haul --help` contient tout ce dont un agent a besoin ; [llms.txt](llms.txt) est le même guide sous forme de fichier. En bref :

```sh
haul info --json "<url>"                                # 1. examiner
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. télécharger exactement ces flux
haul --json -q 720p -c avc,m4a "<url>"                  #    ou laisser les priorités choisir
```

- Avec `--json`, **stdout ne contient qu’un seul document JSON** ; la progression et les journaux vont sur stderr. Son champ `files` liste les fichiers produits, nouveaux ou déjà présents.
- **Rien n’est interactif sans terminal.** Sans terminal, `-i` échoue avec le code de sortie 2 et indique les options à utiliser à la place.
- Les indices de flux d’`info` suivent l’ordre de sélection de haul ; passez les mêmes `-q` / `-c` à `info` et au téléchargement.
- Pour une playlist, une saison, une émission ou une publication à plusieurs vidéos, `info` liste les pages ; `-p <n>` ajoute les flux et les sous-titres de cette page.
- Les pages déjà présentes sur le disque sont signalées par `"status": "skipped", "reason": "exists"` : relancer la commande est donc sans risque.

<details>
<summary><strong>Exemple de réponse JSON</strong> · un téléchargement réussi</summary>

Un téléchargement affiche :

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

C’est la sortie de `haul --json -q 360p -c avc,m4a …`. Ici, les clés suivent l’ordre de lecture et les listes de flux sont réduites aux flux choisis ; le document réel trie ses clés par ordre alphabétique et liste tous les flux.

</details>

Un échec affiche `"ok": false` avec `"error": {"kind", "message", "exitCode"}`, et conserve les pages terminées avant lui.

| Code de sortie | Signification | `error.kind` |
|---|---|---|
| 0 | terminé | – |
| 1 | le téléchargement ou l’extraction a échoué | `failed` |
| 2 | lien non pris en charge ou valeur d’option incorrecte | `input` |
| 3 | un outil requis est absent (ffmpeg, yt-dlp) | `dependency` |
| 4 | connexion nécessaire ou expirée | `auth` |
| 64 | ligne de commande incorrecte (option inconnue, argument manquant) | – |
| 130 | annulé | `cancelled` |

<a id="options"></a>

## Options

`haul download --help` les liste toutes. Les principales :

| Catégorie | Options |
|---|---|
| Général | `--json`, `--config <file>`, `--debug` |
| Flux | `-q, --quality <list>`, `-c, --codec <list>`, `--video-stream <n>`, `--audio-stream <n>`, `-i, --interactive`, `--video-ascending`, `--audio-ascending` |
| Pages | `-p, --pages <spec>`, `--show-all`, `--hide-streams` |
| Contenu | `--audio-only`, `--video-only`, `--subtitle-only`, `--cover-only`, `--skip-subtitle`, `--sub-lang <list>`, `--auto-subtitles`, `--skip-cover`, `--skip-mux` |
| Sortie | `-o, --output <template>`, `--multi-output <template>`, `-w, --work-dir <dir>`, `--lang <code>`, `--no-tags`, `--archive`, `--delay <seconds>` |
| bilibili | `--api web\|tv\|app\|intl`, `--danmaku`, `--danmaku-only`, `--danmaku-format xml,ass`, `--cookie`, `--token` ; d’autres avec `--help-hidden` |
| Outils | `--ffmpeg`, `--yt-dlp`, `--use-mp4box`, `--mp4box`, `--use-aria2c`, `--aria2c`, `--aria2c-args`, `--single-connection` |

<details>
<summary><strong>Priorités de qualité et de codec</strong></summary>

**Qualités et codecs.** `-q` accepte les libellés affichés dans le tableau des flux : `1080p`, `720p60` sur YouTube ; `8K`, `Dolby Vision`, `HDR`, `4K`, `1080P60`, `1080P+`, `1080P`, `720P` sur bilibili. `-c` accepte `av1 vp9 hevc avc` pour la vidéo et `m4a opus flac eac3 mp3` pour l’audio. Ce sont des priorités, pas des filtres : ce qui n’est pas cité vient ensuite, du meilleur au moins bon. `--video-ascending` / `--audio-ascending` inversent l’ordre pour obtenir le téléchargement le plus léger.

</details>

<details>
<summary><strong>Sélection des pages</strong></summary>

**Pages.** `8`, `1,2`, `3-5`, `1-3,10`, `ALL`, `LAST` (la dernière page, c’est-à-dire l’épisode le plus récent d’une émission ; `LATEST` fonctionne aussi). Un lien vers un épisode, ou `?p=N`, sélectionne à lui seul cette page. Les pages sont les parties d’une vidéo bilibili, les épisodes d’une saison, d’une émission ou d’une liste, et les vidéos d’une publication X.

</details>

<details>
<summary><strong>Noms de fichiers et variables de modèle</strong></summary>

**Noms de fichiers.** Un élément à une seule page : `<title>` ; à plusieurs pages (même quand `-p` n’en prend qu’une) : `<title>/[P<pageNumberWithZero>]<pageTitle>`, le numéro étant complété par des zéros selon le nombre de pages. Variables : `<title>` `<pageNumber>` `<pageNumberWithZero>` `<pageTitle>` `<id>` `<site>` `<uploader>` `<uploaderId>` `<quality>` `<resolution>` `<fps>` `<videoCodec>` `<videoBitrate>` `<audioCodec>` `<audioBitrate>` `<publishDate>` `<pageDate>`, plus celles de bilibili : `<bvid>` `<aid>` `<cid>` `<api>`. Les dates acceptent un format : `<publishDate:yyyy-MM-dd>`. L’extension est ajoutée automatiquement.

</details>

<a id="site-notes"></a><a id="notes"></a>

## Notes par site

<details>
<summary><strong>YouTube</strong> · extraction, sous-titres et playlists</summary>

YouTube protège les URL de ses flux par des défis du lecteur qui exigent un moteur JavaScript ; l’extraction est donc confiée à `yt-dlp -J` (qui les exécute dans deno), et tout le reste est fait par haul lui-même. Les sous-titres fournis par l’auteur sont intégrés ; les sous-titres générés automatiquement, seulement avec `--auto-subtitles`. googlevideo ne sert que des plages d’octets limitées : les pistes sont donc récupérées par plages de 10 MB, l’une après l’autre. `watch?v=…&list=…` ne télécharge que la vidéo. Gardez yt-dlp à jour : c’est là qu’arrivent les correctifs pour les changements de YouTube.

</details>

<details>
<summary><strong>X</strong> · publications publiques et audio</summary>

X ne demande aucune connexion pour les publications publiques. Ses vidéos sont des fichiers MP4 uniques contenant l’audio ; le tableau n’affiche donc que la vidéo, et `--audio-only` extrait l’audio.

</details>

<details>
<summary><strong>bilibili</strong> · connexion, qualités supérieures et danmaku</summary>

bilibili est lu via ses propres API web, TV, APP (gRPC) et internationale. Sans connexion, il ne propose que des qualités inférieures (généralement jusqu’à 480P) ; connectez-vous pour 1080P, 4K, HDR, Dolby Vision et l’audio Hi-Res :

```sh
haul login bilibili                # scanner un code QR avec l’application bilibili
haul login bilibili --from-edge    # macOS : réutiliser la connexion de Microsoft Edge (ou --from-chrome ; --profile "Profile 1")
haul login bilibili --tv           # jeton d’accès TV, pour --api tv / --api app
```

Pour l’instant, la lecture de la connexion d’un navigateur ne fonctionne que sous macOS : le terminal doit disposer de l’accès complet au disque, et macOS demande une fois l’accès à l’élément « Safe Storage » du trousseau. `--danmaku` enregistre les commentaires superposés (danmaku) en XML et ASS ; `--danmaku-format ass` n’en garde qu’un.

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · épisodes et métadonnées</summary>

Xiaoyuzhou et Apple Podcasts n’ont pas besoin de yt-dlp : les pages Xiaoyuzhou contiennent le lien audio, et les liens Apple passent par l’API publique iTunes et le flux RSS de l’émission. Les épisodes gardent leur format (`.mp3` ou `.m4a`), avec la couverture intégrée, l’émission comme album et l’animateur comme artiste. Un lien d’émission liste ses derniers épisodes du plus ancien au plus récent ; `-p LAST` désigne donc le plus récent.

</details>

<a id="config"></a>

## Configuration

`~/.config/haul/` (ou `$HAUL_HOME`) contient :

| Fichier | Rôle |
|---|---|
| `config.json` | valeurs par défaut de n’importe quelle option |
| `cookie.txt`, `tv-token.txt`, `app-token.txt` | connexion bilibili |
| `archives.txt` | pages déjà téléchargées (`--archive`) |

`config.json` utilise les noms d’options en camelCase, ceux de bilibili sous `"bilibili"`, et ne contient que ce que vous modifiez ; une liste peut être un tableau ou une chaîne séparée par des virgules. Les options de la ligne de commande ont priorité sur lui :

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

## Développement

```sh
go build ./cmd/haul
go test ./...                                  # hors ligne, quelques secondes
HAUL_LIVE=1 go test -run Live ./internal/...   # contacte aussi bilibili, YouTube, X et Apple Podcasts
```

Les tests hors ligne ne touchent jamais au réseau : ils dialoguent avec un `http.RoundTripper` simulé qui répond à partir de réponses d’API enregistrées et de CDN simulés (serveurs n’acceptant que des plages, connexions coupées, serveurs qui ignorent les plages). Les tests de bout en bout assemblent les pistes avec un vrai ffmpeg et vérifient le résultat avec ffprobe ; ils sont ignorés si ffmpeg n’est pas installé. Architecture et conventions : [AGENTS.md](AGENTS.md).

Pousser un tag `v*` lance, via GitHub Actions, les tests, la compilation croisée et la publication d’une version pour chaque plateforme.

<a id="acknowledgements"></a>

## Remerciements

[yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://ffmpeg.org), [GPAC](https://gpac.io), [aria2](https://aria2.github.io), [pflag](https://github.com/spf13/pflag), [x/term](https://pkg.go.dev/golang.org/x/term), [rsc.io/qr](https://pkg.go.dev/rsc.io/qr), ainsi que les notes sur les API de [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) et de [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api).

<a id="license"></a>

## Licence

[MIT](LICENSE)

---

<p align="center">
  <strong>Un lien. Vos médias.</strong><br>
  <a href="#install">Obtenir haul</a> · <a href="llms.txt">Référence pour les agents</a> · <a href="AGENTS.md">Contribuer</a>
</p>
