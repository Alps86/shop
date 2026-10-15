# language: de
@spec-only @runtime
Funktionalität: Validierter versionierter Demoimport
  Als Betreiber veröffentliche ich nur kohärente Datensätze mit nachvollziehbarer Herkunft.

  Szenariogrundriss: Publikation ohne Teilzustand
    Angenommen der aktive Demodatensatz ist versioniert
      | DatasetVersion | ContentDigest | ProductCount |
      | demo-v1 | digest-a | 4 |
    Wenn ein Importpaket verarbeitet wird
      | DatasetVersion | ContentDigest | Validation |
      | <version> | <inhalt> | <validierung> |
    Dann lautet das Publikationsergebnis
      | Code | ActiveDatasetVersion | PartialPublication | InventoryOverwritten |
      | <code> | <aktiv> | false | false |
    Beispiele:
      | version | inhalt | validierung | code | aktiv |
      | demo-v1 | digest-a | valid | Noop | demo-v1 |
      | demo-v1 | digest-b | valid | Conflict | demo-v1 |
      | demo-v2 | digest-b | valid | OK | demo-v2 |
      | demo-v2 | digest-b | duplicateSKU | DuplicateSKU | demo-v1 |
      | demo-v2 | digest-b | missingFamily | InvalidReference | demo-v1 |
      | demo-v2 | digest-b | invalidAttributeLevel | InvalidAttribute | demo-v1 |
      | demo-v2 | digest-b | unsafeDocumentURL | InvalidInput | demo-v1 |

  Szenario: Unverifizierte reale Lieferantendaten werden nicht erfunden
    Wenn ein aus öffentlicher Referenz abgeleitetes Produkt unvollständige Pflichtdaten besitzt
      | SourceURL | RetrievedAt | LicenseNote | MissingFields | VerificationStatus |
      | https://www.ep.de/p/1045137/ecodecalk-entkalker-500ml | 2026-10-15T00:00:00Z | source-only-rights-unverified | Manufacturer,Safety | pending |
    Dann lautet die Veröffentlichungssicht
      | Purchasable | GeneratedWarnings | ClaimsLegalCompliance | StatusText |
      | false | false | false | Pflichtangaben werden geprüft |
