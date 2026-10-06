#!/usr/bin/env python3
"""Exercise the real CLI through a PTY, without adding test-only CLI flags."""
import errno
import fcntl
import os
from pathlib import Path
import pty
import re
import select
import struct
import subprocess
import sys
import termios
import time

mode = sys.argv[1]
root = Path.cwd() / "repo"
(root / "scripts").mkdir(parents=True)
subprocess.run(["git", "init", "-q", str(root)], check=True)
script = root / "scripts" / "build task.sh"
script.write_text(
    "#!/usr/bin/env bash\nset -euo pipefail\nprintf 'SCRIPT-PROMPT\\n'\nread answer\nprintf 'answer=%s\\n' \"$answer\"\nprintf done > ran.txt\n"
    if mode == "run" else "printf done > ran.txt\nexit 7\n"
)
if mode != "cancel":
    (root / "other.py").write_text("raise Exception('wrong selection')\n")
    (root / "ignored.sh").write_text("exit 99\n")
    (root / ".gitignore").write_text("ignored.sh\n")

if mode == "nonterminal":
    result = subprocess.run(["scriptor"], cwd=root / "scripts", capture_output=True)
    sys.stdout.buffer.write(result.stdout + result.stderr)
    sys.exit(result.returncode)

master, slave = pty.openpty()
fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 24, 100, 0, 0))
before = termios.tcgetattr(slave)
child = subprocess.Popen(["scriptor"], cwd=root / "scripts", stdin=slave, stdout=slave, stderr=slave)
captured = bytearray()
stage = 0
deadline = time.monotonic() + 15
try:
    while time.monotonic() < deadline:
        ready, _, _ = select.select([master], [], [], 0.1)
        if ready:
            try:
                data = os.read(master, 65536)
            except OSError as error:
                if error.errno == errno.EIO:
                    break
                raise
            if not data:
                break
            captured.extend(data)
        if stage == 0 and b"Filter:" in captured:
            if mode == "cancel":
                # A sole result must stay open until explicitly cancelled.
                time.sleep(0.2)
                if child.poll() is not None:
                    raise RuntimeError("sole result was executed automatically")
                os.write(master, b"\x1b")
                stage = 3
            else:
                os.write(master, b"bldtsk")
                stage = 1
        elif stage == 1 and b"1/2 scripts" in captured:
            os.write(master, b"\r")
            stage = 2
        elif stage == 2 and b"SCRIPT-PROMPT" in captured:
            os.write(master, b"hello\n")
            stage = 3
        if child.poll() is not None and not ready:
            break
    if child.poll() is None:
        raise RuntimeError("interactive CLI timed out")
    if termios.tcgetattr(slave) != before:
        raise RuntimeError("CLI did not restore terminal settings")
    if re.search(rb"\x1b\[\?(?:47|1047|1049)[hl]", captured):
        raise RuntimeError("picker switched to the alternate screen")
    if not re.search(rb"\x1b\[[0-9;]*(?:3[0-7]|9[0-7]|38)(?:;[0-9]+)*m", captured):
        raise RuntimeError("picker did not render colors")
    sys.stdout.buffer.write(captured)
    sys.exit(child.returncode)
finally:
    if child.poll() is None:
        child.kill()
        child.wait()
    os.close(master)
    os.close(slave)
