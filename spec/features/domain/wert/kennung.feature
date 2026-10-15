# language: de
@spec-only @core
Funktionalität: Validierte Kennungen und Kontexte
  Als Aufrufer erhalte ich validierte Werte statt stiller Ersatzwerte.

  Szenariogrundriss: Validierung und Kanonisierung
    Wenn ein Wert konstruiert wird
      | Type | Input |
      | <typ> | <eingabe> |
    Dann lautet das Wertergebnis
      | Code | Canonical |
      | <code> | <kanonisch> |
    Beispiele:
      | typ | eingabe | code | kanonisch |
      | ProductID | p-123 | OK | p-123 |
      | CartID | cart_1 | OK | cart_1 |
      | UserID | user-1 | OK | user-1 |
      | OrderID | order-1 | OK | order-1 |
      | ProductFamilyID | f-1 | OK | f-1 |
      | CategoryID | c-1 | OK | c-1 |
      | ReservationID | r-1 | OK | r-1 |
      | DiscountRuleID | rule-1 | OK | rule-1 |
      | ShipmentID | s-1 | OK | s-1 |
      | CommandID | cmd-1 | OK | cmd-1 |
      | EvaluationID | e-1 | OK | e-1 |
      | SKU | Phone_BLUE | OK | Phone_BLUE |
      | SKU | bad sku | InvalidInput | absent |
      | ProductID | empty | InvalidInput | absent |
      | ProductID | /path | InvalidInput | absent |
      | Locale | de-de | OK | de-DE |
      | Locale | zz-ZZ | InvalidLocale | absent |
      | CountryCode | de | OK | DE |
      | CountryCode | ZZ | InvalidInput | absent |
      | Currency | eur | OK | EUR |
      | Currency | ZZZ | InvalidMoney | absent |
      | Decimal | -0 | OK | 0 |
      | Decimal | 1.230000 | OK | 1.23 |
      | Decimal | 01 | InvalidInput | absent |
      | Decimal | 1e3 | InvalidInput | absent |
      | Decimal | +1 | InvalidInput | absent |
      | Decimal | 0.0000001 | InvalidInput | absent |
      | Decimal | 1000000000000 | InvalidInput | absent |

  Szenariogrundriss: Gültiger Wert ist nicht automatisch unterstützt
    Wenn ein Verkaufskontext gewählt wird
      | Locale | Currency | Market |
      | <sprache> | <waehrung> | <land> |
    Dann lautet das Kontextergebnis
      | Code |
      | <code> |
    Beispiele:
      | sprache | waehrung | land | code |
      | de-DE | EUR | DE | OK |
      | fr-FR | EUR | DE | UnsupportedContext |
      | de-DE | USD | DE | UnsupportedContext |
      | de-DE | EUR | FR | UnsupportedContext |
