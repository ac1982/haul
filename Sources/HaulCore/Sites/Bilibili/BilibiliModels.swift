import Foundation

/// Which bilibili playurl backend to ask.
public enum APIType: String, Sendable, Codable, CaseIterable {
    case web, tv, app, intl

    public var label: String { rawValue.uppercased() }
}

/// bilibili quality ids, highest first. The first entry is what "max quality" means.
public enum Quality {
    public static let table: [(id: String, name: String)] = [
        ("127", "8K"), ("126", "Dolby Vision"), ("125", "HDR"), ("120", "4K"), ("116", "1080P60"),
        ("112", "1080P+"), ("100", "AI Enhanced"), ("80", "1080P"), ("74", "720P60"),
        ("64", "720P"), ("48", "720P"), ("32", "480P"), ("16", "360P"),
        ("5", "144P"), ("6", "240P"),
    ]
    public static let maxID = "127"
    public static let dolbyVisionName = "Dolby Vision"

    public static func name(_ id: String) -> String {
        table.first { $0.id == id }?.name ?? "Unknown (\(id))"
    }

    public static func videoCodec(_ codecid: String) -> String {
        switch codecid {
        case "13": return "AV1"
        case "12": return "HEVC"
        case "7": return "AVC"
        default: return "UNKNOWN"
        }
    }
}
