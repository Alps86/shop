"""Feste Task-/Regionszuordnung des vorhandenen Go-Traceformats."""
import json
import re
from schema import Failure, REGIONS, SQL, TOOK, percentile, require, timestamp


class TraceReader:
    def trace(self):
        with (self.root / 'go-trace.txt').open() as source:
            for line in source:
                self.trace_line(line.rstrip())
        self.counts['brokenPairs'] = sum(len(value) for value in self.stacks.values())
        self.counts['missingSync'] = int(self.sync is None)
        self.finish_tasks()

    def trace_line(self, line):
        sync = re.search(r' Sync .* Trace=(\d+) Mono=\d+ Wall=(\S+)', line)
        if sync:
            self.sync = (int(sync[1]), timestamp(sync[2]))
            return
        match = re.match(r'M=-?\d+ P=-?\d+ G=(-?\d+) (TaskBegin|TaskEnd|RegionBegin|RegionEnd|Log) Time=(\d+) (.*)', line)
        if match:
            self.trace_event(match)

    def trace_event(self, match):
        goroutine, kind, moment, detail = match.groups()
        moment = int(moment)
        if kind.startswith('Task'):
            self.task_event(kind, moment, detail)
            return
        task = re.search(r'Task=(\d+)', detail)
        if task is None or int(task[1]) not in self.tasks:
            return
        identifier = int(task[1])
        if kind == 'Log':
            self.trace_log(identifier, moment, detail)
            return
        self.region(identifier, int(goroutine), (kind, moment, detail))

    def task_event(self, kind, moment, detail):
        identifier = int(re.search(r'ID=(\d+)', detail)[1])
        if kind == 'TaskBegin' and 'Type="personal-search"' in detail:
            require(len(self.tasks) < 1000, 'incomplete-data')
            require(identifier not in self.tasks)
            self.tasks[identifier] = dict(start=moment, end=None, request=None, spans=[], pids=[])
            return
        if kind == 'TaskEnd' and identifier in self.tasks:
            self.tasks[identifier]['end'] = moment

    def trace_log(self, identifier, moment, detail):
        match = re.search(r'Category="([a-z.-]+)" Message="(\d+)"$', detail)
        if not match:
            return
        name, value = match[1], int(match[2])
        if name == 'request-id':
            self.tasks[identifier]['request'] = value
        if name in TOOK:
            self.tasks[identifier]['spans'].append((name, moment, moment + value * 1000000))
        if name.removesuffix('.pid') in SQL and name.endswith('.pid'):
            self.tasks[identifier]['pids'].append((name.removesuffix('.pid'), value))

    def region(self, task, goroutine, event):
        kind, moment, detail = event
        name = re.search(r'Type="([a-z.-]+)"$', detail)
        require(name is not None)
        require(name[1] in REGIONS)
        stack = self.stacks.setdefault((task, goroutine), [])
        if kind == 'RegionBegin':
            stack.append((name[1], moment))
            return
        require(bool(stack), 'incomplete-data')
        previous, start = stack.pop()
        require(previous == name[1] and moment >= start, 'incomplete-data')
        self.tasks[task]['spans'].append((previous, start, moment))

    def finish_tasks(self):
        self.counts['requests'] = len(self.tasks)
        identifiers = {task['request'] for task in self.tasks.values() if task['request'] is not None}
        self.counts['requestIds'] = len(identifiers)
        lengths = [task['end'] - task['start'] for task in self.tasks.values() if task['end'] is not None]
        require(bool(lengths), 'incomplete-data')
        threshold = percentile(sorted(lengths), .95)
        for task in self.tasks.values():
            require(task['end'] is not None, 'incomplete-data')
            self.task_metrics(task, task['end'] - task['start'] >= threshold)

    def task_metrics(self, task, tail):
        if tail:
            self.tail.add(task['request'])
        spans = [('http.request', task['start'], task['end'])] + task['spans']
        for name, begin, end in spans:
            self.metric(name, end - begin)
            if tail:
                self.metric(name, end - begin, cohort='http-tail')

    def wall(self, moment):
        require(self.sync is not None, 'incomplete-data')
        return self.sync[1] + moment - self.sync[0]


    def boundaries(self):
        starts, stops = [], []
        for name in ('go-start', 'go-stop', 'jfr-start', 'jfr-stop'):
            value = self.boundary(name)
            if value is None:
                continue
            if name.endswith('start'):
                starts.append(value['finishedUnixNanoseconds'])
            if name.endswith('stop'):
                stops.append(value['startedUnixNanoseconds'])
        self.counts['profileStartsConfirmed'] = len(starts)
        self.counts['profileStopsConfirmed'] = len(stops)
        return max(0, min(self.end, *stops) - max(self.start, *starts))

    def boundary(self, name):
        try:
            value = json.loads((self.root / (name + '.json')).read_text())
            require(value['confirmed'] is True and value['exitCode'] == 0, 'incomplete-data')
            return value
        except Exception:
            if name.startswith('go-'):
                raise Failure('incomplete-data') from None
            self.counts['unmeasuredJVM'] = 1
            return None

