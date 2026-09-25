import Foundation

/// The three request signatures bilibili's endpoints want.
public enum Signing {
    public static let tvAppKey = "4409e2ce8ffd12b8"
    public static let tvAppSecret = "59b43e04ad6965f34319062b478f83dd"
    public static let biliPlusAppKey = "7d089525d3611b1c"
    public static let biliPlusSecret = "acd495b248ec528c2eed1e862d393126"

    private static let mixinKeyTable: [Int] = [
        46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35,
        27, 43, 5, 49, 33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13,
    ]

    /// WBI mixin key from the two image keys the nav endpoint returns.
    public static func wbiMixinKey(imgKey: String, subKey: String) -> String {
        let source = Array(imgKey + subKey)
        return String(mixinKeyTable.compactMap { $0 < source.count ? source[$0] : nil })
    }

    /// Append `w_rid` to a query that already ends in `wts=`.
    public static func wbiSign(_ query: String, key: String) -> String {
        "\(query)&w_rid=\(md5Hex(query + key))"
    }

    /// appkey-style signature: md5(query + secret).
    public static func appSign(_ query: String, secret: String) -> String {
        md5Hex(query + secret)
    }

    /// The file name of a bfs image URL without its extension: `.../abc.png` → `abc`.
    public static func imageKey(from url: String) -> String {
        let name = url.split(separator: "/").last.map(String.init) ?? url
        if let dot = name.lastIndex(of: ".") { return String(name[..<dot]) }
        return name
    }
}
