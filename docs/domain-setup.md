# perfmon.qzz.io — Domain & Cloudflare Setup Guide

## Prerequisites

- Access to wherever `perfmon.qzz.io` is registered (Namecheap, GoDaddy, Porkbun, etc.)
- A Cloudflare account (free tier is fine) — sign up at https://dash.cloudflare.com/sign-up

---

## Step 1: Add your domain to Cloudflare

1. Log in to https://dash.cloudflare.com
2. Click **Add a domain**
3. Enter `perfmon.qzz.io` and click **Continue**
4. Select the **Free** plan and click **Continue**
5. Cloudflare will scan existing DNS records — click **Continue**

---

## Step 2: Find your current registrar

If you don't know where the domain is registered, check:

```bash
whois perfmon.qzz.io | grep -i registrar
```

Typical registrars and where to change nameservers:

| Registrar | Nameserver Settings Location |
|-----------|------------------------------|
| **Namecheap** | Dashboard → Domain List → Manage → Nameservers → Custom DNS |
| **GoDaddy** | My Products → Domains → DNS → Nameservers → Change |
| **Porkbun** | Domains → Details → Nameservers |
| **Google Domains** | DNS → Nameservers → Custom nameservers |
| **Cloudflare Registrar** | Already on Cloudflare — skip Step 3 |

---

## Step 3: Point nameservers to Cloudflare

Cloudflare will show you two nameservers after adding your domain, typically:

```
dns.ns.cloudflare.com
dns.ns.cloudflare.com
```

**Your actual nameservers will differ** — use the ones Cloudflare gives you, not the example above.

1. Go to your registrar's **Nameservers** settings
2. Switch from their default nameservers to **Custom**
3. Enter the two Cloudflare nameservers provided to you
4. Save — propagation takes **1–60 minutes** (usually <5 min)

---

## Step 4: Create DNS records in Cloudflare

Once the domain shows **Active** in Cloudflare:

1. Go to **DNS** → **Records**
2. Add these records:

| Type | Name | Content | Proxy | Purpose |
|------|------|---------|-------|---------|
| **A** | `@` | `192.0.2.1` | Proxied (orange cloud) | Placeholder — replaced by Worker |
| **CNAME** | `get` | `GAM3RG33K.github.io` | DNS only (grey cloud) | For hosting installer page/files on GitHub Pages |

> The `@` record is a placeholder. Once the Worker is deployed, Cloudflare will route traffic through the Worker instead (proxy handles this).

---

## Step 5: Deploy the Cloudflare Worker

The Worker will redirect `https://perfmon.qzz.io` to the raw install script, so users can run:

```bash
curl -sfL https://perfmon.qzz.io | bash
```

### 5a — Create the Worker

1. In Cloudflare dashboard, go to **Workers & Pages** → **Create application**
2. Click **Create Worker**
3. Name it `perfmon-install`
4. Replace the default code with:

```js
export default {
  async fetch(request) {
    const url = new URL(request.url);

    // Redirect root requests to the install script
    if (url.pathname === "/" || url.pathname === "") {
      return Response.redirect(
        "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/install.sh",
        302
      );
    }

    // Redirect /update to the update script
    if (url.pathname === "/update") {
      return Response.redirect(
        "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/update.sh",
        302
      );
    }

    // Everything else: 404
    return new Response("Not found", { status: 404 });
  }
}
```

5. Click **Save and Deploy**

### 5b — Get the Worker URL

After deploying, Cloudflare shows you a `*.workers.dev` URL like:

```
https://perfmon-install.your-subdomain.workers.dev
```

Test it:

```bash
curl -sfL https://perfmon-install.your-subdomain.workers.dev | head -3
```

You should see the first lines of `install.sh`.

---

## Step 6: Connect the Worker to your domain

1. In the Worker page, go to **Triggers** → **Custom Domains**
2. Click **Add Custom Domain**
3. Enter `perfmon.qzz.io`
4. Click **Add domain**

Cloudflare automatically provisions an SSL certificate and routes requests through the Worker.

> **No A record IP needed** — when the Worker is connected to a custom domain via Cloudflare, traffic goes through Cloudflare's edge network directly to the Worker. The `@` A record (`192.0.2.1`) is just a placeholder to keep the domain active in DNS.

---

## Step 7: Verify everything

```bash
# Should download and show the install script
curl -sfL https://perfmon.qzz.io | head -5

# Should show the update script
curl -sfL https://perfmon.qzz.io/update | head -5

# Final one-liner for users:
curl -sfL https://perfmon.qzz.io | bash
```

---

## Step 8 (optional): Update README with the one-liner

Once verified, the install command in the README becomes:

```bash
curl -sfL https://perfmon.qzz.io | bash
```

---

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| `curl: (6) Could not resolve host` | DNS not propagated yet | Wait up to 60 min, check `dig perfmon.qzz.io` |
| `curl: (35) SSL connect error` | SSL cert still provisioning | Wait 1–5 min for Cloudflare to issue cert |
| Worker returns 404 | Wrong path or Worker not deployed | Check Worker code; check Triggers → Custom Domains |
| `Not Found` on root | Worker not connected to domain | Go to Worker → Triggers and add custom domain |
