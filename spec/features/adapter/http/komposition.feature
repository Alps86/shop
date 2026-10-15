# language: de
@spec-only @frontend
Funktionalität: Hugo SSI ESI und sichere Cachegrenzen
  Als Gast erhalte ich vollständig zusammengesetzte Seiten ohne fremde persönliche Daten.

  Szenariogrundriss: Standardengines setzen statische und dynamische Teile zusammen
    Angenommen Hugo hat einen Markdownrahmen erzeugt
      | BuildVersion | SSI | ESI | Locale | Market | Currency |
      | build-a | header,footer | detail | de-DE | DE | EUR |
    Wenn eine Ansicht am öffentlichen Einstieg angefordert wird
      | Class | Cookie | HX-Request |
      | <klasse> | <cookie> | <htmx> |
    Dann lautet der Kompositionsvertrag
      | EngineOrder | RawIncludeMarkers | Shared | MaxAgeSeconds | ContainsForeignGuestData |
      | Nginx-SSI,Varnish-ESI | false | <shared> | <ttl> | false |
    Beispiele:
      | klasse | cookie | htmx | shared | ttl |
      | markdownFrame | absent | false | true | 3600 |
      | categoryNavigation | absent | false | true | 300 |
      | productDetail | absent | false | true | 5 |
      | listing | absent | true | true | 5 |
      | personalizedSearch | guest-a | true | false | 0 |
      | recommendation | guest-a | true | false | 0 |
      | cart | guest-a | false | false | 0 |
      | checkout | guest-a | false | false | 0 |
      | order | guest-a | false | false | 0 |

  Szenariogrundriss: Sichere Fragmentfehler
    Wenn ein Detailfragment an einer definierten Grenze ausfällt
      | Fault |
      | <stoerung> |
    Dann lautet die sichtbare Elternantwort
      | Status | BuyAction | UnverifiedPrice | Message | CacheControl |
      | <status> | false | false | <hinweis> | no-store |
    Beispiele:
      | stoerung | status | hinweis |
      | knownBeforeParentHeaders | 503 | retry |
      | originReturnsSafeFallback | 200 | Daten vorübergehend nicht verfügbar |
      | subrequestFailsAfterParentHeaders | 200 | adjacentFragmentError |

  Szenariogrundriss: Kontextwechsel darf keinen falschen Cachetreffer liefern
    Angenommen ein vollständiges Detailfragment ist gecacht
      | SKU | Locale | Market | Currency | RootVersion | ComplianceMarker |
      | BLUE | de-DE | DE | EUR | 1 | blue-verified |
    Wenn ein abweichender Request am Cache eingeht
      | ChangedContext | Value |
      | <kontext> | <wert> |
    Dann lautet die Cachebeobachtung
      | ReuseOldFragment | ValidateBeforeCache | PersonalDataInKey |
      | false | true | false |
    Beispiele:
      | kontext | wert |
      | SKU | RED |
      | Locale | fr-FR |
      | Market | FR |
      | Currency | USD |
      | RootVersion | 2 |
      | HX-Request | true |

  Szenario: Pflichtdatenwechsel begrenzt die ansonsten gültige TTL
    Angenommen ein öffentliches Angebot wechselt seinen Pflichtdatenstand
      | Now | NextValidityBoundary | DefaultTTLSeconds |
      | 2026-10-31T23:59:58Z | 2026-11-01T00:00:00Z | 5 |
    Wenn das Fragment ausgeliefert wird
    Dann lautet die Cachebegrenzung
      | MaximumTTLSeconds | StaleAfterBoundary |
      | 2 | false |

  Szenario: Persönlicher Badge wird separat geladen
    Wenn zwei Gäste denselben statischen Rahmen laden
      | Guest | CartQuantity |
      | a | 2 |
      | b | 7 |
    Dann lautet die sichtbare Trennung
      | SharedFrameBadge | PrivateBadgeA | PrivateBadgeB | PrivateCacheControl |
      | Warenkorb | 2 | 7 | private,no-store |
