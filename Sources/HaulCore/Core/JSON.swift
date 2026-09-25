import Foundation

public enum JSONError: Error, CustomStringConvertible, Sendable {
    case missingKey(String)
    case invalid(String)

    public var description: String {
        switch self {
        case .missingKey(let k): return "Missing JSON field: \(k)"
        case .invalid(let m): return "Invalid JSON: \(m)"
        }
    }
}

/// Dynamic JSON value. The bilibili responses are traversed ad hoc, so this is the natural shape.
public enum JSON: Sendable, Equatable {
    case object([String: JSON])
    case array([JSON])
    case string(String)
    case int(Int64)
    case double(Double)
    case bool(Bool)
    case null

    public static func parse(_ data: Data) throws -> JSON {
        do { return try JSONDecoder().decode(JSON.self, from: data) }
        catch { throw JSONError.invalid("\(error)") }
    }

    public static func parse(_ text: String) throws -> JSON { try parse(Data(text.utf8)) }

    // MARK: traversal

    public subscript(key: String) -> JSON {
        if case .object(let o) = self { return o[key] ?? .null }
        return .null
    }

    public subscript(index: Int) -> JSON {
        if case .array(let a) = self, a.indices.contains(index) { return a[index] }
        return .null
    }

    /// Like subscript, but a missing key is an error. Use for fields the response must have.
    public func get(_ key: String) throws -> JSON {
        guard case .object(let o) = self, let v = o[key] else { throw JSONError.missingKey(key) }
        return v
    }

    public func has(_ key: String) -> Bool {
        if case .object(let o) = self { return o[key] != nil }
        return false
    }

    public var isNull: Bool { if case .null = self { return true }; return false }
    public var exists: Bool { !isNull }
    public var isObject: Bool { if case .object = self { return true }; return false }
    public var isArray: Bool { if case .array = self { return true }; return false }

    // MARK: scalars

    public var string: String? {
        if case .string(let s) = self { return s }
        return nil
    }

    /// Text form of any value: scalars as text, containers serialized. Matches how the responses are string-searched.
    public var stringValue: String {
        switch self {
        case .string(let s): return s
        case .int(let i): return String(i)
        case .double(let d): return d == d.rounded() && abs(d) < 1e15 ? String(Int64(d)) : String(d)
        case .bool(let b): return b ? "true" : "false"
        case .null: return ""
        case .object, .array: return serialized
        }
    }

    public var int64: Int64? {
        switch self {
        case .int(let i): return i
        case .double(let d): return d == d.rounded() && abs(d) < 9.2e18 ? Int64(d) : nil
        case .string(let s): return Int64(s) ?? Double(s).flatMap { $0 == $0.rounded() ? Int64($0) : nil }
        case .bool(let b): return b ? 1 : 0
        default: return nil
        }
    }

    public var int: Int? { int64.map { Int($0) } }

    public var double: Double? {
        switch self {
        case .int(let i): return Double(i)
        case .double(let d): return d
        case .string(let s): return Double(s)
        default: return nil
        }
    }

    public var bool: Bool? {
        switch self {
        case .bool(let b): return b
        case .int(let i): return i != 0
        case .string(let s): return s == "true" ? true : s == "false" ? false : nil
        default: return nil
        }
    }

    public var array: [JSON] {
        if case .array(let a) = self { return a }
        return []
    }

    public var object: [String: JSON] {
        if case .object(let o) = self { return o }
        return [:]
    }

    public var serialized: String {
        let enc = JSONEncoder()
        enc.outputFormatting = [.withoutEscapingSlashes]
        guard let data = try? enc.encode(self) else { return "" }
        return String(decoding: data, as: UTF8.self)
    }
}

extension JSON: Codable {
    public init(from decoder: Decoder) throws {
        let c = try decoder.singleValueContainer()
        if c.decodeNil() { self = .null; return }
        if let b = try? c.decode(Bool.self) { self = .bool(b); return }
        if let i = try? c.decode(Int64.self) { self = .int(i); return }
        if let d = try? c.decode(Double.self) { self = .double(d); return }
        if let s = try? c.decode(String.self) { self = .string(s); return }
        if let a = try? c.decode([JSON].self) { self = .array(a); return }
        if let o = try? c.decode([String: JSON].self) { self = .object(o); return }
        throw DecodingError.dataCorruptedError(in: c, debugDescription: "Unsupported JSON value")
    }

    public func encode(to encoder: Encoder) throws {
        var c = encoder.singleValueContainer()
        switch self {
        case .object(let o): try c.encode(o)
        case .array(let a): try c.encode(a)
        case .string(let s): try c.encode(s)
        case .int(let i): try c.encode(i)
        case .double(let d): try c.encode(d)
        case .bool(let b): try c.encode(b)
        case .null: try c.encodeNil()
        }
    }
}
