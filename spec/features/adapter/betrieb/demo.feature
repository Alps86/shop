# language: de
@integration
Funktionalität: Reale Demo und ehrliche Performanceabnahme
  Als Betreiber starte ich echte Dienste und unterscheide HTTPmessung vom sichtbaren Browseraufbau.

  @demo-http
  Szenariogrundriss: Echte öffentliche Formulare schließen den Gastkauf idempotent und persistent ab
    Angenommen die echte öffentliche Demolaufzeit ist bereit
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | SeedFile | CacheKey | CSRFKey |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | spec/fixtures/demo/katalog.json | demo-cache-key-0123456789abcdef0123456789 | demo-csrf-key-0123456789abcdef0123456789 |
    Und ein eigener HTTPgast wird verwendet
      | Jar |
      | einkauf |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/warenkorb?add=FP5-BLACK-128 |
    Und der Gast sendet das sichtbare Formular
      | FormAction |
      | /de-DE/gast |
    Und die aktuelle Formularantwort wird begrenzt abgewartet
      | Path | Status | Contains | TimeoutSeconds |
      | /de-DE/warenkorb?add=FP5-BLACK-128 | 200 | Zum Warenkorb hinzufügen | 10 |
    Und der Gast sendet das sichtbare Formular
      | FormAction | action | sku | quantity | HX |
      | /de-DE/warenkorb/aktion | add | FP5-BLACK-128 | 2 | true |
    Dann die öffentliche Antwort enthält
      | Status | Contains | Header | Value |
      | 200 | FP5-BLACK-128 | Cache-Control | private,no-store |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/kasse |
    Und der Gast sendet das sichtbare Formular
      | FormAction | action | recipient | street | postalCode | city | country | return | HX |
      | /de-DE/warenkorb/aktion | address | Demo Gast | Musterweg 1 | 10115 | Berlin | DE | checkout | true |
    Und der Gast sendet das sichtbare Formular
      | FormAction | action | shipping | return | HX |
      | /de-DE/warenkorb/aktion | shipping | <versand> | checkout | true |
    Und der Gast sendet das sichtbare Formular
      | FormAction | action | payment | return | HX |
      | /de-DE/warenkorb/aktion | payment | <zahlung> | checkout | true |
    Und die echte Suchverbindung wird umgeschaltet
      | State |
      | unavailable |
    Und der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/suche?q=Fairphone |
    Dann die öffentliche Antwort enthält
      | Status | Contains |
      | 503 | Daten vorübergehend nicht verfügbar |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/produkt/fairphone-5--p-fairphone5?sku=FP5-BLACK-128 |
    Dann die öffentliche Antwort enthält
      | Status | Contains |
      | 200 | data-sku="FP5-BLACK-128" |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/kasse |
    Und der Gast sendet das sichtbare Formular
      | FormAction | action | return | HX |
      | /de-DE/warenkorb/aktion | confirm | checkout | true |
    Und die Formularanfrage wird für einen Wiederholungsversuch gespeichert
      | FormAction | action | return |
      | /de-DE/warenkorb/aktion | place | checkout |
    Und der unveränderte Abschluss wird erneut gesendet
      | FormAction | HX |
      | /de-DE/warenkorb/aktion | true |
    Dann die öffentliche Antwort enthält
      | Status | Header | Prefix |
      | 200 | HX-Redirect | /de-DE/bestellung/ |
    Dann die Bestellung ist über echte Kaufports beobachtbar
      | PaymentMethod | ShippingMethod | Items | TotalMinor |
      | <zahlung> | <versand> | 1 | <gesamt> |
    Und die öffentliche Antwort enthält
      | Status | Contains |
      | 200 | Simulation – keine echte Bestellung/Zahlung |
      | 200 | Unveränderlicher Kauf |
    Wenn der unveränderte Abschluss wird erneut gesendet
      | FormAction |
      | /de-DE/warenkorb/aktion |
    Dann die Bestellung ist über echte Kaufports beobachtbar
      | PaymentMethod | ShippingMethod | Items | TotalMinor |
      | <zahlung> | <versand> | 1 | <gesamt> |
    Wenn die Kaufdaten und Workflowausführungen werden festgehalten
      | TimeoutSeconds |
      | 10 |
    Und die echte Suchverbindung wird umgeschaltet
      | State |
      | available |
    Und ein eigener HTTPgast wird verwendet
      | Jar |
      | verwaltung |
    Und der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /backend/bestellung/{orderID} |
    Und der Gast sendet das sichtbare Formular
      | FormAction | sku | shipmentID | quantity | status |
      | /backend/lieferschritt | FP5-BLACK-128 | demo-teil-1 | 1 | shipped |
    Und die aktuelle Formularantwort wird begrenzt abgewartet
      | Path | Status | Contains | TimeoutSeconds |
      | /backend/bestellung/{orderID} | 200 | Gekauft 2, versandt 1, geliefert 0 | 10 |
    Und der Gast sendet das sichtbare Formular
      | FormAction | sku | shipmentID | quantity | status |
      | /backend/lieferschritt | FP5-BLACK-128 | demo-teil-1 | 1 | delivered |
    Und ein eigener HTTPgast wird verwendet
      | Jar |
      | einkauf |
    Und die aktuelle Formularantwort wird begrenzt abgewartet
      | Path | Status | Contains | TimeoutSeconds |
      | /de-DE/bestellung/{orderID} | 200 | gekauft 2, versandt 1, geliefert 1 | 10 |
    Dann Kaufdaten und Workflowausführungen sind unverändert
    Wenn die eigene Demolaufzeit wird persistent neu gestartet
    Dann Kaufdaten und Workflowausführungen sind unverändert
    Und die öffentliche Antwort enthält
      | Status | Contains |
      | 200 | Unveränderlicher Kauf |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/warenkorb?add=FP5-BLACK-128 |
    Dann die öffentliche Antwort enthält
      | Status | Contains |
      | 200 | Eigenen Einkauf starten |
    Wenn der Gast sendet das sichtbare Formular
      | FormAction |
      | /de-DE/gast |
    Dann die öffentliche Antwort enthält
      | Status | Header | Value |
      | 303 | Location | /de-DE/warenkorb?add=FP5-BLACK-128 |
    Wenn die aktuelle Formularantwort wird begrenzt abgewartet
      | Path | Status | Contains | TimeoutSeconds |
      | /de-DE/warenkorb?add=FP5-BLACK-128 | 200 | Zum Warenkorb hinzufügen | 10 |
    Und der Gast sendet das sichtbare Formular
      | FormAction | action | sku | quantity | HX |
      | /de-DE/warenkorb/aktion | add | FP5-BLACK-128 | 1 | true |
    Dann der persistierte Gastwarenkorb lautet
      | Items | Quantity | TotalMinor | NewCart |
      | 1 | 1 | 40390 | true |
    Und Kaufdaten und Workflowausführungen sind unverändert
    Beispiele:
      | zahlung | versand | gesamt |
      | invoice | standard | 80290 |
      | creditCard | express | 81290 |

  @demo-http
  Szenario: Zusammengesetzte Demoseiten erhalten SKU Pflichtdaten und teilbare Filter
    Angenommen die echte öffentliche Demolaufzeit ist bereit
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | SeedFile | CacheKey | CSRFKey |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | spec/fixtures/demo/katalog.json | demo-cache-key-0123456789abcdef0123456789 | demo-csrf-key-0123456789abcdef0123456789 |
    Und ein eigener HTTPgast wird verwendet
      | Jar |
      | katalog |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/ |
    Dann die öffentliche Antwort enthält
      | Status | Contains | Excludes |
      | 200 | Lichtblick Technik | <!--# |
      | 200 | Fairphone | <esi: |
      | 200 | hx-trigger="load, cartChanged from:body" | |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/suche?brand=Fairphone&brand=Constructa&sort=priceAsc |
    Dann die öffentliche Antwort enthält
      | Status | Contains | Header | Value |
      | 200 | 3 Modelle | X-Robots-Tag | noindex,follow |
      | 200 | FP5-BLACK-128 | | |
      | 200 | CWF14N30 | | |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/kategorie/smartphones--smartphones?brand=Fairphone&color=sky-blue&storageCapacity.min=256 |
    Dann die öffentliche Antwort enthält
      | Status | Contains | Excludes |
      | 200 | 1 Modelle | ?sku=FP5-BLACK-128 |
      | 200 | ?sku=FP5-BLUE-256 | ?sku=FP5-BLUE-256?sku= |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/produkt/fairphone-5--p-fairphone5?sku=FP5-BLUE-256 |
    Dann die öffentliche Antwort enthält
      | Status | Contains | Excludes |
      | 200 | data-sku="FP5-BLUE-256" | <!--# |
      | 200 | data-frame="hugo" | |
      | 200 | Energieeffizienzklasse | <esi: |
      | 200 | FP5-BLACK-128 | |
      | 200 | FairPhone B.V. | |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path | HX |
      | /de-DE/produkt/fairphone-5--p-fairphone5?sku=FP5-GREEN-128 | true |
    Dann die öffentliche Antwort enthält
      | Status | Contains | Excludes |
      | 200 | data-sku="FP5-GREEN-128" | <!doctype |
      | 200 | ?sku=FP5-TRANSPARENT-256 | |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/produkt/alter-name--p-fairphone5?sku=FP5-GREEN-128 |
    Dann die öffentliche Antwort enthält
      | Status | Header | Value |
      | 301 | Location | /de-DE/produkt/fairphone-5--p-fairphone5?sku=FP5-GREEN-128 |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/produkt/dr-becher-schnell-entkalker--p-drbecher-entkalker?sku=DB-ENTKALKER-500 |
    Dann die öffentliche Antwort enthält
      | Status | Contains |
      | 200 | Gefahr |
      | 200 | H314 |
      | 200 | P305+P351+P338 |
      | 200 | Grundpreis 15,98 EUR / 1 l |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path | Boundary |
      | /backend | public |
    Dann die öffentliche Antwort enthält
      | Status |
      | 404 |

  @demo-http
  Szenario: Freiwilliges Zubehör und das Interessenprofil bleiben im eigenen Gastwarenkorb
    Angenommen die echte öffentliche Demolaufzeit ist bereit
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | SeedFile | CacheKey | CSRFKey |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | spec/fixtures/demo/katalog.json | demo-cache-key-0123456789abcdef0123456789 | demo-csrf-key-0123456789abcdef0123456789 |
    Und ein eigener HTTPgast wird verwendet
      | Jar |
      | eigentuemer |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/warenkorb?add=FP5-BLACK-128 |
    Dann die öffentliche Antwort enthält
      | Status | Contains |
      | 200 | hx-trigger="load, cartChanged from:body" |
    Wenn der Gast sendet das sichtbare Formular
      | FormAction |
      | /de-DE/gast |
    Und die aktuelle Formularantwort wird begrenzt abgewartet
      | Path | Status | Contains | TimeoutSeconds |
      | /de-DE/warenkorb?add=FP5-BLACK-128 | 200 | Zum Warenkorb hinzufügen | 10 |
    Und der Gast sendet das sichtbare Formular
      | FormAction | action | sku | quantity | HX |
      | /de-DE/warenkorb/aktion | add | FP5-BLACK-128 | 1 | true |
    Dann die öffentliche Antwort enthält
      | Status | Contains | Header | Value |
      | 200 | FP5-CASE-BLACK | | |
      | 200 | vorschau=FP5-CASE-BLACK | | |
      | 200 | | HX-Trigger | cartChanged |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/warenkorb/badge |
    Dann die öffentliche Antwort enthält
      | Status | Contains | Header | Value |
      | 200 | Warenkorb (1) | Cache-Control | private,no-store |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/warenkorb?vorschau=FP5-CASE-BLACK |
    Dann die öffentliche Antwort enthält
      | Status | Contains |
      | 200 | Unverbindliche Zubehör- und Rabattvorschau |
      | 200 | Zubehör ablehnen |
    Und der persistierte Gastwarenkorb lautet
      | Items | Quantity | DiscountMinor | TotalMinor |
      | 1 | 1 | 0 | 40390 |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/warenkorb |
    Dann der persistierte Gastwarenkorb lautet
      | Items | Quantity | DiscountMinor | TotalMinor |
      | 1 | 1 | 0 | 40390 |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/warenkorb?vorschau=FP5-CASE-BLACK |
    Und der Gast sendet das sichtbare Formular
      | FormAction | action | sku | quantity | HX |
      | /de-DE/warenkorb/aktion | add | FP5-CASE-BLACK | 1 | true |
    Dann der persistierte Gastwarenkorb lautet
      | Items | Quantity | DiscountMinor | TotalMinor |
      | 2 | 2 | 500 | 41880 |
    Wenn der Gast sendet das sichtbare Formular
      | FormAction | action | sku | HX |
      | /de-DE/warenkorb/aktion | remove | FP5-CASE-BLACK | true |
    Dann der persistierte Gastwarenkorb lautet
      | Items | Quantity | DiscountMinor | TotalMinor |
      | 1 | 1 | 0 | 40390 |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/profil?profile=mobile |
    Dann die öffentliche Antwort enthält
      | Status | Contains |
      | 200 | profile=neutral |
      | 200 | profile=mobile |
      | 200 | profile=home |
      | 200 | profile=coffee |
    Wenn der Gast sendet das sichtbare Formular
      | FormAction | profile |
      | /de-DE/profil | mobile |
    Und ein eigener HTTPgast wird verwendet
      | Jar |
      | anderer-gast |
    Und der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/warenkorb |
    Dann die öffentliche Antwort enthält
      | Status | Contains | Excludes |
      | 200 | Dein Warenkorb ist leer. | FP5-BLACK-128 |
    Wenn ein eigener HTTPgast wird verwendet
      | Jar |
      | eigentuemer |
    Und der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/warenkorb |
    Dann der persistierte Gastwarenkorb lautet
      | Profile | Items | Quantity | DiscountMinor | TotalMinor |
      | mobile | 1 | 1 | 0 | 40390 |
    Und die öffentliche Antwort enthält
      | Status | Contains | Header | Value |
      | 200 | FP5-BLACK-128 | Cache-Control | private,no-store |

  @demo-http
  Szenario: Verwaltungsänderungen werden durch echte WALprojektion und PURGE öffentlich sichtbar
    Angenommen die echte öffentliche Demolaufzeit ist bereit
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | SeedFile | CacheKey | CSRFKey |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | spec/fixtures/demo/katalog.json | demo-cache-key-0123456789abcdef0123456789 | demo-csrf-key-0123456789abcdef0123456789 |
    Und ein eigener HTTPgast wird verwendet
      | Jar |
      | neutral |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/produkt/fairphone-5--p-fairphone5?sku=FP5-GREEN-128 |
    Dann die öffentliche Antwort enthält
      | Status | Contains |
      | 200 | 409,00 EUR |
    Wenn der öffentliche Produktcache wird tatsächlich wiederverwendet
      | Path | MinimumHits | TimeoutSeconds |
      | /de-DE/produkt/fairphone-5--p-fairphone5?sku=FP5-GREEN-128 | 1 | 3 |
    Dann die öffentliche Antwort enthält
      | Status | Contains |
      | 200 | 409,00 EUR |
    Wenn ein eigener HTTPgast wird verwendet
      | Jar |
      | verwaltung |
    Und der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /backend |
    Und der Gast sendet das sichtbare Formular
      | FormAction | sku | price | stock |
      | /backend/inventory | FP5-GREEN-128 | 41900 | 20 |
    Dann die öffentliche Antwort enthält
      | Status | Header | Value |
      | 303 | Location | /backend |
    Wenn ein eigener HTTPgast wird verwendet
      | Jar |
      | neutral |
    Und die aktuelle Formularantwort wird begrenzt abgewartet
      | Path | HX | Status | Contains | TimeoutSeconds |
      | /de-DE/kategorie/smartphones--smartphones?color=moss-green | true | 200 | 419,00 EUR | 2 |
    Und der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /de-DE/produkt/fairphone-5--p-fairphone5?sku=FP5-GREEN-128 |
    Dann die öffentliche Antwort enthält
      | Status | Contains | Excludes |
      | 200 | 419,00 EUR | 409,00 EUR |
    Wenn ein eigener HTTPgast wird verwendet
      | Jar |
      | verwaltung |
    Und der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /backend |
    Und der Gast sendet das sichtbare Formular
      | FormAction | sku | price | stock |
      | /backend/inventory | FP5-GREEN-128 | 40900 | 20 |
    Und ein eigener HTTPgast wird verwendet
      | Jar |
      | neutral |
    Und die aktuelle Formularantwort wird begrenzt abgewartet
      | Path | HX | Status | Contains | TimeoutSeconds |
      | /de-DE/kategorie/smartphones--smartphones?color=moss-green | true | 200 | 409,00 EUR | 2 |

  @demo-http
  Szenario: Zwei bestätigte HTTPgastkäufe konkurrieren um genau eine verfügbare Einheit
    Angenommen die echte öffentliche Demolaufzeit ist bereit
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | SeedFile | CacheKey | CSRFKey |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | spec/fixtures/demo/katalog.json | demo-cache-key-0123456789abcdef0123456789 | demo-csrf-key-0123456789abcdef0123456789 |
    Und ein eigener HTTPgast wird verwendet
      | Jar |
      | verwaltung |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /backend |
    Und der Gast sendet das sichtbare Formular
      | FormAction | sku | price | stock |
      | /backend/inventory | FP5-TRANSPARENT-256 | 46900 | 1 |
    Und eigene HTTPgastkäufe sind bis einschließlich Preisbestätigung vorbereitet
      | Jar | sku | quantity | recipient | street | postalCode | city | country | shipping | payment |
      | gast-eins | FP5-TRANSPARENT-256 | 1 | Erster Demogast | Musterweg 1 | 10115 | Berlin | DE | standard | invoice |
      | gast-zwei | FP5-TRANSPARENT-256 | 1 | Zweiter Demogast | Musterweg 2 | 10115 | Berlin | DE | express | creditCard |
    Und die vorbereiteten Abschlussformulare werden gleichzeitig gesendet
      | TimeoutSeconds |
      | 10 |
    Dann die echten konkurrierenden Käufe ergeben
      | SKU | Orders | Unavailable | Stock | Reserved |
      | FP5-TRANSPARENT-256 | 1 | 1 | 1 | 1 |
    Wenn ein eigener HTTPgast wird verwendet
      | Jar |
      | verwaltung |
    Und der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /backend |
    Und der Gast sendet das sichtbare Formular
      | FormAction | sku | price | stock |
      | /backend/inventory | FP5-TRANSPARENT-256 | 46900 | 20 |

  @demo-http
  Szenario: Ein geänderter Preis verlangt vor dem HTTPabschluss eine erneute ausdrückliche Bestätigung
    Angenommen die echte öffentliche Demolaufzeit ist bereit
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | SeedFile | CacheKey | CSRFKey |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | spec/fixtures/demo/katalog.json | demo-cache-key-0123456789abcdef0123456789 | demo-csrf-key-0123456789abcdef0123456789 |
    Und eigene HTTPgastkäufe sind bis einschließlich Preisbestätigung vorbereitet
      | Jar | sku | quantity | recipient | street | postalCode | city | country | shipping | payment |
      | preisgast | FP5-GREEN-128 | 1 | Preis Demogast | Musterweg 1 | 10115 | Berlin | DE | standard | invoice |
    Und ein eigener HTTPgast wird verwendet
      | Jar |
      | verwaltung |
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /backend |
    Und der Gast sendet das sichtbare Formular
      | FormAction | sku | price | stock |
      | /backend/inventory | FP5-GREEN-128 | 41900 | 20 |
    Und ein vorbereitetes Abschlussformular wird gesendet
      | Jar |
      | preisgast |
    Dann die öffentliche Antwort enthält
      | Status | Contains |
      | 409 | Daten geändert. Bitte neu laden und bestätigen. |
      | 409 | 423,90 EUR |
      | 409 | Aktuelle Positionen und Gesamtpreis ausdrücklich bestätigen |
    Und der persistierte Gastwarenkorb lautet
      | Items | Quantity | DiscountMinor | TotalMinor |
      | 1 | 1 | 0 | 42390 |
    Wenn der Gast sendet das sichtbare Formular
      | FormAction | action | return | HX |
      | /de-DE/warenkorb/aktion | confirm | checkout | true |
    Und die Formularanfrage wird für einen Wiederholungsversuch gespeichert
      | FormAction | action | return |
      | /de-DE/warenkorb/aktion | place | checkout |
    Und der unveränderte Abschluss wird erneut gesendet
      | FormAction |
      | /de-DE/warenkorb/aktion |
    Dann die Bestellung ist über echte Kaufports beobachtbar
      | PaymentMethod | ShippingMethod | Items | TotalMinor |
      | invoice | standard | 1 | 42390 |
    Wenn die Kaufdaten und Workflowausführungen werden festgehalten
      | TimeoutSeconds |
      | 10 |
    Und ein eigener HTTPgast wird verwendet
      | Jar |
      | verwaltung |
    Und der Gast ruft die öffentliche Ansicht auf
      | Path |
      | /backend |
    Und der Gast sendet das sichtbare Formular
      | FormAction | sku | price | stock |
      | /backend/inventory | FP5-GREEN-128 | 40900 | 20 |
    Und ein eigener HTTPgast wird verwendet
      | Jar |
      | preisgast |
    Dann Kaufdaten und Workflowausführungen sind unverändert
    Und die öffentliche Antwort enthält
      | Status | Contains |
      | 200 | 419,00 EUR |
      | 200 | Gesamt 423,90 EUR |

  @demo-performance
  Szenario: Tatsächliche HTTPlast hält Messumfang und Browsergrenze getrennt
    Angenommen der dokumentierte Messdatensatz ist geladen
      | Dataset | Products | Variants | Categories | Profiles | ProfileCodes | Orders | Concurrency | CounterFlushMilliseconds | Network | MinimumCPU | MinimumRAMGiB | Namespace | SKU | Recipient | Street | PostalCode | City | Country | Shipping | Payment | CategoryPath | ListingContains | SearchPath | ProductPath |
      | benchmark-synthetic-v1 | 30 | 90 | 5 | 4 | neutral,mobile,home,coffee | 100 | 10 | 1500 | loopback | 4 | 8 | default | BENCH-01-A | Synthetischer Messgast | Messweg 1 | 10115 | Berlin | DE | standard | invoice | /de-DE/kategorie/benchmark-mobile--mobile | bench-product-01 | /de-DE/suche?q=Synthetischer | /de-DE/produkt/synthetischer-messartikel-01--bench-product-01?sku=BENCH-01-A |
    Wenn die vereinbarten Anzeigerouten gemessen werden
      | Boundary | Cache | WarmupPerRoute | RequestsPerRoute | Routes |
      | requestStart-to-last-composed-HTML-byte | warm | 100 | 1000 | home,category,search,personalized-search,product,cart,checkout,order |
      | requestStart-to-last-composed-HTML-byte | cold-cache-services-ready | 0 | 100 | home,category,search,personalized-search,product,cart,checkout,order |
    Dann muss der Messbericht die Aussagegrenze enthalten
      | TargetP95Milliseconds | RequiredStatistics | RequiredEnvironment | ProvesBrowserLCP | MutatingCommandsIncluded | Measurements |
      | 100 | p50,p95,p99,errorRate | CPU,RAM,OS,images,dataset,TCP,compression | false | false | 16 |
