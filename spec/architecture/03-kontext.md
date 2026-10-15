# 3 Kontext und Abgrenzung

Der Gast nutzt öffentliche HTML-Seiten und seinen geschützten Cart/Status. Der lokale Betreiber nutzt /backend ohne Login ausschließlich im lokalen Demo-Netz. Hersteller-/Lieferantendaten werden kontrolliert importiert; EP liefert eine begrenzte öffentliche Sortiments-/Strukturreferenz, keine Laufzeitabhängigkeit.

```mermaid
flowchart LR
    Guest[Gast] --> Shop[Ein Shop: Storefront und Verwaltung]
    Operator[Lokaler Betreiber] --> Shop
    Markdown[Eigener Markdown Content] --> Shop
    Supplier[Verifizierte Produktdaten] --> Shop
    Shop --> PG[(PostgreSQL Fachzustand und Outbox)]
    Shop --> Temporal[Temporal Cart und Order]
    PG --> CDC[WAL CDC Projektion]
    CDC --> Search[(OpenSearch Lesemodell)]
    Shop --> Search
    Shop -. spaeter .-> External[Externes Abwicklungstool]
```

Im POC ersetzt ein ausdrücklich als Simulation beschrifteter Verwaltungscommand die externe Liefermeldung. Kreditkartenfreigabe ist eine lokale simulierte Activity ohne Kreditkartennummern und ohne fremden Zahlungsdienst. Rechnung hat keinen Zahlungsgateway-Aufruf. Störungen der Suche beeinflussen keinen maßgeblichen Detail-/Kaufzustand.
