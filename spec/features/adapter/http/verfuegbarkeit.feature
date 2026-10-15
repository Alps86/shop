# language: de
@http
Funktionalität: Öffentliche Vorschläge und begrenzte Zusatzdienste
  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Vorschläge verwenden dieselbe validierte Suche
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | <fault> |
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Search":{"Suggestions":[{"Text":"Testmodell","Correction":true}]}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | <method> | <url> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Body | <text> |
    Beispiele:
      | method | url | fault | status | text |
      | GET | /de-DE/vorschlag?q=Test | | 200 | Testmodell |
      | GET | /de-DE/vorschlag?page=0 | | 422 | |
      | GET | /de-DE/vorschlag?q=Test | Search | 503 | Daten vorübergehend nicht verfügbar |
      | POST | /de-DE/vorschlag | | 405 | |

  Szenariogrundriss: Optionale Empfehlungen lassen das Produkt bedienbar
    Angenommen die HTTPkonfiguration verwendet
      | Recommendations | Interests |
      | <recommendations> | <interests> |
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | <fault> |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Testmodell |
      | Header:Cache-Control | private,no-store |
    Beispiele:
      | recommendations | interests | fault |
      | absent | present | |
      | present | absent | |
      | present | present | Interests |
      | present | present | Recommend |

  Szenariogrundriss: Produktnamen und Zubehör sind ergänzende Cartdaten
    Angenommen die HTTPkonfiguration verwendet
      | Variants |
      | <variants> |
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | <fault> |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/warenkorb | test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Warenkorb |
    Beispiele:
      | variants | fault |
      | absent | |
      | present | ProductForSKU |

  Szenariogrundriss: Bestellseiten werden vor dem Leseport validiert
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Listener |
      | GET | /backend?page=<page> | backend |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
    Beispiele:
      | page | status |
      | 2 | 200 |
      | 0 | 422 |
      | abc | 422 |

  Szenariogrundriss: Fehlende Diagnose verhindert keine Verwaltung
    Angenommen die HTTPkonfiguration verwendet
      | Diagnostics |
      | <port> |
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | <fault> |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Listener |
      | GET | /backend | backend |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | <text> |
    Beispiele:
      | port | fault | text |
      | absent | | Lokale Verwaltung |
      | present | Diagnostics | Laufzeitdiagnose vorübergehend nicht verfügbar |

  Szenario: Eine personalisierte Kategorie bleibt privat
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/kategorie/smartphones--c-phone | test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:Cache-Control | private,no-store |
