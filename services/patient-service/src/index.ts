import http from "http";
import { Pool } from "pg";

const port = Number(process.env.PORT || 3000);
const pool = new Pool({ connectionString: process.env.DATABASE_URL });
pool.on("error", (e) => console.error("pool error:", e.message));

function send(res: http.ServerResponse, code: number, body: object) {
  res.writeHead(code, { "Content-Type": "application/json" });
  res.end(JSON.stringify(body));
}

const server = http.createServer(async (req, res) => {
  try {
    if (req.url === "/health") {
      await pool.query("SELECT 1");
      return send(res, 200, { service: "patient-service", status: "ok", db: "up" });
    }
    if (req.url === "/api/v1/patients") {
      const r = await pool.query("SELECT id, full_name FROM patients ORDER BY id");
      return send(res, 200, r.rows);
    }
    return send(res, 404, { error: "not found" });
  } catch (e) {
    return send(res, 503, { service: "patient-service", status: "degraded", db: "down" });
  }
});

server.listen(port, () => console.log(`patient-service listening on :${port}`));

process.on("SIGTERM", () => {
  console.log("SIGTERM received, shutting down gracefully");
  server.close(() => pool.end().then(() => process.exit(0)));
});
