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
  <bdi><strong>Bahasa Indonesia</strong></bdi> ·
  <a href="README.ur.md"><bdi>اردو</bdi></a> ·
  <a href="README.ru.md"><bdi>Русский</bdi></a> ·
  <a href="README.de.md"><bdi>Deutsch</bdi></a> ·
  <a href="README.pcm.md"><bdi>Naijá</bdi></a> ·
  <a href="README.arz.md"><bdi>مصري</bdi></a>
</p>
<!-- languages:end -->

<h1 align="center">Satu tautan. Media milik Anda.</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="Status build"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="Rilis terbaru"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 atau lebih baru">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS, Linux, dan Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="Lisensi MIT"></a>
</p>

<p align="center">
  <strong>Pengunduh video dan audio lewat baris perintah, dibuat untuk manusia maupun agen AI.</strong><br>
  Berikan sebuah tautan. Pilih aliran media Anda. Dapatkan video, audio, subtitel, bab, dan sampul dalam satu berkas.
</p>

<p align="center">
  <a href="#install">Instalasi</a> ·
  <a href="#usage">Mulai cepat</a> ·
  <a href="#sites">Situs yang didukung</a> ·
  <a href="#for-ai-agents-and-scripts">Panduan agen</a> ·
  <a href="#options">Opsi</a> ·
  <a href="../../releases">Rilis</a>
</p>

---

<a id="small-command-complete-download"></a>

## Perintah singkat. Unduhan lengkap.

| ↓ Media Anda, sesuai keinginan Anda | ⌘ Satu biner untuk setiap desktop | { } Siap untuk otomatisasi |
| :--- | :--- | :--- |
| Pilih kualitas, codec, halaman, dan nama berkas dengan istilah yang sama di lima situs | Satu biner Go untuk macOS, Linux, dan Windows, dengan unduhan rentang paralel yang dapat dilanjutkan | Satu dokumen JSON di stdout, log di stderr, kode keluar yang bermakna, dan tanpa pertanyaan interaktif jika tidak ada terminal |

| 01 / PERIKSA | 02 / PILIH | 03 / UNDUH | 04 / GABUNGKAN |
| :--- | :--- | :--- | :--- |
| **Lihat setiap aliran** | **Atur prioritas Anda** | **Bawa pulang** | **Satukan semuanya** |
| Halaman, codec, dan ukuran | Kualitas, audio, dan halaman | Per rentang atau per segmen | Trek, subtitel, dan sampul |
| `haul info "<url>"` | `-q 720p -c avc,m4a` | `haul "<url>"` | `ffmpeg / MP4Box` |

<details>
<summary><strong>Lihat tampilannya di terminal</strong></summary>

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

> Untuk penggunaan pribadi, penelitian, dan penggunaan nonkomersial lainnya. Anda bertanggung jawab menghormati hak cipta dan ketentuan setiap situs.

<a id="install"></a>

## Instalasi

