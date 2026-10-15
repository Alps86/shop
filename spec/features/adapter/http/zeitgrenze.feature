# language: de
@http
Funktionalität: Cachelaufzeit endet vor zeitgebundenen Pflichtdaten
  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |
    Und die HTTPkonfiguration verwendet
      | Clock |
      | fixture |

  Szenariogrundriss: Veröffentlichte Listings verkürzen ihren Cache am nächsten Stichtag
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Now":"2026-10-16T12:00:00Z","Search":{"Cards":[{"SKU":"RED-256","Name":"Zeitangebot","Compliance":<compliance>}]}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/suche |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:Cache-Control | <cache> |
    Beispiele:
      | compliance | cache |
      | {"Context":{"ValidUntil":{"Present":true,"Value":"2026-10-16T12:00:02Z"}}} | public,max-age=2 |
      | {"Context":{"ValidUntil":{"Present":true,"Value":"2026-10-16T11:59:59Z"}}} | private,no-store |
      | {"ApplicableRules":[{"ValidUntil":{"Present":true,"Value":"2026-10-16T12:00:03Z"}}]} | public,max-age=3 |
      | {"ApplicableRules":[{"ValidFrom":{"Present":true,"Value":"2026-10-16T12:00:04Z"}}]} | public,max-age=4 |
      | {"ApplicableRules":[{"ValidFrom":{"Present":true,"Value":"2026-10-16T11:59:59Z"}}]} | public,max-age=5 |
      | {"ApplicableRules":[{}]} | public,max-age=5 |
