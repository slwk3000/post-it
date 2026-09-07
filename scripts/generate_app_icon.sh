#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
TMP_DIR="$(mktemp -d /tmp/postit-icon.XXXXXX)"

trap 'rm -rf "${TMP_DIR}"' EXIT

cat << 'SWIFT' > "${TMP_DIR}/render.swift"
import Cocoa

let size = NSSize(width: 1024, height: 1024)
let image = NSImage(size: size)

image.lockFocus()
guard let context = NSGraphicsContext.current?.cgContext else {
    exit(1)
}

context.clear(CGRect(origin: .zero, size: size))

context.saveGState()
let shadowColor = NSColor(red: 0, green: 0, blue: 0, alpha: 0.22).cgColor
context.setShadow(offset: CGSize(width: 0, height: -24), blur: 36, color: shadowColor)

let paperRect = NSRect(x: 112, y: 112, width: 800, height: 800)
let cornerRadius: CGFloat = 40.0
let paperPath = NSBezierPath(roundedRect: paperRect, xRadius: cornerRadius, yRadius: cornerRadius)

NSColor(red: 0.988, green: 0.961, blue: 0.898, alpha: 1.0).setFill()
paperPath.fill()
context.restoreGState()

NSColor(red: 0.88, green: 0.84, blue: 0.74, alpha: 0.7).setStroke()
paperPath.lineWidth = 4
paperPath.stroke()

let curlPath = NSBezierPath()
curlPath.move(to: NSPoint(x: 832, y: 112))
curlPath.curve(to: NSPoint(x: 912, y: 192), controlPoint1: NSPoint(x: 872, y: 112), controlPoint2: NSPoint(x: 912, y: 152))
curlPath.line(to: NSPoint(x: 872, y: 152))
curlPath.close()
NSColor(red: 0.94, green: 0.90, blue: 0.80, alpha: 0.5).setFill()
curlPath.fill()

context.saveGState()
let tapeRect = NSRect(x: 372, y: 864, width: 280, height: 84)
let tapePath = NSBezierPath(roundedRect: tapeRect, xRadius: 8, yRadius: 8)
NSColor(red: 0.86, green: 0.78, blue: 0.62, alpha: 0.75).setFill()
tapePath.fill()
NSColor(red: 0.68, green: 0.60, blue: 0.44, alpha: 0.5).setStroke()
tapePath.lineWidth = 3
tapePath.stroke()
context.restoreGState()

let penColor = NSColor(red: 0.114, green: 0.306, blue: 0.847, alpha: 0.88)
penColor.setStroke()

func drawHandLine(y: CGFloat, width: CGFloat) {
    let path = NSBezierPath()
    path.lineWidth = 14
    path.lineCapStyle = .round
    path.move(to: NSPoint(x: 232, y: y))
    path.curve(to: NSPoint(x: 232 + width, y: y + 3),
               controlPoint1: NSPoint(x: 232 + width * 0.35, y: y - 5),
               controlPoint2: NSPoint(x: 232 + width * 0.70, y: y + 6))
    path.stroke()
}

drawHandLine(y: 650, width: 560)
drawHandLine(y: 510, width: 520)
drawHandLine(y: 370, width: 440)
drawHandLine(y: 230, width: 260)

image.unlockFocus()

guard let tiffData = image.tiffRepresentation,
      let bitmap = NSBitmapImageRep(data: tiffData),
      let pngData = bitmap.representation(using: .png, properties: [:]) else {
    exit(1)
}

let outPath = CommandLine.arguments.count > 1 ? CommandLine.arguments[1] : "/tmp/postit_icon_1024.png"
try? pngData.write(to: URL(fileURLWithPath: outPath))
SWIFT

swift "${TMP_DIR}/render.swift" "${TMP_DIR}/icon_1024.png"

mkdir -p "${TMP_DIR}/AppIcon.iconset"
sips -z 16 16     "${TMP_DIR}/icon_1024.png" --out "${TMP_DIR}/AppIcon.iconset/icon_16x16.png" >/dev/null
sips -z 32 32     "${TMP_DIR}/icon_1024.png" --out "${TMP_DIR}/AppIcon.iconset/icon_16x16@2x.png" >/dev/null
sips -z 32 32     "${TMP_DIR}/icon_1024.png" --out "${TMP_DIR}/AppIcon.iconset/icon_32x32.png" >/dev/null
sips -z 64 64     "${TMP_DIR}/icon_1024.png" --out "${TMP_DIR}/AppIcon.iconset/icon_32x32@2x.png" >/dev/null
sips -z 128 128   "${TMP_DIR}/icon_1024.png" --out "${TMP_DIR}/AppIcon.iconset/icon_128x128.png" >/dev/null
sips -z 256 256   "${TMP_DIR}/icon_1024.png" --out "${TMP_DIR}/AppIcon.iconset/icon_128x128@2x.png" >/dev/null
sips -z 256 256   "${TMP_DIR}/icon_1024.png" --out "${TMP_DIR}/AppIcon.iconset/icon_256x256.png" >/dev/null
sips -z 512 512   "${TMP_DIR}/icon_1024.png" --out "${TMP_DIR}/AppIcon.iconset/icon_256x256@2x.png" >/dev/null
sips -z 512 512   "${TMP_DIR}/icon_1024.png" --out "${TMP_DIR}/AppIcon.iconset/icon_512x512.png" >/dev/null
sips -z 1024 1024 "${TMP_DIR}/icon_1024.png" --out "${TMP_DIR}/AppIcon.iconset/icon_512x512@2x.png" >/dev/null

mkdir -p "${ROOT_DIR}/assets/icons"
iconutil -c icns "${TMP_DIR}/AppIcon.iconset" -o "${ROOT_DIR}/assets/icons/AppIcon.icns"
cp "${TMP_DIR}/icon_1024.png" "${ROOT_DIR}/assets/icons/AppIcon.png"

echo "Generated AppIcon.icns in ${ROOT_DIR}/assets/icons"
