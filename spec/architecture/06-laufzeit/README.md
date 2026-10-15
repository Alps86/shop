# 6 Laufzeitsicht

PostgreSQL hält fachliche Zustände und Idempotenz. Temporal hält langlebige Ablaufhistorie, Timer und Activitywiederholungen; OpenSearch hält abgeleitete Suchdaten. Keine globale Transaktion überspannt diese Systeme.

- [Temporal Cart und Order](temporal.md): Identität, Übergänge, Zeit und Wiederanlauf.
- [Bestellabschluss](checkout.md): Preisbestätigung, Zahlung, atomare Reservierung und Konkurrenz.
- [Outbox und CDC](projection.md): Commit, Projektion, Versionen, Löschung und Reindex.

```mermaid
sequenceDiagram
    actor Guest as Gast
    participant HTTP as Storefront
    participant Cart as Temporal CartWorkflow
    participant App as Application Activity
    participant PG as PostgreSQL
    participant Order as Temporal OrderWorkflow
    Guest->>HTTP: Cart aendern und Preis bestaetigen
    HTTP->>Cart: Update mit CommandID
    Cart->>App: Aendern und bewerten
    App->>PG: Cart und Bewertung speichern
    Guest->>HTTP: Demoabschluss
    HTTP->>Cart: Place CommandID EvaluationID
    Cart->>App: Zahlung simulieren und Purchase committen
    App->>PG: Reservierung Order Cart Outbox atomar
    PG-->>App: OrderID
    App-->>Cart: Erfolg
    Cart->>Order: EnsureOrder OrderID
    HTTP-->>Guest: Bestellstatus
```

Wenn die Prozessantwort nach Commit verloren geht, ergibt derselbe Command dieselbe OrderID. Der durable Outbox-Prozessauftrag stellt den Orderstart auch nach Workerabsturz sicher. Ablaufentscheidungen werden nicht aus wechselnder aktueller Konfiguration beim Replay neu erzeugt.
