# language: de
@http @frontend
Funktionalität: Stabile lesbare Routen und Canonicals
  Als Käufer teile ich Varianten und Filterzustände ohne Indexexplosion.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Canonical und Indexierung bleiben an dieselbe Identität gebunden
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | <url> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | <canonical> |
      | Header:X-Robots-Tag | <robots> |
    Beispiele:
      | url | canonical | robots |
      | /de-DE/ | href="/de-DE/" | index,follow |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | rel="canonical" href="/de-DE/produkt/testmodell--p-phone" | index,follow |
      | /de-DE/kategorie/smartphones--c-phone?page=2 | rel="canonical" href="/de-DE/kategorie/smartphones--c-phone?page=2" | index,follow |
      | /de-DE/kategorie/smartphones--c-phone?brand=Alpha&brand=Beta | rel="canonical" href="/de-DE/kategorie/smartphones--c-phone" | noindex,follow |
      | /de-DE/kategorie/smartphones--c-phone?brand=Beta&brand=Alpha&brand=Alpha&page=1&sort=relevance | rel="canonical" href="/de-DE/kategorie/smartphones--c-phone" | noindex,follow |
      | /de-DE/suche?q=phone | rel="canonical" href="/de-DE/suche?q=phone" | noindex,follow |

  Szenario: Geteilte Marken-ORauswahl wird normalisiert an den Port weitergereicht
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | HTMX |
      | GET | /de-DE/kategorie/smartphones--c-phone?brand=Beta&brand=Alpha&brand=Alpha&page=1&sort=relevance | true |
    Dann der verwendete Applicationport erhält
      | Brands | Page |
      | Alpha,Beta | 1 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | BodyAll | Alpha,Beta,hx-push-url="true",data-filter-form |
      | Header:Vary | HX-Request |
      | Absent | <!doctype |

  Szenario: Metadaten werden escaped und die aktive SKU ist vollständig SSR
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/produkt/testmodell--p-phone?sku=RED-256 |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | BodyAll | Testmodell &lt;b&gt;,data-sku="RED-256",Galerie RED-256,120&#44;00 EUR,verified-blue |
      | Absent | <title>Testmodell <b> |
      | Absent | hreflang |
      | Absent | shop_session= |
