import Foundation
import QRCodeGenerator

/// A QR code as a module grid, drawable to the terminal or to a PNG.
struct QRImage {
    let size: Int
    private let modules: [[Bool]]

    init(text: String) throws {
        let code = try QRCode.encode(text: text, ecl: .quartile)
        size = code.size
        var rows: [[Bool]] = []
        for y in 0..<code.size {
            rows.append((0..<code.size).map { code.getModule(x: $0, y: y) })
        }
        modules = rows
    }

    func module(_ x: Int, _ y: Int) -> Bool {
        (0..<size).contains(x) && (0..<size).contains(y) ? modules[y][x] : false
    }

    /// Two terminal cells per module, with a quiet zone, using background colours so it scans on any theme.
    func printToConsole() {
        let quiet = 2
        let dark = "\u{1B}[40m  ", light = "\u{1B}[47m  ", reset = "\u{1B}[0m"
        for y in -quiet..<(size + quiet) {
            var line = ""
            for x in -quiet..<(size + quiet) { line += module(x, y) ? dark : light }
            Log.plain(line + reset)
        }
    }

    /// 8-bit greyscale PNG, `scale` pixels per module, 4-module quiet zone.
    func writePNG(to path: String, scale: Int) throws {
        let quiet = 4
        let px = (size + quiet * 2) * scale
        var raw = Data(capacity: px * (px + 1))
        for y in 0..<px {
            raw.append(0) // filter: none
            let my = y / scale - quiet
            for x in 0..<px {
                raw.append(module(x / scale - quiet, my) ? 0x00 : 0xFF)
            }
        }
        var png = Data([0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A])
        var ihdr = Data()
        ihdr.append(contentsOf: Self.be32(UInt32(px)))
        ihdr.append(contentsOf: Self.be32(UInt32(px)))
        ihdr.append(contentsOf: [8, 0, 0, 0, 0]) // depth 8, greyscale, deflate, filter 0, no interlace
        png.append(Self.chunk("IHDR", ihdr))
        png.append(Self.chunk("IDAT", try Gzip.zlibCompress(raw)))
        png.append(Self.chunk("IEND", Data()))
        try png.write(to: URL(fileURLWithPath: path))
    }

    private static func be32(_ v: UInt32) -> [UInt8] {
        [UInt8(v >> 24), UInt8((v >> 16) & 0xFF), UInt8((v >> 8) & 0xFF), UInt8(v & 0xFF)]
    }

    private static func chunk(_ type: String, _ data: Data) -> Data {
        var body = Data(type.utf8)
        body.append(data)
        var out = Data(be32(UInt32(data.count)))
        out.append(body)
        out.append(contentsOf: be32(Gzip.crc32(body)))
        return out
    }
}
