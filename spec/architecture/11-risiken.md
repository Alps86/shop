# 11 Risiken und offene Nachweise

Demo zuerst bedeutet keine erfundene Produktionsreife. Fachverträge sind entschieden, tatsächliche Implementierungs-/Laufzeit-/Performancebelege folgen nach Foundationstories.

| Risiko / offene Grenze | Umgang |
| --- | --- |
| Unter100ms sichtbarer Browseraufbau besonders kalt ambitioniert | Zusammengesetztes HTTP und LCP getrennt messen/als unbelegt berichten; keine Backendgleichsetzung |
| Produktpflichtdaten/Rechte unvollständig | Verifizierte Quelle/Stand, Kauf sperren statt Sicherheitsangaben erfinden; konkrete Lieferanten-/Rechtsprüfung vor echtem Betrieb |
| WAL-Slotrückstau/Versionsfehler | LSN/Verzug sichtbar, Wiederanlauf/Reindex testen; produktives Retentions-/Backupkonzept später |
| Temporal/DB nicht atomar | Persistierte Aufträge/Idempotenz und Wiederholung, keine globale Exactly-once-Behauptung |
| Includeausfall oder Cacheverwechslung | Kontextschlüssel/kurze TTL/no-store, komplette SKUfragmente, sichere Fehlerdarstellung |
| Lokale Dockerberechtigung fehlt | Keine Rechteumgehung; unbeeinträchtigte Entwicklung/echte CI-Dienste, lokalen Start fehlend ausweisen |
| Qualitätsdefaults erfassen nur nativen Buildumfang | Review aktiver Slices/Quellen, zusätzliche vollständige AST-/Buildtag-Antibypasshärtung offen |
| 12h-Prognose bei Reviews/Integration | Frühe reale vertikale Gastbestellung, stabile Adapter parallel, Integrationspuffer; Risiko früh melden |

Spätere Härtung: öffentliches Deployment/TLS/Rate Limits/umfangreiche CSP, Backups/Restore/HA, detaillierte Datenschutz-/Lösch-/Kontowiederherstellung, exhaustive Ausfall-/Lastkombinationen, vollständiger juristischer Betreiber-/Produktfreigabeprozess, Workfloweditor/Migration, echte Payment-/Fulfillment-/Akeneoanbindung. Diese Punkte sind keine stillschweigenden aktuellen Features und blockieren keine korrekt gekennzeichnete lokale Demo. Sichere Gasttrennung, echte Dienste, verbindliche Kaufprüfung und Qualitätsbaseline bleiben Pflicht.
