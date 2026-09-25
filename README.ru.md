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
  <a href="README.fr.md"><bdi>Français</bdi></a><br>
  <a href="README.bn.md"><bdi>বাংলা</bdi></a> ·
  <a href="README.pt.md"><bdi>Português</bdi></a> ·
  <a href="README.id.md"><bdi>Bahasa Indonesia</bdi></a> ·
  <a href="README.ur.md"><bdi>اردو</bdi></a> ·
  <bdi><strong>Русский</strong></bdi> ·
  <a href="README.de.md"><bdi>Deutsch</bdi></a> ·
  <a href="README.pcm.md"><bdi>Naijá</bdi></a> ·
  <a href="README.arz.md"><bdi>مصري</bdi></a>
</p>
<!-- languages:end -->

<h1 align="center">Одна ссылка. Ваши медиа.</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="Статус сборки"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="Последний релиз"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 или новее">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS, Linux и Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="Лицензия MIT"></a>
</p>

<p align="center">
  <strong>Программа командной строки для скачивания видео и аудио, одинаково удобная людям и ИИ-агентам.</strong><br>
  Дайте ей ссылку. Выберите потоки. Получите видео, аудио, субтитры, главы и обложку в одном файле.
</p>

<p align="center">
  <a href="#install">Установка</a> ·
  <a href="#usage">Быстрый старт</a> ·
  <a href="#sites">Поддерживаемые платформы</a> ·
  <a href="#for-ai-agents-and-scripts">Руководство для агентов</a> ·
  <a href="#options">Параметры</a> ·
  <a href="../../releases">Релизы</a>
</p>

---

<a id="small-command-complete-download"></a>

## Короткая команда. Полная загрузка.

| ↓ Ваши медиа — как вам удобно | ⌘ Один бинарный файл для любой настольной ОС | { } Готово к автоматизации |
| :--- | :--- | :--- |
| Качество, кодеки, страницы и имена файлов задаются одними и теми же параметрами на пяти платформах | Один бинарный файл на Go для macOS, Linux и Windows; параллельная загрузка по диапазонам с возобновлением | Один документ JSON в stdout, журналы в stderr, осмысленные коды завершения и никаких запросов без терминала |

| 01 / ПРОСМОТР | 02 / ВЫБОР | 03 / ЗАГРУЗКА | 04 / СБОРКА |
| :--- | :--- | :--- | :--- |
| **Все потоки на виду** | **Задайте приоритеты** | **Скачайте к себе** | **Всё вместе** |
| Страницы, кодеки и размеры | Качество, звук и страницы | Диапазонами или сегментами | Дорожки, субтитры и обложка |
| `haul info "<url>"` | `-q 720p -c avc,m4a` | `haul "<url>"` | `ffmpeg / MP4Box` |

<details>
<summary><strong>Как это выглядит в терминале</strong></summary>

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

> Для личного, исследовательского и иного некоммерческого использования. Вы сами отвечаете за соблюдение авторских прав и условий каждой платформы.

<a id="install"></a>

## Установка

