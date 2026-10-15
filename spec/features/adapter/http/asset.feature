# language: de
@http
Funktionalität: Lokale unveränderliche Clientassets
  Als Gast lade ich die schlichte Oberfläche ohne externe Skript- oder Schriftanbieter.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Echte lokale Inhalte bestimmen Cache und Medientyp
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | <url> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Header:Content-Type | <typ> |
      | Header:Cache-Control | <cache> |
      | Body | <inhalt> |
    Beispiele:
      | url | status | typ | cache | inhalt |
      | /assets/shop.ca28cfc907ed.css | 200 | text/css | public,max-age=31536000,immutable | system-ui |
      | /assets/shop.804f5309a8c8.js | 200 | text/javascript | public,max-age=31536000,immutable | htmx:beforeSwap |
      | /assets/vendor/htmx.min.d6fdc75f204e.js | 200 | text/javascript | public,max-age=31536000,immutable | htmx |
      | /assets/product.dad5b938174f.svg | 200 | image/svg+xml | public,max-age=31536000,immutable | Neutrale eigene Technikdarstellung |
      | /assets/product.svg | 200 | image/svg+xml | public,max-age=3600 | Neutrale eigene Technikdarstellung |
      | /assets/vendor/LICENSE | 200 | text/plain | public,max-age=3600 | Zero-Clause BSD |
      | /assets/missing.js | 404 | text/html | private,no-store | Asset nicht gefunden |
