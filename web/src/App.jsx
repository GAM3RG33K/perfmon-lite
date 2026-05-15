import { useState, useEffect, useRef } from 'react'

const FEATURES = [
  { icon: '⚡', title: 'Instant Start', desc: 'Boot-to-profiling in under a second. No heavy IDEs.' },
  { icon: '📱', title: 'Android & iOS', desc: 'One interface for both platforms. Auto-detect devices.' },
  { icon: '📊', title: 'Live Charts', desc: 'Real-time CPU, memory, and thread sparkline charts.' },
  { icon: '📦', title: 'Export Anywhere', desc: 'Export to JSON, Markdown, or HTML. Perfect for CI/CD.' },
  { icon: '🔋', title: 'Zero Bloat', desc: 'Single 5.5MB binary. No runtime dependencies.' },
  { icon: '🎯', title: 'Target by App', desc: 'Profile specific apps with --id flag.' },
]

const COMMANDS = [
  ['perfmon --mock', 'Try with simulated data'],
  ['perfmon', 'Auto-detect and profile'],
  ['perfmon --id com.example.app', 'Target a specific app'],
  ['perfmon devices', 'List connected devices'],
  ['perfmon update', 'Self-update to latest'],
]

const TYPING_LINES = [
  { text: 'perfmon v1.0.1 — Mobile Performance Monitor', cls: 'dim' },
  { text: '', cls: '' },
  { text: '┌──────────────────────────────────────────────────┐', cls: 'dim' },
  { text: '│  Target: Pixel 8  │  App: com.example.app  [DEBUG]  │', cls: 'dim' },
  { text: '├──────────────────────────────────────────────────┤', cls: 'dim' },
  { text: '│  CPU Utilization (%)                             │', cls: 'cyan' },
  { text: '│  100 ┤      ╭╮                                    │', cls: 'dim' },
  { text: '│   50 ┤  ╭──╯╰─╮╭──╮                              │', cls: 'dim' },
  { text: '│    0 └─╯     ╰╯  ╰────────────────────────────  │', cls: 'dim' },
  { text: '│  Memory Footprint (MB)                            │', cls: 'magenta' },
  { text: '│  210 ┤      ╭───────────────────────────────      │', cls: 'dim' },
  { text: '│    0 └──╯                                         │', cls: 'dim' },
  { text: '│  Peak CPU: 78%  │  Peak RAM: 215 MB               │', cls: 'amber' },
  { text: '├──────────────────────────────────────────────────┤', cls: 'dim' },
  { text: '│  [↑/↓] Navigate  [TAB] Switch  [e] Export  [?] Help  │', cls: 'dim' },
  { text: '└──────────────────────────────────────────────────┘', cls: 'dim' },
]

function useTypewriter(lines, speed = 20) {
  const [visible, setVisible] = useState([])
  const idxRef = useRef(0)
  const charRef = useRef(0)

  useEffect(() => {
    const t = setInterval(() => {
      if (idxRef.current >= lines.length) { clearInterval(t); return }
      const line = lines[idxRef.current].text
      if (charRef.current <= line.length) {
        setVisible(prev => {
          const copy = [...prev]
          if (!copy[idxRef.current]) copy[idxRef.current] = ''
          copy[idxRef.current] = line.slice(0, charRef.current)
          return copy
        })
        charRef.current++
      } else {
        idxRef.current++
        charRef.current = 0
      }
    }, speed)
    return () => clearInterval(t)
  }, [lines, speed])

  return visible
}

