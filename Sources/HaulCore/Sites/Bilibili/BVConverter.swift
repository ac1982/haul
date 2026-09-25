import Foundation

/// av ⇄ BV. Algorithm from https://github.com/Colerar/abv.
public enum BVConverter {
    private static let xorCode: Int64 = 23_442_827_791_579
    private static let maskCode: Int64 = (1 << 51) - 1
    private static let maxAid: Int64 = maskCode + 1
    private static let base: Int64 = 58
    private static let bvLength = 9
    private static let alphabet = Array("FcwAPNKTMug3GV5Lj7EJnHpWsx4tb8haYeviqBz6rkCy12mUSDQX9RdoZf")
    private static let reverse: [Character: Int64] = {
        var d: [Character: Int64] = [:]
        for (i, c) in alphabet.enumerated() { d[c] = Int64(i) }
        return d
    }()

    public static func encode(_ avid: Int64) throws -> String {
        guard avid >= 1 else { throw HaulError.input("av\(avid) is below 1") }
        guard avid < maxAid else { throw HaulError.input("av\(avid) is out of range") }
        var out = [Character](repeating: "0", count: bvLength)
        var tmp = (maxAid | avid) ^ xorCode
        var i = bvLength - 1
        while tmp != 0 {
            out[i] = alphabet[Int(tmp % base)]
            tmp /= base
            i -= 1
        }
        out.swapAt(0, 6)
        out.swapAt(1, 4)
        return "BV1" + String(out)
    }

    /// Takes the 9 characters after `BV1`.
    public static func decode(_ body: String) throws -> Int64 {
        var chars = Array(body)
        guard chars.count == bvLength else { throw HaulError.input("BV1\(body) must be 12 characters") }
        chars.swapAt(0, 6)
        chars.swapAt(1, 4)
        var avid: Int64 = 0
        for c in chars {
            guard let v = reverse[c] else { throw HaulError.input("Invalid character in the BV id: \(c)") }
            avid = avid * base + v
        }
        return (avid & maskCode) ^ xorCode
    }
}
