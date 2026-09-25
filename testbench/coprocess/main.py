import os
import sys
import signal

from concurrent import futures

sys.path.append(os.path.dirname(__file__) + '/../../internal/grpcipc/python')

import grpc
import ipc_pb2 as ipc1
import ipc_pb2_grpc as ipc2

socketname = sys.argv[1] if len(sys.argv) >= 2 else '/tmp/grpcipc.socket'

if os.path.exists(socketname):
    os.remove(socketname)

class Servicer(ipc2.IpcServicer):
    def Process(self, request, context):
        return ipc1.Response(
            result=
                ipc1.RESULT_FAIL
                    if len(request.payload) == 0
                    else ipc1.RESULT_NO
        )

server = grpc.server(futures.ThreadPoolExecutor())

ipc2.add_IpcServicer_to_server(Servicer(), server)

server.add_insecure_port("unix://" + socketname)

def shutdown(*args, **kwargs):
    server.stop(grace=5)

signal.signal(signal.SIGINT, shutdown)
signal.signal(signal.SIGTERM, shutdown)

server.start()
server.wait_for_termination()
