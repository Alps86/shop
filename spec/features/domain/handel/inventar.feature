# language: de
@spec-only @core
Funktionalität: Preis und Bestand einer SKU
  Als Käufer sehe ich maßgebliche Bestände und exakte Geldbeträge.

  Szenariogrundriss: Inventorygrenzen und Reservierungsverbrauch
    Angenommen ein Inventory enthält die Ausgangswerte
      | SKU | Price | Currency | StockQuantity | Reserved | Revision |
      | DEMO | 1000 | EUR | 3 | 1 | 1 |
    Wenn eine Inventoryoperation ausgeführt wird
      | Operation | Amount | Currency | ExpectedRevision |
      | <operation> | <wert> | <waehrung> | <revision> |
    Dann lautet das Inventoryergebnis
      | Code | StockQuantity | Reserved | AvailableQuantity |
      | <code> | <bestand> | <reserviert> | <verfuegbar> |
    Beispiele:
      | operation | wert | waehrung | revision | code | bestand | reserviert | verfuegbar |
      | reserve | 2 | EUR | 1 | OK | 3 | 3 | 0 |
      | reserve | 3 | EUR | 1 | Unavailable | 3 | 1 | 2 |
      | reserve | 0 | EUR | 1 | InvalidInput | 3 | 1 | 2 |
      | setStock | 0 | EUR | 1 | BelowReserved | 3 | 1 | 2 |
      | setStock | -1 | EUR | 1 | InvalidInput | 3 | 1 | 2 |
      | setPrice | 0 | EUR | 1 | OK | 3 | 1 | 2 |
      | setPrice | -1 | EUR | 1 | InvalidMoney | 3 | 1 | 2 |
      | setPrice | 1000 | USD | 1 | UnsupportedContext | 3 | 1 | 2 |
      | setPrice | 1200 | EUR | 0 | Conflict | 3 | 1 | 2 |
      | ship | 1 | EUR | 1 | OK | 2 | 0 | 2 |

  Szenariogrundriss: Geldoperationen bleiben exakt
    Wenn eine Geldoperation ausgeführt wird
      | Operation | Left | Right | CurrencyLeft | CurrencyRight |
      | <operation> | <links> | <rechts> | <wl> | <wr> |
    Dann lautet das Geldergebnis
      | Code | MinorUnits |
      | <code> | <betrag> |
    Beispiele:
      | operation | links | rechts | wl | wr | code | betrag |
      | add | 100 | 200 | EUR | EUR | OK | 300 |
      | add | 100 | 200 | EUR | USD | InvalidMoney | absent |
      | multiply | 9223372036854775807 | 2 | EUR | EUR | AmountOverflow | absent |
      | add | 9223372036854775807 | 1 | EUR | EUR | AmountOverflow | absent |
