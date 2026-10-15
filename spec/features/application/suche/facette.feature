# language: de
@spec-only @application
Funktionalität: Variantenkonsistente disjunktive Suche
  Als Käufer erweitere ich Markenauswahl ohne falsche Variantenkombinationen.

  Hintergrund:
    Angenommen ein synthetischer Smartphonekatalog ist projiziert
      | ProductID | Brand | SKU | Color | StorageGB | Price | Available |
      | p-a | Alpha | a-red | red | 128 | 10000 | true |
      | p-a | Alpha | a-blue | blue | 256 | 15000 | true |
      | p-b | Beta | b-red | red | 256 | 12000 | true |
      | p-c | Gamma | c-blue | blue | 128 | 9000 | true |

  Szenariogrundriss: Treffer und exakte Markenzählungen
    Wenn die öffentliche Suche ausgeführt wird
      | Brand | Color | StorageGB | MaxPrice |
      | <marken> | <farbe> | <speicher> | <preis> |
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
      | Brand | Color | StorageGB |
      | Alpha | red | 256 |
    Dann lauten die kontextuellen Attributzählungen
      | Facet | Value | ProductCount |
      | color | blue | 1 |
      | storageCapacity | 128 | 1 |

  Szenario: Zwei passende Varianten ergeben nur eine Modellkarte
    Wenn die öffentliche Suche ausgeführt wird
      | Brand | Sort |
      | Alpha | priceAsc |
    Dann lautet die Modellkarte
      | ProductID | ChosenSKU | VariantCount | CardCount |
      | p-a | a-red | 2 | 1 |
