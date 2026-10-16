#!/usr/bin/env python3
"""Einmaliger Runnerauswerter mit geschlossenem JSON-Export."""
import csv
import decimal
import json
from pathlib import Path
import re
import sys
from backend import BackendReader
from jvm import JVMReader
from trace import TraceReader
from schema import (COUNTERS, ERRORS, METRICS, PHASES, PROFILES, REGIONS, SQL, TOOK, Failure,
                    atomic, check_pprof, check_status, check_summary, number,
                    require, safe_symbol, statistics, status, timestamp, verify)


class Aggregate(TraceReader, JVMReader, BackendReader):
    def __init__(self, root):
        self.root = root
        self.values = {}
        self.counts = dict.fromkeys(COUNTERS, 0)
        self.tasks = {}
        self.stacks = {}
        self.sync = None
        self.start = timestamp((root / 'load-start.utc').read_text().strip())
        self.end = timestamp((root / 'load-end.utc').read_text().strip())
        require(self.end >= self.start)
        self.rows = []
        self.tail = set()

    def metric(self, name, value, unit='ns', cohort='all'):
        require(name in METRICS)
        number(value)
        self.values.setdefault((name, cohort, unit), []).append(value)

    def csv(self):
        with (self.root / 'load.csv').open() as source:
            reader = csv.DictReader(source)
            require(reader.fieldnames == ['response-time', 'DNS+dialup', 'DNS', 'Request-write', 'Response-delay', 'Response-read', 'status-code', 'offset'])
            for row in reader:
                self.counts['clientSamples'] += 1
                self.counts['clientErrors'] += int(row['status-code'] != '200')
                self.metric('client.response', int(decimal.Decimal(row['response-time']) * 1000000000))

    def pprof(self):
        symbols = set()
        with (self.root / 'symbols.txt').open() as source:
            for line in source:
                parts = line.split()
                if len(parts) == 3 and safe_symbol(parts[2]):
                    symbols.add(parts[2])
        for profile in sorted(PROFILES):
            with (self.root / ('go-' + profile + '-top.txt')).open() as source:
                for line in source:
                    self.pprof_line(profile, line, symbols)
        self.cpu_tags()

    def pprof_line(self, profile, line, symbols):
        pattern = r'\s*([0-9.e+-]+)(?:ns)?\s+([0-9.]+)%\s+([0-9.]+)%\s+([0-9.e+-]+)(?:ns)?\s+([0-9.]+)%\s+(\S+)(?: \(inline\))?\s*'
        match = re.fullmatch(pattern, line)
        if match is None:
            return
        if match[6] not in symbols:
            self.counts['discardedSymbols'] += 1
            return
        self.rows.append(dict(profile=profile, function=match[6], flatNs=float(match[1]), flatPercent=float(match[2]), sumPercent=float(match[3]), cumNs=float(match[4]), cumPercent=float(match[5])))

    def cpu_tags(self):
        active = False
        with (self.root / 'go-cpu-tags.txt').open() as source:
            for line in source:
                if 'Total' in line:
                    active = line.startswith(' phase: Total ')
                match = re.fullmatch(r'\s*(\d+)ns \([0-9.]+%\): ([a-z.-]+)\s*', line)
                if active and match and match[2] in REGIONS:
                    self.metric('cpu.' + match[2], int(match[1]))
        with (self.root / 'go-cpu-top.txt').open() as source:
            for line in source:
                self.cpu_total(line)

    def cpu_total(self, line):
        match = re.search(r'Duration: ([0-9.]+)([a-z]+), Total samples = ([0-9.e+-]+)ns', line)
        if match:
            scale = {'ns': 1, 'us': 1000, 'ms': 1000000, 's': 1000000000}
            require(match[2] in scale)
            self.metric('cpu.duration', float(match[1]) * scale[match[2]])
            self.metric('cpu.total', float(match[3]))

    def optional(self, name, operation):
        before, counts = set(self.values), dict(self.counts)
        try:
            operation()
        except Exception:
            self.values = {key: value for key, value in self.values.items() if key in before}
            self.counts = counts
            self.counts[name] = 1

    def complete(self):
        expected = dict(requests=1000, requestIds=1000, brokenPairs=0, missingSync=0, clientErrors=0, clientSamples=1000)
        return all(self.counts[name] == value for name, value in expected.items())

    def result(self):
        overlap = self.boundaries()
        complete = int(self.complete())
        for name in sorted(REGIONS | TOOK | {'http.request', 'cpu.total', 'cpu.duration'} | {'cpu.' + name for name in REGIONS}):
            self.values.setdefault((name, 'all', 'ns'), [])
        summary = dict(schemaVersion=1, expectedRequests=1000, complete=complete, window=dict(startUnixNs=self.start, endUnixNs=self.end, overlapNs=overlap), counters=self.counts, metrics=[self.statistics(key, value) for key, value in sorted(self.values.items())])
        return summary, dict(schemaVersion=1, rows=self.rows)

    def statistics(self, key, values):
        expected = 1
        if key[0] in REGIONS | TOOK | {'http.request'}:
            expected = self.counts['requests']
        if key[1] == 'http-tail':
            expected = len(self.values.get(('http.request', 'http-tail', 'ns'), []))
        return statistics(key, values, expected)




