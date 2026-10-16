# language: de
@adapter @opensearch
Funktionalität: Ehrliche Suchdienstfehler
  Der Adapter meldet Netzwerk- und Protokollfehler, ohne Produktkarten zu erfinden.

  Grundlage:
    Angenommen eine synthetische Suchquelle mit vollständigen Lieferantendaten besteht
      | Index | Locale | Currency | Market | At | CategoryID | FamilyID | Manufacturer | Street | PostalCode | City | Country | Email |
      | shop-runtime-failure | de-DE | EUR | DE | 2026-10-15T00:00:00Z | phones | smartphones | Testlieferant | Testweg 1 | 12345 | Teststadt | DE | demo@example.invalid |
    Und der synthetische Produktkatalog wird tatsächlich projiziert
      | ProductID | Name | Slug | Brand | SKU | Color | StorageGB | Price | Stock | Published |
      | p-a | Alpha Testtelefon | alpha | Alpha | a-red | red | 128 | 10000 | 1 | true |

  Szenariogrundriss: Eine gestörte Suchantwort liefert keine erfundene Karte
    Angenommen die Suchdienstverbindung eine explizite Protokollstörung liefert
      | Connection | Status | Body | Clock |
      | <Verbindung> | <Status> | <Antwort> | controlled |
    Wenn die echte öffentliche Suche abgefragt wird
      | Text | Page | Sort |
      | A | 1 | relevance |
    Dann liefert die Suchverbindung
      | Code | TotalProducts |
      | <Fehler> | 0 |

    Beispiele:
      | Verbindung | Status | Antwort | Fehler |
      | closed | 503 | {} | SearchUnavailable |
      | invalid | 503 | {} | DependencyUnavailable |
      | open | 503 | {} | SearchUnavailable |
      | open | 200 | invalid-json | SearchUnavailable |
      | open | 200 | {"hits":{"hits":[{"fields":{"product_id":["p-a"]}}]}} | SearchUnavailable |
      | open | 200 | {"hits":{"hits":[{"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWIiLCJTS1UiOiJhLXJlZCJ9"]}}]}}},"fields":{"product_id":["p-a"]}}]}} | SearchUnavailable |
      | open | 200 | {"hits":{"hits":[{"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWEiLCJTS1UiOiJhLWJsdWUifQ=="]}}]}}},"fields":{"product_id":["p-a"]}}]}} | SearchUnavailable |
      | open | 200 | {"aggregations":{"facet_brand":false}} | SearchUnavailable |
      | open | 200 | {"aggregations":{"facet_price":{"variants":{"matching":{"values":{"buckets":[{"key":{}}]}}}}}} | SearchUnavailable |

  Szenariogrundriss: Verwaltungsoperationen melden den tatsächlichen technischen Zustand
    Angenommen die Suchdienstverbindung eine explizite Protokollstörung liefert
      | Connection | Status | Body | Clock |
      | <Verbindung> | <Status> | <Antwort> | system |
    Wenn der Suchadapter die Verwaltungsoperation ausführt
      | Operation | ProductID |
      | <Operation> | p-a |
    Dann liefert die Suchverbindung
      | Code |
      | <Fehler> |

    Beispiele:
      | Operation | Verbindung | Status | Antwort | Fehler |
      | initialize | open | 200 | {} | OK |
      | initialize | open | 503 | {} | DependencyUnavailable |
      | initialize | closed | 503 | {} | SearchUnavailable |
      | initialize | open | 404 | {} | DependencyUnavailable |
      | create | open | 400 | {"error":{"type":"resource_already_exists_exception"}} | OK |
      | create | open | 503 | {} | DependencyUnavailable |
      | create | closed | 503 | {} | SearchUnavailable |
      | apply | closed | 503 | {} | SearchUnavailable |
      | apply | open | 503 | {} | DependencyUnavailable |

  Szenariogrundriss: Inkonsistente Projektionshüllen und Bestände werden abgelehnt
    Wenn eine externe Produktprojektion mit expliziten Metadaten übernommen wird
      | EventProductID | PayloadProductID | RootID | Revision | Present | Stock | Currency |
      | <Ereignis> | p-a | <Produkt> | <Version> | <Vorhanden> | <Bestand> | <Waehrung> |
    Dann liefert die Suchverbindung
      | Code |
      | <Fehler> |

    Beispiele:
      | Ereignis | Produkt | Version | Vorhanden | Bestand | Waehrung | Fehler |
      | ! | p-a | 2 | true | 1 | EUR | InvalidInput |
      | p-a | p-a | 0 | true | 1 | EUR | InvalidInput |
      | p-a | p-a | 2 | false | 1 | EUR | InvalidInput |
      | p-a | p-b | 2 | true | 1 | EUR | InvalidInput |
      | p-a | p-a | 2 | true | -1 | EUR | InvalidInput |
      | p-a | p-a | 2 | true | 1 | USD | UnsupportedContext |
      | p-a | p-a | 2 | true | 1 | EUR | OK |

  Szenariogrundriss: Externe Zusatzattribute behalten ihre Fehlersemantik
    Und die kuratierten Suchdefinitionen stehen bereit
      | Code | Type | Unit | Options | Minimum | Maximum | Top |
      | storageCapacity | measurement | GB | [] | absent | absent | true |
    Wenn eine Produktprojektion mit expliziten Zusatzattributen übernommen wird
      | ProductID | Measurement | Unit | Number | InvalidYear |
      | p-a | <Measurement> | <Unit> | <Number> | <Year> |
    Dann liefert die Suchverbindung
      | Code |
      | <Code> |

    Beispiele:
      | Measurement | Unit | Number | Year | Code |
      | 1 | kg | | | InvalidUnit |
      | | | invalid | | InvalidInput |
      | | | | 10000 | InvalidInput |

  Szenariogrundriss: Normierte Maße und Bildvorrang bleiben auf der öffentlichen Karte sichtbar
    Und die kuratierten Suchdefinitionen stehen bereit
      | Code | Type | Unit | Options | Minimum | Maximum | Top |
      | storageCapacity | measurement | <Target> | [] | absent | absent | true |
    Wenn eine Produktprojektion mit expliziten Zusatzattributen übernommen wird
      | ProductID | Measurement | Unit | ProductImage | VariantImage | VerificationStatus |
      | p-a | 0.128 | TB | https://example.invalid/product.png | <VariantImage> | <Status> |
    Dann liefert die Suchverbindung
      | Code |
      | OK |
    Wenn die öffentliche Suche ausgeführt wird
      | Page | Sort | Brand | Color | StorageGB |
      | 1 | relevance | any | any | <Bound> |
    Dann die ergänzte Modellkarte entspricht
      | Image | AvailableQuantity | TopAttributeCount |
      | <Image> | <Available> | 1 |

    Beispiele:
      | Target | Bound | VariantImage | Image | Status | Available |
      | GB | 128 | https://example.invalid/variant.png | https://example.invalid/variant.png | verified | 1 |
      | TB | 0.128 | | https://example.invalid/product.png | pending | 0 |

  Szenariogrundriss: Ungültige und offene Zahlenintervalle behalten den Suchvertrag
    Und die kuratierten Suchdefinitionen stehen bereit
      | Code | Type | Options | Minimum | Maximum | Top |
      | storageCapacity | number | [] | absent | absent | true |
    Wenn die öffentliche Suche ausgeführt wird
      | Page | Sort | RangeCode | RangeMin | RangeMax |
      | 1 | relevance | storageCapacity | <Minimum> | <Maximum> |
    Dann liefert die Suchverbindung
      | Code | TotalProducts |
      | <Code> | <Count> |

    Beispiele:
      | Minimum | Maximum | Code | Count |
      | invalid | absent | InvalidFilter | 0 |
      | absent | invalid | InvalidFilter | 0 |
      | absent | 128 | OK | 1 |
      | 128 | absent | OK | 1 |
      | 256 | 128 | InvalidFilter | 0 |

  Szenario: Wiederholte Vorschläge verschiedener Gruppen werden stabil dedupliziert
    Angenommen die Suchdienstverbindung eine explizite Protokollstörung liefert
      | Connection | Status | Body | Clock |
      | open | 200 | {"responses":[{"status":200,"hits":{"total":{"value":0},"hits":[]}},{"status":200,"suggest":{"catalog":[{"options":[{"text":"Alpha Z"},{"text":"Beta"},{"text":"Zeta"},{"text":"Alpha A"},{"text":"Alpha A"}]},{"options":[{"text":"Gamma"},{"text":"Alpha B"},{"text":"Alpha C"}]}]}}]} | controlled |
    Wenn Vorschläge angefragt werden
      | Query | Page | Sort |
      | Alpha | 1 | relevance |
    Dann lauten die Vorschläge
      | Names | MaximumCount | Correction | QueryReplaced |
      | Alpha A,Alpha B,Alpha C,Alpha Z,Beta | 5 | true | false |

  Szenariogrundriss: Rebuild transportiert Fehler vor dem Aliaswechsel unverändert
    Angenommen die Rebuildquelle einen expliziten Fehler liefert
      | Phase | Code |
      | <Phase> | SourceUnavailable |
    Und die Suchdienstverbindung eine explizite Protokollstörung liefert
      | Connection | Status | Body | Clock | FailPath |
      | <Connection> | <Status> | {} | controlled | <Path> |
    Wenn ein vollständiger Quellsnapshot in einen neuen physischen Suchindex aufgebaut wird
      | IndexSuffix | Cursor |
      | -candidate | fault-source |
    Dann der Rebuildabschluss beobachtet wird
      | Code | AbortCount | WriterCode |
      | <Code> | <Aborts> | <WriterCode> |

    Beispiele:
      | Phase | Connection | Status | Path | Code | Aborts | WriterCode |
      | none | closed | 200 | | SearchUnavailable | 0 | OK |
      | none | invalid | 200 | | DependencyUnavailable | 0 | OK |
      | none | open | 503 | | Conflict | 0 | OK |
      | none | open | 503 | /_doc/ | DependencyUnavailable | 1 | OK |
      | none | open | 503 | /_aliases | DependencyUnavailable | 1 | OK |
      | dualwrite | open | 503 | -candidate/_doc/p-late | DependencyUnavailable | 1 | DependencyUnavailable |

  Szenario: Ein beidseitiges Preisintervall beschränkt die öffentliche Suche
    Wenn die öffentliche Suche ausgeführt wird
      | Page | Sort | MinPrice | MaxPrice |
      | 1 | priceAsc | 10000 | 10000 |
    Dann liefert die Suchverbindung
      | Code | TotalProducts |
      | OK | 1 |

  Szenario: Eine abgeschnittene Aliasantwort bleibt ein Transportfehler
    Angenommen die Suchdienstverbindung eine explizite Protokollstörung liefert
      | Connection | Status | Body | Clock | FailPath |
      | truncated | 200 | {} | controlled | /_aliases |
    Und die Rebuildquelle einen expliziten Fehler liefert
      | Phase | Code |
      | none | SourceUnavailable |
    Wenn ein vollständiger Quellsnapshot in einen neuen physischen Suchindex aufgebaut wird
      | IndexSuffix | Cursor |
      | -candidate | fault-source |
    Dann der Rebuildabschluss beobachtet wird
      | Code | AbortCount |
      | SearchUnavailable | 1 |

  Szenario: Ein nicht interpretierbares Interessengewicht erzeugt keinen Bonus
    Angenommen vergleichbare synthetische Suchkandidaten liegen vor
      | ProductID | Name | Slug | Brand | SKU | Color | StorageGB | Price | Stock | Published | CategoryID | InterestWeight | PurchaseCost |
      | p-b | Beta Testtelefon | beta | Beta | b-red | red | 128 | 10000 | 1 | true | phones | invalid | 5000 |
    Wenn die Suche sortiert wird
      | Text | Page | Sort |
      | | 1 | relevance |
    Dann liefert die Suchverbindung
      | Code | TotalProducts |
      | OK | 2 |

  Szenariogrundriss: Schreibprojektionen erhalten mehr Zeit während öffentliche Suche ihr Fehlerbudget behält
    Angenommen der echte Suchdienst über eine verzögerte Verbindung angesprochen wird
      | Milliseconds | Client |
      | 600 | <Client> |
    Wenn der Suchadapter die Verwaltungsoperation ausführt
      | Operation | ProductID |
      | apply | p-a |
    Dann liefert die Suchverbindung
      | Code |
      | OK |
    Wenn die echte öffentliche Suche abgefragt wird
      | Text | Page | Sort |
      | Alpha | 1 | relevance |
    Dann liefert die Suchverbindung
      | Code |
      | SearchUnavailable |
    Und bleibt die öffentliche Suchantwort innerhalb des Fehlerbudgets
      | MaximumMilliseconds |
      | 500 |

    Beispiele:
      | Client |
      | default |
      | custom |

  Szenario: Wiederholte vollständige Suchanfragen nutzen den echten Indexrequestcache
    Angenommen der echte Suchantwortkörper zur Transportprüfung beobachtet wird
      | CaptureResponse |
      | true |
    Wenn die echte öffentliche Suche abgefragt wird
      | Text | Page | Sort |
      | Alpha | 1 | relevance |
    Dann liefert die Suchverbindung
      | Code | TotalProducts |
      | OK | 1 |
    Angenommen der echte Suchindex seinen Requestcachezähler vormerkt
    Wenn die identische öffentliche Suche mehrfach ausgeführt wird
      | Attempts |
      | 3 |
    Dann der Suchindex echte Requestcachetreffer meldet
      | MinimumHitDelta |
      | 6 |
    Und liefert die Suchverbindung
      | Code | TotalProducts |
      | OK | 1 |
    Und die echte Suchantwort nur benötigte Rootdaten und die vollständige Variantkarte enthält
      | RootFields | ProductID |
      | 1 | p-a |
