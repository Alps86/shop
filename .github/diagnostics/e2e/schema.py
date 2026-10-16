#!/usr/bin/env python3
"""Einmalige streaming-Auswertung; veröffentlicht ausschließlich geschlossene Aggregate."""
import datetime
import decimal
import json
import math
from pathlib import Path
import re

PHASES = {'guard', 'collectors', 'build', 'recreate', 'warmup', 'profile-start', 'load', 'profile-stop', 'tools', 'aggregate', 'export', 'cleanup', 'finished'}
ERRORS = {'none', 'invocation-rejected', 'command-failed', 'invalid-structure', 'incomplete-data', 'unsafe-output', 'cleanup-failed'}
REGIONS = {'http.query-parse', 'http.query-and-user', 'http.user-interests', 'sql.resolve-session', 'sql.active-cart', 'sql.user-payload', 'os.query-build-validate', 'os.encode', 'os.http-headers', 'os.http-body', 'os.decode-envelope', 'os.cards-facets-suggestions', 'http.prepare-render', 'http.template', 'http.response-write'}
SQL = {'sql.resolve-session', 'sql.active-cart', 'sql.user-payload'}
REGIONS |= {name + '.query' for name in SQL}
TOOK = {'os.outer-took-ms', 'os.main-took-ms', 'os.completion-took-ms'}
CATEGORIES = {'query', 'completion', 'fetch', 'stored-fields', 'compression', 'doc-values', 'json', 'other'}
EVENTS = {'jdk.ExecutionSample', 'jdk.NativeMethodSample', 'jdk.CPULoad', 'jdk.ThreadCPULoad', 'jdk.ObjectAllocationSample', 'jdk.GarbageCollection', 'jdk.GCPhasePause', 'jdk.GCHeapSummary', 'jdk.ThreadPark', 'jdk.JavaMonitorEnter', 'jdk.JavaMonitorWait', 'jdk.Checkpoint', 'jdk.Metadata'}
METRICS = REGIONS | TOOK | {'http.request', 'client.response', 'proxy.nginx', 'proxy.connect', 'proxy.headers', 'proxy.response', 'proxy.varnish', 'jvm.gc', 'jvm.gc-pause', 'jvm.park', 'jvm.monitor', 'jvm.cpu-load', 'jvm.thread-cpu-load'}
METRICS |= {'cpu.' + name for name in REGIONS} | {'cpu.total', 'cpu.duration'}
METRICS |= {prefix + name for prefix in ('jvm.execution.', 'jvm.native.', 'jvm.allocation.') for name in CATEGORIES}
METRICS |= {name + '.' + wait for name in SQL for wait in ('wait-none', 'wait-io', 'wait-lock', 'wait-lwlock', 'wait-client', 'wait-ipc', 'wait-timeout', 'wait-other', 'blocked')}
COUNTERS = {'requests', 'requestIds', 'brokenPairs', 'missingSync', 'missingProxy', 'jvmEvents', 'jvmRimEvents', 'pgSamples', 'pgMatched', 'clientErrors', 'discardedSymbols', 'profileStartsConfirmed', 'profileStopsConfirmed', 'clientSamples', 'unmeasuredProxy', 'unmeasuredPG', 'unmeasuredJVM'}
PROFILES = {'cpu', 'net', 'sync', 'syscall', 'sched'}
PREFIXES = ('github.com/Alps86/shop/internal/', 'github.com/jackc/pgx/v5', 'runtime.', 'runtime/', 'internal/', 'net.', 'net/', 'encoding/json.', 'encoding/base64.', 'html/template.', 'text/template.', 'compress/gzip.', 'compress/flate.', 'reflect.', 'bytes.', 'strings.', 'context.', 'time.', 'sync.', 'sync/', 'syscall.', 'bufio.')


class Failure(Exception):
    pass


def require(condition, kind='invalid-structure'):
    if not condition:
        raise Failure(kind)


def number(value):
    require(type(value) in (int, float))
    require(math.isfinite(value))
    require(value >= 0)
    return value


def exact(value, keys):
    require(type(value) is dict)
    require(set(value) == set(keys))


def timestamp(value):
    match = re.fullmatch(r'(.+?)(?:[.,](\d+))?(Z|[+-]\d\d:\d\d)', value)
    require(match is not None)
    base = datetime.datetime.fromisoformat(match[1] + match[3].replace('Z', '+00:00'))
    fraction = (match[2] or '').ljust(9, '0')
    return int(base.timestamp()) * 1000000000 + int(fraction or '0')


def duration(value):
    match = re.fullmatch(r'PT(?:(\d+)H)?(?:(\d+)M)?(?:([\d.]+)S)?', value)
    require(match is not None)
    seconds = decimal.Decimal(match[1] or 0) * 3600 + decimal.Decimal(match[2] or 0) * 60 + decimal.Decimal(match[3] or 0)
    return int(seconds * 1000000000)


