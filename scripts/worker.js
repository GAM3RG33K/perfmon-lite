// perfmon Cloudflare Worker
// Routes:
//   perfmon.qzz.io       → landing page
//   get.perfmon.qzz.io   → script redirects (install, update, uninstall)

const LANDING_PAGE = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>perfmon — Mobile Performance Monitor</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'SF Mono','SF Pro','Segoe UI',system-ui,sans-serif;background:#0a0a0f;color:#c8c8d0;line-height:1.6;-webkit-font-smoothing:antialiased}
.container{max-width:800px;margin:0 auto;padding:0 24px}
.hero{padding:100px 0 60px;text-align:center}
.hero h1{font-size:clamp(32px,6vw,56px);font-weight:700;color:#fff;letter-spacing:-0.03em;margin-bottom:16px}
.hero h1 span{color:#0ff}
.hero p{font-size:18px;color:#666;max-width:560px;margin:0 auto 40px}
.terminal{background:#12121a;border:1px solid #1e1e2a;border-radius:12px;overflow:hidden;text-align:left;font-size:13px;line-height:1.5;font-family:'SF Mono','Cascadia Code','JetBrains Mono',monospace;margin-bottom:80px}
.terminal-header{display:flex;align-items:center;gap:8px;padding:12px 16px;background:#1e1e2a;border-bottom:1px solid #1e1e2a}
.dot{width:10px;height:10px;border-radius:50%}.dot:nth-child(1){background:#ff5f57}.dot:nth-child(2){background:#ffbd2e}.dot:nth-child(3){background:#28c840}
.terminal-title{color:#666;font-size:11px;margin-left:8px}
.terminal-body{padding:20px;overflow-x:auto;white-space:pre}
.cyan{color:#0ff}.magenta{color:#f0f}.green{color:#0f8}.dim{color:#666}.amber{color:#fb0}
.features{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:20px;margin-bottom:80px}
.card{background:#12121a;border:1px solid #1e1e2a;border-radius:12px;padding:24px}
.card h3{font-size:15px;font-weight:600;color:#fff;margin-bottom:8px}
.card p{font-size:13px;color:#666;line-height:1.5}
.install-box{background:#12121a;border:1px solid #1e1e2a;border-radius:12px;padding:40px;margin-bottom:60px;text-align:center}
.install-box h2{font-size:20px;font-weight:600;color:#fff;margin-bottom:24px}
.code{background:#000;border:1px solid #1e1e2a;border-radius:8px;padding:16px;font-family:'SF Mono','Cascadia Code',monospace;font-size:13px;color:#0f8;overflow-x:auto;text-align:left}
.code .cmt{color:#666}
.code .prmpt{color:#0ff}
.cmd-table{width:100%;border-collapse:collapse;font-size:13px;margin-bottom:60px}
.cmd-table th{text-align:left;padding:10px 16px;color:#666;font-weight:500;border-bottom:1px solid #1e1e2a;font-size:11px;text-transform:uppercase;letter-spacing:1px}
.cmd-table td{padding:10px 16px;border-bottom:1px solid #1e1e2a}
.cmd-table code{font-family:'SF Mono','Cascadia Code',monospace;font-size:12px;color:#0ff}
.footer{text-align:center;padding:40px 0 60px;border-top:1px solid #1e1e2a;font-size:13px;color:#666}
a{color:#0ff;text-decoration:none}a:hover{text-decoration:underline}
</style>
</head>
<body>
<div class="hero"><div class="container">
<h1>perfmon <span>⎈</span></h1>
<p>Blistering-fast, terminal-based mobile app profiling. CPU, memory, and thread telemetry for Android and iOS — right in your terminal.</p>
</div></div>
<div class="container">
<div class="terminal">
<div class="terminal-header"><span class="dot"></span><span class="dot"></span><span class="dot"></span><span class="terminal-title">perfmon</span></div>
<div class="terminal-body"><span class="dim">┌─ perfmon ──────────────────────────────────────┐
│  Target: Pixel 8  │  App: com.example.app  </span><span class="green">[DEBUG]</span><span class="dim">  │
├─────────────────────────────────────────────────┤
│  </span><span class="cyan">CPU Utilization (%)</span><span class="dim">                             │
│ 100 ┤      </span><span class="magenta">╭╮</span><span class="dim">                                       │
│  50 ┤  </span><span class="magenta">╭──╯╰─╮╭──╮</span><span class="dim">                                 │
│   0 └─</span><span class="magenta">╯     ╰╯  ╰</span>─────────────────────────────  │
│  </span><span class="magenta">Memory Footprint (MB)</span><span class="dim">                            │
│ 210 ┤      </span><span class="magenta">╭────────────────────────────────</span><span class="dim">     │
│   0 └─</span><span class="magenta">──╯</span><span class="dim">                                       │
│  </span><span class="amber">Peak CPU: 78%  │  Peak RAM: 215 MB</span><span class="dim">                │
├─────────────────────────────────────────────────┤
│  [↑/↓] Navigate  [TAB] Switch  [e] Export  [?] Help  │
└─────────────────────────────────────────────────┘</span></div>
</div>
<div class="features">
<div class="card"><h3>⚡ Instant Start</h3><p>Boot-to-profiling in under a second. No heavy IDEs.</p></div>
<div class="card"><h3>📱 Android & iOS</h3><p>One interface for both platforms. Auto-detect devices.</p></div>
<div class="card"><h3>📊 Live Charts</h3><p>Real-time CPU, memory, and thread sparkline charts.</p></div>
<div class="card"><h3>📦 Export Anywhere</h3><p>Export to JSON, Markdown, or HTML. Perfect for CI/CD.</p></div>
<div class="card"><h3>🔋 Zero Bloat</h3><p>Single 5.5MB binary. No runtime dependencies.</p></div>
<div class="card"><h3>🎯 Target by App</h3><p>Profile specific apps with --id flag.</p></div>
</div>
<div class="install-box">
<h2>Install in seconds</h2>
<div class="code"><span class="cmt"># macOS / Linux</span>
<span class="prmpt">$</span> curl -sfL https://get.perfmon.qzz.io | bash
</div>
<p style="margin-top:16px;font-size:13px;color:#666"><a href="https://github.com/GAM3RG33K/perfmon-lite">View source</a> · <a href="https://github.com/GAM3RG33K/perfmon-lite/releases">Releases</a></p>
</div>
<table class="cmd-table"><thead><tr><th>Command</th><th>What it does</th></tr></thead>
<tbody>
<tr><td><code>perfmon --mock</code></td><td>Try with simulated data</td></tr>
<tr><td><code>perfmon</code></td><td>Auto-detect and profile</td></tr>
<tr><td><code>perfmon --id com.example.app</code></td><td>Target a specific app</td></tr>
<tr><td><code>perfmon devices</code></td><td>List connected devices</td></tr>
<tr><td><code>perfmon update</code></td><td>Self-update to latest</td></tr>
</tbody></table>
</div>
<div class="footer"><p>Built with <a href="https://go.dev">Go</a>, <a href="https://github.com/charmbracelet/bubbletea">Bubble Tea</a>, and <a href="https://github.com/charmbracelet/lipgloss">Lipgloss</a>.</p><p style="margin-top:8px"><a href="https://github.com/GAM3RG33K/perfmon-lite">GitHub</a> · MIT License</p></div>
</body>
</html>`;

export default {
  async fetch(request) {
    const url = new URL(request.url);
    const host = url.hostname;
    const path = url.pathname;

    // ─── get.perfmon.qzz.io — script redirects ────────────────────────
    if (host === "get.perfmon.qzz.io") {
      if (path === "/" || path === "") {
        return Response.redirect(
          "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/install.sh",
          302
        );
      }
      if (path === "/windows") {
        return Response.redirect(
          "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/install.ps1",
          302
        );
      }
      if (path === "/update") {
        return Response.redirect(
          "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/update.sh",
          302
        );
      }
      if (path === "/update/windows") {
        return Response.redirect(
          "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/update.ps1",
          302
        );
      }
      if (path === "/uninstall") {
        return Response.redirect(
          "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/uninstall.sh",
          302
        );
      }
      if (path === "/uninstall/windows") {
        return Response.redirect(
          "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/uninstall.ps1",
          302
        );
      }
      return new Response("Not found", { status: 404 });
    }

    // ─── perfmon.qzz.io — landing page ────────────────────────────────
    return new Response(LANDING_PAGE, {
      headers: { "content-type": "text/html;charset=UTF-8" },
    });
  }
};
