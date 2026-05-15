import { useState, useEffect, useRef, useCallback } from 'react'

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
  { text: `perfmon v${import.meta.env.VITE_APP_VERSION || 'dev'} — Mobile Performance Monitor`, cls: 'dim' },
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

/* ─── Hooks ─────────────────────────────────────────────── */

function useTypewriter(lines, speed = 18) {
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
      } else { idxRef.current++; charRef.current = 0 }
    }, speed)
    return () => clearInterval(t)
  }, [lines, speed])
  return visible
}

function useScrollY() {
  const [y, setY] = useState(0)
  useEffect(() => {
    const onScroll = () => setY(window.scrollY)
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])
  return y
}

function useMousePos() {
  const [pos, setPos] = useState({ x: 0.5, y: 0.5 })
  const onMove = useCallback((e) => {
    setPos({ x: e.clientX / window.innerWidth, y: e.clientY / window.innerHeight })
  }, [])
  useEffect(() => {
    window.addEventListener('mousemove', onMove)
    return () => window.removeEventListener('mousemove', onMove)
  }, [onMove])
  return pos
}

/* ─── Components ────────────────────────────────────────── */

function ParallaxSection({ children, speed = 0.3, className = '' }) {
  const scrollY = useScrollY()
  const ref = useRef(null)
  const [top, setTop] = useState(0)
  useEffect(() => {
    const el = ref.current
    if (!el) return
    const o = new IntersectionObserver(([e]) => { if (e.isIntersecting) setTop(e.boundingClientRect.top + window.scrollY) })
    o.observe(el)
    return () => o.disconnect()
  }, [])
  const offset = top ? (scrollY - top) * speed : 0
  return (
    <div ref={ref} className={className} style={{ transform: `translateY(${offset}px)` }}>
      {children}
    </div>
  )
}

function Floaters() {
  const scrollY = useScrollY()
  const mouse = useMousePos()
  const symbols = ['</>', '{ }', '()', '[]', '#!', '=>', '/*', '*/', '&&', '||']
  return (
    <div className="floaters" aria-hidden="true">
      {symbols.map((s, i) => {
        const baseX = (i % 5) * 20 + 5
        const baseY = i * 15 + 10
        const driftX = Math.sin(scrollY * 0.001 + i) * 20 + (mouse.x - 0.5) * 40 * (1 + (i % 3) * 0.5)
        const driftY = Math.cos(scrollY * 0.0008 + i * 1.5) * 15
        const opacity = 0.06 + (i % 3) * 0.03
        return (
          <span
            key={i}
            className="floater"
            style={{
              left: `${baseX}%`,
              top: `${baseY}%`,
              transform: `translate(${driftX}px, ${driftY}px)`,
              opacity,
              fontSize: `${12 + (i % 4) * 4}px`,
              animationDelay: `${i * 0.3}s`,
            }}
          >
            {s}
          </span>
        )
      })}
    </div>
  )
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
  useEffect(() => { const t = setInterval(() => setVals(prev => [...prev.slice(1), Math.random() * 60 + 10]), 300); return () => clearInterval(t) }, [])
  return (
    <div className="live-chart" data-label="CPU">
      <Sparkline data={vals} color="#00ffff" />
      <span className="chart-label" style={{ color: '#00ffff' }}>CPU</span>
      <span className="chart-value">{Math.round(vals[vals.length - 1])}%</span>
    </div>
  )
}

function AnimatedMemory() {
  const [vals, setVals] = useState(() => Array.from({ length: 30 }, () => Math.random() * 100 + 100))
  useEffect(() => { const t = setInterval(() => setVals(prev => [...prev.slice(1), Math.random() * 100 + 100]), 400); return () => clearInterval(t) }, [])
  return (
    <div className="live-chart" data-label="RAM">
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
    const o = new IntersectionObserver(([e]) => { if (e.isIntersecting) { setTimeout(() => setVisible(true), delay); o.disconnect() } }, { threshold: 0.1 })
    o.observe(el)
    return () => o.disconnect()
  }, [delay])
  return <div ref={ref} className={`reveal ${visible ? 'revealed' : ''}`}>{children}</div>
}

/* ─── App ────────────────────────────────────────────────── */

