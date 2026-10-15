# language: de
@spec-only @frontend
Funktionalität: Stabile lesbare Routen und Canonicals
  Als Käufer teile ich gültige Produkt- und Filterlinks ohne Indexexplosion.

  Hintergrund:
    Angenommen ein Produkt und eine Kategorie besitzen aktuelle Routen
      | ProductID | ProductSlug | CategoryID | CategorySlug | SKU | ForeignSKU | RemovedProductID |
      | p-phone | testmodell | c-phone | smartphones | BLUE-256 | OTHER | p-removed |

  Szenariogrundriss: Direkte Links Identität und Slugredirect
    Wenn eine öffentliche URL angefordert wird
      | URL |
      | <url> |
    Dann lautet der Routingvertrag
      | Status | Location | Canonical | Robots | ActiveSKU |
      | <status> | <ziel> | <canonical> | <robots> | <sku> |
    Beispiele:
      | url | status | ziel | canonical | robots | sku |
      | / | 302 | /de-DE/ | absent | absent | absent |
      | /de-DE/ | 200 | none | /de-DE/ | index,follow | absent |
      | /de-DE/produkt/testmodell--p-phone | 200 | none | /de-DE/produkt/testmodell--p-phone | index,follow | default |
      | /de-DE/produkt/testmodell--p-phone?sku=BLUE-256 | 200 | none | /de-DE/produkt/testmodell--p-phone | index,follow | BLUE-256 |
      | /de-DE/produkt/alt--p-phone?sku=BLUE-256 | 301 | /de-DE/produkt/testmodell--p-phone?sku=BLUE-256 | absent | absent | absent |
      | /de-DE/produkt/alt--missing | 404 | none | absent | noindex | absent |
      | /de-DE/produkt/alt--p-removed | 410 | none | absent | noindex | absent |
      | /de-DE/produkt/testmodell--p-phone?sku=OTHER | 404 | none | absent | noindex | absent |
      | /de-DE/produkt/testmodell--p-phone?sku=bad%20sku | 422 | none | absent | noindex | absent |
      | /fr-FR/produkt/testmodell--p-phone | 404 | none | absent | noindex | absent |
      | /zz-ZZ/produkt/testmodell--p-phone | 400 | none | absent | noindex | absent |
      | /de-DE/kategorie/smartphones--c-phone?page=2 | 200 | none | /de-DE/kategorie/smartphones--c-phone?page=2 | index,follow | absent |
      | /de-DE/kategorie/smartphones--c-phone?brand=Alpha&brand=Beta | 200 | none | /de-DE/kategorie/smartphones--c-phone | noindex,follow | absent |
      | /de-DE/suche?q=phone | 200 | none | /de-DE/suche?q=phone | noindex,follow | absent |
      | /de-DE/warenkorb | 200 | none | absent | noindex,nofollow | absent |

  Szenariogrundriss: Querynormalisierung und klare Grenzen
    Wenn ein Routingrequest normalisiert wird
      | Input |
      | <eingabe> |
    Dann lautet das Normalisierungsergebnis
      | Status | Normalized |
      | <status> | <normalisiert> |
    Beispiele:
      | eingabe | status | normalisiert |
      | brand=Beta&brand=Alpha&brand=Alpha&page=1&sort=relevance | 200 | brand=Alpha&brand=Beta |
      | filterCount=5 | 422 | absent |
      | valuesPerFacet=11 | 422 | absent |
      | queryLength=101 | 422 | absent |
      | urlLength=2049 | 414 | absent |
      | sort=unknown | 422 | absent |
      | redirect=https://foreign.example | 422 | absent |

  Szenario: Serverseitige Metadaten und Sitemap verwenden dieselbe Identität
    Wenn eine Variante nach vollständiger Includeverarbeitung ausgegeben wird
      | URL |
      | /de-DE/produkt/testmodell--p-phone?sku=BLUE-256 |
    Dann lautet der SEOinhalt
      | SSRActiveSKU | CanonicalProductID | EscapedMetadata | SitemapVariantURLs | SitemapFilteredURLs | UnpublishedHreflang | PersonalTokensInURL |
      | BLUE-256 | p-phone | true | 0 | 0 | 0 | false |
