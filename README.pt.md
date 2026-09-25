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
  <bdi><strong>Português</strong></bdi> ·
  <a href="README.id.md"><bdi>Bahasa Indonesia</bdi></a> ·
  <a href="README.ur.md"><bdi>اردو</bdi></a> ·
  <a href="README.ru.md"><bdi>Русский</bdi></a> ·
  <a href="README.de.md"><bdi>Deutsch</bdi></a> ·
  <a href="README.pcm.md"><bdi>Naijá</bdi></a> ·
  <a href="README.arz.md"><bdi>مصري</bdi></a>
</p>
<!-- languages:end -->

<h1 align="center">Um link. Sua mídia.</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="Status da compilação"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="Versão mais recente"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 ou mais recente">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS, Linux e Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="Licença MIT"></a>
</p>

<p align="center">
  <strong>Um programa de linha de comando para baixar vídeo e áudio, feito tanto para pessoas quanto para agentes de IA.</strong><br>
  Passe um link. Escolha os fluxos. Receba vídeo, áudio, legendas, capítulos e capa em um só arquivo.
</p>

<p align="center">
  <a href="#install">Instalação</a> ·
  <a href="#usage">Primeiros passos</a> ·
  <a href="#sites">Sites compatíveis</a> ·
  <a href="#for-ai-agents-and-scripts">Guia para agentes</a> ·
  <a href="#options">Opções</a> ·
  <a href="../../releases">Versões</a>
</p>

---

<a id="small-command-complete-download"></a>

## Um comando pequeno. Um download completo.

| ↓ Sua mídia, do seu jeito | ⌘ Um binário, qualquer desktop | { } Pronto para automação |
| :--- | :--- | :--- |
| Escolha qualidade, codecs, páginas e nomes de arquivo com o mesmo vocabulário em cinco sites | Um único binário Go para macOS, Linux e Windows, com downloads paralelos por intervalos que podem ser retomados | Um documento JSON no stdout, logs no stderr, códigos de saída com significado e nenhuma pergunta sem terminal |

| 01 / INSPECIONAR | 02 / ESCOLHER | 03 / BAIXAR | 04 / COMBINAR |
| :--- | :--- | :--- | :--- |
| **Veja todos os fluxos** | **Defina suas prioridades** | **Leve para casa** | **Tudo junto** |
| Páginas, codecs e tamanhos | Qualidade, áudio e páginas | Por intervalos ou segmentos | Faixas, legendas e capa |
| `haul info "<url>"` | `-q 720p -c avc,m4a` | `haul "<url>"` | `ffmpeg / MP4Box` |

<details>
<summary><strong>Dê uma olhada no terminal</strong></summary>

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

> Para uso pessoal, de pesquisa e outros fins não comerciais. Você é responsável por respeitar os direitos autorais e os termos de cada site.

<a id="install"></a>

## Instalação