function Sparkline({ data, color, height = 40 }) {
  if (!data.length) return null
  const max = Math.max(...data, 1)
  const w = data.length * 4
  const pts = data.map((v, i) => `${i * 4},${height - (v / max) * height}`).join(' ')
  return (
    <svg width={w} height={height} style={{ display: 'block' }}>
      <polyline points={pts} fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function AnimatedCPU() {
  const [vals, setVals] = useState(() => Array.from({ length: 30 }, () => Math.random() * 60 + 10))
  useEffect(() => {
    const t = setInterval(() => {
      setVals(prev => [...prev.slice(1), Math.random() * 60 + 10])
    }, 300)
    return () => clearInterval(t)
  }, [])
  return (
    <div className="live-chart">
      <Sparkline data={vals} color="#00ffff" />
      <span className="chart-label" style={{ color: '#00ffff' }}>CPU</span>
      <span className="chart-value">{Math.round(vals[vals.length - 1])}%</span>
    </div>
  )
}

function AnimatedMemory() {
  const [vals, setVals] = useState(() => Array.from({ length: 30 }, () => Math.random() * 100 + 100))
  useEffect(() => {
    const t = setInterval(() => {
      setVals(prev => [...prev.slice(1), Math.random() * 100 + 100])
    }, 400)
    return () => clearInterval(t)
  }, [])
  return (
    <div className="live-chart">
      <Sparkline data={vals} color="#ff00ff" />
      <span className="chart-label" style={{ color: '#ff00ff' }}>RAM</span>
      <span className="chart-value">{Math.round(vals[vals.length - 1])} MB</span>
    </div>
  )
}

function ScrollReveal({ children, delay = 0 }) {
  const ref = useRef(null)
  const [visible, setVisible] = useState(false)
  useEffect(() => {
    const el = ref.current
    if (!el) return
    const o = new IntersectionObserver(([e]) => { if (e.isIntersecting) { setTimeout(() => setVisible(true), delay); o.disconnect() } })
    o.observe(el)
    return () => o.disconnect()
  }, [delay])
  return <div ref={ref} className={`reveal ${visible ? 'revealed' : ''}`}>{children}</div>
}

export default function App() {
  const lines = useTypewriter(TYPING_LINES, 18)

  return (
    <div className="app">
      <div className="bg-grid" />
      <div className="bg-glow" />

      <header className="hero">
        <div className="hero-badge">v1.0.1 &nbsp;·&nbsp; Open Source</div>
        <h1>perfmon <span className="accent">⎈</span></h1>
        <p className="hero-desc">Blistering-fast, terminal-based mobile app profiling. CPU, memory, and thread telemetry for Android and iOS — right in your terminal.</p>
        <div className="cta-row">
          <a href="https://get.perfmon.qzz.io" className="btn-primary">Install</a>
          <a href="https://github.com/GAM3RG33K/perfmon-lite" className="btn-secondary">GitHub</a>
        </div>
      </header>

      <section className="terminal-section">
        <div className="terminal">
          <div className="terminal-header">
            <span className="dot"></span><span className="dot"></span><span className="dot"></span>
            <span className="terminal-title">perfmon — live</span>
          </div>
          <div className="terminal-body">
            {lines.map((l, i) => (
              <div key={i} className={`line ${TYPING_LINES[i]?.cls || ''}`}>{l}{i < lines.length - 1 ? '' : <span className="cursor">▌</span>}</div>
            ))}
          </div>
        </div>
      </section>

      <section className="live-section">
        <ScrollReveal>
          <div className="live-panel">
            <AnimatedCPU />
            <AnimatedMemory />
          </div>
        </ScrollReveal>
      </section>

      <section className="features-section">
        <ScrollReveal><h2 className="section-title">Why perfmon?</h2></ScrollReveal>
        <div className="features-grid">
          {FEATURES.map((f, i) => (
            <ScrollReveal key={f.title} delay={i * 80}>
              <div className="card">
                <span className="card-icon">{f.icon}</span>
                <h3>{f.title}</h3>
                <p>{f.desc}</p>
              </div>
            </ScrollReveal>
          ))}
        </div>
      </section>

      <section className="install-section">
        <ScrollReveal>
          <div className="install-box">
            <h2>Install in seconds</h2>
            <div className="code-block">
              <span className="comment"># macOS / Linux</span>
              <br /><span className="prompt">$</span> curl -sfL https://get.perfmon.qzz.io | bash
            </div>
            <p className="install-note">Windows: <code>iwr https://get.perfmon.qzz.io/windows -useb | iex</code></p>
          </div>
        </ScrollReveal>
      </section>

      <section className="commands-section">
        <ScrollReveal><h2 className="section-title">Quick Start</h2></ScrollReveal>
        <table className="cmd-table">
          <thead><tr><th>Command</th><th>What it does</th></tr></thead>
          <tbody>
            {COMMANDS.map(([cmd, desc]) => (
              <tr key={cmd}><td><code>{cmd}</code></td><td>{desc}</td></tr>
            ))}
          </tbody>
        </table>
      </section>

      <footer className="footer">
        <p>Built with <a href="https://go.dev">Go</a>, <a href="https://github.com/charmbracelet/bubbletea">Bubble Tea</a>, and <a href="https://github.com/charmbracelet/lipgloss">Lipgloss</a>.</p>
        <p><a href="https://github.com/GAM3RG33K/perfmon-lite">GitHub</a> · MIT License</p>
      </footer>
    </div>
  )
}
