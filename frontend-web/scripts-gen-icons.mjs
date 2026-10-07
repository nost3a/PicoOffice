// one-off script: generate icon-192.png and icon-512.png
// pure Node builtins; draws blue rounded square + white P, original icon
import zlib from 'node:zlib'
import fs from 'node:fs'
import path from 'node:path'

//  CRC32 (PNG chunk needs it)
const CRC_TABLE = (() => {
  const t = new Uint32Array(256)
  for (let n = 0; n < 256; n++) {
    let c = n
    for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1
    t[n] = c >>> 0
  }
  return t
})()
function crc32(buf) {
  let c = 0xffffffff
  for (const b of buf) c = CRC_TABLE[(c ^ b) & 0xff] ^ (c >>> 8)
  return (c ^ 0xffffffff) >>> 0
}

function chunk(type, data) {
  const len = Buffer.alloc(4)
  len.writeUInt32BE(data.length, 0)
  const typeBuf = Buffer.from(type, 'ascii')
  const body = Buffer.concat([typeBuf, data])
  const crc = Buffer.alloc(4)
  crc.writeUInt32BE(crc32(body), 0)
  return Buffer.concat([len, body, crc])
}

// 7x9 P glyph, 1=white
const P = [
  '.XXXXX.',
  'X.....X',
  'X.....X',
  'X.....X',
  'X.....X',
  '.XXXXX.',
  'X......',
  'X......',
  'X......'
]

function makeIcon(size) {
  const px = Buffer.alloc(size * size * 4)
  const blue = [64, 158, 255] // #409eff
  const white = [255, 255, 255]
  const radius = Math.floor(size * 0.18)

  // glyph area: 7 cols x 9 rows, scaled and centered
  const gridW = 7
  const gridH = 9
  const cell = Math.floor(size * 0.42 / gridH) // use height as baseline
  const glyphW = cell * gridW
  const glyphH = cell * gridH
  const ox = Math.floor((size - glyphW) / 2)
  const oy = Math.floor((size - glyphH) / 2)

  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      const i = (y * size + x) * 4
      // round-corner test: outside corner rect is transparent
      let inside = true
      if (x < radius && y < radius) {
        inside = (x - radius) ** 2 + (y - radius) ** 2 <= radius ** 2
      } else if (x >= size - radius && y < radius) {
        inside = (x - (size - 1 - radius)) ** 2 + (y - radius) ** 2 <= radius ** 2
      } else if (x < radius && y >= size - radius) {
        inside = (x - radius) ** 2 + (y - (size - 1 - radius)) ** 2 <= radius ** 2
      } else if (x >= size - radius && y >= size - radius) {
        inside = (x - (size - 1 - radius)) ** 2 + (y - (size - 1 - radius)) ** 2 <= radius ** 2
      }
      if (!inside) {
        px[i + 3] = 0
        continue
      }
      // check whether inside P glyph
      const gx = Math.floor((x - ox) / cell)
      const gy = Math.floor((y - oy) / cell)
      let color = blue
      if (gx >= 0 && gx < gridW && gy >= 0 && gy < gridH && P[gy][gx] === 'X') {
        color = white
      }
      px[i] = color[0]
      px[i + 1] = color[1]
      px[i + 2] = color[2]
      px[i + 3] = 255
    }
  }

  // assemble PNG: prepend filter byte 0 per row
  const raw = Buffer.alloc((size * 4 + 1) * size)
  for (let y = 0; y < size; y++) {
    const rowStart = y * (size * 4 + 1)
    raw[rowStart] = 0 // filter: none
    px.copy(raw, rowStart + 1, y * size * 4, (y + 1) * size * 4)
  }
  const ihdr = Buffer.alloc(13)
  ihdr.writeUInt32BE(size, 0)
  ihdr.writeUInt32BE(size, 4)
  ihdr[8] = 8 // bit depth
  ihdr[9] = 6 // color type RGBA
  const idat = zlib.deflateSync(raw, { level: 9 })
  const sig = Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a])
  return Buffer.concat([
    sig,
    chunk('IHDR', ihdr),
    chunk('IDAT', idat),
    chunk('IEND', Buffer.alloc(0))
  ])
}

const outDir = path.resolve('public')
fs.writeFileSync(path.join(outDir, 'icon-192.png'), makeIcon(192))
fs.writeFileSync(path.join(outDir, 'icon-512.png'), makeIcon(512))
console.log('icons written:', fs.readdirSync(outDir).filter((f) => f.startsWith('icon')))
