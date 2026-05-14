export default {
  async fetch(request) {
    const url = new URL(request.url);
    const path = url.pathname;

    // macOS / Linux: redirect to bash install script
    if (path === "/" || path === "") {
      return Response.redirect(
        "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/install.sh",
        302
      );
    }

    // Windows: redirect to PowerShell install script
    if (path === "/windows") {
      return Response.redirect(
        "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/install.ps1",
        302
      );
    }

    // Update script (macOS / Linux)
    if (path === "/update") {
      return Response.redirect(
        "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/update.sh",
        302
      );
    }

    // Update script (Windows)
    if (path === "/update/windows") {
      return Response.redirect(
        "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/update.ps1",
        302
      );
    }

    // Uninstall (macOS / Linux)
    if (path === "/uninstall") {
      return Response.redirect(
        "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/uninstall.sh",
        302
      );
    }

    // Uninstall (Windows)
    if (path === "/uninstall/windows") {
      return Response.redirect(
        "https://raw.githubusercontent.com/GAM3RG33K/perfmon-lite/main/scripts/uninstall.ps1",
        302
      );
    }

    return new Response("Not found", { status: 404 });
  }
}
