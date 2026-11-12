#!/usr/bin/env python3
"""Assert the standalone application's ownership and public dependency boundary."""
from pathlib import Path
import re

root = Path(__file__).resolve().parents[1]
errors = []
for name in ('services/identity', 'services/wallet', 'runtime', 'httpapi', 'cmd/dia'):
    if not (root / name).is_dir():
        errors.append(f'application implementation missing: {name}')
module = (root / 'go.mod').read_text()
if re.search(r'^replace\b', module, re.M) or (root / 'go.work').exists():
    errors.append('published consumer must resolve its pinned public dependency independently')
for path in root.rglob('*.go'):
    if any(part in ('.git', 'vendor') for part in path.parts):
        continue
    text = path.read_text()
    if re.search(r'"github\.com/mejango/blossm-toolbox-go/(?:dia|wallet|diaruntime|serviceapp|internal)(?:/|\")', text):
        errors.append(f'application uses framework-owned app code or private API: {path.relative_to(root)}')
if errors:
    raise SystemExit('\n'.join(errors))
print('Independent application ownership boundary passes.')
