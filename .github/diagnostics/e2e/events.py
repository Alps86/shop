#!/usr/bin/env python3
"""Prüft ausschließlich die feste JFR-Ereignisstruktur vor der Freigabe."""
from pathlib import Path
import re
import sys

allowed = {
    "jdk.ExecutionSample", "jdk.NativeMethodSample", "jdk.CPULoad",
    "jdk.ThreadCPULoad", "jdk.ObjectAllocationSample", "jdk.GarbageCollection",
    "jdk.GCPhasePause", "jdk.GCHeapSummary", "jdk.ThreadPark",
    "jdk.JavaMonitorEnter", "jdk.JavaMonitorWait", "jdk.Checkpoint", "jdk.Metadata",
}
for line in Path(sys.argv[1]).read_text().splitlines():
    match = re.match(r"\s*(\S+)\s+(\d+)\s+\d+\s*$", line)
    if match and int(match[2]) and match[1] not in allowed:
        raise SystemExit("unexpected JFR event type")
