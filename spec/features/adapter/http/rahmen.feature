# language: de
@http
Funktionalität: Versionierte interne Rahmen bleiben eine sichere Optimierung
  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |
    Und die HTTPkonfiguration verwendet
      | ContentDirectory |
      | /tmp/shop-nonexistent-http-frame |

  Szenariogrundriss: Fehlende oder ungeeignete Rahmen verwenden den GoRenderer
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | <json> |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | HTMX |
      | GET | <url> | <session> | <htmx> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Testmodell |
      | AbsentHeader:X-Accel-Redirect | |
    Beispiele:
      | json | url | session | htmx |
      | {} | /de-DE/produkt/testmodell--p-phone | | |
      | {} | /de-DE/produkt/testmodell--p-phone | test-session-a | |
      | {} | /de-DE/produkt/testmodell--p-phone | | true |
      | {"Detail":{"Product":{"Revision":0}}} | /de-DE/produkt/testmodell--p-phone | | |
      | {"Detail":{"Product":{"ID":"bad.id"}}} | /de-DE/produkt/testmodell--bad.id | | |
      | {} | /de-DE/kategorie/smartphones--c-phone?brand=Alpha | | |

  Szenariogrundriss: Ausfall der ergänzenden Angebotssuche lässt Navigation bedienbar
    Angenommen die HTTPkonfiguration verwendet
      | Search |
      | <port> |
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | <fault> |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | <url> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | <text> |
      | Absent | data-sku="BLUE-128" |
    Beispiele:
      | port | fault | url | text |
      | absent | | /de-DE/ | Technik |
      | present | Search | /de-DE/ | Angebote vorübergehend nicht verfügbar |
      | present | Search | /de-DE/marken | Angebote vorübergehend nicht verfügbar |