haul berjalan di **macOS, Linux, dan Windows** (amd64 dan arm64). haul memerlukan [ffmpeg](https://ffmpeg.org) untuk penggabungan; YouTube dan X juga memerlukan [yt-dlp](https://github.com/yt-dlp/yt-dlp), dan tantangan pemutar YouTube memerlukan [deno](https://deno.com):

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg pipx unzip       # Linux (Debian / Ubuntu)
pipx install "yt-dlp[default]" && pipx ensurepath
curl -fsSL https://deno.land/install.sh | sh -s -- -y
winget install Gyan.FFmpeg               # Windows, satu paket per perintah
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

Di Linux, pasang yt-dlp dengan pipx atau berkas biner `yt-dlp_linux` dari [rilisnya](https://github.com/yt-dlp/yt-dlp/releases/latest) (`yt-dlp_linux_aarch64` di arm64), bukan dari distribusi Anda: YouTube sering berubah, dan paket distribusi cepat tertinggal. Setelah memasang, buka terminal baru agar alat-alat itu ada di `PATH`.

<a id="get-the-binary"></a><a id="binary"></a>

### Dapatkan biner

Unduh arsip untuk sistem Anda dari [Releases](../../releases) — `haul-<version>-<os>-<arch>.tar.gz` (`.zip` di Windows) — lalu letakkan `haul` di direktori mana pun dalam `PATH` Anda:

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
```

Alur rilis saat ini menandatangani versi macOS dan mengirimkannya untuk notarisasi Apple. Pilih `haul-<version>-darwin-<arch>.pkg` untuk memasang `haul` ke `/usr/local/bin` dengan tiket notarisasi terlampir. Arsip berisi program bertanda tangan yang sama, tetapi pemeriksaan pertama mungkin memerlukan internet. Rilis lama mungkin belum ditandatangani.

<details>
<summary><strong>Bangun dari sumber</strong> · Go 1.26+</summary>

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

atau, dari hasil klon repositori:

```sh
go build -o haul ./cmd/haul
```

</details>

<a id="usage"></a>

## Mulai cepat

Mulailah dengan sebuah tautan. Secara bawaan, haul memilih aliran terbaik yang tersedia.

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

Periksa sebelum mengunduh, atau pilih kualitas dan codec:

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # H.264 + AAC untuk QuickTime
```

<details>
<summary><strong>Referensi perintah</strong></summary>

| Perintah | Fungsi |
|---|---|
| `haul <url> [options]` | unduh (sama dengan `haul download <url>`) |
| `haul info <url> [--urls]` | tampilkan item, halaman, dan alirannya; tidak mengunduh apa pun |
| `haul login bilibili` | login ke bilibili untuk kualitas lebih tinggi |
| `haul templates` | variabel templat nama berkas |
| `haul --help` | situs, penggunaan agen, kode keluar, contoh |
| `haul download --help` | semua opsi |

</details>

<a id="sites"></a>

### Situs yang didukung

| Situs | Tautan | Memerlukan |
|---|---|---|
| **YouTube** | `youtube.com/watch?v=…`, `youtu.be/…`, `/shorts/…`, `/embed/…`, `/live/…` | yt-dlp, deno |
| **X** | `x.com/<user>/status/<id>`, `twitter.com/…`; `/video/<n>` memilih satu video dari sebuah kiriman | yt-dlp |
| **bilibili** | video, bangumi, kursus, koleksi, seri, favorit, ruang pengguna, `b23.tv`, ID polos `BV…` `av…` `ep…` `ss…` `md…` | – |
| **Xiaoyuzhou** | `xiaoyuzhoufm.com/episode/<id>`, `/podcast/<id>` | – |
| **Apple Podcasts** | `podcasts.apple.com/<cc>/podcast/<name>/id<show>`, dengan `?i=<episode>` untuk satu episode | – |

<a id="make-it-yours"></a><a id="examples"></a>

### Sesuaikan dengan kebutuhan Anda

**Audio saja**

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # AAC, bukan Opus
```

**Beberapa bagian, satu musim penuh, atau episode terbaru**

```sh
# Bagian pilihan, disimpan ke folder Movies
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# Semua episode dalam satu musim
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# Episode podcast terbaru
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# Semua video dalam kiriman X
haul -p ALL "https://x.com/<user>/status/<id>"
```

**Berkas yang tertata, subtitel, atau pilihan interaktif**

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # hanya bahasa subtitel ini
haul -i "BV1qt4y1X7TW"                                 # pilih aliran dengan tombol panah
```

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## Untuk agen AI dan skrip

`haul --help` memuat semua yang dibutuhkan agen; [llms.txt](llms.txt) adalah panduan yang sama dalam bentuk berkas. Versi singkatnya:

```sh
haul info --json "<url>"                                # 1. periksa
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. unduh tepat aliran-aliran itu
haul --json -q 720p -c avc,m4a "<url>"                  #    atau biarkan prioritas yang memilih
```

- Dengan `--json`, **stdout hanya berisi satu dokumen JSON**; progres dan log masuk ke stderr. Kolom `files`-nya mencantumkan berkas keluaran, baik yang baru maupun yang sudah ada.
- **Tidak ada yang interaktif tanpa terminal.** `-i` tanpa terminal gagal dengan kode keluar 2 dan menyebutkan opsi yang harus dipakai sebagai gantinya.
- Indeks aliran di `info` mengikuti urutan pemilihan haul; berikan `-q` / `-c` yang sama ke `info` dan ke unduhan.
- Untuk daftar putar, musim, acara, atau kiriman multivideo, `info` mencantumkan halamannya; `-p <n>` menambahkan aliran dan subtitel halaman itu.
- Halaman yang sudah ada di disk dilaporkan sebagai `"status": "skipped", "reason": "exists"`, sehingga aman dijalankan ulang.

<details>
<summary><strong>Contoh respons JSON</strong> · unduhan yang berhasil</summary>

Sebuah unduhan mencetak:

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

Itu hasil `haul --json -q 360p -c avc,m4a …`. Di sini kunci disusun menurut urutan baca dan daftar aliran dipangkas menjadi yang terpilih saja; dokumen aslinya mengurutkan kuncinya menurut abjad dan mencantumkan setiap aliran.

</details>

Kegagalan mencetak `"ok": false` dengan `"error": {"kind", "message", "exitCode"}`, dan tetap menyertakan halaman yang selesai sebelumnya.

| Kode keluar | Arti | `error.kind` |
|---|---|---|
| 0 | selesai | – |
| 1 | unduhan atau ekstraksi gagal | `failed` |
| 2 | tautan tidak didukung atau nilai opsi tidak valid | `input` |
| 3 | alat yang diperlukan tidak ada (ffmpeg, yt-dlp) | `dependency` |
| 4 | perlu login atau login kedaluwarsa | `auth` |
| 64 | baris perintah tidak valid (flag tidak dikenal, argumen kurang) | – |
| 130 | dibatalkan | `cancelled` |

<a id="options"></a>

## Opsi

`haul download --help` mencantumkan semuanya. Opsi utama:

| Grup | Opsi |
|---|---|
| Umum | `--json`, `--config <file>`, `--debug` |
| Aliran | `-q, --quality <list>`, `-c, --codec <list>`, `--video-stream <n>`, `--audio-stream <n>`, `-i, --interactive`, `--video-ascending`, `--audio-ascending` |
| Halaman | `-p, --pages <spec>`, `--show-all`, `--hide-streams` |
| Konten | `--audio-only`, `--video-only`, `--subtitle-only`, `--cover-only`, `--skip-subtitle`, `--sub-lang <list>`, `--auto-subtitles`, `--skip-cover`, `--skip-mux` |
| Keluaran | `-o, --output <template>`, `--multi-output <template>`, `-w, --work-dir <dir>`, `--lang <code>`, `--no-tags`, `--archive`, `--delay <seconds>` |
| bilibili | `--api web\|tv\|app\|intl`, `--danmaku`, `--danmaku-only`, `--danmaku-format xml,ass`, `--cookie`, `--token`; lainnya dengan `--help-hidden` |
| Alat | `--ffmpeg`, `--yt-dlp`, `--use-mp4box`, `--mp4box`, `--use-aria2c`, `--aria2c`, `--aria2c-args`, `--single-connection` |

<details>
<summary><strong>Prioritas kualitas dan codec</strong></summary>

**Kualitas dan codec.** `-q` menerima label yang ditampilkan tabel aliran: `1080p`, `720p60` di YouTube; `8K`, `Dolby Vision`, `HDR`, `4K`, `1080P60`, `1080P+`, `1080P`, `720P` di bilibili. `-c` menerima `av1 vp9 hevc avc` untuk video dan `m4a opus flac eac3 mp3` untuk audio. Keduanya adalah prioritas, bukan filter: apa pun yang tidak dicantumkan menyusul setelahnya, dimulai dari yang terbaik. `--video-ascending` / `--audio-ascending` membalik urutannya untuk unduhan terkecil.

</details>

<details>
<summary><strong>Pemilihan halaman</strong></summary>

**Halaman.** `8`, `1,2`, `3-5`, `1-3,10`, `ALL`, `LAST` (halaman terakhir, yaitu episode terbaru sebuah acara; `LATEST` juga bisa). Tautan ke satu episode, atau `?p=N`, memilih halaman itu saja. Halaman adalah bagian dari video bilibili, episode dari musim, acara, atau daftar, dan video dalam kiriman X.

</details>

<details>
<summary><strong>Nama berkas dan variabel templat</strong></summary>

**Nama berkas.** Item dengan satu halaman: `<title>`; dengan beberapa halaman (bahkan saat `-p` hanya memilih satu): `<title>/[P<pageNumberWithZero>]<pageTitle>`, diberi nol di depan sesuai jumlah halaman. Variabel: `<title>` `<pageNumber>` `<pageNumberWithZero>` `<pageTitle>` `<id>` `<site>` `<uploader>` `<uploaderId>` `<quality>` `<resolution>` `<fps>` `<videoCodec>` `<videoBitrate>` `<audioCodec>` `<audioBitrate>` `<publishDate>` `<pageDate>`, ditambah variabel khusus bilibili `<bvid>` `<aid>` `<cid>` `<api>`. Tanggal dapat diberi format: `<publishDate:yyyy-MM-dd>`. Ekstensi ditambahkan otomatis.

</details>

<a id="site-notes"></a><a id="notes"></a>

## Catatan per situs

<details>
<summary><strong>YouTube</strong> · ekstraksi, subtitel, dan daftar putar</summary>

YouTube melindungi URL alirannya dengan tantangan pemutar yang memerlukan runtime JavaScript, sehingga ekstraksi diserahkan ke `yt-dlp -J` (yang menjalankannya di deno); semua langkah setelahnya dikerjakan haul sendiri. Subtitel unggahan digabungkan ke berkas; subtitel yang dibuat otomatis hanya dengan `--auto-subtitles`. googlevideo hanya melayani rentang byte terbatas, sehingga trek diambil per rentang 10 MB. `watch?v=…&list=…` hanya mengunduh videonya. Selalu perbarui yt-dlp: di sanalah perbaikan untuk perubahan YouTube masuk.

</details>

<details>
<summary><strong>X</strong> · kiriman publik dan audio</summary>

Kiriman publik X tidak memerlukan login. Videonya berupa satu berkas MP4 dengan audio di dalamnya, sehingga tabel hanya mencantumkan video; `--audio-only` mengekstrak audionya.

</details>

<details>
<summary><strong>bilibili</strong> · login, kualitas lebih tinggi, dan danmaku</summary>

bilibili dibaca melalui API web, TV, APP (gRPC), dan internasional miliknya sendiri. Tanpa login, hanya kualitas yang lebih rendah yang tersedia (biasanya hingga 480P); login untuk 1080P, 4K, HDR, Dolby Vision, dan audio Hi-Res:

```sh
haul login bilibili                # pindai kode QR dengan aplikasi bilibili
haul login bilibili --from-edge    # macOS: pakai ulang login Microsoft Edge (atau --from-chrome; --profile "Profile 1")
haul login bilibili --tv           # token akses TV, untuk --api tv / --api app
```

Untuk saat ini, membaca login dari peramban hanya berfungsi di macOS: terminal memerlukan Full Disk Access, dan macOS meminta izin sekali untuk item keychain "Safe Storage". `--danmaku` menyimpan komentar berjalan sebagai XML dan ASS; `--danmaku-format ass` hanya menyimpan salah satunya.

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · episode dan metadata</summary>

Xiaoyuzhou dan Apple Podcasts tidak memerlukan yt-dlp: halaman Xiaoyuzhou memuat tautan audionya, sedangkan tautan Apple diproses lewat API iTunes publik dan feed RSS acara. Episode mempertahankan formatnya (`.mp3` atau `.m4a`) dengan sampul tertanam, nama acara sebagai album, dan pembawa acara sebagai artis. Tautan acara mencantumkan episode terkininya dari yang terlama, sehingga `-p LAST` adalah yang terbaru.

</details>

<a id="config"></a>

## Konfigurasi

`~/.config/haul/` (atau `$HAUL_HOME`) berisi:

| Berkas | Fungsi |
|---|---|
| `config.json` | nilai bawaan untuk opsi apa pun |
| `cookie.txt`, `tv-token.txt`, `app-token.txt` | login bilibili |
| `archives.txt` | halaman yang sudah diunduh (`--archive`) |

`config.json` memakai nama opsi dalam camelCase, opsi bilibili di bawah `"bilibili"`, dan hanya perlu berisi yang diubah; daftar bisa berupa array atau string yang dipisahkan koma. Flag baris perintah mengesampingkannya:

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

## Pengembangan

```sh
go build ./cmd/haul
go test ./...                                  # luring, beberapa detik
HAUL_LIVE=1 go test -run Live ./internal/...   # juga mengakses bilibili, YouTube, X, dan Apple Podcasts
```

Rangkaian pengujian luring tidak pernah menyentuh jaringan: pengujian berbicara dengan stub `http.RoundTripper` yang menjawab dari respons API rekaman dan CDN simulasi (server yang hanya melayani rentang, koneksi terputus, server yang mengabaikan rentang). Pengujian menyeluruh menggabungkan trek dengan ffmpeg asli dan memeriksa hasilnya dengan ffprobe; pengujian ini dilewati jika ffmpeg tidak terpasang. Lihat [AGENTS.md](AGENTS.md) untuk arsitektur dan konvensinya.

Mendorong tag `v*` akan menjalankan pengujian, kompilasi silang, dan penerbitan rilis untuk setiap platform melalui GitHub Actions.

<a id="acknowledgements"></a>

## Ucapan terima kasih

[yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://ffmpeg.org), [GPAC](https://gpac.io), [aria2](https://aria2.github.io), [pflag](https://github.com/spf13/pflag), [x/term](https://pkg.go.dev/golang.org/x/term), [rsc.io/qr](https://pkg.go.dev/rsc.io/qr), serta catatan API dari [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) dan [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api).

<a id="license"></a>

## Lisensi

[MIT](LICENSE)

---

<p align="center">
  <strong>Satu tautan. Media milik Anda.</strong><br>
  <a href="#install">Dapatkan haul</a> · <a href="llms.txt">Referensi agen</a> · <a href="AGENTS.md">Berkontribusi</a>
</p>
