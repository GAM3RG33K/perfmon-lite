// perfmon Cloudflare Worker
// Routes for get.perfmon.qzz.io — script redirects (install, update, uninstall)
// Landing page is served via GitHub Pages at perfmon.qzz.io

export default {
  async fetch(request) {
    const url = new URL(request.url);
    const path = url.pathname;

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
};
