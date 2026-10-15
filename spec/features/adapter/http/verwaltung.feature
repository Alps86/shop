# language: de
@spec-only @integration
Funktionalität: Lokale Verwaltung desselben Shops
  Als Betreiber demonstriere ich Inventory und Lieferung über denselben Kern.

  Szenariogrundriss: Lokale Verwaltungsgrenze und klare Fehlercodes
    Wenn ein Verwaltungsrequest ausgeführt wird
      | Listener | Operation | Input |
      | <listener> | <operation> | <eingabe> |
    Dann lautet die Antwort
      | Status | CacheControl | Robots | SideEffect |
      | <status> | no-store | noindex,nofollow | <wirkung> |
    Beispiele:
      | listener | operation | eingabe | status | wirkung |
      | public | GET:/backend | none | 404 | none |
      | local | GET:/backend | none | 200 | none |
      | local | inventoryPrice | validExpectedRevision | 200 | inventoryRevisionAndOutbox |
      | local | inventoryStock | staleExpectedRevision | 409 | none |
      | local | reportDelivery | invalidQuantity | 422 | none |
      | local | reportDelivery | validPartialShipment | 202 | temporalSignalAccepted |

  Szenario: Temporärer Commandtimeout widerruft keinen angenommenen Auftrag
    Angenommen eine Änderung wurde dauerhaft angenommen
      | CommandID | Outcome |
      | cmd-demo | committed |
    Wenn die HTTPantwort vor Empfang abbricht und derselbe Command wiederholt wird
      | CommandID | Payload |
      | cmd-demo | identical |
    Dann lautet das Ergebnis
      | BusinessEffects | SameResult | NewCommandCreated |
      | 1 | true | false |
