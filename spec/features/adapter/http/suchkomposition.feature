# language: de
@http @http-composition
Funktionalität: Suchseite ohne doppelte Parentfragmentabfrage
  Als Gast erhalte ich die vollständige Suchansicht mit einer Suchportabfrage und unverändertem Fehlerstatus.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |
    Und die HTTPkonfiguration verwendet
      | Clock |
      | fixture |
    Und HTTPportdaten werden ersetzt
      | JSON |
      | {"Now":"2026-10-16T12:00:00Z","Search":{"TotalProducts":25}} |
    Und die Standardengines werden mit Hugo gestartet
      | Nginx | Varnish | CacheKey |
      | deploy/nginx/nginx.conf | deploy/varnish/default.vcl | synthetic-composition-test-key-32-bytes |

  Szenario: Die gewöhnliche öffentliche Suche fragt auch bei zweitem Portfehler nur einmal ab
    Angenommen die Suchaufrufe ab jetzt gezählt werden
      | Fault |
      | Search#2:InternalError |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=phone |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 1 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Testmodell &lt;b&gt; |
      | Body | data-composition="nginx-ssi" |
      | Body | href="/de-DE/marken" |
      | Header:X-Robots-Tag | noindex,follow |
      | Body | rel="canonical" href="/de-DE/suche?q=phone" |
      | Absent | <esi:include |
      | Absent | <!--# |
      | Absent | Daten vorübergehend nicht verfügbar |

    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=phone |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 2 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 503 |
      | Header:Cache-Control | private,no-store |
      | Body | Daten vorübergehend nicht verfügbar |
      | Absent | Testmodell |
      | Absent | <esi:include |
      | Absent | data-composition="nginx-ssi" |

  Szenariogrundriss: Suchfehler und ungültige Eingaben verlieren weder Status noch sichere Fehlerdarstellung
    Angenommen die Suchaufrufe ab jetzt gezählt werden
      | Fault |
      | <fehler> |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | <url> |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | <aufrufe> |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Header:Cache-Control | private,no-store |
      | Header:X-Robots-Tag | noindex,nofollow |
      | Absent | <esi:include |
      | Absent | data-composition="nginx-ssi" |
      | Absent | Testmodell |
    Beispiele:
      | fehler | url | aufrufe | status |
      | Search#1:InternalError | /de-DE/suche?q=phone | 1 | 503 |
      | | /de-DE/suche?sort=invalid | 0 | 422 |
      | | /de-DE/suche?unknown=invalid | 0 | 422 |

  Szenario: Pagination und normalisierte Suchmetadaten bleiben im einmaligen SSR erhalten
    Angenommen die Suchaufrufe ab jetzt gezählt werden
      | Fault |
      | Search#2:InternalError |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=phone&page=2&sort=relevance |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 1 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | <title>Technik entdecken |
      | Body | rel="canonical" href="/de-DE/suche?page=2&amp;q=phone" |
      | Body | aria-current="page" |
      | Body | href="/de-DE/suche?page=3&amp;q=phone" |
      | Body | data-composition="nginx-ssi" |
      | Header:X-Robots-Tag | noindex,follow |

  Szenariogrundriss: HTMX und eigene Gäste behalten getrennte Antworten ohne gemeinsame Includes
    Angenommen die Suchaufrufe ab jetzt gezählt werden
      | Fault |
      | Search#2:InternalError |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie | HTMX |
      | /de-DE/suche?q=phone | <cookie> | <hx> |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 1 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Testmodell &lt;b&gt; |
      | Header:Vary | HX-Request |
      | Header:Cache-Control | <cache> |
      | Absent | <esi:include |
      | Absent | data-composition="nginx-ssi" |
    Beispiele:
      | cookie | hx | cache |
      | | true | public,max-age=0 |
      | shop_session=test-session-a | | private,no-store |
      | shop_session=test-session-a | true | private,no-store |

  Szenario: Pflichtdatenfristen begrenzen die direkte Suchantwort weiterhin
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Search":{"Cards":[{"SKU":"RED-256","Name":"Zeitangebot","Compliance":{"Context":{"ValidUntil":{"Present":true,"Value":"2026-10-16T12:00:02Z"}}}}]}} |
    Und die Suchaufrufe ab jetzt gezählt werden
      | Fault |
      | Search#2:InternalError |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Composition |
      | GET | /de-DE/suche?q=phone | 1 |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 1 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:Cache-Control | public,max-age=2 |
      | Body | Zeitangebot |
      | Body | <!--# include virtual="/_static/header.html" --> |
      | Absent | <esi:include |