haul работает в **macOS, Linux и Windows** (amd64 и arm64). Для объединения дорожек нужен [ffmpeg](https://ffmpeg.org); YouTube и X также требуют [yt-dlp](https://github.com/yt-dlp/yt-dlp), а для проверок плеера YouTube нужен [deno](https://deno.com):

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg pipx unzip       # Linux (Debian / Ubuntu)
pipx install "yt-dlp[default]" && pipx ensurepath
curl -fsSL https://deno.land/install.sh | sh -s -- -y
winget install Gyan.FFmpeg               # Windows, по одному пакету за раз
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

В Linux устанавливайте yt-dlp через pipx или файлом `yt-dlp_linux` из [релизов](https://github.com/yt-dlp/yt-dlp/releases/latest) (`yt-dlp_linux_aarch64` на arm64), а не из дистрибутива: YouTube часто меняется, и пакеты быстро устаревают. После установки откройте новый терминал, чтобы программы оказались в `PATH`.

<a id="get-the-binary"></a><a id="binary"></a>

### Скачать бинарный файл

Скачайте архив для своей системы со страницы [Releases](../../releases) — `haul-<version>-<os>-<arch>.tar.gz` (`.zip` для Windows) — и поместите `haul` в любой каталог из `PATH`:

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
```

Текущий процесс выпуска подписывает версии macOS и отправляет их на нотариализацию Apple. Установщик `haul-<version>-darwin-<arch>.pkg` содержит билет нотариализации и устанавливает `haul` в `/usr/local/bin`. В архиве тот же подписанный файл, но первая проверка может потребовать интернет. Старые выпуски могут быть неподписанными.

<details>
<summary><strong>Сборка из исходников</strong> · Go 1.26+</summary>

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

или из клона репозитория:

```sh
go build -o haul ./cmd/haul
```

</details>

<a id="usage"></a>

## Быстрый старт

Начните со ссылки. По умолчанию haul выбирает лучшие доступные потоки.

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

Посмотрите сведения перед скачиванием или выберите качество и кодек:

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # H.264 + AAC для QuickTime
```

<details>
<summary><strong>Справочник команд</strong></summary>

| Команда | Назначение |
|---|---|
| `haul <url> [options]` | скачать (то же, что `haul download <url>`) |
| `haul info <url> [--urls]` | показать материал, его страницы и потоки; ничего не скачивать |
| `haul login bilibili` | войти в bilibili ради более высокого качества |
| `haul templates` | переменные шаблонов имён файлов |
| `haul --help` | платформы, работа с агентами, коды завершения, примеры |
| `haul download --help` | все параметры |

</details>

<a id="sites"></a>

### Поддерживаемые платформы

| Платформа | Ссылки | Требуется |
|---|---|---|
| **YouTube** | `youtube.com/watch?v=…`, `youtu.be/…`, `/shorts/…`, `/embed/…`, `/live/…` | yt-dlp, deno |
| **X** | `x.com/<user>/status/<id>`, `twitter.com/…`; `/video/<n>` выбирает одно видео публикации | yt-dlp |
| **bilibili** | видео, бангуми, курсы, коллекции, серии, избранное, страницы пользователей, `b23.tv`, голые ID `BV…` `av…` `ep…` `ss…` `md…` | – |
| **Xiaoyuzhou** | `xiaoyuzhoufm.com/episode/<id>`, `/podcast/<id>` | – |
| **Apple Podcasts** | `podcasts.apple.com/<cc>/podcast/<name>/id<show>`, с `?i=<episode>` для одного выпуска | – |

<a id="make-it-yours"></a><a id="examples"></a>

### Под ваши задачи

**Только аудио**

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # AAC вместо Opus
```

**Несколько частей, сезон целиком или последний выпуск**

```sh
# Выбранные части, сохранённые в папку Movies
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# Все серии сезона
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# Новейший выпуск подкаста
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# Все видео из публикации X
haul -p ALL "https://x.com/<user>/status/<id>"
```

**Упорядоченные файлы, субтитры или интерактивный выбор**

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # только эти языки субтитров
haul -i "BV1qt4y1X7TW"                                 # выбор потоков клавишами со стрелками
```

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## Для ИИ-агентов и скриптов

`haul --help` содержит всё, что нужно агенту; [llms.txt](llms.txt) — то же руководство в виде файла. Кратко:

```sh
haul info --json "<url>"                                # 1. посмотреть
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. скачать именно эти потоки
haul --json -q 720p -c avc,m4a "<url>"                  #    или доверить выбор приоритетам
```

- С `--json` **stdout содержит только один документ JSON**; прогресс и журналы идут в stderr. Его поле `files` перечисляет выходные файлы — новые и уже существующие.
- **Без терминала ничего интерактивного нет.** `-i` без терминала завершается с кодом 2 и называет параметры, которые нужно использовать вместо него.
- Индексы потоков в `info` следуют порядку, в котором выбирает haul; передавайте одинаковые `-q` / `-c` и в `info`, и при скачивании.
- Для плейлиста, сезона, шоу или публикации с несколькими видео `info` перечисляет страницы; `-p <n>` добавляет потоки и субтитры этой страницы.
- Страницы, уже сохранённые на диске, отмечаются как `"status": "skipped", "reason": "exists"`, поэтому повторный запуск безопасен.

<details>
<summary><strong>Пример ответа JSON</strong> · успешное скачивание</summary>

Скачивание выводит:

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

Это результат `haul --json -q 360p -c avc,m4a …`. Здесь ключи идут в порядке чтения, а списки потоков сокращены до выбранных; реальный документ сортирует ключи по алфавиту и перечисляет все потоки.

</details>

При ошибке выводится `"ok": false` с `"error": {"kind", "message", "exitCode"}`, а страницы, завершённые до неё, сохраняются.

| Код завершения | Значение | `error.kind` |
|---|---|---|
| 0 | готово | – |
| 1 | скачивание или извлечение не удалось | `failed` |
| 2 | неподдерживаемая ссылка или неверное значение параметра | `input` |
| 3 | отсутствует нужный инструмент (ffmpeg, yt-dlp) | `dependency` |
| 4 | требуется вход или срок входа истёк | `auth` |
| 64 | неверная командная строка (неизвестный флаг, нет аргумента) | – |
| 130 | отменено | `cancelled` |

<a id="options"></a>

## Параметры

`haul download --help` перечисляет все. Основные:

| Группа | Параметры |
|---|---|
| Общие | `--json`, `--config <file>`, `--debug` |
| Потоки | `-q, --quality <list>`, `-c, --codec <list>`, `--video-stream <n>`, `--audio-stream <n>`, `-i, --interactive`, `--video-ascending`, `--audio-ascending` |
| Страницы | `-p, --pages <spec>`, `--show-all`, `--hide-streams` |
| Содержимое | `--audio-only`, `--video-only`, `--subtitle-only`, `--cover-only`, `--skip-subtitle`, `--sub-lang <list>`, `--auto-subtitles`, `--skip-cover`, `--skip-mux` |
| Вывод | `-o, --output <template>`, `--multi-output <template>`, `-w, --work-dir <dir>`, `--lang <code>`, `--no-tags`, `--archive`, `--delay <seconds>` |
| bilibili | `--api web\|tv\|app\|intl`, `--danmaku`, `--danmaku-only`, `--danmaku-format xml,ass`, `--cookie`, `--token`; другие — через `--help-hidden` |
| Инструменты | `--ffmpeg`, `--yt-dlp`, `--use-mp4box`, `--mp4box`, `--use-aria2c`, `--aria2c`, `--aria2c-args`, `--single-connection` |

<details>
<summary><strong>Приоритеты качества и кодеков</strong></summary>

**Качество и кодеки.** `-q` принимает метки из таблицы потоков: `1080p`, `720p60` на YouTube; `8K`, `Dolby Vision`, `HDR`, `4K`, `1080P60`, `1080P+`, `1080P`, `720P` на bilibili. `-c` принимает `av1 vp9 hevc avc` для видео и `m4a opus flac eac3 mp3` для аудио. Это приоритеты, а не фильтры: всё, что не указано, идёт следом, начиная с лучшего. `--video-ascending` / `--audio-ascending` меняют порядок на обратный — для самой маленькой загрузки.

</details>

<details>
<summary><strong>Выбор страниц</strong></summary>

**Страницы.** `8`, `1,2`, `3-5`, `1-3,10`, `ALL`, `LAST` (последняя страница, то есть новейший выпуск шоу; `LATEST` тоже работает). Ссылка на один выпуск или `?p=N` выбирает только эту страницу. Страницы — это части видео bilibili, серии сезона, выпуски шоу или элементы списка, а также видео публикации X.

</details>

<details>
<summary><strong>Имена файлов и переменные шаблонов</strong></summary>

**Имена файлов.** Материал с одной страницей: `<title>`; с несколькими (даже если `-p` выбирает одну): `<title>/[P<pageNumberWithZero>]<pageTitle>`, номер дополняется нулями по числу страниц. Переменные: `<title>` `<pageNumber>` `<pageNumberWithZero>` `<pageTitle>` `<id>` `<site>` `<uploader>` `<uploaderId>` `<quality>` `<resolution>` `<fps>` `<videoCodec>` `<videoBitrate>` `<audioCodec>` `<audioBitrate>` `<publishDate>` `<pageDate>`, а также переменные bilibili `<bvid>` `<aid>` `<cid>` `<api>`. Для дат можно задать формат: `<publishDate:yyyy-MM-dd>`. Расширение добавляется автоматически.

</details>

<a id="site-notes"></a><a id="notes"></a>

## Особенности платформ

<details>
<summary><strong>YouTube</strong> · извлечение, субтитры и плейлисты</summary>

YouTube защищает URL потоков проверками плеера, которым нужна среда выполнения JavaScript, поэтому извлечение поручено `yt-dlp -J` (он выполняет их в deno); всё остальное haul делает сам. Загруженные автором субтитры встраиваются в файл, автоматически созданные — только с `--auto-subtitles`. googlevideo отдаёт только ограниченные диапазоны байтов, поэтому дорожки скачиваются диапазонами по 10 MB. `watch?v=…&list=…` скачивает только само видео. Обновляйте yt-dlp: именно туда попадают исправления под изменения YouTube.

</details>

<details>
<summary><strong>X</strong> · публичные публикации и аудио</summary>

Для публичных публикаций X вход не нужен. Видео там — отдельные файлы MP4 со встроенным звуком, поэтому таблица показывает только видео; `--audio-only` извлекает аудио.

</details>

<details>
<summary><strong>bilibili</strong> · вход, высокое качество и данмаку</summary>

bilibili читается через собственные web-, TV-, APP- (gRPC) и международные API. Без входа доступно только более низкое качество (обычно до 480P); войдите, чтобы получить 1080P, 4K, HDR, Dolby Vision и Hi-Res-аудио:

```sh
haul login bilibili                # отсканируйте QR-код приложением bilibili
haul login bilibili --from-edge    # macOS: взять вход из Microsoft Edge (или --from-chrome; --profile "Profile 1")
haul login bilibili --tv           # TV-токен доступа для --api tv / --api app
```

Чтение входа из браузера пока работает только в macOS: терминалу нужен полный доступ к диску, а macOS один раз запрашивает доступ к элементу «Safe Storage» в Связке ключей. `--danmaku` сохраняет экранные комментарии в XML и ASS; `--danmaku-format ass` оставляет только один из форматов.

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · выпуски и метаданные</summary>

Xiaoyuzhou и Apple Podcasts не требуют yt-dlp: страницы Xiaoyuzhou содержат ссылку на аудио, ссылки Apple обрабатываются через публичный API iTunes и RSS-ленту шоу. Выпуски сохраняют свой формат (`.mp3` или `.m4a`) со встроенной обложкой, названием шоу в качестве альбома и ведущим в качестве исполнителя. Ссылка на шоу перечисляет последние выпуски от старых к новым, поэтому `-p LAST` — самый новый.

</details>

<a id="config"></a>

## Настройки

`~/.config/haul/` (или `$HAUL_HOME`) содержит:

| Файл | Назначение |
|---|---|
| `config.json` | значения по умолчанию для любых параметров |
| `cookie.txt`, `tv-token.txt`, `app-token.txt` | вход в bilibili |
| `archives.txt` | уже скачанные страницы (`--archive`) |

`config.json` использует имена параметров в camelCase, параметры bilibili — внутри `"bilibili"`, и содержит только то, что меняет; список может быть массивом или строкой через запятую. Флаги командной строки имеют приоритет:

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

## Разработка

```sh
go build ./cmd/haul
go test ./...                                  # офлайн, несколько секунд
HAUL_LIVE=1 go test -run Live ./internal/...   # также обращается к bilibili, YouTube, X и Apple Podcasts
```

Офлайн-набор тестов никогда не обращается к сети: тесты работают с подставным `http.RoundTripper`, который отвечает записанными ответами API и имитациями CDN (серверы, отдающие только диапазоны; обрывы соединения; серверы, игнорирующие диапазоны). Сквозные тесты объединяют дорожки настоящим ffmpeg и проверяют результат через ffprobe; если ffmpeg не установлен, они пропускаются. Архитектура и соглашения описаны в [AGENTS.md](AGENTS.md).

Отправка тега `v*` запускает через GitHub Actions тесты, кросс-компиляцию и публикацию релиза для всех платформ.

<a id="acknowledgements"></a>

## Благодарности

[yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://ffmpeg.org), [GPAC](https://gpac.io), [aria2](https://aria2.github.io), [pflag](https://github.com/spf13/pflag), [x/term](https://pkg.go.dev/golang.org/x/term), [rsc.io/qr](https://pkg.go.dev/rsc.io/qr), а также заметки об API из [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) и [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api).

<a id="license"></a>

## Лицензия

[MIT](LICENSE)

---

<p align="center">
  <strong>Одна ссылка. Ваши медиа.</strong><br>
  <a href="#install">Установить haul</a> · <a href="llms.txt">Справка для агентов</a> · <a href="AGENTS.md">Участие в разработке</a>
</p>
