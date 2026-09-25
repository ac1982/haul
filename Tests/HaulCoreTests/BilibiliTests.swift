import Foundation
import Testing
@testable import HaulCore

@Suite struct BVConverterTests {
    @Test func knownPairs() throws {
        #expect(try BVConverter.encode(2) == "BV1xx411c7mD")
        #expect(try BVConverter.decode("xx411c7mD") == 2)
        #expect(try BVConverter.encode(170001) == "BV17x411w7KC")
        #expect(try BVConverter.decode("7x411w7KC") == 170001)
    }

    @Test func roundTrip() throws {
        for aid: Int64 in [1, 12345, 999_999_999, 1_000_000_000_000] {
            let bv = try BVConverter.encode(aid)
            #expect(bv.hasPrefix("BV1"))
            #expect(try BVConverter.decode(String(bv.dropFirst(3))) == aid)
        }
    }
}

@Suite struct SigningTests {
    @Test func wbiMixinKeyMatchesDocumentedExample() {
        let key = Signing.wbiMixinKey(imgKey: "7cd084941338484aae1ad9425b84077c", subKey: "4932caff0ff746eab6f01bf08b70ac45")
        #expect(key == "ea1db124af3c7062474693fa704f4ff8")
    }

    @Test func md5() {
        #expect(md5Hex("") == "d41d8cd98f00b204e9800998ecf8427e")
        #expect(md5Hex("abc") == "900150983cd24fb0d6963f7d28e17f72")
    }

    @Test func imageKey() {
        #expect(Signing.imageKey(from: "https://i0.hdslb.com/bfs/wbi/7cd084941338484aae1ad9425b84077c.png") == "7cd084941338484aae1ad9425b84077c")
    }
}
