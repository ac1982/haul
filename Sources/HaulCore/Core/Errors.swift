import Foundation

/// A user-facing failure. The message is what gets printed; the kind decides the exit code and the `error.kind`
/// of `--json` output, so scripts and agents can tell a bad link from a missing tool from a failed download.
public struct HaulError: Error, CustomStringConvertible, LocalizedError, Sendable {
    public enum Kind: String, Sendable, Codable {
        /// The download or extraction failed (network, site, muxing).
        case failed
        /// The input cannot be handled: an unsupported or malformed link, a bad option value.
        case input
        /// An external tool (ffmpeg, yt-dlp…) is missing.
        case dependency
        /// The content needs a login, or the stored login is no longer valid.
        case auth
        /// The person cancelled an interactive choice.
        case cancelled

        /// Exit codes: 1 failed, 2 input, 3 dependency, 4 auth, 130 cancelled. Bad command-line flags exit with 64.
        public var exitCode: Int32 {
            switch self {
            case .failed: return 1
            case .input: return 2
            case .dependency: return 3
            case .auth: return 4
            case .cancelled: return 130
            }
        }
    }

    public let kind: Kind
    public let message: String

    public init(_ message: String, kind: Kind = .failed) {
        self.kind = kind
        self.message = message
    }

    public var description: String { message }
    public var errorDescription: String? { message }

    public static func input(_ message: String) -> HaulError { HaulError(message, kind: .input) }
    public static func dependency(_ message: String) -> HaulError { HaulError(message, kind: .dependency) }
    public static func auth(_ message: String) -> HaulError { HaulError(message, kind: .auth) }
    public static let cancelled = HaulError("Cancelled", kind: .cancelled)

    /// A link no site recognises.
    public static func unsupported(_ input: String) -> HaulError {
        .input("Unsupported link: \(input)\nSupported sites: " + Site.allCases.map(\.name).joined(separator: ", ")
            + " (run `haul --help` for the link forms)")
    }
}

extension Error {
    /// The message we want a person to read, without the "The operation couldn't be completed" noise.
    public var readableMessage: String {
        if let e = self as? HaulError { return e.message }
        if let e = self as? JSONError { return e.description }
        if let e = self as? LocalizedError, let d = e.errorDescription { return d }
        return String(describing: self)
    }

    /// How the failure is reported: `HaulError`s carry their kind, everything else is a failed download.
    public var haulKind: HaulError.Kind { (self as? HaulError)?.kind ?? .failed }
}
