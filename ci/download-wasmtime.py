# Helper script to download a precompiled binary of the wasmtime dll for the
# current platform. Currently always downloads the dev release of wasmtime.

import urllib.request
import zipfile
import tarfile
import io
import os
import shutil
import glob

urls = [
    ['wasmtime-dev-x86_64-mingw-c-api.zip', 'windows-x86_64'],
    ['wasmtime-dev-x86_64-linux-c-api.tar.xz', 'linux-x86_64'],
    ['wasmtime-dev-x86_64-macos-c-api.tar.xz', 'macos-x86_64'],
    ['wasmtime-dev-aarch64-linux-c-api.tar.xz', 'linux-aarch64'],
]

try:
    shutil.rmtree('build')
except FileNotFoundError:
    pass

os.makedirs('build')

for i, arr in enumerate(urls):
    filename, dirname = arr
    url = 'https://github.com/bytecodealliance/wasmtime/releases/download/dev/'
    url += filename
    print('Download', url)

    with urllib.request.urlopen(url) as f:
        contents = f.read()

    if filename.endswith('.zip'):
        z = zipfile.ZipFile(file=io.BytesIO(contents))
        z.extractall()
    else:
        t = tarfile.open(fileobj=io.BytesIO(contents))
        t.extractall()

    src = filename.replace('.zip', '').replace('.tar.xz', '')
    if i == 0:
        os.rename(src + '/include', 'build/include')

    os.rename(src + '/lib', 'build/' + dirname)
    shutil.rmtree(src)

for dylib in glob.glob("build/**/*.dll"):
    os.remove(dylib)
for dylib in glob.glob("build/**/*.dll.a"):
    os.remove(dylib)
for dylib in glob.glob("build/**/*.dylib"):
    os.remove(dylib)
for dylib in glob.glob("build/**/*.so"):
    os.remove(dylib)

for subdir, dirs, files in os.walk("build"):
    dir_name = os.path.basename(os.path.normpath(subdir))
    file_path = os.path.join(subdir, "empty.go")
    with open(file_path, "w") as empty_file:
        empty_file.write("package %s\n" % dir_name.replace("-", "_"))
