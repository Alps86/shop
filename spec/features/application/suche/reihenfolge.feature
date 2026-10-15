# language: de
@application
Funktionalität: Sortierung Vorschläge und Ausfallgrenzen
  Als Käufer behalte ich nachvollziehbare Suchreihenfolge und verfügbare Kaufpfade.

  @opensearch
  Szenariogrundriss: Explizite Sortierung gewinnt gegen Boosts
    Angenommen eine synthetische Suchquelle mit vollständigen Lieferantendaten besteht
      | Index | Locale | Currency | Market | At | CategoryID | FamilyID | Manufacturer | Street | PostalCode | City | Country | Email |
      | shop-runtime-ranking-scored | de-DE | EUR | DE | 2026-10-15T00:00:00Z | phones | smartphones | Testlieferant | Testweg 1 | 12345 | Teststadt | DE | demo@example.invalid |
    Und vergleichbare synthetische Suchkandidaten liegen vor
      | ProductID | Name | Slug | Brand | SKU | Color | StorageGB | Price | Stock | Published | CategoryID | PurchaseCost | InterestWeight |
      | p-a | Alpha | alpha | Alpha | a-red | red | 128 | 10000 | 1 | true | phones | 10000 | 0 |
      | p-b | Beta | beta | Beta | b-red | red | 128 | 12000 | 1 | true | home | 0 | 1 |
    Wenn die Suche sortiert wird
      | Text | Sort | Page |
      | | <sortierung> | 1 |
    Dann lautet die öffentliche Ergebnisreihenfolge
      | ProductIDs | ExposedMargin | ExposedInterestWeight |
      | <reihenfolge> | false | false |
    Beispiele:
      | sortierung | reihenfolge |
      | relevance | p-b,p-a |
      | priceAsc | p-a,p-b |
      | priceDesc | p-b,p-a |
      | nameAsc | p-a,p-b |

  @opensearch
  Szenariogrundriss: Präfix und begrenzte Tippfehler
    Angenommen eine synthetische Suchquelle mit vollständigen Lieferantendaten besteht
      | Index | Locale | Currency | Market | At | CategoryID | FamilyID | Manufacturer | Street | PostalCode | City | Country | Email |
      | shop-runtime-suggestions-scored | de-DE | EUR | DE | 2026-10-15T00:00:00Z | phones | smartphones | Testlieferant | Testweg 1 | 12345 | Teststadt | DE | demo@example.invalid |
    Und veröffentlichte Suchbegriffe sind vorhanden
      | ProductID | Name | Slug | Brand | SKU | Color | StorageGB | Price | Stock | Published |
      | p-phone | Smartphone | smartphone | Alpha | phone-red | red | 128 | 10000 | 1 | true |
      | p-watch | Smartwatch | smartwatch | Alpha | watch-red | red | 128 | 10000 | 1 | true |
    Wenn Vorschläge angefragt werden
      | Query | Sort | Page |
      | <eingabe> | relevance | 1 |
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

  @app
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

  @commerce-runtime
  Szenario: Suchausfall sperrt den maßgeblichen Kaufpfad nicht
    Angenommen nur OpenSearch ist nicht erreichbar
      | PostgreSQL | Temporal | OpenSearch | SearchEndpoint | Fixture | Namespace | Locale | Currency | Market |
      | ready | ready | down | http://127.0.0.1:1 | spec/fixtures/runtime/katalog.json | default | de-DE | EUR | DE |
    Wenn die öffentlichen Ports aufgerufen werden
      | Operation | ProductID | SKU | StartCommandID | ChangeCommandID | AddressCommandID | PaymentCommandID | ConfirmCommandID | PlaceCommandID | Recipient | Street | PostalCode | City | Country | PaymentMethod | Quantity |
      | SearchPort.Search | | | | | | | | | | | | | | | |
      | CatalogReader.Product | p-phone | PHONE | | | | | | | | | | | | | |
      | CheckoutService.Place | p-phone | PHONE | start | add | address | payment | confirm | place | Demo Gast | Testweg 1 | 12345 | Teststadt | DE | invoice | 1 |
    Dann lauten die Portantworten
      | Operation | Code | SearchTotal | Cards | ProductID | SKU | OrderCount | Quantity | Total | Pending | ProcessState |
      | SearchPort.Search | SearchUnavailable | 0 | 0 | | | | | | | |
      | CatalogReader.Product | OK | | | p-phone | | | | | | |
      | CheckoutService.Place | OK | | | | PHONE | 1 | 1 | 10490 | false | running |

  @app
  Szenariogrundriss: Typisierte Auswahl und Bereichsgrenzen werden vor dem Suchadapter geprüft
    Wenn ein strukturierter Suchrequest angefragt wird
      | Type | Attribute | Options | Minimum | Maximum | CurrencyMinimum | CurrencyMaximum | Fault |
      | <typ> | <attribut> | <optionen> | <minimum> | <maximum> | <wl> | <wr> | <fehler> |
    Dann lautet das Suchvalidierungsergebnis
      | Code | PageSize |
      | <code> | 12 |
    Beispiele:
      | typ | attribut | optionen | minimum | maximum | wl | wr | fehler | code |
      | options | color | red,blue | absent | absent | EUR | EUR | none | OK |
      | options | color | red,red | absent | absent | EUR | EUR | none | InvalidFilter |
      | options | color | green | absent | absent | EUR | EUR | none | InvalidFilter |
      | options | color | [] | absent | absent | EUR | EUR | none | InvalidFilter |
      | range | storageCapacity | [] | 128 | 256 | EUR | EUR | none | OK |
      | range | storageCapacity | [] | 256 | 128 | EUR | EUR | none | InvalidFilter |
      | range | storageCapacity | [] | absent | 256 | EUR | EUR | none | OK |
      | range | storageCapacity | [] | 128 | absent | EUR | EUR | none | OK |
      | range | storageCapacity | [] | absent | absent | EUR | EUR | none | OK |
      | range | storageCapacity | [] | invalid | 256 | EUR | EUR | none | InvalidFilter |
      | range | storageCapacity | [] | 128 | invalid | EUR | EUR | none | InvalidFilter |
      | range | storageCapacity | [] | dirty | 256 | EUR | EUR | none | InvalidFilter |
      | range | color | [] | 1 | 2 | EUR | EUR | none | InvalidFilter |
      | range | notCurated | [] | 1 | 2 | EUR | EUR | none | InvalidFilter |
      | price | none | [] | 100 | 200 | EUR | EUR | none | OK |
      | price | none | [] | 200 | 100 | EUR | EUR | none | InvalidFilter |
      | price | none | [] | 100 | 200 | EUR | USD | none | InvalidFilter |
      | price | none | [] | 100 | 200 | USD | USD | none | InvalidFilter |
      | price | none | [] | absent | 200 | EUR | EUR | none | OK |
      | price | none | [] | 100 | absent | EUR | EUR | none | OK |
      | price | none | [] | -1 | 200 | EUR | EUR | none | InvalidFilter |
      | price | none | [] | 100 | 200 | ZZZ | EUR | none | InvalidFilter |
      | price | none | [] | dirty | 200 | EUR | EUR | none | InvalidFilter |
      | dirtyPrice | none | [] | absent | absent | EUR | EUR | none | InvalidFilter |
      | plain | none | [] | absent | absent | EUR | EUR | search | SearchUnavailable |

  @app
  Szenariogrundriss: Der Suchkontext bleibt vom gültigen ISO-Wert getrennt
    Wenn die Suche mit einem Verkaufskontext angefragt wird
      | Locale | Currency | Market |
      | <sprache> | <waehrung> | <markt> |
    Dann lautet das Suchvalidierungsergebnis
      | Code | PageSize |
      | <code> | 12 |
    Beispiele:
      | sprache | waehrung | markt | code |
      | de-DE | EUR | DE | OK |
      | fr-FR | EUR | DE | UnsupportedContext |
      | de-DE | USD | DE | UnsupportedContext |
      | de-DE | EUR | FR | UnsupportedContext |
