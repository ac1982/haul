import Foundation

/// gRPC length-prefixed message framing: 1 byte compressed flag + 4 byte big-endian length + payload.
public enum GRPCFrame {
    public static func pack(_ message: Data) throws -> Data {
        let payload = try Gzip.compress(message)
        var out = Data(capacity: 5 + payload.count)
        out.append(1)
        var len = UInt32(payload.count).bigEndian
        withUnsafeBytes(of: &len) { out.append(contentsOf: $0) }
        out.append(payload)
        return out
    }

    public static func unpack(_ data: Data) throws -> Data {
        guard data.count >= 5 else { throw HaulError("Incomplete gRPC response (\(data.count) bytes)") }
        let start = data.startIndex
        let flag = data[start]
        let len = data[(start + 1)..<(start + 5)].withUnsafeBytes { UInt32(bigEndian: $0.loadUnaligned(as: UInt32.self)) }
        let end = min(data.endIndex, start + 5 + Int(len))
        let body = data.subdata(in: (start + 5)..<end)
        return flag == 1 ? try Gzip.decompress(body) : body
    }
}
