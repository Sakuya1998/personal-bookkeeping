"""PaddleOCR HTTP API — 轻量 OCR 识别服务"""
import json
import os
import tempfile
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import urlparse
from cgi import FieldStorage
import io

from paddleocr import PaddleOCR

ocr = PaddleOCR(use_angle_cls=True, lang='ch', use_gpu=False)


class OCRHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        parsed = urlparse(self.path)
        if parsed.path == '/health':
            self._json({"status": "ok"}, 200)
        else:
            self._json({"error": "not found"}, 404)

    def do_POST(self):
        if urlparse(self.path).path != '/api/v1/ocr':
            self._json({"error": "not found"}, 404)
            return

        content_type = self.headers.get('Content-Type', '')
        if 'multipart/form-data' not in content_type:
            self._json({"error": "multipart/form-data required"}, 400)
            return

        try:
            form = FieldStorage(
                fp=self.rfile,
                headers=self.headers,
                environ={'REQUEST_METHOD': 'POST'},
            )
            file_item = form['file']
            data = file_item.file.read()

            suffix = '.png'
            name = file_item.filename or ''
            if name.lower().endswith(('.jpg', '.jpeg')):
                suffix = '.jpg'

            with tempfile.NamedTemporaryFile(suffix=suffix, delete=False) as f:
                f.write(data)
                tmp = f.name

            try:
                result = ocr.ocr(tmp, cls=True)
                items = []
                if result and result[0]:
                    for line in result[0]:
                        text = line[1][0]
                        conf = float(line[1][1])
                        items.append({"text": text, "confidence": conf})
                self._json({"result": items, "msg": "success"}, 200)
            finally:
                os.unlink(tmp)

        except Exception as e:
            self._json({"error": str(e)}, 500)

    def _json(self, data, code=200):
        self.send_response(code)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Access-Control-Allow-Origin', '*')
        self.end_headers()
        self.wfile.write(json.dumps(data, ensure_ascii=False).encode())

    def log_message(self, fmt, *args):
        pass


if __name__ == '__main__':
    port = int(os.environ.get('PORT', 9000))
    server = HTTPServer(('0.0.0.0', port), OCRHandler)
    print(f"PaddleOCR API running on http://0.0.0.0:{port}")
    server.serve_forever()
