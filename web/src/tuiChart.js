const BLOCK_UP = [
  ' ', '▗', '▗', '▐', '▐',
  '▖', '▄', '▄', '▟', '▟',
  '▖', '▄', '▄', '▟', '▟',
  '▌', '▙', '▙', '█', '█',
  '▌', '▙', '▙', '█', '█',
]

function graphLevel(value, curHigh, curLow, mod) {
  if (value >= curHigh) return 4
  if (value <= curLow) return 0
  if (curHigh === curLow) return 0
  let lv = Math.round(((value - curLow) * 4) / (curHigh - curLow) + mod)
  return Math.max(0, Math.min(4, lv))
}

export function miniGauge(pct, width = 12) {
  const filled = Math.round((pct / 100) * width)
  return '█'.repeat(filled) + '░'.repeat(width - filled)
}

export function btopGraphRows(data, width, height = 8) {
  const padded = Array(width).fill(0)
  const start = Math.max(0, data.length - width)
  for (let i = 0; i < data.length - start; i++) {
    padded[width - (data.length - start) + i] = data[start + i]
  }

  const rows = Array.from({ length: height }, () => Array(width).fill(' '))
  const mod = height === 1 ? 0.3 : 0.1
  let last = 0

  for (let x = 0; x < width; x++) {
    const cur = padded[x]
    for (let horizon = 0; horizon < height; horizon++) {
      let curHigh = 100
      let curLow = 0
      if (height > 1) {
        curHigh = Math.round((100 * (height - horizon)) / height)
        curLow = Math.round((100 * (height - (horizon + 1))) / height)
      }
      const prevLv = graphLevel(last, curHigh, curLow, mod)
      const curLv = graphLevel(cur, curHigh, curLow, mod)
      rows[horizon][x] = prevLv + curLv === 0 ? ' ' : BLOCK_UP[prevLv * 5 + curLv]
    }
    last = cur
  }

  return rows.map((row, horizon) =>
    row.map((ch) => {
      if (ch !== ' ') return ch
      if (height > 1 && (horizon === Math.floor(height / 4) || horizon === Math.floor(height / 2) || horizon === Math.floor((3 * height) / 4))) {
        return '·'
      }
      return ' '
    }).join('')
  )
}

export function renderAreaBlock(data, width, height, labelW = 6, maxLabel = '100') {
  const rows = btopGraphRows(data, width, height)
  const lines = []
  for (let h = 0; h < height; h++) {
    let axis = '      '
    if (h === 0) axis = String(maxLabel).padStart(5) + ' │'
    else if (h === Math.floor(height / 2)) axis = String(Math.round(parseFloat(maxLabel) / 2)).padStart(5) + ' │'
    else axis = '      │'
    lines.push(axis + rows[h])
  }
  lines.push('      └' + '─'.repeat(width))
  const left = '100s ago'
  const gap = Math.max(1, width - left.length - 3)
  lines.push('      ' + left + ' '.repeat(gap) + 'now')
  return lines
}
