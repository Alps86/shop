"""Echter Übergangsbeleg mit temporärem Cucumber-Vertrag, nie Shop-Produktivcode."""
from pathlib import Path


class NativeRegression:
    def __init__(self, regression):
        self.regression = regression

    def prepare(self):
        probe = self.regression
        probe.write('unexpected/source.go', 'package unexpected\nfunc Value() int { return 1 }\n')
        probe.write('spec/features/probe.feature', probe.feature())
        probe.phase('implementation', {'unexpected': '@slice-probe'})
        path = Path('cucumber/feature_test.go')
        source = (probe.gate.root / path).read_text()
        source = source.replace('Name: "shop", Options: options', 'Name: "shop", Options: options, ScenarioInitializer: registerProbe')
        probe.write(str(path), source)
        probe.write('cucumber/probe_test.go', self.binding())

    def binding(self):
        return '''package cucumber
import (
 "fmt"
 "github.com/cucumber/godog"
 "github.com/Alps86/shop/unexpected"
)
func registerProbe(ctx *godog.ScenarioContext) {
 ctx.Step(`^eine unbekannte Probe ausgeführt wird$`, func() error {
  if unexpected.Value() != 1 { return fmt.Errorf("unerwarteter Probewert") }
  return nil
 })
}
'''

    def execute(self):
        self.prepare()
        probe = self.regression
        operation = lambda: probe.gate.production_coverage(probe.gate.inventory()[0])
        probe.expect('echte Cucumber-Bindung und vollständige native Coverage', operation)
        probe.write('unexpected/uncovered.go', 'package unexpected\nfunc Uncovered() int { return 2 }\n')
        probe.expect('echter nicht ausgeführter Produktivblock', operation, True)
        (probe.gate.root / 'unexpected/uncovered.go').unlink()
        probe.write('unimported/source.go', 'package unimported\nfunc Uncovered() int { return 2 }\n')
        probe.phase('implementation', {'unexpected': '@slice-probe', 'unimported': '@slice-probe'})
        probe.expect('echtes nicht importiertes Produktivpackage', operation, True)
