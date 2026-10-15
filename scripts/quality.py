#!/usr/bin/env python3
"""Reproduzierbare Quality-Gates; Produktivinventar unabhängig von go list."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile

from inventory import Inventory

os.environ.setdefault('GOCACHE', '/tmp/shop-go-build')
ROOT = Path(__file__).resolve().parents[1]
RESULTS = ROOT / '.quality/results'
JOBS = ('static', 'build', 'cucumber-and-coverage', 'race')


class Quality(Inventory):
    @property
    def root(self):
        return ROOT

    def run(self, *args, **kwargs):
        return subprocess.run(args, cwd=ROOT, check=True, text=True, **kwargs)

    def read(self, path):
        return json.loads((ROOT / path).read_text())

    def fingerprint(self):
        digest = hashlib.sha256()
        for path in sorted(ROOT.rglob('*')):
            if self.fingerprinted(path):
                digest.update(str(path.relative_to(ROOT)).encode())
                digest.update(path.read_bytes())
        return digest.hexdigest()

    def fingerprinted(self, path):
        if not path.is_file():
            return False
        relative = path.relative_to(ROOT)
        if '.git' in relative.parts or '__pycache__' in relative.parts:
            return False
        return not str(relative).startswith(('.quality/results/', '.qlty/out/'))

    def execute(self, job):
        sources, phase = self.inventory()
        RESULTS.mkdir(exist_ok=True)
        (RESULTS / (job + '.json')).unlink(missing_ok=True)
        print(f"Phase: {phase['phase']}; Produktivdateien: {len(sources)}", flush=True)
        getattr(self, job.replace('-', '_'))(sources)
        RESULTS.mkdir(exist_ok=True)
        result = dict(job=job, fingerprint=self.fingerprint(), phase=phase['phase'], sources=sources)
        (RESULTS / (job + '.json')).write_text(json.dumps(result, indent=2) + '\n')

    def static(self, sources):
        self.diff()
        self.run('qlty', 'config', 'validate')
        self.run('qlty', 'check', '--all', '--no-fix', '--no-cache', '--summary')
        report = self.run('qlty', 'smells', '--all', '--json', capture_output=True)
        if json.loads(report.stdout):
            raise ValueError('Offene Maintainability-Befunde: ' + report.stdout)
        self.selftest(sources)

    def diff(self, base=None):
        self.run('git', 'diff', '--check')
        self.run('git', 'diff', '--cached', '--check')
        base = base or os.environ.get('SHOP_BASE_SHA', 'origin/main')
        self.run('git', 'diff', '--check', base + '...HEAD')

    def tool_command(self, race=False):
        command = ['go', 'run']
        if race:
            command.append('-race')
        command.extend(str(p) for p in sorted((ROOT / '.qlty/checks/type-spacing').glob('*.go')))
        return command

    def build(self, sources):
        with tempfile.TemporaryDirectory() as directory:
            files = self.tool_command()[2:]
            self.run('go', 'build', '-o', directory + '/stylecheck', *files)
            self.run('go', 'test', '-c', '-o', directory + '/cucumber.test', './cucumber')
        if sources:
            self.run('go', 'build', './...')

    def syntax(self):
        self.run('go', 'test', './cucumber', '-run', '^TestFeatureSyntax$', '-count=1')
        features = list((ROOT / 'spec').rglob('*.feature'))
        print(f'Gherkin-Syntax: {len(features)} Dateien; kein fachlicher Verhaltensnachweis')

    def cucumber_and_coverage(self, sources):
        self.syntax()
        self.selftest(sources)
        if not sources:
            print('Produktive Coverage: nicht anwendbar; keine fachlichen Szenarien ausgeführt')
            return
        self.production_coverage(sources)

    def package_names(self, sources):
        module = self.run('go', 'list', '-m', capture_output=True).stdout.strip()
        expected = {module + '/' + str(Path(p).parent) for p in sources}
        expected = {p.removesuffix('/.') for p in expected}
        listed = set(self.run('go', 'list', './...', capture_output=True).stdout.split())
        if not expected.issubset(listed):
            raise ValueError('Go-Inventarlücke: ' + str(expected - listed))
        self.verify_visible_sources(sources)
        return ','.join(sorted(expected))

    def verify_visible_sources(self, sources):
        template = '{{.Dir}}|{{join .GoFiles ","}}|{{join .CgoFiles ","}}'
        result = self.run('go', 'list', '-f', template, './...', capture_output=True)
        visible = set()
        for line in result.stdout.splitlines():
            directory, normal, cgo = line.split('|')
            names = filter(None, (normal + ',' + cgo).split(','))
            visible.update(str((Path(directory) / name).relative_to(ROOT)) for name in names)
        if visible != set(sources):
            raise ValueError('Von Go nicht gemessene Produktivdateien: ' + str(set(sources) - visible))

    def production_coverage(self, sources):
        with tempfile.TemporaryDirectory() as directory:
            packages = self.package_names(sources)
            inventory = directory + '/inventory.out'
            profile = directory + '/coverage.out'
            report = directory + '/cucumber.json'
            self.run('go', 'test', './...', '-run', '^$', '-coverpkg=' + packages,
                     '-covermode=atomic', '-coverprofile=' + inventory)
            environment = dict(os.environ, SHOP_CUCUMBER_REPORT=report)
            self.run('go', 'test', './cucumber', '-run', '^TestFeatures$', '-count=1',
                     '-coverpkg=' + packages, '-covermode=atomic', '-coverprofile=' + profile, env=environment)
            self.verify_report(json.loads(Path(report).read_text()))
            self.verify_coverage(Path(inventory).read_text(), Path(profile).read_text())
            print('Produktive Go-Statement-Coverage: 100%; alle inventarisierten Blöcke ausgeführt')

    def verify_report(self, report):
        scenarios = [s for feature in report for s in feature.get('elements', [])]
        steps = [s for scenario in scenarios for s in scenario.get('steps', [])]
        if not steps:
            raise ValueError('Keine tatsächlich ausgeführten fachlichen Schritte')
        if any(s.get('result', {}).get('status') != 'passed' for s in steps):
            raise ValueError('Fehlende Bindung oder fachlicher Cucumber-Fehler')

    def blocks(self, text):
        lines = text.strip().splitlines()
        if not lines or lines[0] != 'mode: atomic':
            raise ValueError('Natives atomic-Coverageprofil fehlt')
        result = {}
        for line in lines[1:]:
            location, statements, count = line.split()
            previous = result.get(location, (int(statements), 0))
            if previous[0] != int(statements):
                raise ValueError('Widersprüchlicher Coverageblock: ' + location)
            result[location] = (int(statements), previous[1] + int(count))
        return result

    def verify_coverage(self, inventory, actual):
        expected = self.blocks(inventory)
        measured = self.blocks(actual)
        if not expected:
            raise ValueError('Produktiver Messumfang enthält keine Blöcke')
        if set(expected) != set(measured):
            raise ValueError('Fehlende oder unerwartete Coverageblöcke')
        for location, (statements, _) in expected.items():
            count = measured[location]
            if count[0] != statements or count[1] <= 0:
                raise ValueError('Produktivblock nicht vollständig ausgeführt: ' + location)

    def race(self, sources):
        self.run(*self.tool_command(True), '--regression', '.qlty/checks/type-spacing/fixtures.json')
        self.run('go', 'test', '-race', './cucumber', '-run', '^TestFeatureSyntax$', '-count=1')
        if sources:
            self.run('go', 'test', '-race', './cucumber', '-run', '^TestFeatures$', '-count=1',
                     env=dict(os.environ, SHOP_CUCUMBER_REPORT=str(RESULTS / 'race-cucumber.json')))
        print('Race: echtes Prüftooling' + (' und fachliche Cucumber-Verträge' if sources else '; keine Shop-Implementierung'))

    def quality_gate(self, sources):
        for job in JOBS:
            report = self.read('.quality/results/' + job + '.json')
            if report['fingerprint'] != self.fingerprint():
                raise ValueError('Veralteter/abweichender Nachweis: ' + job)
            if report['sources'] != sources:
                raise ValueError('Abweichendes Produktivinventar: ' + job)
        print('Alle vier Teilgates mit identischem Stand und Inventar erfolgreich')

    def selftest(self, sources):
        self.run(*self.tool_command(), '--regression', '.qlty/checks/type-spacing/fixtures.json')
        self.run('python3', 'scripts/quality_test.py')


if __name__ == '__main__':
    try:
        Quality().execute(sys.argv[1])
    except (ValueError, subprocess.CalledProcessError, OSError, KeyError) as error:
        print('Quality-Gate fehlgeschlagen:', error, file=sys.stderr)
        sys.exit(1)
