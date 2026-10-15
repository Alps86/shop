# Langlebige Cart- und Orderprozesse

CartWorkflow-ID `cart/{CartID}`, OrderWorkflow-ID `order/{OrderID}`, TaskQueue `shop`. Workflow-ID-Reuse nach terminalem Abschluss abweisen. Ensure-Aufrufe dürfen AlreadyStarted als Erfolg behandeln, niemals zweite Instanz erzeugen. IDs enthalten keine Gasttoken. Workflowdaten enthalten nur nötige Kennungen/gebundene Versionsstände, fachliche Detaildaten liest Application aus PostgreSQL.

## Cart und Checkout

Start erzeugt/übernimmt gespeicherten Cart mit Definition `cart-v1` und Regelstand `demo-v1`. POC unterstützt feste Schritte: Artikelwahl → Bewertung → Adresse/Versand/Zahlung → Bestätigung → simulierte Zahlung → verbindlicher Abschluss. Keine DSL/Editor, kein eigener Checkoutroot. Cart.Checkout hält fachliche Prozessdaten, PostgreSQL ist Wahrheit.

| Zustand | Eingang | Wirkung / nächster Zustand |
| --- | --- | --- |
| open/checkingOut | Add/SetQuantity/Remove | Application validiert, speichert neue Revision/LastActivityAt, Bewertung/Bestätigung/Autorisierung ungültig; bisheriger Zustand bleibt |
| open/checkingOut | Address/Shipping/Payment | Änderung vor verbindlicher Order erlaubt, neue Bewertung/Bestätigung nötig; checkingOut |
| open/checkingOut | Confirm(EvaluationID, Revision) | exakt aktuelle Bewertung binden, bleibt checkingOut |
| checkingOut | Place(CommandID) | Kreditkarte simuliert autorisieren bzw. Rechnung direkt, dann CommitPurchase; Erfolg ordered |
| open/checkingOut | Inaktivität ≥ 168 h | Application prüft LastActivityAt erneut, setzt expired, terminal |
| ordered/expired | Mutation | AlreadyOrdered/Expired, keine Wiederöffnung; neuer Einkauf braucht neue CartID |

Temporal Updates `ChangeCart`, `ConfirmEvaluation`, `PlaceOrder` liefern quittierte fachliche Ergebnisse; öffentliche API wartet bis 2 s, danach 202/Pending. Workflowquery `ProcessStatus` ist reine Prozessdiagnose, öffentliche Kaufdaten kommen aus PostgreSQL. Liefermeldungen Order über Signal `ReportDelivery`, CommandID dedupliziert in Application; Bestätigung über Statusabfrage. Kein Polling schlägt als Aktivität zu Buche.

Sieben Tage sind exakt 168h seit letzter erfolgreicher Nutzeränderung/Bestätigung. Bei t<LastActivity+168h wiederaufnehmbar; bei t=Grenze und danach expired. Kontrollierte fachliche Uhr als Applicationwert; Workflowtimer nutzt Temporalzeit. Ablauf-Activity sperrt Cart und prüft aktuelle Zeit erneut, damit alter Timer neuere Aktivität nicht überschreibt. Ablauf und Place konkurrieren unter demselben Cartlock: am Grenzzeitpunkt darf Place nicht mehr committen. Reads/Duplikate verlängern nicht.

Während simulierte Freigabe läuft sind Änderungen weiterhin vor Commit möglich; jede Cartrevision entwertet die Autorisierung/Bestätigung des alten Versuchs. Commit meldet Conflict, Simulation wird verworfen, neue Bewertung erforderlich. Keine echte Zahlung zurückzubuchen. Bei PriceConfirmationRequired/Unavailable bleibt checkingOut für aktive Korrektur.

## Order und Lieferdemo

Orderstart bindet `order-v1` und ursprünglichen Regelstand, Kauf-Snapshot existiert bereits. Prozesszustände `accepted`, `partiallyShipped`, `shipped`, `delivered`; packed ist optionale Fortschrittsmeldung. ReportDelivery enthält ShipmentID und positive Positionsmengen, nur bekannte SKUs, kumulativ nie mehr als gekauft. shipped verbraucht Reservierung/Bestand atomar, delivered setzt zuvor versandte Mengen als geliefert. Mehrere Shipments erlauben Teillieferung. Wiederholte identische Meldung ist Noop, abweichende Payload zur selben CommandID/Shipmentstatus Conflict. delivered terminal, Snapshots unverändert.

Externe Abwicklungsgrenze erhält `OrderPlaced` mit OrderID/Kauf-Snapshot und später dieselben FulfillmentUpdates. In der Demo sendet ausschließlich lokales /backend die Meldungen mit sichtbarer Simulationskennzeichnung. Kein Versandlabel, Carrier, Retouren- oder Lagerprozess.

## Wiederholung, Versionierung und Fehler

Activities StartToClose 10s, ScheduleToClose 2min, Retry exponential 1s→10s, maximal5 Versuche. Fachliche Fehler nicht retriable; technische Timeouts/DB-Ausfälle retriable. Nach Ausschöpfung bleibt Prozess `needsAttention` als technische Diagnose, keine fälschliche Order-/Zahlungsbestätigung. Retry per Verwaltung mit identischer CommandID nach Reparatur; bei bereits vorhandenem Commit Ergebnis lesen. Worker-Neustart replayt History; nichtdeterministische IO nur Activities, Timer/Workflowzeit aus Temporal.

Start/Update-Aufträge werden mit fachlicher Änderung als Outboxauftrag gespeichert. Dispatcher startet/signaliert idempotent und quittiert erst nach Annahme; ungeklärte Aufträge bleiben sichtbar und werden wiederholt. Dadurch heilt ein Absturz zwischen DB-Commit und Temporalstart. Keine Exactly-once-Zusage.

Definition/Regelstand bleiben je Cartlauf gebunden. Neue Konfiguration gilt nur für neue Carts; keine automatische Migration laufender Prozesse im POC. Order übernimmt gebundene Versionen. Workflowcodeänderungen verwenden Temporal-Versionierung/Patching, alte History bleibt replaybar; neue Konfiguration beim Replay nicht frei neu lesen. Späterer Editor/validierte Publikation und explizite Migration sind Folgearbeit, kein aktueller Kernvertrag bleibt dafür offen.
