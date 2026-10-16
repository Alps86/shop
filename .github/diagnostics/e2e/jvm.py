"""Streaming-Reduktion der festen JFR-Ereignisse auf numerische Kategorien."""
import json
import subprocess
from schema import EVENTS, duration, number, require, timestamp


class JVMReader:
    def jfr(self):
        command = ['jq', '--stream', '-c', '.', str(self.root / 'jfr-events.json')]
        with (self.root / 'jq-private.stderr').open('w') as errors:
            process = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=errors, text=True)
            try:
                self.jfr_leaves(process.stdout)
            finally:
                process.stdout.close()
                process.wait()
        require(process.returncode == 0)

    def jfr_leaves(self, source):
        current, event = None, {}
        for line in source:
            leaf = json.loads(line)
            if len(leaf) != 2 or len(leaf[0]) < 4:
                continue
            path, value = leaf
            require(path[:2] == ['recording', 'events'])
            if path[2] != current:
                self.jvm_event(event)
                current, event = path[2], {}
            self.jfr_leaf(event, path[3:], value)
        self.jvm_event(event)

    def jfr_leaf(self, event, path, value):
        fields = {('type',): 'type', ('values', 'startTime'): 'time', ('values', 'duration'): 'duration', ('values', 'weight'): 'weight', ('values', 'jvmUser'): 'user', ('values', 'jvmSystem'): 'system', ('values', 'user'): 'user', ('values', 'system'): 'system'}
        key = fields.get(tuple(path))
        if key:
            event[key] = value
        if len(path) == 7 and path[:3] == ['values', 'stackTrace', 'frames'] and path[4:] == ['method', 'type', 'name']:
            event.setdefault('classes', []).append(value)

    def jvm_event(self, event):
        if not event:
            return
        require(event.get('type') in EVENTS)
        if 'time' not in event:
            return
        self.counts['jvmEvents'] += 1
        moment = timestamp(event['time'])
        if not self.start <= moment <= self.end:
            self.counts['jvmRimEvents'] += 1
            return
        self.jvm_value(event)

    def jvm_value(self, event):
        name = event['type']
        categories = {'jdk.ExecutionSample': 'jvm.execution.', 'jdk.NativeMethodSample': 'jvm.native.', 'jdk.ObjectAllocationSample': 'jvm.allocation.'}
        if name in categories:
            category = stack_category(event.get('classes', []))
            self.jvm_sample(name, categories[name] + category, event)
        names = {'jdk.GarbageCollection': 'jvm.gc', 'jdk.GCPhasePause': 'jvm.gc-pause', 'jdk.ThreadPark': 'jvm.park', 'jdk.JavaMonitorEnter': 'jvm.monitor', 'jdk.JavaMonitorWait': 'jvm.monitor'}
        if name in names and 'duration' in event:
            self.metric(names[name], duration(event['duration']))
        if name in ('jdk.CPULoad', 'jdk.ThreadCPULoad'):
            require('user' in event and 'system' in event)
            phase = {'jdk.CPULoad': 'jvm.cpu-load', 'jdk.ThreadCPULoad': 'jvm.thread-cpu-load'}[name]
            self.metric(phase, number(event['user']) + number(event['system']), 'fraction')

    def jvm_sample(self, name, phase, event):
        if name == 'jdk.ObjectAllocationSample':
            require('weight' in event)
            self.metric(phase, event['weight'], 'bytes')
            return
        self.metric(phase, 1, 'count')



def stack_category(classes):
    for name in classes:
        category = class_category(name)
        if category != 'other':
            return category
    return 'other'


def class_category(name):
    if name.startswith('org/apache/lucene/codecs/') and 'StoredFields' in name:
        return 'stored-fields'
    if name.startswith('org/apache/lucene/codecs/') and 'DocValues' in name:
        return 'doc-values'
    prefixes = [('completion', ('org/apache/lucene/search/suggest/', 'org/opensearch/search/suggest/')), ('fetch', ('org/opensearch/search/fetch/',)), ('compression', ('net/jpountz/lz4/', 'org/apache/lucene/util/compress/')), ('stored-fields', ('org/apache/lucene/codecs/compressing/',)), ('doc-values', ('org/apache/lucene/index/DocValues',)), ('json', ('com/fasterxml/jackson/', 'org/opensearch/core/xcontent/')), ('query', ('org/apache/lucene/search/', 'org/opensearch/search/query/'))]
    for category, choices in prefixes:
        if name.startswith(choices):
            return category
    return 'other'

