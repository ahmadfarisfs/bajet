#!/usr/bin/env node
// Generates solid-color PNG icons for the PWA using only Node.js built-ins.
const zlib = require('node:zlib')
const fs   = require('node:fs')
const path = require('node:path')

function crc32(buf) {
  const t = new Uint32Array(256)
  for (let i = 0; i < 256; i++) {
    let c = i
    for (let j = 0; j < 8; j++) c = (c & 1) ? (0xEDB88320 ^ (c >>> 1)) : (c >>> 1)
    t[i] = c
  }
  let crc = 0xFFFFFFFF
  for (const b of buf) crc = t[(crc ^ b) & 0xFF] ^ (crc >>> 8)
  return (crc ^ 0xFFFFFFFF) >>> 0
}

function chunk(type, data) {
  const tb = Buffer.from(type)
  const len = Buffer.alloc(4); len.writeUInt32BE(data.length)
  const crcInput = Buffer.concat([tb, data])
  const crcBuf = Buffer.alloc(4); crcBuf.writeUInt32BE(crc32(crcInput))
  return Buffer.concat([len, tb, data, crcBuf])
}

function solidPNG(size, r, g, b) {
  const sig = Buffer.from([137,80,78,71,13,10,26,10])

  const ihdr = Buffer.alloc(13)
  ihdr.writeUInt32BE(size, 0); ihdr.writeUInt32BE(size, 4)
  ihdr[8] = 8; ihdr[9] = 2 // 8-bit RGB

  const rowLen = size * 3 + 1
  const raw = Buffer.alloc(size * rowLen)
  for (let y = 0; y < size; y++) {
    raw[y * rowLen] = 0 // filter None
    for (let x = 0; x < size; x++) {
      raw[y * rowLen + x*3 + 1] = r
      raw[y * rowLen + x*3 + 2] = g
      raw[y * rowLen + x*3 + 3] = b
    }
  }

  return Buffer.concat([
    sig,
    chunk('IHDR', ihdr),
    chunk('IDAT', zlib.deflateSync(raw)),
    chunk('IEND', Buffer.alloc(0)),
  ])
}

const OUT = path.join(__dirname, 'public')
fs.mkdirSync(OUT, { recursive: true })

// Brand color: #4f46e5 (indigo)
const [R, G, B] = [0x4f, 0x46, 0xe5]

for (const size of [192, 512]) {
  const file = path.join(OUT, `pwa-${size}x${size}.png`)
  fs.writeFileSync(file, solidPNG(size, R, G, B))
  console.log(`✓ ${file}`)
}

// apple-touch-icon (180x180)
fs.writeFileSync(path.join(OUT, 'apple-touch-icon.png'), solidPNG(180, R, G, B))
console.log(`✓ apple-touch-icon.png`)
