#!/usr/bin/env python3
import json
import os
from http import HTTPStatus
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

from sentence_transformers import SentenceTransformer


HOST = os.getenv("ST_SERVER_HOST", "127.0.0.1")
PORT = int(os.getenv("ST_SERVER_PORT", "7008"))
MODEL_NAME = os.getenv(
    "ST_MODEL",
    "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2",
)
DEVICE = os.getenv("ST_DEVICE") or None

print(f"loading model: {MODEL_NAME}")
MODEL = SentenceTransformer(MODEL_NAME, device=DEVICE)


def encode_texts(payload):
    texts = payload.get("texts") or payload.get("inputs") or payload.get("sentences")
    if not isinstance(texts, list) or not texts:
        raise ValueError("texts must be a non-empty list")

    texts = [str(item).strip() for item in texts if str(item).strip()]
    if not texts:
        raise ValueError("texts must contain at least one non-empty string")

    requested_model = str(payload.get("model") or "").strip()
    if requested_model and requested_model != MODEL_NAME:
        raise ValueError(f"loaded model is {MODEL_NAME}, request asked for {requested_model}")

    task = str(payload.get("task") or "").strip().lower()
    normalize = bool(payload.get("normalize", True))

    if task == "query" and hasattr(MODEL, "encode_query"):
        embeddings = MODEL.encode_query(texts, normalize_embeddings=normalize)
    elif task == "document" and hasattr(MODEL, "encode_document"):
        embeddings = MODEL.encode_document(texts, normalize_embeddings=normalize)
    else:
        embeddings = MODEL.encode(texts, normalize_embeddings=normalize)

    return {
        "model": MODEL_NAME,
        "task": task or "default",
        "embeddings": embeddings.tolist(),
    }


class Handler(BaseHTTPRequestHandler):
    server_version = "sentence-transformers-http/1.0"

    def do_GET(self):
        if self.path != "/healthz":
            self.send_error(HTTPStatus.NOT_FOUND)
            return
        self.send_json(HTTPStatus.OK, {"status": "ok", "model": MODEL_NAME})

    def do_POST(self):
        if self.path != "/embed":
            self.send_error(HTTPStatus.NOT_FOUND)
            return

        try:
            length = int(self.headers.get("Content-Length", "0"))
        except ValueError:
            self.send_json(HTTPStatus.BAD_REQUEST, {"error": "invalid content length"})
            return

        try:
            body = self.rfile.read(length)
            payload = json.loads(body or b"{}")
            result = encode_texts(payload)
        except json.JSONDecodeError:
            self.send_json(HTTPStatus.BAD_REQUEST, {"error": "invalid json body"})
            return
        except ValueError as exc:
            self.send_json(HTTPStatus.BAD_REQUEST, {"error": str(exc)})
            return
        except Exception as exc:  # noqa: BLE001
            self.send_json(HTTPStatus.INTERNAL_SERVER_ERROR, {"error": str(exc)})
            return

        self.send_json(HTTPStatus.OK, result)

    def send_json(self, status, payload):
        data = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(data)))
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.end_headers()
        self.wfile.write(data)

    def do_OPTIONS(self):
        self.send_response(HTTPStatus.NO_CONTENT)
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.end_headers()

    def log_message(self, format, *args):
        return


if __name__ == "__main__":
    print(f"listening on http://{HOST}:{PORT}")
    ThreadingHTTPServer((HOST, PORT), Handler).serve_forever()
