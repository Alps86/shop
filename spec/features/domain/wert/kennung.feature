# language: de
@core @value
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

  Szenariogrundriss: Register- und Präzisionsgrenzen
    Wenn ein Wert konstruiert wird
      | Type | Input |
      | <typ> | <eingabe> |
    Dann lautet das Wertergebnis
      | Code | Canonical |
      | <code> | <kanonisch> |
    Beispiele:
      | typ | eingabe | code | kanonisch |
      | Locale | und | InvalidLocale | absent |
      | CountryCode | EU | InvalidInput | absent |
      | CountryCode | ?? | InvalidInput | absent |
      | CountryCode | DEU | InvalidInput | absent |
      | CountryCode | XK | InvalidInput | absent |
      | Decimal | 999999999999.999999 | OK | 999999999999.999999 |
      | Money | 100:eur | OK | 100:EUR |
      | Money | 100:ZZZ | InvalidMoney | absent |
      | Money | -1:EUR | InvalidMoney | absent |

  Szenariogrundriss: Explizite JSON-Präsenz an der Eingangsgrenze
    Wenn ein JSON-Wert geprüft wird
      | Kind | Input |
      | <art> | <eingabe> |
    Dann lautet das Wertergebnis
      | Code | Canonical |
      | <code> | <kanonisch> |
    Beispiele:
      | art | eingabe | code | kanonisch |
      | required | 0 | OK | 0 |
      | required | broken | InvalidInput | absent |
      | replacement | "text" | OK | text |
      | replacement | broken | InvalidInput | absent |

  Szenario: Technischer Fehler wird nicht als fachlicher Erfolg ausgegeben
    Wenn ein technischer Fehler eingeordnet wird
      | Message |
      | storage unavailable |
    Dann lautet das Wertergebnis
      | Code | Canonical |
      | InternalError | absent |

  Szenariogrundriss: Exakte skalierte Darstellung ohne stillen Ersatzwert
    Wenn eine skalierte Zahl angefragt wird
      | Input |
      | <eingabe> |
    Dann lautet das Wertergebnis
      | Code | Canonical |
      | <code> | <kanonisch> |
    Beispiele:
      | eingabe | code | kanonisch |
      | 1.23 | OK | 1230000 |
      | invalid | InvalidInput | absent |

  Szenariogrundriss: Exakte Dezimalvergleiche validieren beide Operanden
    Wenn zwei exakte Dezimalwerte verglichen werden
      | Left | Right |
      | <left> | <right> |
    Dann lautet das Wertergebnis
      | Code | Canonical |
      | <code> | <result> |
    Beispiele:
      | left | right | code | result |
      | 1.00 | 1 | OK | 0 |
      | 0.9 | 1 | OK | -1 |
      | 1.1 | 1 | OK | 1 |
      | invalid | 1 | InvalidInput | absent |
      | 1 | invalid | InvalidInput | absent |
