#!/usr/bin/env python

import argparse
import json
from http.server import BaseHTTPRequestHandler, HTTPServer

class MockHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        response = [
            {
                'useDate': '2024-06-03',
                'name': 'てすと明細Z',
                'price': 1234,
                'registDate': '2024-06-05',
                'importJudgeDate': '2024-06-05 15:00:00',
                # 'importDate': ,
            },
            {
                'useDate': '2024-06-02',
                'name': 'てすと明細Y',
                'price': 1234,
                'registDate': '2024-06-04',
                'importJudgeDate': '2024-06-04 15:00:00',
                'importDate': '2024-06-04 15:00:00'
            },
            {
                'useDate': '2024-06-01',
                'name': 'てすと明細X',
                'price': 1234,
                'registDate': '2024-06-03',
                'importJudgeDate': '2024-06-03 15:00:00',
                'importDate': '2024-06-03 15:00:00'
            },
            ]
        responseBody = json.dumps(response)

        self.wfile.write(responseBody.encode('utf-8'))

def import_args():
    parser = argparse.ArgumentParser("mock server start")

    parser.add_argument('--host', '-H', required=False, default='0.0.0.0')
    parser.add_argument('--port', '-P', required=False, type=int, default=8080)

    args = parser.parse_args()

    return args.host, args.port

def run(server_class=HTTPServer, handler_class=MockHandler, server_name='localhost', port=20010):

    server = server_class((server_name, port), handler_class)
    server.serve_forever()

def main():
    host, port = import_args()
    run(server_name=host, port=port)

if __name__ == '__main__':
    main()
