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

  Szenario: Öffentliche Empfehlungen begrenzen auch das vollständige Produktfragment
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Now":"2026-10-16T12:00:00Z","Recommendations":{"Cards":[{"SKU":"CASE","Name":"Zeitgebundene Empfehlung","URL":"/de-DE/produkt/testmodell--p-phone?sku=RED-256","UnitPrice":{"MinorUnits":2000,"Currency":"EUR"},"Compliance":{"Context":{"VerificationStatus":"verified","ValidUntil":{"Present":true,"Value":"2026-10-16T12:00:02Z"}}}}]}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /fragments/de-DE/produkt/testmodell--p-phone?sku=RED-256&rootVersion=1 |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Zeitgebundene Empfehlung |
      | Header:Cache-Control | public,max-age=2 |

  Szenario: Echter öffentlicher ESIcache endet an der kürzeren Empfehlungsgrenze
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Now":"2026-10-16T12:00:00Z","Recommendations":{"Cards":[{"SKU":"CASE","Name":"Alte zeitgebundene Empfehlung","URL":"/de-DE/produkt/testmodell--p-phone?sku=RED-256","UnitPrice":{"MinorUnits":2000,"Currency":"EUR"},"Compliance":{"Context":{"VerificationStatus":"verified","ValidUntil":{"Present":true,"Value":"2026-10-16T12:00:02Z"}}}}]}} |
    Und die Standardengines werden mit Hugo gestartet
      | Nginx | Varnish | CacheKey |
      | deploy/nginx/nginx.conf | deploy/varnish/default.vcl | synthetic-composition-test-key-32-bytes |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Alte zeitgebundene Empfehlung |
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Recommendations":{"Cards":[{"SKU":"CASE","Name":"Neue Empfehlung nach Ablauf","URL":"/de-DE/produkt/testmodell--p-phone?sku=RED-256","UnitPrice":{"MinorUnits":2100,"Currency":"EUR"},"Compliance":{"Context":{"VerificationStatus":"verified"}}}]}} |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Body | Alte zeitgebundene Empfehlung |
    Wenn die Fragmentcachefrist verstreicht
      | Milliseconds |
      | 2200 |
    Und die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Neue Empfehlung nach Ablauf |
      | Absent | Alte zeitgebundene Empfehlung |
      | Header:Cache-Control | public,max-age=0 |
