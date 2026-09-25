import Foundation
#if canImport(FoundationXML)
import FoundationXML
#endif

public enum DanmakuFormat: String, Sendable, Codable, CaseIterable {
    case xml, ass
}

public struct DanmakuItem: Sendable {
    public enum Mode: Int, Sendable { case scroll = 1, top = 2, bottom = 3 }

    public var seconds: Double
    public var mode: Mode
    public var fontSize: String
    /// RRGGBB, upper-case hex.
    public var color: String
    public var content: String

    /// Parse the `p` attribute (`time,mode,size,color,timestamp,pool,uid,rowid`) plus text.
    init?(attributes: String, content: String) {
        let parts = attributes.split(separator: ",", omittingEmptySubsequences: false).map(String.init)
        guard parts.count >= 8 else { return nil }
        seconds = Double(parts[0]) ?? 0
        switch parts[1] {
        case "4": mode = .bottom
        case "5": mode = .top
        default: mode = .scroll
        }
        fontSize = parts[2]
        color = String(format: "%06X", Int(parts[3]) ?? 0xFFFFFF)
        self.content = content
    }
}

/// bilibili comment XML → items, and items → ASS.
public enum Danmaku {
    static let screenWidth = 1920
    static let screenHeight = 1080
    static let fontSize = 40
    static let scrollSeconds = 8.0
    static let staticSeconds = 4.0
    static let protectedPercent = 50

    public static func parseXML(at path: String) -> [DanmakuItem]? {
        guard let data = FileManager.default.contents(atPath: path) else { return nil }
        let parser = XMLParser(data: data)
        let delegate = DanmakuXMLDelegate()
        parser.delegate = delegate
        guard parser.parse() else {
            Log.debug("Danmaku XML parse error: \(parser.parserError.map { "\($0)" } ?? "unknown")")
            return nil
        }
        return delegate.items
    }

    public static func writeASS(_ items: [DanmakuItem], to path: String) throws {
        var out = ""
        out += "[Script Info]\n"
        out += "Script Updated By: haul\n"
        out += "ScriptType: v4.00+\n"
        out += "PlayResX: \(screenWidth)\n"
        out += "PlayResY: \(screenHeight)\n"
        out += "Aspect Ratio: \(screenWidth):\(screenHeight)\n"
        out += "Collisions: Normal\n"
        out += "WrapStyle: 2\n"
        out += "ScaledBorderAndShadow: yes\n"
        out += "YCbCr Matrix: TV.601\n"
        out += "[V4+ Styles]\n"
        out += "Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n"
        out += "Style: Danmaku, 黑体, \(fontSize), &H00FFFFFF, &H00FFFFFF, &H00000000, &H00000000, 0, 0, 0, 0, 100, 100, 0.00, 0.00, 1, 2, 0, 7, 0, 0, 0, 0\n"
        out += "[Events]\n"
        out += "Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n"

        var controller = PositionController()
        for d in items.sorted(by: { $0.seconds < $1.seconds }) {
            let length = d.content.count
            let height = controller.place(mode: d.mode, time: d.seconds, length: length)
            if height < 0 { continue }
            let end = d.seconds + (d.mode == .scroll ? scrollSeconds : staticSeconds)
            var effect: String
            switch d.mode {
            case .bottom: effect = "\\an8\\pos(\(screenWidth / 2), \(screenHeight - fontSize - height))"
            case .top: effect = "\\an8\\pos(\(screenWidth / 2), \(height))"
            case .scroll: effect = "\\move(\(screenWidth), \(height), \(-length * fontSize), \(height))"
            }
            if d.color != "FFFFFF" { effect += "\\c&H\(assColor(d.color))&" }
            out += "Dialogue: 2,\(assTime(d.seconds)),\(assTime(end)),Danmaku,,0000,0000,0000,,{\(effect)}\(d.content)\n"
        }
        try out.write(toFile: path, atomically: true, encoding: .utf8)
    }

    /// RRGGBB → BBGGRR, the byte order ASS colour overrides use.
    static func assColor(_ rgb: String) -> String {
        let c = Array(rgb)
        guard c.count == 6 else { return rgb }
        return String(c[4...5] + c[2...3] + c[0...1])
    }

    /// `h:mm:ss.cc`
    static func assTime(_ seconds: Double) -> String {
        let h = Int(seconds) / 3600
        let m = (Int(seconds) % 3600) / 60
        let s = seconds - Double(h * 3600 + m * 60)
        return String(format: "%d:%02d:%05.2f", h, m, s)
    }

    /// Tracks when each row frees up so comments don't pile onto each other.
    struct PositionController {
        private let rows = Danmaku.screenHeight * Danmaku.protectedPercent / Danmaku.fontSize / 100
        private var scroll: [Double]
        private var top: [Double]
        private var bottom: [Double]

        init() {
            scroll = Array(repeating: 0, count: rows)
            top = scroll
            bottom = scroll
        }

        /// Returns the y offset for the comment, or -1 when every row is busy.
        mutating func place(mode: DanmakuItem.Mode, time: Double, length: Int) -> Int {
            var display = Danmaku.staticSeconds
            if mode == .scroll {
                display = Danmaku.scrollSeconds * Double(length + 5) * Double(Danmaku.fontSize)
                    / (Double(Danmaku.screenWidth) + Double(length) * Danmaku.scrollSeconds)
            }
            func find(_ queue: inout [Double]) -> Int {
                for i in 0..<rows where time >= queue[i] {
                    queue[i] = time + display
                    return i * Danmaku.fontSize
                }
                return -1
            }
            switch mode {
            case .bottom: return find(&bottom)
            case .top: return find(&top)
            case .scroll: return find(&scroll)
            }
        }
    }
}

private final class DanmakuXMLDelegate: NSObject, XMLParserDelegate {
    var items: [DanmakuItem] = []
    private var currentAttributes: String?
    private var text = ""

    func parser(_ parser: XMLParser, didStartElement name: String, namespaceURI: String?, qualifiedName: String?, attributes: [String: String]) {
        if name == "d" {
            currentAttributes = attributes["p"]
            text = ""
        }
    }

    func parser(_ parser: XMLParser, foundCharacters string: String) {
        if currentAttributes != nil { text += string }
    }

    func parser(_ parser: XMLParser, didEndElement name: String, namespaceURI: String?, qualifiedName: String?) {
        guard name == "d", let attrs = currentAttributes else { return }
        if let item = DanmakuItem(attributes: attrs, content: text) { items.append(item) }
        currentAttributes = nil
    }
}
