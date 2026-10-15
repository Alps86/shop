# language: de
@http
Funktionalität: Veröffentlichte Routen und Crawlerzugriff
  Als Suchmaschine finde ich nur freigegebene Basisrouten und keine persönlichen Zustände.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenario: Sitemap verwendet ausschließlich die veröffentlichte Allowlist
    Angenommen die HTTPkonfiguration verwendet
      | Routes |
      | [{"Path":"/de-DE/"},{"Path":"/de-DE/produkt/testmodell--p-phone"},{"Path":"/de-DE/kategorie/smartphones--c-phone"},{"Path":"/de-DE/service"},{"Path":"/de-DE/marken"}] |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /sitemap.xml |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:Content-Type | application/xml |
      | BodyAll | http://localhost:8080/de-DE/produkt/testmodell--p-phone,http://localhost:8080/de-DE/kategorie/smartphones--c-phone |
      | Absent | ?sku= |
      | Absent | ?brand= |
      | Absent | bestellung |

  Szenariogrundriss: Persönliche oder uneindeutige Routen werden schon bei Initialisierung abgewiesen
    Angenommen die HTTPkonfiguration verwendet
      | Routes |
      | [{"Path":"<route>"}] |
    Wenn der HTTPadapter wird initialisiert
      | Listener |
      | public |
    Dann die Adapterinitialisierung meldet
      | Error |
      | öffentliche Route |
    Beispiele:
      | route |
      | % |
      | http://foreign.example/de-DE/ |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 |
      | /de-DE/bestellung/order-a |
      | /backend |

  Szenario: Robots sperrt persönliche Pfade und lässt Ressourcen renderbar
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /robots.txt |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | BodyAll | Disallow: /backend,Disallow: /de-DE/warenkorb,Disallow: /de-DE/kasse,Disallow: /de-DE/bestellung/,Sitemap: http://localhost:8080/sitemap.xml |
      | Absent | Disallow: /assets |
      | Absent | Disallow: /de-DE/kategorie |

  Szenariogrundriss: Crawlerdateien erlauben keine Mutation
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | POST | <url> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 405 |
    Beispiele:
      | url |
      | /robots.txt |
      | /sitemap.xml |
