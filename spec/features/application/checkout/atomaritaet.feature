# language: de
@spec-only @application
Funktionalität: Atomarer und idempotenter Kauf
  Als Gast erzeuge ich trotz Konkurrenz und Wiederholung höchstens einen Kauf.

  Szenariogrundriss: Wiederholung und Prozesslücken
    Angenommen ein gültiger Abschlussversuch ist vorbereitet
      | CartID | CommandID | EvaluationID | SKU | Quantity | StockQuantity |
      | c-a | cmd-a | e-a | PHONE | 1 | 2 |
    Wenn der Abschluss mit einer Störung und Wiederholung ausgeführt wird
      | Fault | RetryCommandID | RetryPayload |
      | <stoerung> | <command> | <payload> |
    Dann lautet das dauerhafte Kaufergebnis
      | Code | OrderCount | Reservations | CartState | OrderID |
      | <code> | <orders> | <reserviert> | <zustand> | <order> |
    Beispiele:
      | stoerung | command | payload | code | orders | reserviert | zustand | order |
      | none | cmd-a | identical | OK | 1 | 1 | ordered | o-a |
      | afterCommitBeforeResponse | cmd-a | identical | OK | 1 | 1 | ordered | o-a |
      | beforeCommit | cmd-a | identical | OK | 1 | 1 | ordered | o-a |
      | afterCommitBeforeOrderStart | cmd-a | identical | OK | 1 | 1 | ordered | o-a |
      | none | cmd-b | identical | AlreadyOrdered | 1 | 1 | ordered | o-a |
      | none | cmd-a | different | Conflict | 1 | 1 | ordered | o-a |

  Szenario: Zwei Gäste konkurrieren um den letzten Bestand
    Angenommen zwei bestätigte Carts benötigen dieselbe letzte Einheit
      | CartID | UserID | SKU | Quantity | AvailableQuantity |
      | c-a | u-a | LAST | 1 | 1 |
      | c-b | u-b | LAST | 1 | 1 |
    Wenn beide Kauftransaktionen gleichzeitig freigegeben werden
    Dann lautet das gemeinsame Ergebnis unabhängig vom Gewinner
      | SuccessfulOrders | UnavailableResults | Reserved | AvailableQuantity | PartialOrders |
      | 1 | 1 | 1 | 0 | 0 |

  Szenario: Fehler innerhalb der Kauftransaktion hinterlässt keine Teilwirkung
    Angenommen ein bestätigter Cart benötigt mehrere SKUs
      | SKU | Quantity |
      | A | 1 |
      | B | 1 |
    Wenn der Commit nach dem ersten Reservierungsschritt fehlschlägt
      | Fault |
      | transactionRollback |
    Dann lautet der persistierte Zustand
      | OrderCount | ReservationCount | OrderOutboxCount | CartState |
      | 0 | 0 | 0 | checkingOut |

  Szenario: Reine Bestandsrevision entwertet keinen unveränderten Preis
    Angenommen eine bestätigte Bewertung ist weiterhin verfügbar
      | UnitPrice | AvailableBefore | AvailableNow | InventoryRevisionBefore | InventoryRevisionNow |
      | 10000 | 3 | 2 | 1 | 2 |
    Wenn derselbe bestätigte Kauf abgeschlossen wird
    Dann lautet das Abschlussergebnis
      | Code | OrderCount |
      | OK | 1 |
