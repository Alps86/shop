# language: de
@http
Funktionalität: Routing und sichere Fehlerantworten
  Als Gast erhalte ich konsistente Statuscodes und teilbare HTMLansichten.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Identität und Locale bestimmen die Route
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | <url> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Header:Location | <ziel> |
    Beispiele:
      | url | status | ziel |
      | / | 302 | /de-DE/ |
      | /de-DE/produkt/alt--p-phone?sku=RED-256 | 301 | /de-DE/produkt/testmodell--p-phone?sku=RED-256 |
      | /de-DE/kategorie/alt--c-phone?brand=Beta | 301 | /de-DE/kategorie/smartphones--c-phone?brand=Beta |
      | /de-DE/produkt/alt--missing | 404 | |
      | /de-DE/produkt/alt--p-removed | 410 | |
      | /de-DE/produkt/testmodell--p-phone?sku=OTHER | 404 | |
      | /de-DE/produkt/testmodell--p-phone?sku=bad%20sku | 422 | |
      | /de-DE/produkt/testmodell--p-phone?redirect=https://foreign.example | 422 | |
      | /fr-FR/ | 404 | |
      | /zz-ZZ/ | 400 | |
      | /de-DE/unbekannt | 404 | |
      | /de-DE/kategorie/alt--missing | 404 | |

  Szenariogrundriss: Queryfehler werden nicht still ersetzt
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | <url> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 422 |
      | Header:Cache-Control | private,no-store |
    Beispiele:
      | url |
      | /de-DE/suche?sort=unknown |
      | /de-DE/suche?page=abc |
      | /de-DE/suche?page=0 |
      | /de-DE/suche?view=unknown |
      | /de-DE/suche?redirect=https://foreign.example |
      | /de-DE/kategorie/smartphones--c-phone?color=invalid |
      | /de-DE/suche?brand=a&brand=b&brand=c&brand=d&brand=e&brand=f&brand=g&brand=h&brand=i&brand=j&brand=k |

  Szenariogrundriss: Lesende Navigation löst keine Commands aus
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | <methode> | <url> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 405 |
    Beispiele:
      | methode | url |
      | POST | /de-DE/ |
      | POST | /de-DE/marken |
      | POST | /de-DE/kategorie/smartphones--c-phone |
      | POST | /de-DE/produkt/testmodell--p-phone |
      | POST | /de-DE/warenkorb |
      | POST | /de-DE/kasse |
      | POST | /de-DE/bestellung/order-a |
      | POST | /de-DE/warenkorb/badge |
      | GET | /de-DE/gast |
      | GET | /de-DE/warenkorb/aktion |

  Szenariogrundriss: Reale Portstörung bleibt sichtbar
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | <port> |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | Listener |
      | GET | <url> | <session> | <listener> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Body | <hinweis> |
    Beispiele:
      | port | url | session | listener | status | hinweis |
      | Search | /de-DE/suche?q=phone | | public | 503 | Daten vorübergehend nicht verfügbar |
      | Product | /de-DE/produkt/testmodell--p-phone | | public | 503 | Daten vorübergehend nicht verfügbar |
      | Category | /de-DE/kategorie/smartphones--c-phone | | public | 503 | Daten vorübergehend nicht verfügbar |
      | Open | /de-DE/warenkorb | test-session-a | public | 503 | Daten vorübergehend nicht verfügbar |
      | GetCart | /de-DE/kasse | test-session-a | public | 503 | Daten vorübergehend nicht verfügbar |
      | GetOrder | /de-DE/bestellung/order-a | test-session-a | public | 503 | Daten vorübergehend nicht verfügbar |
      | Snapshot | /backend | | backend | 503 | Daten vorübergehend nicht verfügbar |
      | ListOrders | /backend | | backend | 503 | Daten vorübergehend nicht verfügbar |
      | AdminGetOrder | /backend/bestellung/order-a | | backend | 503 | Daten vorübergehend nicht verfügbar |