def aggregate(root, approved):
    current = Aggregate(root)
    for operation in (current.trace, current.csv, current.pprof):
        operation()
    for name, operation in (('unmeasuredProxy', current.proxy), ('unmeasuredPG', current.pg), ('unmeasuredJVM', current.jfr)):
        current.optional(name, operation)
    summary, profiles = current.result()
    check_summary(summary)
    check_pprof(profiles)
    atomic(approved / 'summary.json', summary)
    atomic(approved / 'pprof.json', profiles)
    error = 'none'
    if any(current.counts[name] for name in ('unmeasuredProxy', 'unmeasuredPG', 'unmeasuredJVM')):
        error = 'incomplete-data'
    status(approved, 'aggregate', error, 1)
    require(summary['complete'] == 1, 'incomplete-data')


def export_aggregate(root, approved):
    try:
        aggregate(root, approved)
    except Exception as error:
        kind = 'invalid-structure'
        if isinstance(error, Failure) and str(error) in ERRORS:
            kind = str(error)
        present = int((approved / 'summary.json').exists() and (approved / 'pprof.json').exists())
        if not present:
            (approved / 'summary.json').unlink(missing_ok=True)
            (approved / 'pprof.json').unlink(missing_ok=True)
        status(approved, 'aggregate', kind, present)
        raise Failure(kind) from None


def finish(arguments):
    path, code, phase, cleaned = arguments
    output = Path(path)
    current = json.loads((output / 'status.json').read_text())
    if int(code) and current['errorClass'] == 'none':
        current.update(phase=phase, errorClass='command-failed')
        if phase == 'guard':
            current['errorClass'] = 'invocation-rejected'
    if int(cleaned):
        current.update(phase='cleanup', errorClass='cleanup-failed')
    current['success'] = int(not int(code) and not int(cleaned) and current['dataReady'] == 1 and current['errorClass'] == 'none')
    if current['success']:
        current.update(phase='finished', errorClass='none')
    check_status(current)
    atomic(output / 'status.json', current)


def phase_status(path, phase):
    require(phase in PHASES)
    value = json.loads((path / 'status.json').read_text())
    check_status(value)
    value['phase'] = phase
    atomic(path / 'status.json', value)


def main(arguments):
    if arguments[0] == '--verify-approved':
        verify(Path(arguments[1]))
        return
    if arguments[0] == '--status':
        status(Path(arguments[1]), arguments[2], arguments[3])
        return
    if arguments[0] == '--phase':
        phase_status(Path(arguments[1]), arguments[2])
        return
    if arguments[0] == '--finish':
        finish(arguments[1:])
        return
    export_aggregate(Path(arguments[0]), Path(arguments[1]))


if __name__ == '__main__':
    try:
        main(sys.argv[1:])
    except Exception:
        sys.exit(1)
