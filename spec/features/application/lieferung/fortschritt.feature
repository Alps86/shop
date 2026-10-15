# language: de
@application @app
Funktionalität: Begrenzte Liefer- und Teilliefersimulation
  Als Gast sehe ich Lieferfortschritt getrennt von unveränderlichen Kaufwerten.

  Szenariogrundriss: Mengen und Reihenfolge von Liefermeldungen
    Angenommen eine angenommene Bestellung wurde noch nicht versandt
      | OrderID | SKU | Purchased | Reserved | Shipped | Delivered |
      | o-a | PHONE | 2 | 2 | 0 | 0 |
    Wenn eine lokale simulierte Liefermeldung verarbeitet wird
      | ShipmentID | CommandID | Status | Quantity |
      | s-a | cmd-a | <meldung> | <menge> |
    Dann lautet die getrennte Lieferansicht
      | Code | State | Shipped | Delivered | Reserved | PurchaseSnapshotChanged |
      | <code> | <zustand> | <versandt> | <geliefert> | <reserviert> | false |
    Beispiele:
      | meldung | menge | code | zustand | versandt | geliefert | reserviert |
      | shipped | 1 | OK | partiallyShipped | 1 | 0 | 1 |
      | shipped | 2 | OK | shipped | 2 | 0 | 0 |
      | shipped | 3 | InvalidQuantity | accepted | 0 | 0 | 2 |
      | shipped | 0 | InvalidQuantity | accepted | 0 | 0 | 2 |
      | delivered | 1 | InvalidTransition | accepted | 0 | 0 | 2 |

  Szenariogrundriss: Wiederholung und vollständige Lieferung
    Angenommen eine Teilmenge wurde bereits versandt
      | OrderID | ShipmentID | CommandID | SKU | Purchased | Shipped | Delivered |
      | o-a | s-a | cmd-a | PHONE | 2 | 1 | 0 |
    Wenn eine weitere Liefermeldung verarbeitet wird
      | ShipmentID | CommandID | Status | SKU | Quantity |
      | <sendung> | <command> | <status> | <sku> | <menge> |
    Dann lautet die Lieferbeobachtung
      | Code | Shipped | Delivered |
      | <code> | <versandt> | <geliefert> |
    Beispiele:
      | sendung | command | status | sku | menge | code | versandt | geliefert |
      | s-a | cmd-a | shipped | PHONE | 1 | OK | 1 | 0 |
      | s-a | cmd-a | shipped | PHONE | 2 | Conflict | 1 | 0 |
      | s-b | cmd-b | shipped | PHONE | 1 | OK | 2 | 0 |
      | s-a | cmd-c | delivered | PHONE | 1 | OK | 1 | 1 |
      | s-a | cmd-c | delivered | UNKNOWN | 1 | InvalidQuantity | 1 | 0 |

  Szenario: Zwei Teillieferungen erreichen den terminalen Lieferstatus
    Angenommen eine Bestellung mit reservierter Kaufmenge ist angenommen
      | OrderID | SKU | Purchased |
      | o-a | PHONE | 2 |
    Wenn die lokale Liefersimulation nacheinander Meldungen verarbeitet
      | ShipmentID | CommandID | Status | Quantity |
      | s-a | cmd-1 | shipped | 1 |
      | s-a | cmd-2 | delivered | 1 |
      | s-b | cmd-3 | shipped | 1 |
      | s-b | cmd-4 | delivered | 1 |
    Dann lautet die endgültige Lieferansicht
      | State | Shipped | Delivered | Reserved | PurchaseSnapshotChanged |
      | delivered | 2 | 2 | 0 | false |

  Szenario: Kauf- und Lieferansicht bleiben getrennt und geordnet
    Angenommen eine Bestellung enthält unterschiedliche Positionen
      | OrderID | UserID | SKU | Purchased |
      | o-a | u-a | Z-CABLE | 1 |
      | o-a | u-a | A-PHONE | 2 |
    Wenn der Eigentümer die Bestellansicht liest
      | Session | OrderID |
      | a | o-a |
    Dann enthält die separate Lieferansicht geordnete Positionen
      | SKU | Purchased | Shipped | Delivered |
      | A-PHONE | 2 | 0 | 0 |
      | Z-CABLE | 1 | 0 | 0 |