haul funciona em **macOS, Linux e Windows** (amd64 e arm64). Precisa do [ffmpeg](https://ffmpeg.org) para combinar as faixas; YouTube e X também precisam do [yt-dlp](https://github.com/yt-dlp/yt-dlp), e os desafios do player do YouTube precisam do [deno](https://deno.com):

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg yt-dlp           # Debian / Ubuntu; deno: https://deno.com
winget install Gyan.FFmpeg               # Windows, um pacote de cada vez
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

<a id="get-the-binary"></a><a id="binary"></a>

### Obter o binário

Baixe o arquivo compactado do seu sistema em [Releases](../../releases) — `haul-<version>-<os>-<arch>.tar.gz` (`.zip` no Windows) — e coloque `haul` em qualquer diretório do seu `PATH`:

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
```

O fluxo atual assina e notariza as versões macOS com a Apple. Escolha `haul-<version>-darwin-<arch>.pkg` para instalar `haul` em `/usr/local/bin` com o comprovante de notarização anexado. O arquivo compactado contém o mesmo executável assinado, mas a primeira verificação pode precisar de internet. Versões antigas podem não estar assinadas.

<details>
<summary><strong>Compilar a partir do código-fonte</strong> · Go 1.26+</summary>

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

ou, a partir de um clone do repositório:

```sh
go build -o haul ./cmd/haul
```

</details>

<a id="usage"></a>

## Primeiros passos

Comece com um link. Por padrão, haul escolhe os melhores fluxos disponíveis.

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

Consulte as informações antes de baixar, ou escolha uma qualidade e um codec:

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # H.264 + AAC para o QuickTime
```

<details>
<summary><strong>Referência de comandos</strong></summary>

| Comando | Significado |
| --- | --- |
| `haul <url> [options]` | Baixar; equivale a `haul download <url>` |
| `haul info <url> [--urls]` | Mostrar o item, suas páginas e fluxos, sem baixar nada |
| `haul login bilibili` | Entrar no bilibili para obter qualidades mais altas |
| `haul templates` | Variáveis dos modelos de nome de arquivo |
| `haul --help` | Sites, uso por agentes, códigos de saída e exemplos |
| `haul download --help` | Todas as opções |

</details>

<a id="sites"></a>

### Sites compatíveis

| Site | Links | Requer |
|---|---|---|
| **YouTube** | `youtube.com/watch?v=…`, `youtu.be/…`, `/shorts/…`, `/embed/…`, `/live/…` | yt-dlp, deno |
| **X** | `x.com/<user>/status/<id>`, `twitter.com/…`; `/video/<n>` escolhe um vídeo da publicação | yt-dlp |
| **bilibili** | vídeos, bangumi, cursos, coleções, séries, favoritos, espaços de usuário, `b23.tv` e identificadores avulsos `BV…` `av…` `ep…` `ss…` `md…` | – |
| **Xiaoyuzhou** | `xiaoyuzhoufm.com/episode/<id>`, `/podcast/<id>` | – |
| **Apple Podcasts** | `podcasts.apple.com/<cc>/podcast/<name>/id<show>`, com `?i=<episode>` para um único episódio | – |

<a id="make-it-yours"></a><a id="examples"></a>

### Do seu jeito

**Só o áudio**

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # AAC em vez de Opus
```

**Algumas partes, uma temporada inteira ou o episódio mais recente**

```sh
# Partes selecionadas, salvas na sua pasta Movies
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# Todos os episódios de uma temporada
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# O episódio de podcast mais recente
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# Todos os vídeos de uma publicação do X
haul -p ALL "https://x.com/<user>/status/<id>"
```

**Arquivos organizados, legendas ou uma escolha interativa**

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # só estes idiomas de legenda
haul -i "BV1qt4y1X7TW"                                 # escolha os fluxos com as teclas de seta
```

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## Para agentes de IA e scripts

`haul --help` traz tudo de que um agente precisa; [llms.txt](llms.txt) é o mesmo guia em forma de arquivo. Em resumo:

```sh
haul info --json "<url>"                                # 1. consultar
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. baixar exatamente esses fluxos
haul --json -q 720p -c avc,m4a "<url>"                  #    ou deixar as prioridades escolherem
```

- Com `--json`, **o stdout contém apenas um documento JSON**; progresso e logs vão para o stderr. O campo `files` lista os arquivos de saída, novos ou já existentes.
- **Nada é interativo sem terminal.** Sem terminal, `-i` falha com o código de saída 2 e indica as opções a usar no lugar.
- Os índices de fluxo de `info` seguem a ordem de escolha do haul; passe os mesmos `-q` / `-c` para `info` e para o download.
- Em uma playlist, temporada, programa ou publicação com vários vídeos, `info` lista as páginas; `-p <n>` acrescenta os fluxos e as legendas daquela página.
- Páginas que já estão no disco são informadas como `"status": "skipped", "reason": "exists"`, então executar de novo é seguro.

<details>
<summary><strong>Exemplo de resposta JSON</strong> · um download bem-sucedido</summary>

Um download imprime:

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

Essa é a saída de `haul --json -q 360p -c avc,m4a …`. Aqui as chaves estão na ordem de leitura e as listas de fluxos foram reduzidas aos escolhidos; o documento real ordena as chaves alfabeticamente e lista todos os fluxos.

</details>

Uma falha imprime `"ok": false` com `"error": {"kind", "message", "exitCode"}` e mantém as páginas concluídas antes dela.

| Código de saída | Significado | `error.kind` |
|---|---|---|
| 0 | concluído | – |
| 1 | o download ou a extração falhou | `failed` |
| 2 | link não compatível ou valor de opção inválido | `input` |
| 3 | falta uma ferramenta necessária (ffmpeg, yt-dlp) | `dependency` |
| 4 | login necessário ou expirado | `auth` |
| 64 | linha de comando inválida (opção desconhecida, argumento ausente) | – |
| 130 | cancelado | `cancelled` |

<a id="options"></a>

## Opções

`haul download --help` lista todas. As principais:

| Grupo | Opções |
|---|---|
| Geral | `--json`, `--config <file>`, `--debug` |
| Fluxos | `-q, --quality <list>`, `-c, --codec <list>`, `--video-stream <n>`, `--audio-stream <n>`, `-i, --interactive`, `--video-ascending`, `--audio-ascending` |
| Páginas | `-p, --pages <spec>`, `--show-all`, `--hide-streams` |
| Conteúdo | `--audio-only`, `--video-only`, `--subtitle-only`, `--cover-only`, `--skip-subtitle`, `--sub-lang <list>`, `--auto-subtitles`, `--skip-cover`, `--skip-mux` |
| Saída | `-o, --output <template>`, `--multi-output <template>`, `-w, --work-dir <dir>`, `--lang <code>`, `--no-tags`, `--archive`, `--delay <seconds>` |
| bilibili | `--api web\|tv\|app\|intl`, `--danmaku`, `--danmaku-only`, `--danmaku-format xml,ass`, `--cookie`, `--token`; mais opções com `--help-hidden` |
| Ferramentas | `--ffmpeg`, `--yt-dlp`, `--use-mp4box`, `--mp4box`, `--use-aria2c`, `--aria2c`, `--aria2c-args`, `--single-connection` |

<details>
<summary><strong>Prioridades de qualidade e codec</strong></summary>

**Qualidades e codecs.** `-q` aceita os rótulos mostrados na tabela de fluxos: `1080p`, `720p60` no YouTube; `8K`, `Dolby Vision`, `HDR`, `4K`, `1080P60`, `1080P+`, `1080P`, `720P` no bilibili. `-c` aceita `av1 vp9 hevc avc` para vídeo e `m4a opus flac eac3 mp3` para áudio. São prioridades, não filtros: o que não estiver na lista vem depois, do melhor para o pior. `--video-ascending` / `--audio-ascending` invertem a ordem para obter o menor download.

</details>

<details>
<summary><strong>Seleção de páginas</strong></summary>

**Páginas.** `8`, `1,2`, `3-5`, `1-3,10`, `ALL`, `LAST` (a última página, que é o episódio mais recente de um programa; `LATEST` também funciona). Um link para um episódio, ou `?p=N`, seleciona só essa página. Páginas são as partes de um vídeo do bilibili, os episódios de uma temporada, programa ou lista, e os vídeos de uma publicação do X.

</details>

<details>
<summary><strong>Nomes de arquivo e variáveis de modelo</strong></summary>

**Nomes de arquivo.** Um item com uma página: `<title>`; com várias (mesmo quando `-p` escolhe uma só): `<title>/[P<pageNumberWithZero>]<pageTitle>`, com o número completado com zeros de acordo com o total de páginas. Variáveis: `<title>` `<pageNumber>` `<pageNumberWithZero>` `<pageTitle>` `<id>` `<site>` `<uploader>` `<uploaderId>` `<quality>` `<resolution>` `<fps>` `<videoCodec>` `<videoBitrate>` `<audioCodec>` `<audioBitrate>` `<publishDate>` `<pageDate>`, além das do bilibili: `<bvid>` `<aid>` `<cid>` `<api>`. As datas aceitam um formato: `<publishDate:yyyy-MM-dd>`. A extensão é adicionada automaticamente.

</details>

<a id="site-notes"></a><a id="notes"></a>

## Notas por site

<details>
<summary><strong>YouTube</strong> · extração, legendas e playlists</summary>

O YouTube protege as URLs dos fluxos com desafios do player que exigem um runtime de JavaScript, por isso a extração fica a cargo de `yt-dlp -J` (que os executa no deno); todo o resto é feito pelo próprio haul. Legendas enviadas pelo autor são incluídas no arquivo; as geradas automaticamente, só com `--auto-subtitles`. O googlevideo só fornece intervalos de bytes limitados, então as faixas são baixadas em intervalos de 10 MB, um de cada vez. `watch?v=…&list=…` baixa só o vídeo. Mantenha o yt-dlp atualizado: é nele que chegam as correções para mudanças do YouTube.

</details>

<details>
<summary><strong>X</strong> · publicações públicas e áudio</summary>

O X não exige login para publicações públicas. Os vídeos são arquivos MP4 únicos com o áudio incluído, por isso a tabela mostra só o vídeo; `--audio-only` extrai o áudio.

</details>

<details>
<summary><strong>bilibili</strong> · login, qualidades mais altas e danmaku</summary>

O bilibili é lido pelas suas próprias APIs web, TV, APP (gRPC) e internacional. Sem login, oferece só qualidades mais baixas (normalmente até 480P); entre para obter 1080P, 4K, HDR, Dolby Vision e áudio Hi-Res:

```sh
haul login bilibili                # escaneie um QR code com o app do bilibili
haul login bilibili --from-edge    # macOS: reutiliza o login do Microsoft Edge (ou --from-chrome; --profile "Profile 1")
haul login bilibili --tv           # token de acesso de TV, para --api tv / --api app
```

Por enquanto, ler o login de um navegador só funciona no macOS: o terminal precisa de Acesso Total ao Disco, e o macOS pede uma vez acesso ao item “Safe Storage” das Chaves. `--danmaku` salva os comentários sobrepostos (danmaku) em XML e ASS; `--danmaku-format ass` mantém só um deles.

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · episódios e metadados</summary>

Xiaoyuzhou e Apple Podcasts não precisam do yt-dlp: as páginas do Xiaoyuzhou trazem o link do áudio, e os links da Apple passam pela API pública do iTunes e pelo feed RSS do programa. Os episódios mantêm o formato (`.mp3` ou `.m4a`), com a capa incorporada, o programa como álbum e o apresentador como artista. O link de um programa lista os episódios mais recentes do mais antigo para o mais novo, então `-p LAST` é o mais recente.

</details>

<a id="config"></a>

## Configuração

`~/.config/haul/` (ou `$HAUL_HOME`) contém:

| Arquivo | Finalidade |
|---|---|
| `config.json` | valores padrão para qualquer opção |
| `cookie.txt`, `tv-token.txt`, `app-token.txt` | login do bilibili |
| `archives.txt` | páginas já baixadas (`--archive`) |

`config.json` usa os nomes das opções em camelCase, com as do bilibili em `"bilibili"`, e só precisa do que você quer mudar; uma lista pode ser um array ou uma string separada por vírgulas. As opções da linha de comando têm prioridade sobre ele:

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

## Desenvolvimento

```sh
go build ./cmd/haul
go test ./...                                  # offline, poucos segundos
HAUL_LIVE=1 go test -run Live ./internal/...   # também acessa bilibili, YouTube, X e Apple Podcasts
```

Os testes offline nunca acessam a rede: eles conversam com um `http.RoundTripper` simulado que responde a partir de respostas de API gravadas e de CDNs simuladas (servidores que só aceitam intervalos, conexões interrompidas, servidores que ignoram intervalos). Os testes de ponta a ponta combinam as faixas com um ffmpeg real e conferem o resultado com ffprobe; eles são ignorados quando o ffmpeg não está instalado. Veja a arquitetura e as convenções em [AGENTS.md](AGENTS.md).

Enviar uma tag `v*` faz o GitHub Actions testar, compilar para cada plataforma e publicar uma versão para todas elas.

<a id="acknowledgements"></a>

## Agradecimentos

[yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://ffmpeg.org), [GPAC](https://gpac.io), [aria2](https://aria2.github.io), [pflag](https://github.com/spf13/pflag), [x/term](https://pkg.go.dev/golang.org/x/term), [rsc.io/qr](https://pkg.go.dev/rsc.io/qr) e as notas sobre APIs de [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) e [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api).

<a id="license"></a>

## Licença

[MIT](LICENSE)

---

<p align="center">
  <strong>Um link. Sua mídia.</strong><br>
  <a href="#install">Obtenha o haul</a> · <a href="llms.txt">Referência para agentes</a> · <a href="AGENTS.md">Contribuir</a>
</p>
