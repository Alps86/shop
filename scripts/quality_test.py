#!/usr/bin/env python3
"""Technische Positiv-/Negativproben in einem entfernten temporären Repository."""
import json
from pathlib import Path
import shutil
import tempfile

import quality
from native_regression import NativeRegression
from diff_regression import DiffRegression


class Regression:
    def __init__(self, directory):
        self.original = quality.ROOT
        quality.ROOT = Path(directory)
        quality.RESULTS = quality.ROOT / '.quality/results'
        shutil.copytree(self.original, directory, dirs_exist_ok=True,
                        ignore=self.ignored)
        self.gate = quality.Quality()
        self.passed = 0

    def ignored(self, directory, names):
        ignored = shutil.ignore_patterns('.git', 'results', '__pycache__')(directory, names)
        if Path(directory) == self.original:
            ignored.add('.golangci.yml')
        if Path(directory) == self.original / '.qlty':
            ignored.update(('logs', 'out', 'results', 'plugin_cachedir'))
        return ignored

    def expect(self, name, operation, fails=False):
        failed = False
        try:
            operation()
        except (ValueError, OSError, quality.subprocess.CalledProcessError) as error:
            failed = True
            if not fails:
                print(error, flush=True)
        if failed != fails:
            raise ValueError('Regression fehlgeschlagen: ' + name)
        self.passed += 1
        print('Regression erfolgreich:', name, flush=True)

    def write(self, name, text):
        path = quality.ROOT / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(text)

    def phase(self, phase, slices):
        self.write('.quality/phase.json', json.dumps(dict(phase=phase, slices=slices)))

    def inventory(self):
        self.expect('leerer Bootstrap', self.gate.inventory)
        self.write('unexpected/source.go', 'package unexpected\nfunc Value() int { return 1 }\n')
        self.expect('unbekannter Produktivpfad erzwingt Transition', self.gate.inventory, True)
        self.phase('implementation', {'unexpected': '@slice-probe'})
        self.expect('fehlender fachlicher Vertrag', self.gate.inventory, True)
        self.write('spec/features/probe.feature', self.feature())
        self.expect('vollständige Transition', self.gate.inventory)
        self.write('spec/features/probe.feature', '@spec-only\n' + self.feature())
        self.expect('aktiver Slice hinter spec-only', self.gate.inventory, True)
        self.write('spec/features/probe.feature', self.feature())
        self.inventory_limits()

    def inventory_limits(self):
        self.write('unexpected/source.go', '//go:build hidden\n\npackage unexpected\n')
        self.expect('nicht gemessene Build-Tags', self.gate.inventory, True)
        self.write('unexpected/source.go', 'package unexpected\nfunc Value() int { return 1 }\n')
        self.write('extra/go.mod', 'module hidden\ngo 1.26\n')
        self.expect('verstecktes zweites Modul', self.gate.inventory, True)
        (quality.ROOT / 'extra/go.mod').unlink()
        (quality.ROOT / 'unexpected/source.go').unlink()
        self.expect('unzulässiger Rückfall in leeres Inventar', self.gate.inventory, True)
        self.phase('bootstrap', {})
        self.write('.hidden/source.go', 'package hidden\n')
        self.expect('von go list übergangener Pfad', self.gate.inventory, True)
        (quality.ROOT / '.hidden/source.go').unlink()
        self.tooling_limits()

    def tooling_limits(self):
        tooling = self.gate.read('.quality/tooling.json')
        changed = dict(tooling, **{'unexpected/source.go': 'fake'})
        self.write('.quality/tooling.json', json.dumps(changed))
        self.expect('erweiterte Tooling-Ausnahme', self.gate.inventory, True)
        self.write('.quality/tooling.json', json.dumps(tooling))
        self.write('.qlty/checks/type-spacing/main.go', 'package main\n')
        self.expect('geänderter Toolinginhalt', self.gate.inventory, True)
        original = self.original / '.qlty/checks/type-spacing/main.go'
        self.write('.qlty/checks/type-spacing/main.go', original.read_text())

    def coverage(self):
        inventory = 'mode: atomic\na.go:1.1,2.1 1 0\nb.go:1.1,2.1 2 0\n'
        measured = inventory.replace(' 0', ' 1')
        self.expect('vollständige native Blöcke', lambda: self.gate.verify_coverage(inventory, measured))
        self.expect('fehlender Block', lambda: self.gate.verify_coverage(inventory, measured.split('b.go')[0]), True)
        self.expect('nicht ausgeführter Block', lambda: self.gate.verify_coverage(inventory, inventory), True)
        self.expect('falsche Statementzahl', lambda: self.gate.verify_coverage(inventory, measured.replace(' 2 ', ' 3 ')), True)
        self.expect('fehlende echte Szenarien', lambda: self.gate.verify_report([]), True)
        self.expect('undefinierter Schritt', lambda: self.gate.verify_report([{'elements': [{'steps': [{'result': {'status': 'undefined'}}]}]}]), True)

    def results(self):
        paths = []
        for job in quality.JOBS:
            name = '.quality/results/' + job + '.json'
            self.write(name, json.dumps(dict(fingerprint=self.gate.fingerprint(), sources=[])))
            paths.append(quality.ROOT / name)
        self.expect('Teilgate-Nachweise am identischen Stand', lambda: self.gate.quality_gate([]))
        paths[0].write_text(json.dumps(dict(fingerprint='stale', sources=[])))
        self.expect('veralteter Gate-Nachweis', lambda: self.gate.quality_gate([]), True)
        paths[0].unlink()
        self.expect('fehlender Gate-Nachweis', lambda: self.gate.quality_gate([]), True)

    def feature(self):
        return '# language: de\n@slice-probe\nFunktionalität: Technische Probe\n  Szenario: Bindung prüfen\n    Dann eine unbekannte Probe ausgeführt wird\n'

    def binding(self):
        self.write('spec/features/probe.feature', self.feature())
        report = str(quality.ROOT / 'undefined.json')
        environment = dict(quality.os.environ, SHOP_CUCUMBER_REPORT=report)
        operation = lambda: self.gate.run('go', 'test', './cucumber', '-run', '^TestFeatures$', '-count=1', env=environment)
        self.expect('echte Godog-Ausführung ohne Stepbindung', operation, True)
        if 'undefined' not in Path(report).read_text():
            raise ValueError('Fehlschlag war kein fehlender Step-Vertrag')
        self.expect('echter Gherkin-Parser gültig', self.gate.syntax)
        self.write('spec/features/probe.feature', '# language: de\nFunktionalität: Probe\n  Szenario: ungültig\n    Dann Probe\n      | offen |\n      | kaputt | zusatz |\n')
        self.expect('echter Gherkin-Parser ungültig', self.gate.syntax, True)

    def execute(self):
        self.inventory()
        self.coverage()
        self.results()
        self.binding()
        NativeRegression(self).execute()
        DiffRegression(self).execute()
        print(f'{self.passed} technische Gate-Regressionen erfolgreich; Proben entfernt nach Prozessende')


if __name__ == '__main__':
    with tempfile.TemporaryDirectory(prefix='shop-quality-') as directory:
        Regression(directory).execute()
