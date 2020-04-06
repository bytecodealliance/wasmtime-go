# Helper script to download a precompiled binary of the wasmtime dll for the
# current platform. Currently always downloads the dev release of wasmtime.

import urllib.request
import zipfile
import tarfile
import io
import sys
import os
import shutil
import glob

urls = [
    ['wasmtime-dev-x86_64-linux-c-api.tar.xz', 'linux-x86_64'],
    ['wasmtime-dev-x86_64-macos-c-api.tar.xz', 'macos-x86_64'],
]

try:
    shutil.rmtree('pkg/wasmtime/build')
except FileNotFoundError:
    pass

os.makedirs('pkg/wasmtime/build')

for i, arr in enumerate(urls):
    filename, dirname = arr
    url = 'https://github.com/bytecodealliance/wasmtime/releases/download/dev/'
    url += filename
    print('Download', url)

    with urllib.request.urlopen(url) as f:
        contents = f.read()

    t = tarfile.open(fileobj=io.BytesIO(contents))
    t.extractall()

    src = filename.replace('.zip', '').replace('.tar.xz', '')
    if i == 0:
        os.rename(src + '/include', 'pkg/wasmtime/build/include')

    os.rename(src + '/lib', 'pkg/wasmtime/build/' + dirname)
    shutil.rmtree(src)

for dylib in glob.glob("pkg/wasmtime/build/**/*.dll"):
    os.remove(dylib)
for dylib in glob.glob("pkg/wasmtime/build/**/*.dylib"):
    os.remove(dylib)
for dylib in glob.glob("pkg/wasmtime/build/**/*.so"):
    os.remove(dylib)
