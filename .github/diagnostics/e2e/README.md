# Einmalige E2E-Diagnose

Aufruf ausschließlich im unabhängig freigegebenen Quality-Lauf 48, Job `race`, Versuch 1, Ereignis `pull_request/synchronize` auf `codex/shop-performance`:

```sh
bash .github/diagnostics/e2e/run.sh
```

Der Workflow führt zuvor die unveränderte Pflichtsuite vollständig aus. Der Runner benötigt die vorhandenen `SHOP_PERFORMANCE_*`-Werte, `SHOP_HEY`, Docker Compose, Go, Python, jq, pidstat und den vorhandenen passwortfreien CI-Zugriff auf `sudo nsenter`. Er startet genau ein zusätzliches Fenster mit 100 Aufwärmrequests und 1000 Messrequests bei Parallelität 10. Fehlgeschlagene oder begonnene Aufnahmen werden nicht erneut gestartet.

Der Go-Overlaybuild ersetzt nur den Benchmark-Appprozess und ergänzt feste Regionen, technische Request- und Backend-PIDs sowie numerische `took`-Werte. Der interne Kontrolllistener liegt ausschließlich auf dessen Containerloopback. CPUprofil und Executiontrace enden gemeinsam bei explizitem Stop oder spätestens nach 60 Sekunden. Ein eigenes frisches Container-IDset erfasst den neuen Prozess; die ursprünglichen Collector bleiben unverändert. Die drei SQL-Regionen messen Auflösung der Sitzung, aktiven Warenkorb und Benutzerpayload. Die ergänzende PG-Beobachtung enthält keine SQL-Texte oder Parameter und kann kurze Wartezustände verpassen.

Alle Rohdateien, Zustände, gerenderten Konfigurationen und Fehlerlogs bleiben unter `/tmp/shop-e2e-private` mit eingeschränkten Rechten. Nur die feste Liste in `approve.sh` gelangt nach Prüfung der Ereignisstruktur und privater Sentinelwerte nach `/tmp/shop-e2e-approved`. CPU-Flat-/Cum-/Sampleausgaben, vollständige geparste Go-Traceereignisse, die vier Standard-Warteprofile sowie sichere JFR-Ereignisse mit Zeit und Stacks erhalten die kausale Zuordnung. Die numerischen hey-CSV bleiben für Statuszählung und Clientperzentile erhalten. Start-/Stopmetadaten begrenzen jeden erfolgreichen Go- und JFR-Kontrollaufruf mit UTC-Unixnanosekunden; sie sind Aufrufgrenzen und keine behaupteten atomaren gemeinsamen Profilstarts. Der Go-Stopzustand, Binaryhash, Buildinfo, relative Overlayhashes sowie Container-/PID-/comm-Zuordnung ergänzen den exakten Build- und Fensterbeleg. Ungeprüfte Rohprofile werden nicht freigegeben. Die kleine Transportverschlüsselung und der Upload liegen beim Workflow.

Die Diagnose verändert Wärme, Prozessidentität und Laufzeitkosten. Ihre Zeiten ersetzen weder die ursprünglichen 28 Messberichte noch deren p95. Ein fehlender Beleg ist kein erfolgreicher Befund. Bestands-, Bestell- und Workflowhashes müssen vor und nach dem Fenster übereinstimmen. Beim Cleanup werden Profiler, begrenzte Collector und die temporären App-/Proxyprozesse beendet; der bestehende Workflow entfernt anschließend das Compose-Projekt. Overlay und Assets werden nach diesem einmaligen Lauf wieder entfernt. Die tatsächlichen Produktquellen bleiben unverändert.
