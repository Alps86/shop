"""Die Diffprüfung muss ungecommitete, staged und PR-Änderungen tatsächlich sehen."""


class DiffRegression:
    def __init__(self, regression):
        self.regression = regression

    def prepare(self):
        gate = self.regression.gate
        gate.run('git', 'init', '-q')
        gate.run('git', 'config', 'user.name', 'Quality Regression')
        gate.run('git', 'config', 'user.email', 'quality@example.invalid')
        gate.run('git', 'add', '.')
        gate.run('git', 'commit', '-qm', 'Technischer Ausgangspunkt')
        gate.run('git', 'update-ref', 'refs/remotes/origin/main', 'HEAD')

    def execute(self):
        self.prepare()
        probe = self.regression
        operation = lambda: probe.gate.diff('origin/main')
        probe.expect('sauberer tatsächlicher Diff', operation)
        probe.write('whitespace.md', 'Probe mit Leerraum  \n')
        probe.gate.run('git', 'add', 'whitespace.md')
        probe.expect('Whitespace im staged Diff', operation, True)
        probe.gate.run('git', 'commit', '-qm', 'Technische Negativprobe')
        probe.expect('Whitespace im tatsächlichen PR-Diff', operation, True)
        probe.gate.run('git', 'update-ref', 'refs/remotes/origin/main', 'HEAD')
        probe.write('README.md', 'Unstaged Leerraum  \n')
        probe.expect('Whitespace im unstaged Diff', operation, True)
