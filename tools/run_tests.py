#!/usr/bin/env python

import os
import sys
import yaml
import subprocess

with open(sys.argv[1]) as handle:
    config = yaml.load(handle.read())

basedir = os.path.dirname(os.path.abspath(sys.argv[1]))

cwl_cmd = "./bin/ktl-cwl"

TEST = [
0, 1, 2
]

for i, test in enumerate(config):
    if i in TEST:
        tool = os.path.join(basedir, test['tool'])
        job = os.path.join(basedir, test['job'])
        print "Test %d" % (i)
        print tool
        print job

        p = subprocess.Popen([cwl_cmd, "--quiet", tool, job], stdout=subprocess.PIPE)
        out, err = p.communicate()
        if p.returncode != 0:
            print "Exit on Error, %s %s " % (tool, job)
        print test['doc']
        print "output", out
        print "expected", test['output']
        print "-=-=-=-=-=-=-=-=-=-"
