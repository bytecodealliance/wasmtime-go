# Helper script to download a precompiled binary of the wasmtime dll for the
# current platform. Currently always downloads the dev release of wasmtime.

import urllib.request
import zipfile
import tarfile
import io
import os
import shutil
import glob


version = '2.0.0'
urls = [
    ['wasmtime-v{}-x86_64-mingw-c-api.zip', 'windows-x86_64'],
    ['wasmtime-v{}-x86_64-linux-c-api.tar.xz', 'linux-x86_64'],
    ['wasmtime-v{}-x86_64-macos-c-api.tar.xz', 'macos-x86_64'],
    ['wasmtime-v{}-aarch64-linux-c-api.tar.xz', 'linux-aarch64'],
    ['wasmtime-v{}-aarch64-macos-c-api.tar.xz', 'macos-aarch64'],
]

try:
    # keep directory structure, only iremove files
    for subdir, dirs, files in os.walk("build"):
        dir_name = os.path.basename(os.path.normpath(subdir))
        for file in files:
            os.unlink(os.path.join(subdir, file))
except FileNotFoundError:
    os.makedirs('build')

for i, arr in enumerate(urls):
    filename, dirname = arr
    filename = filename.format(version)
    url = 'https://github.com/bytecodealliance/wasmtime/releases/download/v{}/'
    url += filename
    url = url.format(version)
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
        shutil.copytree(src + '/include', 'build/include', dirs_exist_ok=True)

    shutil.copytree(src + '/lib', 'build/' + dirname, dirs_exist_ok=True)
    shutil.rmtree(src)

for dylib in glob.glob("build/**/*.dll"):
    os.remove(dylib)
for dylib in glob.glob("build/**/*.dll.a"):
    os.remove(dylib)
for dylib in glob.glob("build/**/*.dylib"):
    os.remove(dylib)

for subdir, dirs, files in os.walk("build"):
    dir_name = os.path.basename(os.path.normpath(subdir))
    file_path = os.path.join(subdir, "empty.go")
    with open(file_path, "w") as empty_file:
        empty_file.write("package %s\n" % dir_name.replace("-", "_"))
