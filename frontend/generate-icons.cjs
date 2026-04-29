#!/usr/bin/env node
// Generates PWA icon PNGs using only Node.js built-ins.
// Renders a bar-chart design on an indigo background.
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
  const tb  = Buffer.from(type)
  const len = Buffer.alloc(4); len.writeUInt32BE(data.length)
  const crcBuf = Buffer.alloc(4); crcBuf.writeUInt32BE(crc32(Buffer.concat([tb, data])))
  return Buffer.concat([len, tb, data, crcBuf])
}

function encodePNG(size, pixels) {
  const sig  = Buffer.from([137,80,78,71,13,10,26,10])
  const ihdr = Buffer.alloc(13)
  ihdr.writeUInt32BE(size, 0); ihdr.writeUInt32BE(size, 4)
  ihdr[8] = 8; ihdr[9] = 2 // 8-bit RGB

  const rowLen = size * 3 + 1
  const raw = Buffer.alloc(size * rowLen)
  for (let y = 0; y < size; y++) {
    raw[y * rowLen] = 0
    for (let x = 0; x < size; x++) {
      const pi = (y * size + x) * 3
      raw[y * rowLen + x*3 + 1] = pixels[pi]
      raw[y * rowLen + x*3 + 2] = pixels[pi + 1]
      raw[y * rowLen + x*3 + 3] = pixels[pi + 2]
    }
  }
  return Buffer.concat([sig, chunk('IHDR', ihdr), chunk('IDAT', zlib.deflateSync(raw)), chunk('IEND', Buffer.alloc(0))])
}

function blend(bg, fg, alpha) {
  return Math.round(bg * (1 - alpha) + fg * alpha)
}

function drawRect(pixels, size, x, y, w, h, r, g, b, opacity = 1) {
  for (let py = Math.max(0, y); py < Math.min(size, y + h); py++) {
    for (let px = Math.max(0, x); px < Math.min(size, x + w); px++) {
      const i = (py * size + px) * 3
      pixels[i]   = blend(pixels[i],   r, opacity)
      pixels[i+1] = blend(pixels[i+1], g, opacity)
      pixels[i+2] = blend(pixels[i+2], b, opacity)
    }
  }
}

function makeIcon(size) {
  const pixels = new Uint8Array(size * size * 3)
  const s = size / 512

  // Background: #4f46e5 (indigo)
  drawRect(pixels, size, 0, 0, size, size, 0x4f, 0x46, 0xe5)

  // Bar chart (coordinates designed for 512×512, scaled by s)
  const bars = [
    { x: 88,  y: 210, w: 96, h: 215, op: 0.65 },
    { x: 208, y: 118, w: 96, h: 307, op: 1.00 },
    { x: 328, y: 158, w: 96, h: 267, op: 0.85 },
  ]
  for (const b of bars) {
    drawRect(pixels, size,
      Math.round(b.x*s), Math.round(b.y*s),
      Math.round(b.w*s), Math.round(b.h*s),
      255, 255, 255, b.op)
  }

  // Base line
  drawRect(pixels, size,
    Math.round(68*s), Math.round(430*s),
    Math.round(376*s), Math.max(1, Math.round(10*s)),
    255, 255, 255, 0.45)

  return pixels
}

const OUT = path.join(__dirname, 'public')
fs.mkdirSync(OUT, { recursive: true })

for (const [size, name] of [[180, 'apple-touch-icon.png'], [192, 'pwa-192x192.png'], [512, 'pwa-512x512.png']]) {
  const file = path.join(OUT, name)
  fs.writeFileSync(file, encodePNG(size, makeIcon(size)))
  console.log(`✓ ${name}`)
}
