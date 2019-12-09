from BaseHTTPServer import HTTPServer, BaseHTTPRequestHandler
import json
import pymongo

class RequestHandler(BaseHTTPRequestHandler):
    def _set_headers(self):
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        myclient = pymongo.MongoClient('mongodb://localhost:27017/')
        mydb = myclient['test']
        self.mycol = mydb['accelerations']

    def do_GET(self):
        response = {
            'status':'SUCCESS',
            'data':'hello from server'
        }
        self._set_headers()
        self.wfile.write(json.dumps(response))

    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        print 'post data from client:'
        print post_data
        data = json.loads(post_data)
        x = data["X"]
        y = data["Y"]
        letter = data["letter"]
        mydict = { "X": x, "Y": y, "letter": letter}
        result = self.mycol.insert_one(mydict)
        response = {
            'status':'SUCCESS',
            'data':'server got your post data',
            'result': result
        }
        self._set_headers()
        self.wfile.write(json.dumps(response))

def run():
    port = 8080
    print('Listening on localhost:%s' % port)
    
    server = HTTPServer(('', port), RequestHandler)
    server.serve_forever()

run()
