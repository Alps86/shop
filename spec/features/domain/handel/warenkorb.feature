# language: de
@spec-only @core
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
