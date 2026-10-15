# language: de
@application @app
Funktionalität: Atomarer und idempotenter Kauf
  Als Gast erzeuge ich trotz Konkurrenz und Wiederholung höchstens einen Kauf.

  Grundlage:
    Angenommen die synthetischen Kaufvertragsgrunddaten gelten
      | CartID | UserID | CommandID | EvaluationID | SKU | Quantity | StockQuantity | UnitPrice | CartRevision | Recipient | Street | PostalCode | City | Country | ShippingMethod | PaymentMethod | SimulatedAuthorization |
      | c-a | u-a | cmd-a | e-a | PHONE | 1 | 2 | 10000 | 1 | Demo | Testweg 1 | 12345 | Teststadt | DE | standard | invoice | approved |

  Szenariogrundriss: Wiederholung und Prozesslücken
    Angenommen ein gültiger Abschlussversuch ist vorbereitet
      | CartID | CommandID | EvaluationID | SKU | Quantity | StockQuantity |
      | c-a | cmd-a | e-a | PHONE | 1 | 2 |
    Wenn der Abschluss mit einer Störung und Wiederholung ausgeführt wird
      | Fault | RetryCommandID | RetryPayload | FirstCode | FirstOrders |
      | <stoerung> | <command> | <payload> | <erstcode> | <erstorders> |
    Dann lautet das dauerhafte Kaufergebnis
      | Code | OrderCount | Reservations | CartState | OrderID | OrderOutboxCount |
      | <code> | <orders> | <reserviert> | <zustand> | <order> | 1 |
    Beispiele:
      | stoerung | command | payload | code | orders | reserviert | zustand | order | erstcode | erstorders |
      | none | cmd-a | identical | OK | 1 | 1 | ordered | o-a | OK | 1 |
      | afterCommitBeforeResponse | cmd-a | identical | OK | 1 | 1 | ordered | o-a | DependencyUnavailable | 1 |
      | beforeCommit | cmd-a | identical | OK | 1 | 1 | ordered | o-a | DependencyUnavailable | 0 |
      | afterCommitBeforeOrderStart | cmd-a | identical | OK | 1 | 1 | ordered | o-a | DependencyUnavailable | 1 |
      | none | cmd-b | identical | AlreadyOrdered | 1 | 1 | ordered | o-a | OK | 1 |
      | none | cmd-a | different | Conflict | 1 | 1 | ordered | o-a | OK | 1 |

  Szenario: Zwei Gäste konkurrieren um den letzten Bestand
    Angenommen zwei bestätigte Carts benötigen dieselbe letzte Einheit
      | CartID | UserID | SKU | Quantity | AvailableQuantity |
      | c-a | u-a | LAST | 1 | 1 |
      | c-b | u-b | LAST | 1 | 1 |
    Wenn beide Kauftransaktionen gleichzeitig freigegeben werden
    Dann lautet das gemeinsame Ergebnis unabhängig vom Gewinner
      | SuccessfulOrders | UnavailableResults | Reserved | AvailableQuantity | PartialOrders | OrderCount |
      | 1 | 1 | 1 | 0 | 0 | 1 |

  Szenario: Fehler innerhalb der Kauftransaktion hinterlässt keine Teilwirkung
    Angenommen ein bestätigter Cart benötigt mehrere SKUs
      | SKU | Quantity |
      | A | 1 |
      | B | 1 |
    Wenn der Commit nach dem ersten Reservierungsschritt fehlschlägt
      | Fault |
      | transactionRollback |
    Dann lautet der persistierte Zustand
      | Code | OrderCount | ReservationCount | OrderOutboxCount | CartState |
      | DependencyUnavailable | 0 | 0 | 0 | checkingOut |

  Szenario: Reine Bestandsrevision entwertet keinen unveränderten Preis
    Angenommen eine bestätigte Bewertung ist weiterhin verfügbar
      | UnitPrice | AvailableBefore | AvailableNow | InventoryRevisionBefore | InventoryRevisionNow |
      | 10000 | 3 | 2 | 1 | 2 |
    Wenn derselbe bestätigte Kauf abgeschlossen wird
    Dann lautet das Abschlussergebnis
      | Code | OrderCount |
      | OK | 1 |

  Szenario: Gegenläufige Positionspreise benötigen trotz gleicher Gesamtsumme Bestätigung
    Angenommen ein bestätigter Cart benötigt mehrere SKUs
      | SKU | Quantity |
      | A | 1 |
      | B | 1 |
    Wenn sich die Positionspreise vor dem Abschluss ändern
      | SKU | UnitPrice |
      | A | 10100 |
      | B | 9900 |
    Und derselbe bestätigte Kauf abgeschlossen wird
    Dann lautet das dauerhafte Kaufergebnis
      | Code | OrderCount | Reservations | CartState | OrderID | OrderOutboxCount |
      | PriceConfirmationRequired | 0 | 0 | checkingOut | absent | 0 |

  Szenario: Kreditkartenfreigabe wird beim identischen Retry nicht wiederholt
    Angenommen ein gültiger Abschlussversuch ist vorbereitet
      | CartID | CommandID | EvaluationID | SKU | Quantity | StockQuantity | PaymentMethod |
      | c-a | cmd-a | e-a | PHONE | 1 | 2 | creditCard |
    Wenn der Abschluss mit einer Störung und Wiederholung ausgeführt wird
      | Fault | RetryCommandID | RetryPayload | FirstCode | FirstOrders |
      | afterCommitBeforeResponse | cmd-a | identical | DependencyUnavailable | 1 |
    Dann lautet das dauerhafte Kaufergebnis
      | Code | OrderCount | Reservations | CartState | OrderID | OrderOutboxCount | AuthorizationCalls |
      | OK | 1 | 1 | ordered | o-a | 1 | 1 |
