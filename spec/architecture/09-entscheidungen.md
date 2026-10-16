# 9 Architekturentscheidungen

| Entscheidung | Begründung und Folge |
| --- | --- |
| Ein modularer Monolith/ein Betreiber | Kohäsiver Demo-Kern, keine hypothetische Händler-/Tenantplattform; getrennte Adapter ohne getrennte UI-App |
| Fünf Verkaufsroots plus versionierte Konfigurationsroots | Bestätigte Grenzen bleiben; Family/Category/DiscountRule haben eigene Publikationskonsistenz, keine Einbettung fremder veränderlicher Graphen |
| PostgreSQL-Transaktion über Abschlussroots | Reservierung/Order/Cart/Outbox müssen im konkreten Abschluss zusammen konsistent sein; DDD verbietet keine explizit begründete Applicationtransaktion |
| Inventorypreis und Bestand zusammen | Keine konkurrierende Preiswahrheit im Product; Suchkarten sind Lesemodelle |
| Vollständige Compliancepakete je Variantenoverride | Eindeutige Auflösung statt riskantem partiellen Merge; fehlende Pflichtdaten blockieren Kauf |
| Exakte Minor Units/Decimal, Rabatt auf Zubehör | Deterministische Rundung und sichtbare Zuordnung, kein doppelter Abzug |
| Temporal Cart UND Order | Langlebigkeit/Timer/Neustart, fachliche Wahrheit bleibt DB; Outbox heilt Startlücke |
| Gebundene Versionen, keine automatische Migration | Replay/Preisbestätigung bleiben stabil; Editor/DSL später |
| Logisches pgoutput-CDC ohne Broker | Echte Commitänderungsquelle mit kleinem Compose-Footprint; Adapter muss LSN/Versionen/Neuaufbau korrekt behandeln |
| Nested Varianten und disjunktive Facetten | Keine falschen SKUkombinationen, Marken-OR erweitert Treffer; explizite Sortierung schlägt Boost |
| Synthetische Co-Purchase-Daten | Demo nachvollziehbar ohne echte Käuferbehauptung, Profile beeinflussen nur begrenzt Reihenfolge |
| Hugo+Nginx SSI+Varnish ESI+Go/HTMX | Vorgegebene statisch/dynamische Komposition mit Standardengines; persönliche Antworten umgehen shared Cache |
| ID im lesbaren Slugpfad | Umbenennungen ohne Identitätsverlust; ein kanonisches Modell mit adressierbaren Varianten |
| Sessionfähigkeit getrennt von Profil | Öffentlicher Profilumschalter darf keine fremden Carts/Orders öffnen |
| Schlichtes eigenständiges Design | EP als Strukturreferenz, Performance/Demo vor Pixelkopie; keine Browsermatrix |
| Vorab festgelegte lokale Vier-CPU-Vollstackreferenz | Kontrollierbare Hardware und nachgewiesene gemeinsame Ressourcenhülle erlauben eine ausdrücklich begrenzte lokale Leistungsabnahme. Derselbe vollständige fachliche Stack und unveränderte Last gelten; die technische Ausführung über unprivilegierte OCI-Binärdateien ist gesondert auf Lifecycle- und Versionsäquivalenz zu prüfen. Cloudgates und sämtliche Cloudlatenzberichte bleiben erhalten. Daraus folgen weder eine bewiesene externe CPUursache noch Cloud- oder Hardwareunabhängigkeitsgarantien. |

Dies sind bewusst jetzt getroffene Konkretisierungen früherer Vorschläge, keine Behauptung historischer Einzelabstimmung. Implementierung darf technische Details verfeinern, muss verhaltensrelevante Vertragsänderungen vorher synchronisieren.
