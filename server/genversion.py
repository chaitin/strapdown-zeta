#!/usr/bin/env python2
#-*- coding:utf-8 -*-

import os, sys, json, subprocess

CWD = os.path.dirname(os.path.realpath(__file__))

def get_version():
    version = json.load(open(os.path.join(CWD, '..', 'package.json')))
    return version['version']

def get_build():
    os.chdir(CWD)
    stdout = subprocess.check_output(['git', 'rev-parse', '--short=11', 'HEAD'])
    return stdout.strip()

if __name__ == '__main__':
    try:
        version = get_version()
        if len(sys.argv) > 1 and sys.argv[1] == 'build':
            version += '+' + get_build()
        print version
    except:
        import traceback
        exc_type, exc_obj, exc_tb = sys.exc_info()
        tbs = ''.join(traceback.format_exception(exc_type, exc_obj, exc_tb))
        print tbs
        sys.exit(10)