def percentile(values, fraction):
    return values[max(0, math.ceil(len(values) * fraction) - 1)]


def statistics(key, values, expected=1):
    name, cohort, unit = key
    ordered = sorted(values)
    result = dict(phase=name, cohort=cohort, unit=unit, count=len(values), presence=0, missingCount=expected, sum=0, min=0, p50=0, p95=0, p99=0, max=0)
    if not ordered:
        return result
    result.update(presence=1, missingCount=max(0, expected - len(values)), sum=sum(values), min=ordered[0], p50=percentile(ordered, .5), p95=percentile(ordered, .95), p99=percentile(ordered, .99), max=ordered[-1])
    return result


def safe_symbol(value):
    if not value.startswith(PREFIXES):
        return False
    if not re.fullmatch(r'[A-Za-z0-9_./*()\[\],]+', value):
        return False
    return re.search(r'\d+(?:\.\d+){3}', value) is None


def atomic(path, value):
    temporary = path.with_suffix('.temporary')
    temporary.write_text(json.dumps(value, allow_nan=False, separators=(',', ':')) + '\n')
    temporary.chmod(0o600)
    temporary.replace(path)


def status(path, phase, error, ready=0):
    require(phase in PHASES)
    require(error in ERRORS)
    path.mkdir(mode=0o700, parents=True, exist_ok=True)
    atomic(path / 'status.json', dict(schemaVersion=1, phase=phase, errorClass=error, success=0, dataReady=ready))


def check_status(value):
    exact(value, ('schemaVersion', 'phase', 'errorClass', 'success', 'dataReady'))
    require(type(value['schemaVersion']) is int and value['schemaVersion'] == 1)
    require(value['phase'] in PHASES and value['errorClass'] in ERRORS)
    for key in ('success', 'dataReady'):
        require(type(value[key]) is int and value[key] in (0, 1))
    if value['success']:
        require(value['dataReady'] == 1 and value['errorClass'] == 'none')


def check_metric(value):
    exact(value, ('phase', 'cohort', 'unit', 'count', 'presence', 'missingCount', 'sum', 'min', 'p50', 'p95', 'p99', 'max'))
    require(value['phase'] in METRICS)
    require(value['cohort'] in ('all', 'http-tail'))
    require(value['unit'] in ('ns', 'count', 'bytes', 'fraction'))
    for key in ('count', 'presence', 'missingCount', 'sum', 'min', 'p50', 'p95', 'p99', 'max'):
        number(value[key])
    require(value['presence'] in (0, 1))
    for key in ('count', 'presence', 'missingCount'):
        require(type(value[key]) is int)
    require(value['min'] <= value['p50'] <= value['p95'] <= value['p99'] <= value['max'])


def check_summary(value):
    exact(value, ('schemaVersion', 'expectedRequests', 'complete', 'window', 'counters', 'metrics'))
    require(type(value['schemaVersion']) is int and value['schemaVersion'] == 1)
    require(type(value['expectedRequests']) is int and value['expectedRequests'] == 1000)
    require(type(value['complete']) is int and value['complete'] in (0, 1))
    exact(value['window'], ('startUnixNs', 'endUnixNs', 'overlapNs'))
    for item in value['window'].values():
        number(item)
    exact(value['counters'], COUNTERS)
    for item in value['counters'].values():
        require(type(item) is int)
        number(item)
    require(type(value['metrics']) is list and len(value['metrics']) <= len(METRICS) * 2)
    for item in value['metrics']:
        check_metric(item)


def check_pprof(value):
    exact(value, ('schemaVersion', 'rows'))
    require(type(value['schemaVersion']) is int and value['schemaVersion'] == 1)
    require(type(value['rows']) is list and len(value['rows']) <= 150)
    for row in value['rows']:
        check_pprof_row(row)


def check_pprof_row(row):
    exact(row, ('profile', 'function', 'flatNs', 'flatPercent', 'sumPercent', 'cumNs', 'cumPercent'))
    require(row['profile'] in PROFILES)
    require(type(row['function']) is str and safe_symbol(row['function']), 'unsafe-output')
    for key in ('flatNs', 'flatPercent', 'sumPercent', 'cumNs', 'cumPercent'):
        number(row[key])


def verify(path):
    require(path.is_dir() and not path.is_symlink())
    names = {item.name for item in path.iterdir()}
    require(names in ({'status.json'}, {'status.json', 'summary.json', 'pprof.json'}), 'unsafe-output')
    for name in names:
        require((path / name).is_file() and not (path / name).is_symlink(), 'unsafe-output')
    state = json.loads((path / 'status.json').read_text())
    check_status(state)
    require(state['dataReady'] == int('summary.json' in names))
    if state['dataReady']:
        summary = json.loads((path / 'summary.json').read_text())
        check_summary(summary)
        if state['success']:
            require(summary['complete'] == 1)
        check_pprof(json.loads((path / 'pprof.json').read_text()))


