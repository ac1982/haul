// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "haul",
    platforms: [.macOS(.v15)],
    products: [
        .executable(name: "haul", targets: ["haul"]),
        .library(name: "HaulCore", targets: ["HaulCore"]),
    ],
    dependencies: [
        .package(url: "https://github.com/apple/swift-argument-parser.git", from: "1.5.0"),
        .package(url: "https://github.com/apple/swift-protobuf.git", from: "1.28.0"),
        .package(url: "https://github.com/apple/swift-crypto.git", from: "3.0.0"),
        .package(url: "https://github.com/dagronf/swift-qrcode-generator.git", from: "2.0.0"),
    ],
    targets: [
        .systemLibrary(
            name: "CZlib",
            path: "Sources/CZlib",
            providers: [.apt(["zlib1g-dev"]), .brew(["zlib"])]
        ),
        .target(
            name: "HaulCore",
            dependencies: [
                "CZlib",
                .product(name: "SwiftProtobuf", package: "swift-protobuf"),
                .product(name: "Crypto", package: "swift-crypto", condition: .when(platforms: [.linux])),
                .product(name: "QRCodeGenerator", package: "swift-qrcode-generator"),
            ],
            exclude: ["Sites/Bilibili/Proto"],
            swiftSettings: [.swiftLanguageMode(.v6)]
        ),
        .executableTarget(
            name: "haul",
            dependencies: [
                "HaulCore",
                .product(name: "ArgumentParser", package: "swift-argument-parser"),
            ],
            swiftSettings: [.swiftLanguageMode(.v6)]
        ),
        .testTarget(
            name: "HaulCoreTests",
            dependencies: ["HaulCore"],
            resources: [.copy("Fixtures")],
            swiftSettings: [.swiftLanguageMode(.v6)]
        ),
    ]
)
