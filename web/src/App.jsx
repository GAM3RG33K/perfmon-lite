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

/* ─── Device detection ─────────────────────────────────── */

function useDeviceInfo() {
  const [info] = useState(() => {
    const ua = navigator.userAgent
    let device = 'Desktop'
    let platform = 'unknown'
    if (/iPhone|iPad|iPod/.test(ua)) { device = 'iPhone'; platform = 'ios' }
    else if (/Android/.test(ua)) { device = 'Android Device'; platform = 'android' }
    else if (/Mac/.test(ua)) { device = 'Mac'; platform = 'darwin' }
    else if (/Linux/.test(ua)) { device = 'Linux PC'; platform = 'linux' }
    else if (/Windows/.test(ua)) { device = 'Windows PC'; platform = 'windows' }
    return { device, platform, cores: navigator.hardwareConcurrency || 4 }
  })
  return info
}

/* ─── Live metrics hook ────────────────────────────────── */

function useLiveMetrics(deviceName) {
  const totalMem = navigator.deviceMemory || 8
  const isChrome = typeof performance?.memory?.usedJSHeapSize === 'number'

  const [metrics, setMetrics] = useState(() => {
    let initialMem = totalMem * 0.45
    if (isChrome) {
      // Real JS heap ratio scaled to total RAM
      const heapRatio = performance.memory.usedJSHeapSize / performance.memory.jsHeapSizeLimit
      initialMem = Math.round(totalMem * heapRatio * 100) / 100
      if (initialMem < 0.5 || initialMem > totalMem * 0.9) initialMem = totalMem * 0.45
    }
    return {
      device: deviceName,
      cpuCores: navigator.hardwareConcurrency || 4,
      cpuLoad: 28,
      memUsed: initialMem,
      memTotal: totalMem.toString(),
      threads: 42,
      uptime: '0:00',
    }
  })

  useEffect(() => {
    const total = parseFloat(totalMem.toString())
    let prevLoad = 28
    let prevMem = 0

    const tick = () => {
      // Smooth CPU: drift slowly ±3% per tick, cap at 5-80%
      prevLoad += (Math.random() - 0.5) * 6
      prevLoad = Math.max(5, Math.min(80, prevLoad))
      const load = Math.round(prevLoad)

      // Smooth memory: very slow drift ±2% of total per tick
      if (isChrome) {
        // Real JS heap data
        const heapRatio = performance.memory.usedJSHeapSize / performance.memory.jsHeapSizeLimit
        const target = total * heapRatio
        prevMem = prevMem === 0 ? target : prevMem + (target - prevMem) * 0.1
      } else {
        prevMem === 0 ? prevMem = total * 0.45 : prevMem += (Math.random() - 0.5) * total * 0.02
        prevMem = Math.max(total * 0.2, Math.min(total * 0.85, prevMem))
      }

      setMetrics(prev => ({
        ...prev,
        cpuLoad: load,
        memUsed: Math.round(prevMem * 100) / 100,
        threads: Math.floor(38 + Math.sin(Date.now() / 5000) * 4),
      }))
    }
    tick()
    const t = setInterval(tick, 2000)
    return () => clearInterval(t)
  }, [isChrome, totalMem])

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

// Draw a mini line chart using Unicode half-blocks
function lineChartStr(pct) {
  const w = 30
  const filled = Math.round(pct / 100 * w)
  let s = ''
  for (let i = 0; i < w; i++) {
    if (i < filled) s += '█'
    else s += '░'
  }
  return s
}

function LiveTerminal({ mouse }) {
  const device = useDeviceInfo()
  const m = useLiveMetrics(device.device)
  const ver = import.meta.env.VITE_APP_VERSION || 'dev'

  const lines = [
    `┌────────────────────────────────────────────────────────────┐`,
    `│ perfmon-tool v${(ver + ' ').padEnd(14)} Device: ${m.device.padEnd(15)} Uptime: ${m.uptime.padEnd(5)} │`,
    `├────────────────────────────────────────────────────────────┤`,
    `│ [Dashboard]                                          (q) quit │`,
    `├────────────────────────────────────────────────────────────┤`,
    `│ ${('CPU: ' + m.cpuCores + ' cores').padEnd(20)} ${('Mem: ' + m.memTotal + ' GB').padEnd(15)} App: com.example.app [DEBUG] │`,
    `├────────────────────────────────────────────────────────────┤`,
    `│ CPU Utilization (overall)  ${m.cpuLoad}%`,
    `│ ${lineChartStr(m.cpuLoad)} ${String(m.cpuLoad).padStart(3)}%`,
    `│`,
    `│ Memory (Total: ${m.memTotal} GB)  ${m.memUsed.toFixed(1)} GB used`,
    `│ ${lineChartStr(Math.round(m.memUsed / parseFloat(m.memTotal) * 100))} ${Math.round(m.memUsed / parseFloat(m.memTotal) * 100)}%`,
    `│`,
    `│ Threads: ${m.threads}  │  Peak CPU: ${m.cpuLoad + 12}%  │  Peak RAM: ${(m.memUsed + 0.5).toFixed(1)} GB`,
    `├────────────────────────────────────────────────────────────┤`,
    `│ [↑/↓] Scroll  [TAB] Switch  [e] Export  [?] Help  [q] Quit  │`,
    `├────────────────────────────────────────────────────────────┤`,
    `│ INFO  System ready  |  INFO  Polling every 1s  |  OK  12 samples │`,
    `└────────────────────────────────────────────────────────────┘`,
  ]

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
          {lines.map((l, i) => (
            <div key={i} className={`line ${i >= 6 && i <= 8 ? 'cyan' : i >= 9 && i <= 11 ? 'magenta' : i === 13 ? 'amber' : i >= 15 ? 'dim' : 'dim'}`}>
              {l}
            </div>
          ))}
          <span className="cursor">▌</span>
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
