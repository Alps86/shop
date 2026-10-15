# language: de
@spec-only @application
Funktionalität: Sortierung Vorschläge und Ausfallgrenzen
  Als Käufer behalte ich nachvollziehbare Suchreihenfolge und verfügbare Kaufpfade.

  Szenariogrundriss: Explizite Sortierung gewinnt gegen Boosts
    Angenommen vergleichbare synthetische Suchkandidaten liegen vor
      | ProductID | Name | Price | TextScore | InterestWeight | MarginSignal |
      | p-a | Alpha | 10000 | 1 | 0 | 0 |
      | p-b | Beta | 12000 | 1 | 1 | 1 |
    Wenn die Suche sortiert wird
      | Sort |
      | <sortierung> |
    Dann lautet die öffentliche Ergebnisreihenfolge
      | ProductIDs | ExposedMargin | ExposedInterestWeight |
      | <reihenfolge> | false | false |
    Beispiele:
      | sortierung | reihenfolge |
      | relevance | p-b,p-a |
      | priceAsc | p-a,p-b |
      | priceDesc | p-b,p-a |
      | nameAsc | p-a,p-b |

  Szenariogrundriss: Präfix und begrenzte Tippfehler
    Angenommen veröffentlichte Suchbegriffe sind vorhanden
      | Name | ID |
      | Smartphone | p-phone |
      | Smartwatch | p-watch |
    Wenn Vorschläge angefragt werden
      | Query |
      | <eingabe> |
    Dann lauten die Vorschläge
      | Names | Correction | QueryReplaced | MaximumCount |
      | <namen> | <korrektur> | false | 5 |
    Beispiele:
      | eingabe | namen | korrektur |
      | s | [] | false |
      | Sm | Smartphone,Smartwatch | false |
      | Smartphnoe | Smartphone | true |
      | Xmartphone | [] | false |
      | empty | [] | false |

  Szenariogrundriss: Kuratierte Filter und Eingabegrenzen
    Wenn ein Suchrequest validiert wird
      | Field | Value |
      | <feld> | <wert> |
    Dann lautet das Suchvalidierungsergebnis
      | Code | PageSize |
      | <code> | 12 |
    Beispiele:
      | feld | wert | code |
      | Page | 1 | OK |
      | Page | 0 | InvalidFilter |
      | Page | 101 | InvalidFilter |
      | Sort | random | InvalidFilter |
      | Attribute | notCurated | InvalidFilter |
      | FilterOperator | unknown | InvalidFilter |

  Szenario: Suchausfall sperrt den maßgeblichen Kaufpfad nicht
    Angenommen nur OpenSearch ist nicht erreichbar
      | PostgreSQL | Temporal | OpenSearch |
      | ready | ready | down |
    Wenn die öffentlichen Ports aufgerufen werden
      | Operation |
      | SearchPort.Search |
      | CatalogReader.Product |
      | CheckoutService.Place |
    Dann lauten die Portantworten
      | Operation | Code |
      | SearchPort.Search | SearchUnavailable |
      | CatalogReader.Product | OK |
      | CheckoutService.Place | OK |
