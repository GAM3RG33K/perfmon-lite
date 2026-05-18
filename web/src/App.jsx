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
  ['perfmon-tool --mock', 'Try with simulated data'],
  ['perfmon-tool', 'Auto-detect and profile'],
  ['perfmon-tool --id com.example.app', 'Target a specific app'],
  ['perfmon-tool devices', 'List connected devices'],
  ['perfmon-tool update', 'Self-update to latest'],
]

function useTypewriter(lines, speed = 12) {
  const [visible, setVisible] = useState([])
  const idxRef = useRef(0)
  const charRef = useRef(0)
  useEffect(() => {
    const t = setInterval(() => {
      if (idxRef.current >= lines.length) { clearInterval(t); return }
      const line = lines[idxRef.current]
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

/* ─── Live metrics hook ────────────────────────────────── */

function useLiveMetrics() {
  const [metrics, setMetrics] = useState(() => ({
    cpuCores: navigator.hardwareConcurrency || 4,
    cpuLoad: 34,
    memUsed: 3.2,
    memTotal: (performance.memory ? performance.memory.jsHeapSizeLimit / (1024*1024*1024) : 8).toFixed(1),
    threads: 42,
    uptime: '0:00',
  }))

  useEffect(() => {
    const tick = () => {
      const load = Math.floor(Math.random() * 40 + 15)
      const mem = performance.memory
        ? (performance.memory.usedJSHeapSize / (1024*1024*1024)).toFixed(1)
        : (2 + Math.random() * 3).toFixed(1)
      setMetrics(prev => {
        const totalMin = parseFloat(prev.memTotal) * 0.3
        const memVal = Math.max(totalMin, parseFloat(mem) || 0)
        return {
          ...prev,
          cpuLoad: load,
          memUsed: parseFloat(memVal.toFixed(1)),
          threads: Math.floor(35 + Math.random() * 20),
        }
      })
    }
    const t = setInterval(tick, 3000)
    tick()
    return () => clearInterval(t)
  }, [])

  return metrics
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
  return <div ref={ref} className={className} style={{ transform: `translateY(${offset}px)` }}>{children}</div>
}

function Floaters() {
  const scrollY = useScrollY()
  const mouse = useMousePos()
  const symbols = ['</>', '{ }', '()', '[]', '#!', '=>', '/*', '*/', '&&', '||']
  return (
    <div className="floaters" aria-hidden="true">
      {symbols.map((s, i) => {
        const driftX = Math.sin(scrollY * 0.001 + i) * 20 + (mouse.x - 0.5) * 40 * (1 + (i % 3) * 0.5)
        const driftY = Math.cos(scrollY * 0.0008 + i * 1.5) * 15
        return (
          <span key={i} className="floater" style={{
            left: `${(i % 5) * 20 + 5}%`,
            top: `${i * 15 + 10}%`,
            transform: `translate(${driftX}px, ${driftY}px)`,
            opacity: 0.06 + (i % 3) * 0.03,
            fontSize: `${12 + (i % 4) * 4}px`,
          }}>{s}</span>
        )
      })}
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

function CopyButton({ text }) {
  const [copied, setCopied] = useState(false)
  return (
    <button className={`copy-btn ${copied ? 'copied' : ''}`}
      onClick={() => { navigator.clipboard.writeText(text); setCopied(true); setTimeout(() => setCopied(false), 2000) }}
      title="Copy command">
      {copied ? '✓' : '⎘'}
    </button>
  )
}

/* ─── Terminal with live metrics ────────────────────────── */

function Bar({ pct, color }) {
  const w = Math.max(Math.round(pct / 100 * 44), 1)
  const bar = '█'.repeat(w) + '─'.repeat(Math.max(44 - w, 0))
  return <span style={{ color }}>{bar}</span>
}

function LiveTerminal({ mouse }) {
  const m = useLiveMetrics()
  const ver = import.meta.env.VITE_APP_VERSION || 'dev'
  const memPct = Math.round((m.memUsed / parseFloat(m.memTotal)) * 100)
  const cachePct = Math.max(Math.round(memPct * 0.3), 2)

  const lines = [
    `┌─────────────────────────────────────────────────────────────────────┐`,
    `│ perfmon-tool v${ver.padEnd(14)} Device: Pixel 8            Uptime: ${m.uptime.padEnd(5)} │`,
    `├─────────────────────────────────────────────────────────────────────┤`,
    `│ [Dashboard]  [Processes]  [System Logs]                      (q) quit │`,
    `├─────────────────────────────────────────────────────────────────────┤`,
    `│ App: com.example.app  [DEBUG]  │  CPU: ${m.cpuCores} cores  │  Mem: ${m.memTotal} GB total  │`,
    `├─────────────────────────────────────────────────────────────────────┤`,
    `│ CPU Utilization (overall)  ${m.cpuLoad}%`,
    `│ ┌─────────────────────────────────────────────────────────────────┐ │`,
    `│ │ ${Bar({ pct: m.cpuLoad, color: '#0ff' })} ${String(m.cpuLoad).padStart(3)}% │ │`,
    `│ └─────────────────────────────────────────────────────────────────┘ │`,
    `│`,
    `│ Memory (Total: ${m.memTotal} GB)  ${m.memUsed.toFixed(1)} GB used`,
    `│ ┌─────────────────────────────────────────────────────────────────┐ │`,
    `│ │ Used:  ${Bar({ pct: memPct, color: '#f0f' })} ${String(memPct).padStart(3)}% │ │`,
    `│ │ Cache: ${Bar({ pct: cachePct, color: '#0f8' })} ${String(cachePct).padStart(3)}% │ │`,
    `│ └─────────────────────────────────────────────────────────────────┘ │`,
    `│`,
    `│ Threads: ${m.threads}  │  Peak CPU: ${m.cpuLoad + 12}%  │  Peak RAM: ${(m.memUsed + 0.5).toFixed(1)} GB  │  Samples: 300`,
    `├─────────────────────────────────────────────────────────────────────┤`,
    `│ [↑/↓] Navigate  [TAB] Switch  [e] Export  [?] Help  [q] Quit        │`,
    `└─────────────────────────────────────────────────────────────────────┘`,
  ]

  const typed = useTypewriter(lines, 10)

  return (
    <section className="terminal-section">
      <div className="terminal" style={{
        transform: `rotateX(${(mouse.y - 0.5) * 3}deg) rotateY(${(mouse.x - 0.5) * 3}deg)`,
      }}>
        <div className="terminal-header">
          <span className="dot"></span><span className="dot"></span><span className="dot"></span>
          <span className="terminal-title">perfmon-tool — live</span>
        </div>
        <div className="terminal-body">
          {typed.map((l, i) => (
            <div key={i} className={`line ${i >= 6 && i <= 9 ? 'cyan' : i >= 11 && i <= 15 ? 'magenta' : i === 17 ? 'amber' : 'dim'}`}>
              {l}{i < typed.length - 1 ? '' : <span className="cursor">▌</span>}
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}

/* ─── App ────────────────────────────────────────────────── */

export default function App() {
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
          <LiveTerminal mouse={mouse} />
        </ParallaxSection>
      </div>

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

      <ParallaxSection speed={0.05}>
        <section className="install-section">
          <ScrollReveal>
            <div className="install-box">
              <h2>Install in seconds</h2>
              <div className="install-platform-tabs">
                <span className={`install-tab ${installTab === 'unix' ? 'active' : ''}`} onClick={() => setInstallTab('unix')}>macOS / Linux</span>
                <span className={`install-tab ${installTab === 'win' ? 'active' : ''}`} onClick={() => setInstallTab('win')}>Windows</span>
              </div>
              <div className="code-wrap" style={{ transform: `perspective(400px) rotateX(${(mouse.y - 0.5) * 2}deg)` }}>
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
                  : 'iwr https://get.perfmon.qzz.io/windows -useb | iex'} />
              </div>
            </div>
          </ScrollReveal>
        </section>
      </ParallaxSection>

      <section className="download-section">
        <ScrollReveal>
          <h2 className="section-title">Manual Download</h2>
          <div className="download-box">
            <p className="download-desc">Download the latest binary from <a href="https://github.com/GAM3RG33K/perfmon-lite/releases">GitHub Releases</a>.</p>
            <table className="dl-table">
              <thead><tr><th>Platform</th><th>File</th></tr></thead>
              <tbody>
                <tr><td>macOS (Intel)</td><td><code>perfmon-tool-{import.meta.env.VITE_APP_VERSION || 'dev'}-darwin-amd64</code></td></tr>
                <tr><td>macOS (Apple Silicon)</td><td><code>perfmon-tool-{import.meta.env.VITE_APP_VERSION || 'dev'}-darwin-arm64</code></td></tr>
                <tr><td>Linux (x86_64)</td><td><code>perfmon-tool-{import.meta.env.VITE_APP_VERSION || 'dev'}-linux-amd64</code></td></tr>
                <tr><td>Linux (ARM64)</td><td><code>perfmon-tool-{import.meta.env.VITE_APP_VERSION || 'dev'}-linux-arm64</code></td></tr>
                <tr><td>Windows (x86_64)</td><td><code>perfmon-tool-{import.meta.env.VITE_APP_VERSION || 'dev'}-windows-amd64.exe</code></td></tr>
                <tr><td>Windows (ARM64)</td><td><code>perfmon-tool-{import.meta.env.VITE_APP_VERSION || 'dev'}-windows-arm64.exe</code></td></tr>
              </tbody>
            </table>
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

      <section className="usage-section">
        <ScrollReveal>
          <h2 className="section-title">Usage</h2>
          <div className="usage-box">
            <p>See the <a href="https://github.com/GAM3RG33K/perfmon-lite/blob/main/USAGE.md">full usage guide</a> on GitHub.</p>
            <ul className="usage-list">
              <li>Interactive TUI keybindings and navigation</li>
              <li>Exporting telemetry to JSON, Markdown, and HTML</li>
              <li>Targeting devices and apps with <code>--device</code> and <code>--id</code></li>
              <li>Environment variables for configuration</li>
              <li>Platform prerequisites and troubleshooting</li>
            </ul>
          </div>
        </ScrollReveal>
      </section>

      <footer className="footer">
        <p>Built with <a href="https://go.dev">Go</a>, <a href="https://github.com/charmbracelet/bubbletea">Bubble Tea</a>, and <a href="https://github.com/charmbracelet/lipgloss">Lipgloss</a>.</p>
        <p><a href="https://github.com/GAM3RG33K/perfmon-lite">GitHub</a> · MIT License</p>
      </footer>
    </div>
  )
}
