#!/usr/bin/env python2

import os
import sys
import unittest
import random
import string
import tempfile
import subprocess
import socket
import requests
import time


def check_port(p):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    return sock.connect_ex(('127.0.0.1', p)) == 0


class Test(unittest.TestCase):

    def writefile(self, fp, s):
        f = open(os.path.join(self.cwd, fp), 'w')
        f.write(s)
        f.close()

    def setUp(self):
        self.cwd = tempfile.mkdtemp()
        self.title = ''.join([random.choice(string.printable[:62]) for x in range(20)])
        self.ports = [random.randint(60000, 65535) for x in range(4)]
        self.writefile(".md", "# Wiki Index Page\n\nStrapdown Rocks!\n\n")
        if not os.path.exists("./server"):
            print './server not found'
            sys.exit(10)

        args = ["./server", "-dir=" + self.cwd, "-toc=true", "-title=" + self.title, "-init", "-heading_number=i", "-addr=" + ','.join(map(lambda x: '127.0.0.1:%d' % x, self.ports))]
        print args
        self.proc = subprocess.Popen(args, stdout=subprocess.PIPE)

        while True:
            for i in ports:
                if check_port(i):
                    print("test port true")
                    break
            else:
                continue
            break
        # wait other ports avaliable
        time.sleep(0.5)
        self.ports = filter(check_port, self.ports)
        assert len(self.ports) > 0

    def test_index(self):
        r = requests.get("http://127.0.0.1:%d/" % self.ports[0])
        assert self.title in r.text

    def tearDown(self):
        self.proc.terminate()
        self.proc.wait()

if __name__ == '__main__':
    if os.path.dirname(sys.argv[0]):
        os.chdir(os.path.dirname(sys.argv[0]))
    suite = unittest.TestLoader().loadTestsFromTestCase(Test)
    rs = unittest.TextTestRunner(verbosity=2).run(suite)
    if len(rs.errors) > 0 or len(rs.failures) > 0:
        sys.exit(10)
    else:
        sys.exit(0)
