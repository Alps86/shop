"""Vollständige Produktivquellen und geprüfter Phasenübergang."""
import hashlib
from pathlib import Path


class Inventory:
    def inventory(self):
        tooling = self.read('.quality/tooling.json')
        self.verify_tooling(tooling)
        sources = sorted(str(p.relative_to(self.root)) for p in self.root.rglob('*.go')
                         if self.is_product(p, tooling))
        self.verify_sources(sources)
        phase = self.read('.quality/phase.json')
        expected = 'implementation' if sources else 'bootstrap'
        if phase['phase'] != expected:
            raise ValueError('Unzulässiger Phasenwechsel: tatsächliches Inventar widerspricht phase.json')
        self.verify_slices(sources, phase['slices'])
        return sources, phase

    def is_product(self, path, tooling):
        relative = str(path.relative_to(self.root))
        if '.git' in path.relative_to(self.root).parts:
            return False
        if relative in tooling:
            return False
        return not path.name.endswith('_test.go')

    def verify_tooling(self, tooling):
        expected = {'.qlty/checks/type-spacing/' + name for name in
                    ('main.go', 'type.go', 'checker.go', 'regression.go')}
        if set(tooling) != expected:
            raise ValueError('Tooling-Ausnahme darf nur den versionierten AST-Stilchecker enthalten')
        for path, digest in tooling.items():
            if hashlib.sha256((self.root / path).read_bytes()).hexdigest() != digest:
                raise ValueError('Tooling geändert: Herkunft/Zweck prüfen und Fingerprint aktualisieren: ' + path)

    def verify_sources(self, sources):
        modules = sorted(str(p.relative_to(self.root)) for p in self.root.rglob('go.mod'))
        if modules != ['go.mod']:
            raise ValueError('Genau ein Root-Go-Modul erforderlich')
        for name in sources:
            source = (self.root / name).read_text()
            if '//go:build' in source or '// +build' in source:
                raise ValueError('Produktive Build-Tags benötigen zuerst vollständige Messmatrix: ' + name)
            if any(part.startswith(('.', '_')) for part in Path(name).parts):
                raise ValueError('Produktivquelle in von Go übergangenem Pfad: ' + name)
            if Path(name).parent == Path('cucumber'):
                raise ValueError('Harness ausschließlich in _test.go; sonst Produktivquelle: ' + name)

    def verify_slices(self, sources, slices):
        packages = {str(Path(p).parent) for p in sources}
        if set(slices) != packages:
            raise ValueError('Slicevertrag muss jedes tatsächliche Produktivpackage exakt enthalten')
        for tag in slices.values():
            self.verify_slice(tag)

    def verify_slice(self, tag):
        if not tag.startswith('@slice-'):
            raise ValueError('Slice benötigt expliziten @slice-Tag')
        matches = [p for p in (self.root / 'spec').rglob('*.feature') if tag in p.read_text().split()]
        if not matches:
            raise ValueError('Fachlicher Vertrag fehlt: ' + tag)
        for path in matches:
            if '@spec-only' in path.read_text():
                raise ValueError('Aktiver Slice darf nicht spec-only bleiben: ' + str(path))
