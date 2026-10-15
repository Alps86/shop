# language: de
@application @opensearch
Funktionalität: Variantenkonsistente disjunktive Suche
  Als Käufer erweitere ich Markenauswahl ohne falsche Variantenkombinationen.

  Hintergrund:
    Angenommen eine synthetische Suchquelle mit vollständigen Lieferantendaten besteht
      | Index | Locale | Currency | Market | At | CategoryID | FamilyID | Manufacturer | Street | PostalCode | City | Country | Email |
      | shop-runtime-facets-scored | de-DE | EUR | DE | 2026-10-15T00:00:00Z | phones | smartphones | Testlieferant | Testweg 1 | 12345 | Teststadt | DE | demo@example.invalid |
    Und die kuratierten Suchdefinitionen stehen bereit
      | Code | Type | Options | Minimum | Maximum | Top |
      | color | select | red,blue | absent | absent | true |
      | storageCapacity | number | [] | 1 | 1024 | true |
    Und der synthetische Produktkatalog wird tatsächlich projiziert
      | ProductID | Name | Slug | Brand | SKU | Color | StorageGB | Price | Stock | Published |
      | p-a | Alpha Testtelefon | alpha | Alpha | a-red | red | 128 | 10000 | 1 | true |
      | p-a | Alpha Testtelefon | alpha | Alpha | a-blue | blue | 256 | 15000 | 1 | true |
      | p-b | Beta Testtelefon | beta | Beta | b-red | red | 256 | 12000 | 1 | true |
      | p-c | Gamma Testtelefon | gamma | Gamma | c-blue | blue | 128 | 9000 | 1 | true |

  Szenariogrundriss: Treffer und exakte Markenzählungen
    Wenn die öffentliche Suche ausgeführt wird
      | Brand | Color | StorageGB | MaxPrice | Sort | Page |
      | <marken> | <farbe> | <speicher> | <preis> | relevance | 1 |
    Dann lautet das Suchergebnis
      | ProductIDs | TotalProducts | BrandAlpha | BrandBeta | BrandGamma | SelectedZeroChip |
      | <treffer> | <anzahl> | <alpha> | <beta> | <gamma> | <nullchip> |
    Beispiele:
      | marken | farbe | speicher | preis | treffer | anzahl | alpha | beta | gamma | nullchip |
      | [] | any | any | any | p-a,p-b,p-c | 3 | 1 | 1 | 1 | none |
      | Alpha | any | any | any | p-a | 1 | 1 | 1 | 1 | none |
      | Alpha,Beta | any | any | any | p-a,p-b | 2 | 1 | 1 | 1 | none |
      | [] | red | 256 | any | p-b | 1 | 0 | 1 | 0 | none |
      | Alpha | red | 256 | any | [] | 0 | 0 | 1 | 0 | Alpha |
      | [] | any | 256 | 13000 | p-b | 1 | 0 | 1 | 0 | none |

  Szenario: Facetten ignorieren jeweils nur ihren eigenen Filter
    Wenn die öffentliche Suche ausgeführt wird
      | Brand | Color | StorageGB | Sort | Page |
      | Alpha | red | 256 | relevance | 1 |
    Dann lauten die kontextuellen Attributzählungen
      | Facet | Value | ProductCount |
      | color | blue | 1 |
      | storageCapacity | 128 | 1 |

  Szenario: Zwei passende Varianten ergeben nur eine Modellkarte
    Wenn die öffentliche Suche ausgeführt wird
      | Brand | Sort | Page |
      | Alpha | priceAsc | 1 |
    Dann lautet die Modellkarte
      | ProductID | ChosenSKU | VariantCount | CardCount |
      | p-a | a-red | 2 | 1 |
