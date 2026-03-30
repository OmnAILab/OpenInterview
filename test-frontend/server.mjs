import { createReadStream, existsSync, statSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { extname, join, normalize } from "node:path";
import http from "node:http";

const host = process.env.TEST_FRONTEND_HOST || "127.0.0.1";
const port = Number(process.env.TEST_FRONTEND_PORT || 4173);
const root = normalize(fileURLToPath(new URL(".", import.meta.url)));

const mimeTypes = {
  ".css": "text/css; charset=utf-8",
  ".html": "text/html; charset=utf-8",
  ".js": "application/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".svg": "image/svg+xml",
};

function resolvePath(urlPath) {
  const pathname = new URL(urlPath || "/", "http://localhost").pathname;
  const safePath = pathname === "/" ? "/index.html" : pathname;
  const filePath = normalize(join(root, safePath));
  if (!filePath.startsWith(root)) {
    return null;
  }
  return filePath;
}

const server = http.createServer((req, res) => {
  const filePath = resolvePath(req.url || "/");
  if (!filePath || !existsSync(filePath) || statSync(filePath).isDirectory()) {
    res.writeHead(404, { "Content-Type": "text/plain; charset=utf-8" });
    res.end("Not found");
    return;
  }

  const ext = extname(filePath).toLowerCase();
  res.writeHead(200, {
    "Cache-Control": "no-store",
    "Content-Type": mimeTypes[ext] || "application/octet-stream",
  });
  createReadStream(filePath).pipe(res);
});

server.listen(port, host, () => {
  console.log(`Test frontend running at http://${host}:${port}`);
});
