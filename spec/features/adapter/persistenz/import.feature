# language: de
@runtime @postgres
Funktionalität: Zweistufige atomare Datensatzveröffentlichung
  Fehlerhafte Pakete werden vollständig abgelehnt; dieselbe Version bleibt inhaltlich eindeutig.

  Hintergrund:
    Angenommen ein kohärenter Demodatensatz in echter PostgreSQL vorbereitet ist
      | DatasetVersion | ProductID | SKU | Name | Price | StockQuantity |
      | demo-v1 | p-phone | PHONE | Synthetisches Testmodell | 10000 | 2 |

  Szenariogrundriss: Fehler vor der Publikation hinterlassen keine Stammdaten oder Bestände
    Angenommen ein Importfeld vor der Veröffentlichung geändert wird
      | Field | Value |
      | <feld> | <wert> |
    Wenn der Datensatz veröffentlicht wird
    Dann lautet die atomare Importbeobachtung
      | Code | Products | Inventory |
      | <code> | 0 | 0 |
    Beispiele:
      | feld | wert | code |
      | DatasetVersion | invalid version | InvalidInput |
      | SourceURL | ://invalid | InvalidInput |
      | SourceURL | http://example.invalid | InvalidInput |
      | SourceURL | https:example | InvalidInput |
      | LicenseNote | | InvalidInput |
      | RetrievedAt | missing | InvalidInput |
      | FamilyVersion | 0 | InvalidAttribute |
      | FamilyReference | f-missing | InvalidReference |
      | AttributeLevel | variant | InvalidAttribute |
      | CategoryParent | missing | InvalidReference |
      | Inventory | -1 | InvalidInput |
      | Inventory | missing | InvalidReference |
      | Inventory | duplicate | DuplicateSKU |
      | Inventory | orphan | InvalidReference |
      | Duplicate | family | Conflict |
      | Duplicate | category | Conflict |
      | Duplicate | product | Conflict |
      | SKU | PHONE | DuplicateSKU |
      | DocumentURL | javascript:alert(1) | InvalidInput |
      | DocumentURL | https:missing | InvalidInput |
      | DocumentURL | https://user:secret@example.invalid/manual | InvalidInput |
      | DocumentURL | ://invalid | InvalidInput |
      | Discount | 0 | InvalidMoney |

  Szenario: Identische Version ist ein Noop ohne Bestandsüberschreibung
    Wenn der Datensatz veröffentlicht wird
    Und der Datensatz veröffentlicht wird
    Dann lautet die atomare Importbeobachtung
      | Code | Result | Products | Inventory |
      | OK | Noop | 1 | 1 |

  Szenario: Andere Inhalte unter derselben Version sind ein Konflikt
    Wenn der Datensatz veröffentlicht wird
    Und ein Importfeld vor der Veröffentlichung geändert wird
      | Field | Value |
      | ProductName | Geänderter Name |
    Und der Datensatz veröffentlicht wird
    Dann lautet die atomare Importbeobachtung
      | Code | Products | Inventory |
      | Conflict | 1 | 1 |

  Szenariogrundriss: Jeder fehlgeschlagene Publikationsschritt hinterlässt das ganze Paket unpubliziert
    Angenommen ein kompatibles synthetisches Zubehör mit Rabatt und Paarstatistik ergänzt wird
      | ProductID | SKU | Name | Price | Stock | Discount | BasketCount | WindowStart | WindowEnd | StatisticsVersion |
      | p-case | CASE | Synthetische Hülle | 2000 | 3 | 500 | 2 | 2026-01-01T00:00:00Z | 2026-10-15T00:00:00Z | synthetic-v1 |
    Und ein echter SQL-Schreibfehler gezielt vorbereitet ist
      | Table | Operation | MatchField | MatchValue | SQLState |
      | <tabelle> | INSERT | <feld> | <wert> | P0001 |
    Wenn der Datensatz veröffentlicht wird
    Dann lautet die atomare Importbeobachtung
      | Code | Products | Inventory |
      | DependencyUnavailable | 0 | 0 |
    Beispiele:
      | tabelle | feld | wert |
      | records | kind | family |
      | records | kind | category |
      | records | kind | discount |
      | records | kind | ruleset |
      | records | kind | active-ruleset |
      | records | kind | statistics |
      | records | kind | inventory |
      | projection_versions | | |
      | sku_registry | | |
      | records | kind | product |
      | outbox | | |
      | datasets | | |

  Szenariogrundriss: Empfehlungen stammen nur aus publizierten kompatiblen Zubehörreferenzen
    Angenommen ein kompatibles synthetisches Zubehör mit Rabatt und Paarstatistik ergänzt wird
      | ProductID | SKU | Name | Price | Stock | Discount | BasketCount | WindowStart | WindowEnd | StatisticsVersion |
      | p-case | CASE | Synthetische Hülle | 2000 | 3 | 500 | 2 | 2026-01-01T00:00:00Z | 2026-10-15T00:00:00Z | synthetic-v1 |
    Wenn der Datensatz veröffentlicht wird
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | ID | SKU |
      | recommendations | <product> | <sku> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | Count | Degraded | Label |
      | OK | <anzahl> | <degradiert> | <label> |
    Beispiele:
      | product | sku | anzahl | degradiert | label |
      | p-phone | PHONE | 1 | false | Häufig zusammen gekauft |
      | p-phone | UNKNOWN | 0 | false | |
      | p-case | CASE | 0 | false | |
      | p-missing | UNKNOWN | 0 | true | |

  Szenariogrundriss: Neue Datensätze bewahren den aktiven Stand bei abgelehnten Paketen
    Wenn der Datensatz veröffentlicht wird
    Und ein Importfeld vor der Veröffentlichung geändert wird
      | Field | Value |
      | DatasetVersion | demo-v2 |
    Und ein Importfeld vor der Veröffentlichung geändert wird
      | Field | Value |
      | <feld> | <wert> |
    Und der Datensatz veröffentlicht wird
    Dann lautet die atomare Importbeobachtung
      | Code | Products | Inventory |
      | <code> | 1 | 1 |
    Und lautet der weiterhin maßgebliche Datensatz
      | ActiveDatasetVersion | Products | Price | Stock |
      | <aktiv> | 1 | 10000 | 2 |
    Beispiele:
      | feld | wert | code | aktiv |
      | ProductName | Neuer Name | OK | demo-v2 |
      | SKU | PHONE | DuplicateSKU | demo-v1 |
      | FamilyReference | f-missing | InvalidReference | demo-v1 |
      | AttributeLevel | variant | InvalidAttribute | demo-v1 |
      | DocumentURL | javascript:alert(1) | InvalidInput | demo-v1 |

  Szenario: Eine neue Publikation überschreibt keine zwischenzeitlichen Bestands- und Preisänderungen
    Wenn der Datensatz veröffentlicht wird
    Und der Bestand atomar geändert wird
      | SKU | CommandID | Revision | SetPrice | Price | SetStock | Stock |
      | PHONE | inventory-a | 1 | true | 12000 | true | 5 |
    Und ein Importfeld vor der Veröffentlichung geändert wird
      | Field | Value |
      | DatasetVersion | demo-v2 |
    Und der Datensatz veröffentlicht wird
    Dann lautet der weiterhin maßgebliche Datensatz
      | ActiveDatasetVersion | Products | Price | Stock |
      | demo-v2 | 1 | 12000 | 5 |

  Szenario: Unvollständige öffentliche Referenzdaten erhalten keine erfundenen Pflichtangaben
    Angenommen unverifizierte öffentliche Referenzdaten ohne Pflichtangabenergänzung übernommen werden
      | SourceURL | RetrievedAt | LicenseNote | VerificationStatus |
      | https://www.ep.de/p/1045137/ecodecalk-entkalker-500ml | 2026-10-15T00:00:00Z | source-only-rights-unverified | pending |
    Wenn der Datensatz veröffentlicht wird
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | ID | SKU |
      | product | p-phone | PHONE |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | Available | Warnings | VerificationStatus |
      | OK | 0 | 0 | pending |
    Und bleibt die importierte Herkunft dauerhaft nachvollziehbar
      | SourceURL | RetrievedAt | LicenseNote |
      | https://www.ep.de/p/1045137/ecodecalk-entkalker-500ml | 2026-10-15T00:00:00Z | source-only-rights-unverified |

  Szenariogrundriss: Regeln und synthetische Statistik müssen intern referenzierbar und zeitlich gültig sein
    Angenommen ein kompatibles synthetisches Zubehör mit Rabatt und Paarstatistik ergänzt wird
      | ProductID | SKU | Name | Price | Stock | Discount | BasketCount | WindowStart | WindowEnd | StatisticsVersion |
      | p-case | CASE | Synthetische Hülle | 2000 | 3 | 500 | 2 | 2026-01-01T00:00:00Z | 2026-10-15T00:00:00Z | synthetic-v1 |
    Und eine Importregel oder Statistikreferenz geändert wird
      | Field | Value |
      | <feld> | <wert> |
    Wenn der Datensatz veröffentlicht wird
    Dann lautet die atomare Importbeobachtung
      | Code | Products | Inventory |
      | <code> | 0 | 0 |
    Beispiele:
      | feld | wert | code |
      | Source | p-missing | InvalidReference |
      | Target | p-missing | InvalidReference |
      | MainRule | p-missing | InvalidReference |
      | AccessoryRule | p-missing | InvalidReference |
      | Count | -1 | InvalidInput |
      | Window | empty | InvalidInput |
      | Version | invalid version | InvalidInput |
