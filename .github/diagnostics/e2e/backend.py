"""Private Zuordnung numerischer Proxy- und SQL-Beobachtungen."""
import decimal
import json
import re
from schema import require, timestamp


class BackendReader:
    def proxy(self):
        identifiers = {task['request'] for task in self.tasks.values()}
        observed = set()
        with (self.root / 'nginx-times.txt').open() as source:
            for line in source:
                self.proxy_line(line, identifiers, observed)
        self.counts['missingProxy'] = len(identifiers - observed)
        self.varnish()

    def proxy_line(self, line, identifiers, observed):
        parts = line.split()
        require(len(parts) == 7)
        require(all(re.fullmatch(r'[0-9.]+', part) for part in parts))
        identifier = int(parts[0])
        require(identifier in identifiers and identifier not in observed)
        require(int(parts[1]) == 200, 'incomplete-data')
        observed.add(identifier)
        for name, value in zip(('proxy.nginx', 'proxy.connect', 'proxy.headers', 'proxy.response'), parts[3:]):
            self.proxy_metric(identifier, name, int(decimal.Decimal(value) * 1000000000))

    def varnish(self):
        identifier = None
        elapsed = None
        with (self.root / 'varnish-times.txt').open() as source:
            for line in source:
                if '<< Request >>' in line:
                    self.varnish_value(identifier, elapsed)
                    identifier, elapsed = None, None
                found = re.search(r'VCL_Log\s+\S+\s+diag:(\d+)\s*$', line)
                if found:
                    identifier = int(found[1])
                found = re.search(r'Timestamp\s+\S+\s+Resp:\s+[\d.]+\s+([\d.]+)', line)
                if found:
                    elapsed = int(decimal.Decimal(found[1]) * 1000000000)
        self.varnish_value(identifier, elapsed)

    def varnish_value(self, identifier, elapsed):
        if identifier is None or elapsed is None:
            return
        if identifier in {task['request'] for task in self.tasks.values()}:
            self.proxy_metric(identifier, 'proxy.varnish', elapsed)

    def proxy_metric(self, identifier, name, value):
        self.metric(name, value)
        if identifier in self.tail:
            self.metric(name, value, cohort='http-tail')

    def pg(self):
        with (self.root / 'pg-waits.jsonl').open() as source:
            for line in source:
                record = json.loads(line)
                moment = timestamp(record['utc'])
                self.counts['pgSamples'] += 1
                for activity in record['activity']:
                    self.pg_activity(moment, activity)

    def pg_activity(self, moment, activity):
        for task in self.tasks.values():
            self.pg_task(task, moment, activity)

    def pg_task(self, task, moment, activity):
        for name, pid in task['pids']:
            if pid == activity['pid']:
                self.pg_match(task, name, moment, activity)

    def pg_match(self, task, name, moment, activity):
        spans = [span for span in task['spans'] if span[0] == name + '.query']
        if not any(self.wall(begin) <= moment <= self.wall(end) for _, begin, end in spans):
            return
        self.counts['pgMatched'] += 1
        category = {'IO': 'io', 'Lock': 'lock', 'LWLock': 'lwlock', 'Client': 'client', 'IPC': 'ipc', 'Timeout': 'timeout', None: 'none'}.get(activity['wait_event_type'], 'other')
        self.metric(name + '.wait-' + category, 1, 'count')
        if task['request'] in self.tail:
            self.metric(name + '.wait-' + category, 1, 'count', 'http-tail')
        if activity['blockers']:
            self.metric(name + '.blocked', 1, 'count')