function CopyButton({ text }) {
  const [copied, setCopied] = useState(false)
  return (
    <button
      className={`copy-btn ${copied ? 'copied' : ''}`}
      onClick={() => { navigator.clipboard.writeText(text); setCopied(true); setTimeout(() => setCopied(false), 2000) }}
      title="Copy command"
    >
      {copied ? '✓' : '⎘'}
    </button>
  )
}

export default function App() {
  const lines = useTypewriter(TYPING_LINES, 18)
  const mouse = useMousePos()
  const [installTab, setInstallTab] = useState('unix')

  return (
    <div className="app">
      <div className="parallax-bg" style={{ transform: `translateY(${useScrollY() * 0.15}px)` }} />
      <div className="bg-grid" />
      <div className="bg-glow" style={{ transform: `translate(${(mouse.x - 0.5) * 20}px, ${(mouse.y - 0.5) * 20}px)` }} />
      <Floaters />

      <ParallaxSection speed={-0.08}>
        <header className="hero">
          <div className="hero-badge">v{import.meta.env.VITE_APP_VERSION || 'dev'} &nbsp;·&nbsp; Beta</div>
          <h1>perfmon <span className="accent">⎈</span></h1>
          <p className="hero-desc">Blistering-fast, terminal-based mobile app profiling. CPU, memory, and thread telemetry for Android and iOS — right in your terminal.</p>
          <div className="cta-row">
            <a href="https://get.perfmon.qzz.io" className="btn-primary">Install</a>
            <a href="https://github.com/GAM3RG33K/perfmon-lite" className="btn-secondary">GitHub</a>
          </div>
        </header>
      </ParallaxSection>

      <div className="terminal-wrapper" style={{ perspective: '800px' }}>
        <ParallaxSection speed={0.06}>
          <section className="terminal-section">
            <div
              className="terminal"
              style={{
                transform: `rotateX(${(mouse.y - 0.5) * 3}deg) rotateY(${(mouse.x - 0.5) * 3}deg)`,
                transition: 'transform 0.1s ease-out',
              }}
            >
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
        </ParallaxSection>
      </div>

      <ParallaxSection speed={-0.04}>
        <section className="live-section">
          <ScrollReveal>
            <div className="live-panel">
              <AnimatedCPU />
              <AnimatedMemory />
            </div>
          </ScrollReveal>
        </section>
      </ParallaxSection>

      <section className="features-section">
        <ScrollReveal><h2 className="section-title">Why perfmon?</h2></ScrollReveal>
        <div className="features-grid">
          {FEATURES.map((f, i) => (
            <ScrollReveal key={f.title} delay={i * 80}>
              <div className="card" style={{ transform: `translateY(${Math.sin(i * 1.2) * 4}px)` }}>
                <span className="card-icon">{f.icon}</span>
                <h3>{f.title}</h3>
                <p>{f.desc}</p>
              </div>
            </ScrollReveal>
          ))}
        </div>
      </section>

      <ParallaxSection speed={0.05}>
        <section className="install-section">
          <ScrollReveal>
            <div className="install-box">
              <h2>Install in seconds</h2>
              <div className="install-platform-tabs">
                <span className={`install-tab ${installTab === 'unix' ? 'active' : ''}`} onClick={() => setInstallTab('unix')}>macOS / Linux</span>
                <span className={`install-tab ${installTab === 'win' ? 'active' : ''}`} onClick={() => setInstallTab('win')}>Windows</span>
              </div>
              <div className="code-wrap" style={{
                transform: `perspective(400px) rotateX(${(mouse.y - 0.5) * 2}deg)`,
              }}>
                <div className="code-block">
                  {installTab === 'unix' ? (
                    <><span className="comment"># macOS / Linux</span>
                    <br /><span className="prompt">$</span> curl -sfL https://get.perfmon.qzz.io | bash</>
                  ) : (
                    <><span className="comment"># Windows (PowerShell)</span>
                    <br /><span className="prompt">PS&gt;</span> iwr https://get.perfmon.qzz.io/windows -useb | iex</>
                  )}
                </div>
                <CopyButton text={installTab === 'unix'
                  ? 'curl -sfL https://get.perfmon.qzz.io | bash'
                  : 'iwr https://get.perfmon.qzz.io/windows -useb | iex'
                } />
              </div>
            </div>
          </ScrollReveal>
        </section>
      </ParallaxSection>

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
