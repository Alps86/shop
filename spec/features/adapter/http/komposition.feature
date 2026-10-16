# language: de
@http @frontend
Funktionalität: Originfallback und klar getrennte Fragmentantworten
  Als Gast erhalte ich bei Fragmentfehlern keine ungeprüfte Kaufaktion.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Nur erlaubte öffentliche Fragmente liefern HTML
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | <methode> | <url> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | <form> | <!doctype |
    Beispiele:
      | methode | url | status | form |
      | GET | /fragments/de-DE/ | 200 | Absent |
      | GET | /fragments/de-DE/suche?q=phone | 200 | Absent |
      | GET | /fragments/de-DE/kategorie/smartphones--c-phone | 200 | Absent |
      | GET | /fragments/de-DE/produkt/testmodell--p-phone?sku=RED-256&rootVersion=1 | 200 | Absent |
      | POST | /fragments/de-DE/ | 405 | Body |
      | GET | /fragments/de-DE/warenkorb | 404 | Body |

  Szenario: Origin liefert bei bekannter Fragmentstörung sicheres ErsatzHTML
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | Product |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /fragments/de-DE/produkt/testmodell--p-phone?sku=RED-256 |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:Cache-Control | private,no-store |
      | Body | Daten vorübergehend nicht verfügbar |
      | Absent | In den Warenkorb |
      | Absent | 120,00 EUR |

  Szenario: Persönliche Anfragen werden auch bei Kompositionsheader nicht zerlegt
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | Composition |
      | GET | /de-DE/warenkorb | test-session-a | 1 |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:Cache-Control | private,no-store |
      | Absent | <esi:include |

  Szenario: Nicht dynamische Route benötigt keinen ESIabruf
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Composition |
      | GET | /de-DE/service | 1 |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Absent | <esi:include |
