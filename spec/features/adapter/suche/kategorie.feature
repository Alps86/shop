# language: de
@adapter @opensearch
Funktionalität: Kategorien in der echten Suchprojektion
  Oberkategorien umfassen ihre Produktunterkategorien und liefern veröffentlichte Vorschläge.

  Szenario: Oberkategorie und Kategorienvorschlag stammen aus demselben veröffentlichten Produkt
    Angenommen eine synthetische Suchquelle mit vollständigen Lieferantendaten besteht
      | Index | Locale | Currency | Market | At | CategoryID | FamilyID | Manufacturer | Street | PostalCode | City | Country | Email |
      | shop-runtime-category | de-DE | EUR | DE | 2026-10-15T00:00:00Z | phones | smartphones | Testlieferant | Testweg 1 | 12345 | Teststadt | DE | demo@example.invalid |
    Und der veröffentlichte Kategorienbaum steht bereit
      | ID | ParentID | Name |
      | electronics | absent | Elektronik |
      | phones | electronics | Telefone |
    Und der synthetische Produktkatalog wird tatsächlich projiziert
      | ProductID | Name | Slug | Brand | SKU | Color | StorageGB | Price | Stock | Published |
      | p-a | Alpha | alpha | Alpha | a-red | red | 128 | 10000 | 1 | true |
    Wenn die Suche sortiert wird
      | Text | Sort | Page | CategoryID |
      | | relevance | 1 | electronics |
    Dann lauten die tatsächlichen Modellkarten
      | ProductIDs | TotalProducts |
      | p-a | 1 |
    Wenn Vorschläge angefragt werden
      | Query | Sort | Page |
      | Tel | relevance | 1 |
    Dann lauten die Vorschläge
      | Names | Correction | QueryReplaced | MaximumCount |
      | Telefone | false | false | 5 |
