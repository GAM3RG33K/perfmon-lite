# perfmon — Brand Guidelines & Design System

> **Project:** perfmon — Mobile Performance Monitor & Profiler
> **Tagline:** Blistering-fast, terminal-based mobile app profiling

---

## 1. Brand Essence

| Attribute | Description |
|-----------|-------------|
| **Personality** | Technical, precise, fast, minimal, developer-first |
| **Tone** | Confident, clear, no-nonsense, slightly edgy |
| **Archetype** | The Wizard / The Toolmaker |
| **Vibe** | Terminal aesthetics meets modern product design — "what if `htop` and a premium dev tool had a baby" |

---

## 2. Logo Design Requirements

### Symbol

A single, iconic mark that works at 16px favicon and 1024px hero:

- **Core motif:** A terminal cursor `▌` or prompt `>` combined with a signal/wave (representing telemetry data)
- **Alternative:** A stylized ship's helm / radar dish merged with a terminal bracket `⎈` (the Unicode symbol already used in the landing page)
- **Style:** Monoline, geometric, open paths — no filled shapes
- **Structure:** Asymmetric but balanced — feels like motion / data flowing

### Color Palette

| Color | Hex | Usage | Role |
|-------|-----|-------|------|
| **Cyan** | `#00FFFF` | Primary brand color | Headers, links, selection highlights, logo primary |
| **Magenta** | `#FF00FF` | Secondary accent | Charts, peak indicators, data visualization |
| **Green** | `#00FF88` | Success / debug | Badges, status indicators, "all good" signals |
| **Amber** | `#FFB000` | Warning | Alerts, release badges, threshold indicators |
| **White** | `#FFFFFF` | Text primary | Headings, key values |
| **Dim White** | `#888888` | Text secondary | Labels, descriptions, subtle UI |
| **Background** | `#0A0A0F` | Surface | Dark theme base — near-black with slight blue tint |
| **Surface** | `#12121A` | Card/panel | Elevated surfaces, code blocks, terminal bg |

### Typography

| Usage | Font | Fallback |
|-------|------|----------|
| **Logo/Display** | `SF Mono` or `JetBrains Mono` | `Cascadia Code`, `Fira Code` |
| **Headings** | `SF Pro Display` or `Inter` | `-apple-system`, `Segoe UI` |
| **Body** | `SF Pro Text` or `Inter` | `-apple-system`, `Segoe UI` |
| **Code/Terminal** | `SF Mono` or `JetBrains Mono` | `Fira Code`, `Courier New` |

### Spacing

- 4px grid base unit
- Content max-width: 800px (docs), 700px (terminal mockup)
- 24px / 32px / 44px padding rhythm
- Border radius: 8px (small UI), 12px (cards), 100px (badges)

---

## 3. Logo Generation Prompt

Use this prompt with Midjourney / DALL·E / Stable Diffusion / Leonardo AI:

```
A professional tech logo for "perfmon" — a mobile app performance monitoring tool.
The logo should combine two elements:
1. A terminal prompt symbol (like ">" or "_" cursor or "⎈" ship's wheel)
2. A heartbeat / signal wave line representing live telemetry data

Style:
- Monoline, geometric, open-contour line art — no filled shapes
- Clean minimal tech aesthetic, inspired by htop/btop terminal UIs
- Asymmetric composition with dynamic forward motion feeling
- Pure black line on transparent background (for dark theme usage)
- Suitable for both 16px favicon and full-size hero display
- The wave should have 3-4 smooth peaks suggesting CPU/data flow
- Merge the prompt symbol with the start of the wave line

Avoid:
- No gradients, no shadows, no 3D, no glossy effects
- No circular badges or enclosing borders
- No text / lettermarks — symbol only
- No complex multi-line paths — keep it to 2-3 clean strokes max

The final output should be a clean SVG-able black line logo
```

---

## 4. Design Usage Rules

| Context | Logo Style | Color |
|---------|-----------|-------|
| Landing page hero | Full symbol | Cyan `#00FFFF` on dark bg |
| Favicon / tab icon | Symbol only | White on transparent |
| Terminal title bar | Inline symbol (⎈) | Cyan terminal text |
| Social preview / OG | Symbol + "perfmon" wordmark | Cyan + White |
| Light theme (future) | Symbol only | Black on white bg |
| Monochrome / print | Symbol only | Solid black |

---

## 5. Application Examples

### Terminal integration
```
┌─ perfmon v0.0.7 ⎈ ─────────────────────────────┐
│  Target: Pixel 8  │  App: com.example.app       │
├────────────────────────────────────────────────┤
```

### Favicon
A 16×16 pixel rendering of the symbol — must be recognizable at this size.

### Social card
Centered symbol (120px) above "perfmon" wordmark in SF Mono, tagline below in smaller weight. Dark background with cyan symbol.
