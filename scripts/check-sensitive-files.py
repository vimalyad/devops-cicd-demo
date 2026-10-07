#!/usr/bin/env python3
"""Reject committed credential-file names without reading or printing their contents."""
import json
from pathlib import PurePosixPath
import subprocess
import sys

paths = subprocess.check_output(['git', 'ls-files', '-z']).decode().split('\0')
tracked = [path for path in paths if path]
blocked = []
for path in tracked:
    name = PurePosixPath(path).name.lower()
    if name == '.env' or name.startswith('.env.') or name.endswith(('.pem', '.key')):
        blocked.append(path)

print(f'Checked {len(tracked)} tracked paths for environment and private-key files.')
if blocked:
    for path in blocked:
        print('Blocked tracked path: ' + json.dumps(path))
    print('FAIL: remove these files from the committed tree before building an image.')
    sys.exit(1)
print('PASS: no tracked .env, .env.*, .pem or .key files.')
