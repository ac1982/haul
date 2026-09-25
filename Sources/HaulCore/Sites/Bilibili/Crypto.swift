import Foundation
#if canImport(CryptoKit)
import CryptoKit
#else
import Crypto
#endif

private let hexDigits: [Character] = Array("0123456789abcdef")

public func md5Hex(_ text: String) -> String {
    let digest = Insecure.MD5.hash(data: Data(text.utf8))
    var out = ""
    out.reserveCapacity(32)
    for byte in digest {
        out.append(hexDigits[Int(byte >> 4)])
        out.append(hexDigits[Int(byte & 0x0F)])
    }
    return out
}
