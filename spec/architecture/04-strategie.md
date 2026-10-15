# 4 Lösungsstrategie

Ein kohäsiver Fachkern entscheidet auf Fakten, Application koordiniert Ports und Transaktionen. Adapter übersetzen IO und besitzen Ressourcen/Timeouts. Konstruktorinjektion reicht; Strategy ist nur bei tatsächlich austauschbaren Suchbestandteilen sinnvoll, kein Interface für jedes Value Object.

| Baustein | Strategie |
| --- | --- |
| Katalog | Product enthält Varianten; Family/Category sind referenzierte versionierte Konfiguration, SKU-Register schützt Eindeutigkeit |
| Verkauf | Inventory hält Preis und Menge; Abschluss reserviert mit Order/Outbox in einer PostgreSQL-Transaktion |
| Prozesse | Temporal CartWorkflow und OrderWorkflow; Activities delegieren Application, Datenbank bleibt maßgeblich |
| Suche | OpenSearch nested Varianten; getrennte Filter-, Facetten-, Ranking-, Sortier- und Vorschlagsverantwortung |
| Oberfläche | Hugo-Markdownrahmen, Nginx SSI, Varnish ESI, Go-Templates und HTMX; kein Clientframework |
| Änderungsprojektion | PostgreSQL-logisches WAL-CDC liest transaktionale Outbox; kein zusätzlicher Broker im Demo nötig |
| Qualität | Öffentliche Domain-/Port-Cucumberbasis, wenige reale Adapterintegrationen und HTTP-/Templateabnahmen |

## Größere anschließende Pakete

Nach arc42 folgt die vollständige Cucumber-Story. Dann: (A) Katalog/Handel-Domain einschließlich Pflichtdaten und Bewertung, (B) Application/Gast/Abschlussports samt stabilen DTOs. Auf B bauen parallel (C) PostgreSQL/Outbox/WAL-CDC/OpenSearch und (D) Temporal Cart/Order auf. Unabhängiger Content/Hugo-Rahmen kann nach stabilen Seiten-DTOs parallel entstehen; Storefront/Backend verdrahten anschließend die echten Ports. Ein Integrationspaket verbindet Compose, Demoimport, HTTP-Abnahme und Messung. Keine neue Vorratsissueflut; jedes tatsächlich begonnene Paket vorher unabhängig planprüfen, danach frisch abnehmen.

Arbeitsziel laut Auftrag: 16.10.2026 09:36 MESZ, äußerer Rahmen 21:36 MESZ. Zeitblöcke ab 15.10. 21:36: Grundlagen 0–2 h, Domain/Application 2–4 h, Adapter und früher realer Gastabschluss 4–7 h, Oberfläche/Content 6–9 h, Gesamtdemo/Messung/Reviewpuffer 9–12 h. Der Plan ist eine Prognose, keine Fertiggarantie; fehlende Laufzeitnachweise oder Zielabweichungen offen berichten. Keine Spezifikationsperfektion oder Grafikmatrix vor frühem durchgängigem Abschluss.
