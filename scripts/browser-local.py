#!/usr/bin/env python3
"""Optional Ubuntu 24.04 amd64 WebKit dependencies, without administrator access."""
import hashlib,json,pathlib,platform,shlex,subprocess,tempfile
ROOT=pathlib.Path(__file__).resolve().parent.parent
if platform.machine()!='x86_64' or 'VERSION_ID="24.04"' not in pathlib.Path('/etc/os-release').read_text():raise SystemExit('This optional fallback supports Ubuntu 24.04 amd64 only; use Playwright install-deps on other hosts.')
directory=ROOT/'.runtime/browser-deps';directory.mkdir(parents=True,exist_ok=True);target=directory/'root';target.mkdir(exist_ok=True)
for package in json.loads((ROOT/'scripts/toolchain/browser-deps-ubuntu24-amd64.json').read_text()):
 fields=dict(line.split(': ',1) for line in package['metadata'].splitlines())
 with tempfile.TemporaryDirectory(prefix='gati-browser-deb-') as temporary:
  subprocess.run(['apt-get','download',fields['Package']+'='+fields['Version']],cwd=temporary,check=True)
  files=list(pathlib.Path(temporary).glob('*.deb'))
  if len(files)!=1 or hashlib.sha256(files[0].read_bytes()).hexdigest()!=package['sha256']:raise SystemExit('browser dependency checksum mismatch')
  subprocess.run(['dpkg-deb','-x',str(files[0]),str(target)],check=True)
subprocess.run(['npx','playwright','install','firefox','webkit'],cwd=ROOT/'web',check=True)
executable=subprocess.check_output(['node','--input-type=module','-e','import {webkit} from "playwright";console.log(webkit.executablePath())'],cwd=ROOT/'web',text=True).strip()
wk=pathlib.Path(executable).parent/'minibrowser-wpe';libs=target/'usr/lib/x86_64-linux-gnu'
# The upstream bundle wrapper replaces LD_LIBRARY_PATH. This separate wrapper keeps
# its original binaries and adds the verified local libraries, excluding Snap paths.
wrapper=directory/'webkit-local'
wrapper.write_text('#!/bin/sh\nset -eu\nunset GTK_PATH GTK_EXE_PREFIX GTK_DATA_PREFIX GIO_MODULE_DIR GIO_EXTRA_MODULES GSETTINGS_SCHEMA_DIR LD_PRELOAD\n'+
 '\n'.join('export '+k+'='+shlex.quote(str(v)) for k,v in {'WEBKIT_EXEC_PATH':wk/'bin','WEBKIT_INJECTED_BUNDLE_PATH':wk/'lib','WEBKIT_INSPECTOR_RESOURCES_PATH':wk/'share','WEBKIT_FORCE_COMPLEX_TEXT':'1','LD_LIBRARY_PATH':f'{libs}:{wk}/lib:{wk}/sys/lib','GST_PLUGIN_PATH':libs/'gstreamer-1.0'}.items())+'\nexec '+shlex.quote(str(wk/'bin/MiniBrowser'))+' "$@"\n');wrapper.chmod(0o700)
print('Run: GATI_WEBKIT_EXECUTABLE='+shlex.quote(str(wrapper))+' make test-browser-matrix')
