# language: de
@opensearch
Funktionalität: Explizite Wortkorrektur statt automatisch unscharfer Haupttreffer
  Als Käufer sehe ich normale deutsche Texttreffer und wähle eine angebotene Tippfehlerkorrektur bewusst aus.

  Szenariogrundriss: Hauptliste und vollständiges Korrekturangebot bleiben getrennt
    Angenommen eine synthetische Suchquelle mit vollständigen Lieferantendaten besteht
      | Index | Locale | Currency | Market | At | CategoryID | FamilyID | Manufacturer | Street | PostalCode | City | Country | Email |
      | shop-word-correction | de-DE | EUR | DE | 2026-10-15T00:00:00Z | phones | smartphones | Testlieferant | Testweg 1 | 12345 | Teststadt | DE | demo@example.invalid |
    Und veröffentlichte Suchbegriffe sind vorhanden
      | ProductID | Name | Slug | Brand | SKU | Color | StorageGB | Price | Stock | Published |
      | p-phone | Fairphone 5 | fairphone | Fairphone | PHONE | red | 128 | 10000 | 1 | true |
      | p-case | Fairphone Schutzhuelle | huelle | Fairphone | CASE | red | 128 | 1000 | 1 | true |
      | p-cleaner | Schnell Entkalker | entkalker | Becher | CLEANER | red | 128 | 799 | 1 | true |
      | p-private | Geheimzauber | geheim | Geheimmarke | PRIVATE | red | 128 | 999 | 1 | false |
    Wenn Vorschläge angefragt werden
      | Query | Sort | Page |
      | <Eingabe> | relevance | 1 |
    Dann lauten die tatsächlichen Modellkarten
      | ProductIDs | TotalProducts |
      | <Produkte> | <Anzahl> |
    Und lauten die Vorschläge
      | Names | Correction | QueryReplaced | MaximumCount |
      | <Vorschlag> | <Korrektur> | false | 5 |
    Beispiele:
      | Eingabe | Produkte | Anzahl | Vorschlag | Korrektur |
      | Fairphone Entkalkre | p-case,p-phone | 2 | Fairphone Entkalker | true |
      | Entkalkre | | 0 | Entkalker | true |
      | Fairphone Entkalker | p-case,p-cleaner,p-phone | 3 | [] | false |
      | Fairphone Unbekannt Entkalkre | p-case,p-phone | 2 | Fairphone Unbekannt Entkalker | true |
      | fairphone Entkalkre | p-case,p-phone | 2 | fairphone Entkalker | true |
      | Geheimzaubrr | | 0 | [] | false |
      | Fairphone | p-case,p-phone | 2 | Fairphone 5,Fairphone Schutzhuelle | false |
      | Fairphone Unbekannt | p-case,p-phone | 2 | [] | false |
      | Fairphone xy | p-case,p-phone | 2 | Fairphone Schutzhuelle | true |

  Szenariogrundriss: Entzogene Wortquellen verschwinden ohne Wörterbuchleck
    Angenommen eine synthetische Suchquelle mit vollständigen Lieferantendaten besteht
      | Index | Locale | Currency | Market | At | CategoryID | FamilyID | Manufacturer | Street | PostalCode | City | Country | Email |
      | shop-revoked-words | de-DE | EUR | DE | 2026-10-15T00:00:00Z | phones | smartphones | Testlieferant | Testweg 1 | 12345 | Teststadt | DE | demo@example.invalid |
    Und veröffentlichte Suchbegriffe sind vorhanden
      | ProductID | Name | Slug | Brand | SKU | Color | StorageGB | Price | Stock | Published |
      | p-cleaner | Schnell Entkalker | entkalker | Becher | CLEANER | red | 128 | 799 | 1 | true |
    Wenn Vorschläge angefragt werden
      | Query | Sort | Page |
      | Entkalkre | relevance | 1 |
    Dann lauten die Vorschläge
      | Names | Correction | QueryReplaced | MaximumCount |
      | Entkalker | true | false | 5 |
    Wenn die öffentliche Wortquelle erneut projiziert wird
      | ProductID | Revision | Published | Deleted |
      | p-cleaner | 2 | false | <Geloescht> |
    Und Vorschläge angefragt werden
      | Query | Sort | Page |
      | Entkalkre | relevance | 1 |
    Dann lauten die Vorschläge
      | Names | Correction | QueryReplaced | MaximumCount |
      | [] | false | false | 5 |
    Beispiele:
      | Geloescht |
      | false |
      | true |

  Szenariogrundriss: Deutscher Wortstamm und höherer Namensscore bleiben erhalten
    Angenommen eine synthetische Suchquelle mit vollständigen Lieferantendaten besteht
      | Index | Locale | Currency | Market | At | CategoryID | FamilyID | Manufacturer | Street | PostalCode | City | Country | Email |
      | shop-german-text-score | de-DE | EUR | DE | 2026-10-15T00:00:00Z | phones | smartphones | Testlieferant | Testweg 1 | 12345 | Teststadt | DE | demo@example.invalid |
    Und veröffentlichte Suchbegriffe sind vorhanden
      | ProductID | Name | Slug | Brand | SKU | Color | StorageGB | Price | Stock | Published |
      | p-name | Telefon | telefon | Alpha | PHONE | red | 128 | 10000 | 1 | true |
      | p-description | Zubehoer | zubehoer | Beta | EXTRA | red | 128 | 1000 | 1 | true |
    Wenn die deutsche Suchbeschreibung projiziert wird
      | ProductID | Revision | Description |
      | p-description | 2 | Telefon |
    Und die Suche sortiert wird
      | Text | Sort | Page |
      | <Eingabe> | relevance | 1 |
    Dann lautet die öffentliche Ergebnisreihenfolge
      | ProductIDs | ExposedMargin | ExposedInterestWeight |
      | p-name,p-description | false | false |
    Beispiele:
      | Eingabe |
      | Telefon |
      | Telefonen |

  Szenario: Haupttext und öffentliche Wortkorrekturen teilen denselben Multirequest
    Angenommen ein beobachtbarer Suchprotokollorigin vorbereitet ist
      | ResponseFixture |
      | spec/fixtures/opensearch/msearch-response.json |
    Wenn die Suche über den Protokollorigin ausgeführt wird
      | Text |
      | Fairphone Entkalkre |
    Dann der Suchdrahtvertrag entspricht
      | Path | ContentType | Calls | Lines | FinalNewline |
      | /protocol/_msearch | application/x-ndjson | 1 | 4 | true |
    Und die NDJSON-Teilabfragen entsprechen
      | Line | Path | JSON |
      | 0 | request_cache | true |
      | 2 | request_cache | true |
      | 1 | query.bool.must | [{"multi_match":{"query":"Fairphone Entkalkre","fields":["name^3","brand.text^2","description"]}}] |
      | 3 | suggest.catalog.prefix | "Fairphone Entkalkre" |
      | 3 | suggest.word0.prefix | "Fairphone" |
      | 3 | suggest.word0.completion.field | "words" |
      | 3 | suggest.word0.completion.contexts.visibility | ["public"] |
      | 3 | suggest.word1.prefix | "Entkalkre" |
      | 3 | suggest.word1.completion.fuzzy | {"fuzziness":2,"min_length":4,"prefix_length":2,"transpositions":true} |
      | 3 | suggest.word1.completion.size | 25 |
