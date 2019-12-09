import socket

HOST, PORT = '', 8888

def respond(conn, response):
    conn.send('HTTP/1.1 200 OK\n')
    conn.send('Content-Type: text/html\n')
    conn.send('Connection: close\n\n')
    conn.sendall(response + '\n\n')

def check(request):
    if request.find('display message') > 0:
        return 'displaying message'
    elif request.find('led on') > 0:
        return 'turnning on led'
    elif request.find('led off') > 0:
        return 'turnning off led'
    elif request.find('display time') > 0:
        return 'displaying time'
    else:
        return 'not success'

listen_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
listen_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 5)
listen_socket.bind((HOST, PORT))
listen_socket.listen(1)
print 'Serving HTTP on port %s ...' % PORT
while True:
    conn, client_address = listen_socket.accept()
    request = conn.recv(1024)
    result = check(request)

    respond(conn, 'success ' + result)
    conn.close()

