# Helper script to download a precompiled binary of the wasmtime dll for the
# current platform. Currently always downloads the dev release of wasmtime.

import urllib.request
import zipfile
import tarfile
import io
import sys
import os
import shutil

is_zip = False
if sys.platform == 'linux':
    filename = 'wasmtime-dev-x86_64-linux-c-api.tar.xz'
elif sys.platform == 'win32':
    filename = 'wasmtime-dev-x86_64-windows-c-api.zip'
    is_zip = True
elif sys.platform == 'darwin':
    filename = 'wasmtime-dev-x86_64-macos-c-api.tar.xz'
else:
    raise RuntimeError("unknown platform: " + sys.platform)

url = 'https://github.com/bytecodealliance/wasmtime/releases/download/dev/'
url += filename
print('Download', url)

with urllib.request.urlopen(url) as f:
    contents = f.read()

if is_zip:
    t = zipfile.ZipFile(io.BytesIO(contents))
    t.extractall()
else:
    t = tarfile.open(fileobj=io.BytesIO(contents))
    t.extractall()

try:
    shutil.rmtree('pkg/wasmtime/build')
except FileNotFoundError:
    pass
os.rename(filename.replace('.zip', '').replace('.tar.xz', ''), 'pkg/wasmtime/build')
