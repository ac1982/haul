import CZlib
import Foundation

/// gzip (RFC 1952) via the system zlib. Used for gRPC message compression.
public enum Gzip {
    public static func compress(_ input: Data) throws -> Data {
        var stream = z_stream()
        var status = deflateInit2_(&stream, Z_DEFAULT_COMPRESSION, Z_DEFLATED, 15 + 16, 8, Z_DEFAULT_STRATEGY,
                                   ZLIB_VERSION, Int32(MemoryLayout<z_stream>.size))
        guard status == Z_OK else { throw HaulError("zlib deflateInit failed: \(status)") }
        defer { deflateEnd(&stream) }

        var output = Data()
        let chunk = 64 * 1024
        var buffer = [UInt8](repeating: 0, count: chunk)
        try input.withUnsafeBytes { (inBuf: UnsafeRawBufferPointer) in
            stream.next_in = inBuf.baseAddress.map { UnsafeMutablePointer(mutating: $0.assumingMemoryBound(to: Bytef.self)) }
            stream.avail_in = uInt(input.count)
            repeat {
                try buffer.withUnsafeMutableBufferPointer { out in
                    stream.next_out = out.baseAddress
                    stream.avail_out = uInt(chunk)
                    status = deflate(&stream, Z_FINISH)
                    guard status == Z_OK || status == Z_STREAM_END || status == Z_BUF_ERROR else {
                        throw HaulError("zlib deflate failed: \(status)")
                    }
                    output.append(out.baseAddress!, count: chunk - Int(stream.avail_out))
                }
            } while status != Z_STREAM_END
        }
        return output
    }

    public static func decompress(_ input: Data) throws -> Data {
        var stream = z_stream()
        var status = inflateInit2_(&stream, 15 + 32, ZLIB_VERSION, Int32(MemoryLayout<z_stream>.size))
        guard status == Z_OK else { throw HaulError("zlib inflateInit failed: \(status)") }
        defer { inflateEnd(&stream) }

        var output = Data()
        let chunk = 64 * 1024
        var buffer = [UInt8](repeating: 0, count: chunk)
        try input.withUnsafeBytes { (inBuf: UnsafeRawBufferPointer) in
            stream.next_in = inBuf.baseAddress.map { UnsafeMutablePointer(mutating: $0.assumingMemoryBound(to: Bytef.self)) }
            stream.avail_in = uInt(input.count)
            repeat {
                try buffer.withUnsafeMutableBufferPointer { out in
                    stream.next_out = out.baseAddress
                    stream.avail_out = uInt(chunk)
                    status = inflate(&stream, Z_NO_FLUSH)
                    guard status == Z_OK || status == Z_STREAM_END || status == Z_BUF_ERROR else {
                        throw HaulError("zlib inflate failed: \(status)")
                    }
                    output.append(out.baseAddress!, count: chunk - Int(stream.avail_out))
                }
                if status == Z_BUF_ERROR && stream.avail_in == 0 { break }
            } while status != Z_STREAM_END
        }
        return output
    }

    /// zlib-wrapped deflate (RFC 1950), as PNG IDAT wants it.
    public static func zlibCompress(_ input: Data) throws -> Data {
        var destLen = compressBound(uLong(input.count))
        var output = Data(count: Int(destLen))
        let status = output.withUnsafeMutableBytes { (outBuf: UnsafeMutableRawBufferPointer) -> Int32 in
            input.withUnsafeBytes { (inBuf: UnsafeRawBufferPointer) -> Int32 in
                compress2(outBuf.baseAddress!.assumingMemoryBound(to: Bytef.self), &destLen,
                          inBuf.baseAddress!.assumingMemoryBound(to: Bytef.self), uLong(input.count), Z_DEFAULT_COMPRESSION)
            }
        }
        guard status == Z_OK else { throw HaulError("zlib compress failed: \(status)") }
        output.count = Int(destLen)
        return output
    }

    public static func crc32(_ data: Data) -> UInt32 {
        data.withUnsafeBytes { buf in
            UInt32(CZlib.crc32(0, buf.baseAddress?.assumingMemoryBound(to: Bytef.self), uInt(data.count)))
        }
    }
}
