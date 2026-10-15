# language: de
@core @sale
Funktionalität: Warenkorbpositionen und Kaufkopien
  Als Gast ändere ich Mengen vor dem Kauf und behalte danach unveränderliche Kaufwerte.

  Szenariogrundriss: Mengenänderung reserviert nicht
    Angenommen ein offener Warenkorb enthält eine Position
      | SKU | Quantity | CartRevision | PositionCount | Reserved |
      | PHONE | <vorher> | 1 | <positionen> | 0 |
    Wenn eine Warenkorboperation ausgeführt wird
      | Operation | SKU | Quantity | ExpectedRevision |
      | <operation> | <sku> | <menge> | 1 |
    Dann lautet das Warenkorbergebnis
      | Code | QuantityPHONE | PositionCount | Reserved | ConfirmationPresent |
      | <code> | <nachher> | <anzahl> | 0 | false |
    Beispiele:
      | vorher | positionen | operation | sku | menge | code | nachher | anzahl |
      | 1 | 1 | add | PHONE | 2 | OK | 3 | 1 |
      | 99 | 1 | add | PHONE | 1 | InvalidInput | 99 | 1 |
      | 1 | 1 | setQuantity | PHONE | 0 | InvalidInput | 1 | 1 |
      | 1 | 1 | setQuantity | PHONE | 100 | InvalidInput | 1 | 1 |
      | 1 | 1 | setQuantity | PHONE | -1 | InvalidInput | 1 | 1 |
      | 1 | 1 | remove | PHONE | absent | OK | 0 | 0 |
      | 1 | 50 | add | NEW | 1 | InvalidInput | 1 | 50 |

  Szenario: Kauf-Snapshot bleibt bei späteren Änderungen erhalten
    Angenommen eine abgeschlossene Bestellung besitzt Kaufwerte
      | OrderID | SKU | ProductName | Color | Quantity | UnitPrice | Discount | Recipient | Street | PostalCode | City | Country | DefinitionVersion | RuleSetVersion | ShippingMethod | PaymentMethod | ShippingCost | Taxes | Total |
      | o-1 | PHONE | Testmodell | blue | 1 | 10000 | 0 | Demo Gast | Testweg 1 | 12345 | Teststadt | DE | order-v1 | demo-v1 | standard | invoice | 490 | 1675 | 10490 |
    Wenn Katalog und Inventory später geändert werden
      | ProductName | Color | UnitPrice |
      | Neuer Modellname | red | 15000 |
    Dann enthält die öffentliche Kaufansicht unveränderte Werte
      | OrderID | SKU | ProductName | Color | Quantity | UnitPrice | Discount | Recipient | Street | PostalCode | City | Country | DefinitionVersion | RuleSetVersion | ShippingMethod | PaymentMethod | ShippingCost | Taxes | Total |
      | o-1 | PHONE | Testmodell | blue | 1 | 10000 | 0 | Demo Gast | Testweg 1 | 12345 | Teststadt | DE | order-v1 | demo-v1 | standard | invoice | 490 | 1675 | 10490 |

  Szenariogrundriss: Terminaler Cart wird nicht wieder geöffnet
    Angenommen ein Cart besitzt einen terminalen Zustand
      | State | CartRevision | ItemCount |
      | <zustand> | 5 | 1 |
    Wenn eine Warenkorbänderung angefordert wird
      | Operation | SKU | Quantity |
      | add | PHONE | 1 |
    Dann lautet die unveränderte Antwort
      | Code | State | CartRevision | ItemCount |
      | <code> | <zustand> | 5 | 1 |
    Beispiele:
      | zustand | code |
      | ordered | AlreadyOrdered |
      | expired | Expired |

  Szenariogrundriss: Gastwarenkorb benötigt eigene Kennungen
    Wenn ein Gastwarenkorb angelegt wird
      | CartID | UserID |
      | <cart> | <user> |
    Dann lautet der fachliche Fehlercode
      | Code |
      | <code> |
    Beispiele:
      | cart | user | code |
      | empty | u-a | InvalidInput |
      | bad id | u-a | InvalidInput |
      | c-a | empty | InvalidInput |
      | c-a | bad id | InvalidInput |
      | c-a | u-a | OK |

  Szenariogrundriss: Änderungen berücksichtigen Revision und Zeitgrenze
    Angenommen ein offener Warenkorb enthält eine Position
      | SKU | Quantity | CartRevision | PositionCount | Reserved |
      | PHONE | 1 | 1 | 2 | 0 |
    Wenn eine zeitbezogene Cartoperation ausgeführt wird
      | Operation | SKU | Quantity | ExpectedRevision | HoursSinceActivity |
      | <operation> | <sku> | <menge> | <revision> | <stunden> |
    Dann lautet der zeitbezogene Cartzustand
      | Code | State | ItemCount | QuantityPHONE |
      | <code> | <zustand> | <anzahl> | <telefonmenge> |
    Beispiele:
      | operation | sku | menge | revision | stunden | code | zustand | anzahl | telefonmenge |
      | add | PHONE | 0 | 1 | 0 | InvalidInput | open | 2 | 1 |
      | add | empty | 1 | 1 | 0 | InvalidInput | open | 2 | 1 |
      | add | PHONE | 1 | 0 | 0 | Conflict | open | 2 | 1 |
      | remove | PHONE | 0 | 0 | 0 | Conflict | open | 2 | 1 |
      | remove | PHONE | 0 | 1 | 0 | OK | open | 1 | 0 |
      | add | PHONE | 1 | 1 | 167 | OK | open | 2 | 2 |
      | add | PHONE | 1 | 1 | 168 | Expired | open | 2 | 1 |
      | expire | PHONE | 1 | 1 | 167 | OK | open | 2 | 1 |
      | expire | PHONE | 1 | 1 | 168 | OK | expired | 2 | 1 |
      | expire | PHONE | 1 | 1 | 169 | OK | expired | 2 | 1 |

  Szenariogrundriss: Checkoutdaten vor Bestätigung validieren
    Wenn Checkoutdaten validiert werden
      | Recipient | Street | PostalCode | City | Country | ShippingMethod | PaymentMethod |
      | <empfaenger> | <strasse> | <plz> | <stadt> | <land> | <versand> | <zahlung> |
    Dann lautet der fachliche Fehlercode
      | Code |
      | <code> |
    Beispiele:
      | empfaenger | strasse | plz | stadt | land | versand | zahlung | code |
      | Demo | Testweg 1 | 12345 | Teststadt | DE | standard | invoice | OK |
      | Demo | Testweg 1 | 12345 | Teststadt | DE | express | creditCard | OK |
      | empty | Testweg 1 | 12345 | Teststadt | DE | standard | invoice | InvalidInput |
      | Demo | empty | 12345 | Teststadt | DE | standard | invoice | InvalidInput |
      | Demo | Testweg 1 | 12345 | empty | DE | standard | invoice | InvalidInput |
      | Demo | Testweg 1 | 1234 | Teststadt | DE | standard | invoice | InvalidInput |
      | Demo | Testweg 1 | 12345 | Teststadt | AT | standard | invoice | UnsupportedContext |
      | Demo | Testweg 1 | 12345 | Teststadt | DE | invalid | invoice | InvalidInput |
      | Demo | Testweg 1 | 12345 | Teststadt | DE | standard | invalid | InvalidInput |

  Szenario: Bestätigter Kauf und terminale Mutationen
    Angenommen ein offener Warenkorb enthält eine Position
      | SKU | Quantity | CartRevision | PositionCount | Reserved |
      | PHONE | 1 | 1 | 1 | 0 |
    Wenn der fachliche Kaufablauf schrittweise ausgeführt wird
      | Operation | EvaluationID | Revision | RuleSetVersion | ExpectedCode | ExpectedState |
      | configure | e-a | 0 | demo-v1 | Conflict | open |
      | configureInvalid | e-a | 1 | demo-v1 | InvalidInput | open |
      | configure | e-a | 1 | demo-v1 | OK | checkingOut |
      | evaluate | e-a | 1 | demo-v1 | Conflict | checkingOut |
      | evaluate | e-a | 2 | invalid | Conflict | checkingOut |
      | confirm | e-a | 0 | demo-v1 | Conflict | checkingOut |
      | confirm | e-a | 2 | demo-v1 | PriceConfirmationRequired | checkingOut |
      | markOrdered | e-a | 2 | demo-v1 | PriceConfirmationRequired | checkingOut |
      | createOrder | e-a | 2 | demo-v1 | PriceConfirmationRequired | checkingOut |
      | evaluate | e-a | 2 | demo-v1 | OK | checkingOut |
      | confirm | e-other | 2 | demo-v1 | PriceConfirmationRequired | checkingOut |
      | confirm | e-a | 2 | demo-v1 | OK | checkingOut |
      | createEmptyOrder | e-a | 2 | demo-v1 | InvalidInput | checkingOut |
      | createOrder | e-a | 2 | demo-v1 | OK | checkingOut |
      | markOrdered | e-a | 2 | demo-v1 | OK | ordered |
      | expire | e-a | 3 | demo-v1 | AlreadyOrdered | ordered |
      | configure | e-a | 3 | demo-v1 | AlreadyOrdered | ordered |
    Dann enthält der erzeugte Kauf seine eigenen Preiswerte
      | OrderID | SKU | Quantity | UnitPrice | Total |
      | o-new | PHONE | 1 | 10000 | 10490 |

  Szenariogrundriss: Lieferfortschritt ist ein separater Fachzustand
    Angenommen eine Bestellung wird für die Lieferdemo verwendet
      | OrderID | SKU | Purchased |
      | o-a | PHONE | 2 |
    Wenn fachliche Liefermeldungen nacheinander angewendet werden
      | OrderID | ShipmentID | CommandID | Status | SKU | Quantity | Code |
      | o-a | s-a | cmd-a | <status> | <sku> | <menge> | <code> |
    Dann lautet der fachliche Lieferzustand
      | State | Shipped | Delivered |
      | <zustand> | <versandt> | <geliefert> |
    Beispiele:
      | status | sku | menge | code | zustand | versandt | geliefert |
      | shipped | PHONE | 1 | OK | partiallyShipped | 1 | 0 |
      | shipped | PHONE | 2 | OK | shipped | 2 | 0 |
      | shipped | PHONE | 3 | InvalidQuantity | accepted | 0 | 0 |
      | shipped | PHONE | 0 | InvalidQuantity | accepted | 0 | 0 |
      | shipped | UNKNOWN | 1 | InvalidQuantity | accepted | 0 | 0 |
      | delivered | PHONE | 1 | InvalidTransition | accepted | 0 | 0 |
      | invalid | PHONE | 1 | InvalidTransition | accepted | 0 | 0 |

  Szenario: Wiederholte Meldungen und zwei Teillieferungen
    Angenommen eine Bestellung wird für die Lieferdemo verwendet
      | OrderID | SKU | Purchased |
      | o-a | PHONE | 2 |
    Wenn fachliche Liefermeldungen nacheinander angewendet werden
      | OrderID | ShipmentID | CommandID | Status | SKU | Quantity | Code |
      | o-b | s-a | cmd-a | shipped | PHONE | 1 | InvalidReference |
      | o-a | s-a | cmd-a | shipped | none | 0 | InvalidQuantity |
      | o-a | s-a | cmd-a | shipped | PHONE | 1 | OK |
      | o-a | s-a | cmd-a | shipped | PHONE | 1 | OK |
      | o-a | s-a | cmd-a | shipped | PHONE | 2 | Conflict |
      | o-a | s-a | cmd-b | delivered | PHONE | 1 | OK |
      | o-a | s-b | cmd-c | shipped | PHONE | 1 | OK |
      | o-a | s-b | cmd-d | delivered | PHONE | 1 | OK |
    Dann lautet der fachliche Lieferzustand
      | State | Shipped | Delivered |
      | delivered | 2 | 2 |

  Szenariogrundriss: Kaufkopien müssen dem bestätigten Warenkorb entsprechen
    Angenommen ein offener Warenkorb enthält eine Position
      | SKU | Quantity | CartRevision | PositionCount | Reserved |
      | PHONE | 1 | 1 | 1 | 0 |
    Wenn der fachliche Kaufablauf schrittweise ausgeführt wird
      | Operation | EvaluationID | Revision | RuleSetVersion | ExpectedCode | ExpectedState |
      | configure | e-a | 1 | demo-v1 | OK | checkingOut |
      | evaluate | e-a | 2 | demo-v1 | OK | checkingOut |
      | confirm | e-a | 2 | demo-v1 | OK | checkingOut |
    Und eine Bestellung mit expliziten Kaufkopien angefordert wird
      | OrderID | ItemCount | SKU | Quantity | UnitPrice | State | HoursSinceActivity |
      | <kennung> | <anzahl> | <sku> | <menge> | <preis> | <zustand> | <stunden> |
    Dann lautet der fachliche Fehlercode
      | Code |
      | <code> |
    Beispiele:
      | kennung | anzahl | sku | menge | preis | zustand | stunden | code |
      | empty | 1 | PHONE | 1 | 10000 | checkingOut | 0 | InvalidInput |
      | o-a | 2 | PHONE | 1 | 10000 | checkingOut | 0 | Conflict |
      | o-a | 1 | PHONE | 2 | 10000 | checkingOut | 0 | Conflict |
      | o-a | 1 | PHONE | 1 | 9999 | checkingOut | 0 | Conflict |
      | o-a | 1 | OTHER | 1 | 10000 | checkingOut | 0 | Conflict |
      | o-a | 1 | PHONE | 1 | 10000 | ordered | 0 | AlreadyOrdered |
      | o-a | 1 | PHONE | 1 | 10000 | expired | 0 | Expired |
      | o-a | 1 | PHONE | 1 | 10000 | checkingOut | 168 | Expired |

  Szenario: Ein abgelaufener Warenkorb wird niemals bestellt
    Angenommen ein Cart besitzt einen terminalen Zustand
      | State | CartRevision | ItemCount |
      | expired | 5 | 1 |
    Wenn der fachliche Kaufablauf schrittweise ausgeführt wird
      | Operation | EvaluationID | Revision | RuleSetVersion | ExpectedCode | ExpectedState |
      | markOrdered | e-a | 5 | demo-v1 | Expired | expired |

  Szenario: Eine Liefermeldung darf dieselbe Position nicht mehrfach zählen
    Angenommen eine Bestellung wird für die Lieferdemo verwendet
      | OrderID | SKU | Purchased |
      | o-a | PHONE | 2 |
    Wenn fachliche Liefermeldungen nacheinander angewendet werden
      | OrderID | ShipmentID | CommandID | Status | SKU | Quantity | Code |
      | o-a | s-a | cmd-a | shipped | PHONE,PHONE | 2 | InvalidQuantity |
    Dann lautet der fachliche Lieferzustand
      | State | Shipped | Delivered |
      | accepted | 0 | 0 |

  Szenario: Zugeordnete Kauf-Rabatte besitzen eigene verschachtelte Kopien
    Angenommen eine synthetische Rabattregel gilt für Testprodukte
      | RuleID | MainProductID | AccessoryID | Scope | Amount | Currency |
      | r-demo | p-phone | p-case | pair | 500 | EUR |
    Und die bewertbaren Positionen lauten
      | SKU | ProductID | Quantity | UnitPrice | Currency |
      | PHONE | p-phone | 1 | 10000 | EUR |
      | CASE | p-case | 1 | 2000 | EUR |
    Wenn der Warenkorb bewertet wird
      | ShippingMethod | Profile | TaxRate |
      | standard | neutral | 19 |
    Und die Bewertung als Bestellung bestätigt wird
      | CartID | UserID | OrderID | EvaluationID | Recipient | Street | PostalCode | City | Country | ShippingMethod | PaymentMethod |
      | c-copy | u-copy | o-copy | e-copy | Demo | Testweg 1 | 12345 | Teststadt | DE | standard | invoice |
    Und die ursprünglichen Rabattdaten geändert werden
      | Amount | SKU |
      | 1 | CHANGED |
    Dann besitzt die Kaufkopie unveränderte Rabattdaten
      | OrderID | RuleID | Amount | SKU |
      | o-copy | r-demo | 500 | CASE |

  Szenariogrundriss: Verpacken ist eine optionale Meldung ohne Versandverbrauch
    Angenommen eine Bestellung wird für die Lieferdemo verwendet
      | OrderID | SKU | Purchased |
      | o-a | PHONE | 2 |
    Wenn fachliche Liefermeldungen nacheinander angewendet werden
      | OrderID | ShipmentID | CommandID | Status | SKU | Quantity | Code |
      | o-a | s-a | cmd-a | packed | PHONE | <menge> | <code> |
    Dann lautet der fachliche Lieferzustand
      | State | Shipped | Delivered | PackedQuantity |
      | accepted | 0 | 0 | <gepackt> |
    Beispiele:
      | menge | code | gepackt |
      | 1 | OK | 1 |
      | 2 | OK | 2 |
      | 3 | InvalidQuantity | 0 |

  Szenario: Verpackung und Versand bleiben getrennte Mengen
    Angenommen eine Bestellung wird für die Lieferdemo verwendet
      | OrderID | SKU | Purchased |
      | o-a | PHONE | 2 |
    Wenn fachliche Liefermeldungen nacheinander angewendet werden
      | OrderID | ShipmentID | CommandID | Status | SKU | Quantity | Code |
      | o-a | s-a | cmd-a | packed | PHONE | 2 | OK |
      | o-a | s-a | cmd-a | packed | PHONE | 2 | OK |
      | o-a | s-a | cmd-b | shipped | PHONE | 2 | OK |
    Dann lautet der fachliche Lieferzustand
      | State | Shipped | Delivered | PackedQuantity |
      | shipped | 2 | 0 | 2 |

  Szenario: Verpackungsmengen überschreiten die Kaufmenge nicht
    Angenommen eine Bestellung wird für die Lieferdemo verwendet
      | OrderID | SKU | Purchased |
      | o-a | PHONE | 2 |
    Wenn fachliche Liefermeldungen nacheinander angewendet werden
      | OrderID | ShipmentID | CommandID | Status | SKU | Quantity | Code |
      | o-a | s-a | cmd-a | packed | PHONE | 1 | OK |
      | o-a | s-b | cmd-b | packed | PHONE | 2 | InvalidQuantity |
    Dann lautet der fachliche Lieferzustand
      | State | Shipped | Delivered | PackedQuantity |
      | accepted | 0 | 0 | 1 |

  Szenario: Vollständige Lieferung ist auch gegenüber Verpackungsmeldungen terminal
    Angenommen eine Bestellung wird für die Lieferdemo verwendet
      | OrderID | SKU | Purchased |
      | o-a | PHONE | 2 |
    Wenn fachliche Liefermeldungen nacheinander angewendet werden
      | OrderID | ShipmentID | CommandID | Status | SKU | Quantity | Code |
      | o-a | s-a | cmd-a | shipped | PHONE | 2 | OK |
      | o-a | s-a | cmd-b | delivered | PHONE | 2 | OK |
      | o-a | s-b | cmd-c | packed | PHONE | 1 | InvalidTransition |
    Dann lautet der fachliche Lieferzustand
      | State | Shipped | Delivered | PackedQuantity |
      | delivered | 2 | 2 | 0 |
