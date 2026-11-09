"""使用真实本地 HTTP 路径验证鉴权、请求体与重定向边界。"""
import http.server
import importlib.util
import json
from pathlib import Path
import threading
import unittest

spec = importlib.util.spec_from_file_location("forum_http_client", Path(__file__).with_name("bridge.py"))
bridge = importlib.util.module_from_spec(spec)
spec.loader.exec_module(bridge)


class ClientTests(unittest.TestCase):
    def test_real_http_authorization_and_no_redirect(self):
        requests = []

        class Handler(http.server.BaseHTTPRequestHandler):
            def do_POST(self):
                body = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
                requests.append((self.path, self.headers["Authorization"], self.headers["X-Telegram-User-ID"], body))
                self.send_response(200)
                self.end_headers()
                self.wfile.write(b'{"code":0,"data":{"id":"operation"}}')

            def do_GET(self):
                requests.append(self.path)
                self.send_response(302)
                self.send_header("Location", "/credential-leak")
                self.end_headers()

            def log_message(self, *args):
                pass

        server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        try:
            client = bridge.Client(f"http://127.0.0.1:{server.server_port}", "test-download-token", 123)
            self.assertEqual(client.call("POST", "/operations", {"event_id": "event"}), {"id": "operation"})
            self.assertEqual(requests[0], ("/api/v1/integrations/telegram-forum/operations", "Bearer test-download-token", "123", {"event_id": "event"}))
            with self.assertRaises(bridge.APIError):
                client.call("GET", "/receipts")
            self.assertEqual(requests[1:], ["/api/v1/integrations/telegram-forum/receipts"])
        finally:
            server.shutdown()
            server.server_close()
            thread.join(timeout=5)


if __name__ == "__main__":
    unittest.main()
