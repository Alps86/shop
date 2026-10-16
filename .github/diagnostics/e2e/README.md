# Einmalige E2E-Diagnose mit Runneraggregaten

Genau Quality-Lauf 50, Job `race`, Versuch 1, `pull_request/synchronize` auf `codex/shop-performance`:

```sh
bash .github/diagnostics/e2e/run.sh
python3 .github/diagnostics/e2e/aggregate.py --verify-approved /tmp/shop-e2e-approved
```

Vorher läuft die vollständige unveränderte Pflichtsuite. Aufnahmeaufbau und Last aus dem vorherigen Diagnoseschnitt bleiben gleich: tatsächlicher temporärer App-Overlaybuild, Go-CPUprofil und Executiontrace, zeitgleiche feste JFR-Aufnahme, 100 Aufwärmrequests und 1000 Messrequests bei Parallelität 10 mit den bisherigen Flags und derselben privaten Sitzung. Kein weiterer Lauf und kein automatisches Ausweichen auf eine andere Runnummer.

Der Runner beendet und prüft nach dem ersten sicheren Status die ursprünglichen Docker-event-Prozessgruppen, bevor Diagnose-execs starten. Nginx schreibt Diagnosezugriffe und Fehler in private Containerdateien. Freie Toolausgaben und sämtliche Rohprofile bleiben ausschließlich im privaten Verzeichnis. Ein CLI-Auswerter mit kohäsiven Schema-, Trace-, JVM- und Backendlesern reduziert sie auf drei geschlossene JSON-Dateien: `summary.json`, `pprof.json`, `status.json`. Bei frühem Fehler ist ausschließlich der feste Status erlaubt. Das separate Gate prüft Dateinamen, Typen, Enums und die erlaubten Funktionssymbole; keine Rohtextfreigabe durch Regex.

Exportiert werden feste Phasen/Kategorien und numerische Zeiten/Zähler. Nur pprof-Funktionssymbole aus expliziten Namensräumen und der tatsächlichen Binarysymboltabelle dürfen als zusätzliche Nutzstrings erscheinen. Alle übrigen Symbolnamen, Labels, Header, Dateipfade, Queries, Cookies, Adressen, IDs/PIDs, Events, Samples, Stacktexte, Buildinfo und Manifeste bleiben ausgeschlossen. Keine Verschlüsselung und kein Raw-Fallback. JFR-Modulversionen werden strukturell verworfen; die frühere IPv4-Regel wird nicht aufgeweicht.

Go-Sync liefert die Trace-Wallzeitbasis. SQL-PIDs und Proxyrequest-IDs dienen ausschließlich intern der Zuordnung. Pro Phase entstehen inklusive Wallzeitverteilungen und eine feste HTTP-p95-Kohorte; verschachtelte Zeiten nicht addieren. Die CPU-Phasenlabels liefern separat numerische Sampleszeit. JVM-Execution-/Native-Samples sind getrennte Zählungen, keine exakte CPUzeit; Allocationgewichte und GC-/Park-/Monitorzeiten bleiben getrennt. Nicht beobachtete Werte haben presence=0 beziehungsweise missingCount. PG-Samples können kurze Sperren verpassen. Hey-Zeilen werden nicht als Request-ID interpretiert, getrennte Perzentile werden nicht subtrahiert.

EXIT-Cleanup stoppt Profiler/Collector, entfernt die eigenen temporären App-/Proxycontainer, löscht feste JFR-Dateien in der OS-Instanz und sämtliche privaten Hostdaten einschließlich Overlay/Binary/Configs. Der sichere Status bleibt auch bei Fehler erhalten. Äußeres Workflow-Cleanup und kurzlebiger Runner sichern Timeout-/Killgrenzen zusätzlich ab. Upload ausschließlich nach erfolgreichem separatem Gate; niemals Raw-, Compose- oder freie Diagnose-stdout-Dateien hinzufügen.

Die Auswertung verändert Profilingkosten und Wärme. Sie ersetzt weder die ursprünglichen Messberichte noch den weiterhin negativen SLA-Befund. Vor Ausführung benötigt der konkrete Aufbau die unabhängige Implementierungsreview.
