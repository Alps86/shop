# language: de
@adapter @opensearch
Funktionalität: Echter Suchdienst
  Szenario: Ein noch nicht initialisierter Index wird als technischer Fehler sichtbar
    Angenommen eine echte OpenSearchverbindung besteht
      | Index | Locale | Currency | Market |
      | shop-runtime-missing-index | de-DE | EUR | DE |
    Wenn die echte öffentliche Suche abgefragt wird
      | Text | Page | Sort |
      | Telefon | 1 | relevance |
    Dann liefert die Suchverbindung
      | Code | TotalProducts |
      | SearchUnavailable | 0 |

  Szenario: Eine leere echte Suchprojektion liefert ein gültiges leeres Ergebnis
    Angenommen eine echte OpenSearchverbindung besteht
      | Index | Locale | Currency | Market |
      | shop-runtime-empty-mapped-scored | de-DE | EUR | DE |
    Und der Suchindex als leere echte Fixture angelegt wird
    Wenn die echte öffentliche Suche abgefragt wird
      | Text | Page | Sort |
      | Telefon | 1 | relevance |
    Dann liefert die Suchverbindung
      | Code | TotalProducts |
      | OK | 0 |

  Szenario: Eine vollständige Produktprojektion liefert genau eine Karte je Modell
    Angenommen eine synthetische Suchquelle mit vollständigen Lieferantendaten besteht
      | Index | Locale | Currency | Market | At | CategoryID | FamilyID | Manufacturer | Street | PostalCode | City | Country | Email |
      | shop-runtime-projection-scored | de-DE | EUR | DE | 2026-10-15T00:00:00Z | phones | smartphones | Testlieferant | Testweg 1 | 12345 | Teststadt | DE | demo@example.invalid |
    Und der synthetische Produktkatalog wird tatsächlich projiziert
      | ProductID | Name | Slug | Brand | SKU | Color | StorageGB | Price | Stock | Published |
      | p-a | Alpha Testtelefon | alpha | Alpha | a-red | red | 128 | 10000 | 1 | true |
      | p-a | Alpha Testtelefon | alpha | Alpha | a-blue | blue | 256 | 15000 | 1 | true |
      | p-b | Beta Testtelefon | beta | Beta | b-red | red | 256 | 12000 | 1 | true |
      | p-c | Gamma Testtelefon | gamma | Gamma | c-blue | blue | 128 | 9000 | 1 | true |
    Wenn die echte öffentliche Suche abgefragt wird
      | Text | Page | Sort |
      | | 1 | relevance |
    Dann lauten die tatsächlichen Modellkarten
      | ProductIDs | TotalProducts |
      | p-a,p-b,p-c | 3 |

  Szenario: Ein neuer physischer Index übernimmt den öffentlichen Suchalias vollständig
    Angenommen eine synthetische Suchquelle mit vollständigen Lieferantendaten besteht
      | Index | Locale | Currency | Market | At | CategoryID | FamilyID | Manufacturer | Street | PostalCode | City | Country | Email |
      | shop-runtime-rebuild | de-DE | EUR | DE | 2026-10-15T00:00:00Z | phones | smartphones | Testlieferant | Testweg 1 | 12345 | Teststadt | DE | demo@example.invalid |
    Und der synthetische Produktkatalog wird tatsächlich projiziert
      | ProductID | Name | Slug | Brand | SKU | Color | StorageGB | Price | Stock | Published |
      | p-a | Alpha Testtelefon | alpha | Alpha | a-red | red | 128 | 10000 | 1 | true |
    Wenn ein vollständiger Quellsnapshot in einen neuen physischen Suchindex aufgebaut wird
      | IndexSuffix | Cursor |
      | -v2 | fixture-snapshot |
    Dann liefert die Suchverbindung
      | Code |
      | OK |
    Wenn die echte öffentliche Suche abgefragt wird
      | Text | Page | Sort |
      | Alpha | 1 | relevance |
    Dann lauten die tatsächlichen Modellkarten
      | ProductIDs | TotalProducts |
      | p-a | 1 |
    Angenommen nach dem Snapshot eine laufende Projektion das Produkt löscht
      | ProductID | Revision |
      | p-a | 2 |
    Wenn ein vollständiger Quellsnapshot in einen neuen physischen Suchindex aufgebaut wird
      | IndexSuffix | Cursor |
      | -v3 | fixture-snapshot-next |
    Dann liefert die Suchverbindung
      | Code |
      | OK |
    Wenn die echte öffentliche Suche abgefragt wird
      | Text | Page | Sort |
      | Alpha | 1 | relevance |
    Dann lautet das Suchergebnis
      | TotalProducts | ProductIDs |
      | 0 | [] |

  Szenariogrundriss: Ein gescheiterter Neuaufbau gibt Ressourcen frei und lässt den vorhandenen Alias nutzbar
    Angenommen eine echte OpenSearchverbindung besteht
      | Index | Locale | Currency | Market |
      | shop-runtime-rebuild-fault | de-DE | EUR | DE |
    Und der Suchindex als leere echte Fixture angelegt wird
    Und die Rebuildquelle einen expliziten Fehler liefert
      | Phase | Code |
      | <Phase> | SourceUnavailable |
    Wenn ein vollständiger Quellsnapshot in einen neuen physischen Suchindex aufgebaut wird
      | IndexSuffix | Cursor | Source |
      | <Suffix> | controlled-source | <Source> |
    Dann der Rebuildabschluss beobachtet wird
      | Code | AbortCount |
      | <Code> | <Aborts> |
    Wenn die echte öffentliche Suche abgefragt wird
      | Text | Page | Sort |
      | | 1 | relevance |
    Dann liefert die Suchverbindung
      | Code | TotalProducts |
      | OK | 0 |

    Beispiele:
      | Phase | Suffix | Source | Code | Aborts |
      | snapshot | -v2 | present | SourceUnavailable | 0 |
      | catchup | -v2 | present | SourceUnavailable | 1 |
      | abort | -v2 | present | SourceUnavailable | 1 |
      | replay | -v2 | present | InvalidInput | 1 |
      | concurrent | -v2 | present | Conflict | 1 |
      | none | invalid | present | InvalidInput | 0 |
      | none | -v2 | absent | DependencyUnavailable | 0 |
