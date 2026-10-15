# language: de
@application @app
Funktionalität: Bestätigter Gastabschluss
  Als Gast bestelle ich bewusst mit aktuellen Preisen und simulierter Zahlung.

  Hintergrund:
    Angenommen ein eigener Gastcart ist vollständig vorbereitet
      | UserID | CartID | SKU | Quantity | UnitPrice | StockQuantity | CartRevision | EvaluationID | Recipient | Street | PostalCode | City | Country |
      | u-a | c-a | PHONE | 1 | 10000 | 2 | 4 | e-a | Demo Gast | Testweg 1 | 12345 | Teststadt | DE |

  Szenariogrundriss: Versand und simulierte Zahlungen
    Wenn der Gast den bewerteten Gesamtpreis bestätigt und abschließt
      | ShippingMethod | PaymentMethod | SimulatedAuthorization | EvaluationID | CommandID |
      | <versand> | <zahlung> | <freigabe> | e-a | cmd-a |
    Dann lautet das Abschlussergebnis
      | Code | OrderCount | Reserved | Total | AuthorizationCalls | RealPaymentCalls |
      | <code> | <orders> | <reserviert> | <gesamt> | <aufrufe> | 0 |
    Beispiele:
      | versand | zahlung | freigabe | code | orders | reserviert | gesamt | aufrufe |
      | standard | invoice | absent | OK | 1 | 1 | 10490 | 0 |
      | express | invoice | absent | OK | 1 | 1 | 11490 | 0 |
      | standard | creditCard | approved | OK | 1 | 1 | 10490 | 1 |
      | express | creditCard | approved | OK | 1 | 1 | 11490 | 1 |
      | standard | creditCard | declined | PaymentDeclined | 0 | 0 | absent | 1 |

  Szenariogrundriss: Änderungen erfordern erneute Bestätigung
    Angenommen die Bewertung ist ausdrücklich bestätigt
      | EvaluationID | CartRevision | Total |
      | e-a | 4 | 10490 |
    Wenn vor dem Abschluss eine Änderung stattfindet
      | Change | NewValue |
      | <aenderung> | <wert> |
    Und der alte Abschlussversuch ausgeführt wird
      | CommandID | EvaluationID |
      | cmd-a | e-a |
    Dann lautet das Abschlussergebnis
      | Code | OrderCount | Reserved | NewTotal | ConfirmationPresent | CartPresent | ReturnedConfirmationPresent | ReturnedEvaluationPresent |
      | <code> | 0 | 0 | <gesamt> | false | true | false | true |
    Beispiele:
      | aenderung | wert | code | gesamt |
      | inventoryPrice | 11000 | PriceConfirmationRequired | 11490 |
      | inventoryPrice | 9000 | PriceConfirmationRequired | 9490 |
      | shipping | express | PriceConfirmationRequired | 11490 |
      | quantity | 2 | PriceConfirmationRequired | 20490 |
      | address | Testweg 2 | PriceConfirmationRequired | 10490 |
      | payment | creditCard | PriceConfirmationRequired | 10490 |

  Szenario: Aktive Ersatzwahl statt automatischem Austausch
    Angenommen die bestätigte SKU wurde zwischenzeitlich unverfügbar
      | SKU | AvailableQuantity | SameProductAlternative | AlternativePrice |
      | PHONE | 0 | PHONE-BLUE-256 | 12000 |
    Wenn der bestätigte Abschluss versucht wird
      | CartID | CommandID |
      | c-a | cmd-a |
    Dann wird ein unveränderter Cart mit Ersatzangebot geliefert
      | Code | CartSKU | ProposedSKU | OrderCount | Reserved | UnavailableSKU | AlternativeCount |
      | Unavailable | PHONE | PHONE-BLUE-256 | 0 | 0 | PHONE | 1 |
    Wenn der Gast die Ersatzvariante wählt und den neuen Preis bestätigt
      | SKU | ConfirmedTotal | CommandID |
      | PHONE-BLUE-256 | 12490 | cmd-b |
    Dann entsteht die Bestellung mit der aktiv gewählten SKU
      | SKU | Total | OrderCount |
      | PHONE-BLUE-256 | 12490 | 1 |

  Szenariogrundriss: Ungültige Checkoutdaten
    Wenn Checkoutdaten validiert werden
      | Field | Value |
      | <feld> | <wert> |
    Dann lautet das Validierungsergebnis
      | Code | OrderCount |
      | <code> | 0 |
    Beispiele:
      | feld | wert | code |
      | Recipient | empty | InvalidInput |
      | Street | empty | InvalidInput |
      | City | empty | InvalidInput |
      | PostalCode | 1234 | InvalidInput |
      | PostalCode | abcde | InvalidInput |
      | Country | FR | UnsupportedContext |
      | ShippingMethod | pickup | InvalidInput |
      | PaymentMethod | paypal | InvalidInput |

  Szenario: Neuer Gesamtpreis wird ausdrücklich bestätigt und erfolgreich gekauft
    Angenommen der maßgebliche Preis wurde nach alter Bestätigung geändert
      | OldEvaluationID | OldTotal | NewEvaluationID | NewTotal | NewCartRevision |
      | e-a | 10490 | e-b | 11490 | 5 |
    Wenn der Gast die neue Bewertung bestätigt und abschließt
      | EvaluationID | ExpectedRevision | CommandID |
      | e-b | 5 | cmd-b |
    Dann lautet das Abschlussergebnis
      | Code | OrderCount | Total | Reserved |
      | OK | 1 | 11490 | 1 |

  Szenario: Fehlende Bestätigung verhindert den Kauf
    Wenn ein vorbereiteter Cart ohne Bestätigung abgeschlossen wird
      | EvaluationID | ConfirmedEvaluation.Present | CommandID |
      | e-a | false | cmd-a |
    Dann lautet das Abschlussergebnis
      | Code | OrderCount | Reserved |
      | PriceConfirmationRequired | 0 | 0 |

  Szenario: Der Prozesseingang authentisiert vor dauerhaftem Workflowaufruf
    Wenn der Gast den Abschluss über den Prozesseingang anfordert
      | Session | CartID | CommandID |
      | a | c-a | cmd-a |
    Dann enthält der Prozessaufruf ausschließlich autorisierte Fachkennungen
      | ProcessCalls |
      | 1 |


  Szenario: Ersatzangebote gehören zum selben Produkt und sind verfügbar
    Angenommen der maßgebliche Variantenbestand lautet
      | SKU | ProductID | UnitPrice | StockQuantity |
      | PHONE | p-phone | 10000 | 0 |
      | PHONE-ALT | p-phone | 12000 | 1 |
      | PHONE-EMPTY | p-phone | 9000 | 0 |
      | OTHER-CHEAP | p-other | 1000 | 10 |
    Wenn der bestätigte Abschluss versucht wird
      | CartID | CommandID |
      | c-a | cmd-a |
    Dann wird ein unveränderter Cart mit Ersatzangebot geliefert
      | Code | CartSKU | UnavailableSKU | ProposedSKU | AlternativeCount | OrderCount | Reserved |
      | Unavailable | PHONE | PHONE | PHONE-ALT | 1 | 0 | 0 |
