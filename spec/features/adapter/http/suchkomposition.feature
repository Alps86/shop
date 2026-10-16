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

  Szenario: Ein gültiger anonymer Suchcachehit benötigt bei zweitem Portfehler keinen Originaufruf
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
      | 1 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:Cache-Control | public,max-age=0 |
      | Absent | Daten vorübergehend nicht verfügbar |
      | Body | Testmodell &lt;b&gt; |
      | Absent | <esi:include |
      | Body | data-composition="nginx-ssi" |

    Angenommen die Suchaufrufe ab jetzt gezählt werden
      | Fault |
      | Search:InternalError |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=phone |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 0 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:Age | 0 |
      | Body | Testmodell &lt;b&gt; |
    Wenn die Fragmentcachefrist verstreicht
      | Milliseconds |
      | 5500 |
    Und die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=phone |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 503 |
      | Header:Cache-Control | private,no-store |
      | Body | Daten vorübergehend nicht verfügbar |
      | Absent | Testmodell |
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
      | Absent | Testmodell |

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
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | <url> |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | <wiederholte-aufrufe> |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Header:Cache-Control | private,no-store |
    Beispiele:
      | fehler | url | aufrufe | status | wiederholte-aufrufe |
      | Search:InternalError | /de-DE/suche?q=phone | 1 | 503 | 2 |
      | | /de-DE/suche?sort=invalid | 0 | 422 | 0 |
      | | /de-DE/suche?unknown=invalid | 0 | 422 | 0 |

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

  Szenariogrundriss: Query Host und HTMX bilden getrennte öffentliche Cacheobjekte
    Angenommen die Suchaufrufe ab jetzt gezählt werden
      | Fault |
      | Search#3:InternalError |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Host |
      | /de-DE/suche?q=phone | first.localhost |
    Und die zusammengesetzte HTTPansicht wird angefordert
      | URL | Host |
      | /de-DE/suche?q=phone | first.localhost |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 1 |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Host | HTMX |
      | <url> | <host> | <hx> |
    Und die zusammengesetzte HTTPansicht wird angefordert
      | URL | Host | HTMX |
      | <url> | <host> | <hx> |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 2 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Testmodell &lt;b&gt; |
      | <metadatenfeld> | <metadaten> |
    Beispiele:
      | url | host | hx | metadatenfeld | metadaten |
      | /de-DE/suche?q=tablet | first.localhost | | Body | rel="canonical" href="/de-DE/suche?q=tablet" |
      | /de-DE/suche?q=phone&page=2 | first.localhost | | Body | rel="canonical" href="/de-DE/suche?page=2&amp;q=phone" |
      | /de-DE/suche?q=phone | second.localhost | | Body | rel="canonical" href="/de-DE/suche?q=phone" |
      | /de-DE/suche?q=phone | first.localhost | true | Absent | data-composition="nginx-ssi" |

  Szenariogrundriss: Cookies und Authorization umgehen den öffentlichen Suchhit
    Angenommen die Suchaufrufe ab jetzt gezählt werden
      | Fault |
      | |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=phone |
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Search":{"Cards":[{"ProductID":"p-private","SKU":"PRIVATE","Name":"Nur privater Origin"}]}} |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie | Authorization |
      | /de-DE/suche?q=phone | <cookie> | <authorization> |
    Und die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie | Authorization |
      | /de-DE/suche?q=phone | <cookie> | <authorization> |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 3 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Nur privater Origin |
      | Absent | Testmodell |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=phone |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 3 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Testmodell &lt;b&gt; |
      | Absent | Nur privater Origin |
    Beispiele:
      | cookie | authorization |
      | shop_session=test-session-a | |
      | unrelated=still-private | |
      | | Bearer private-probe |

  Szenario: Ein globaler Ban erreicht auch ein Suchobjekt ohne SKUtags
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Search":{"TotalProducts":0,"Cards":[]}} |
    Und die Suchaufrufe ab jetzt gezählt werden
      | Fault |
      | |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=new-product |
    Und die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=new-product |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 1 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:Surrogate-Key | shop:public-search |
      | Absent | Testmodell |
    Wenn der komponierte Cache invalidiert wird
      | Key | SKUs |
      | synthetic-composition-test-key-32-bytes | |
    Und die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=new-product |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 2 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:Surrogate-Key | shop:public-search |

  Szenario: Zwei Gastcookiejars und ihre Interessen überschreiben keinen anonymen Suchhit
    Angenommen getrennte Gastcookiejars werden eingerichtet
      | Jar | Session |
      | first | test-session-a |
      | second | test-session-b |
    Und die Suchaufrufe ab jetzt gezählt werden
      | Fault |
      | |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=phone |
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Interests":[{"CategoryID":"c-phone","Weight":"0.75"}],"Search":{"Cards":[{"ProductID":"p-private-first","SKU":"PRIVATE-FIRST","Name":"Erster privater Suchinhalt"}]}} |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Jar |
      | /de-DE/suche?q=phone | first |
    Dann der verwendete Applicationport erhält
      | SearchInterests |
      | [{c-phone 0.75}] |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:Cache-Control | private,no-store |
      | Body | Erster privater Suchinhalt |
      | Absent | Testmodell |
      | Absent | 0.75 |
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Interests":[{"CategoryID":"c-phone","Weight":"0.25"}],"Search":{"Cards":[{"ProductID":"p-private-second","SKU":"PRIVATE-SECOND","Name":"Zweiter privater Suchinhalt"}]}} |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Jar |
      | /de-DE/suche?q=phone | second |
    Dann der verwendete Applicationport erhält
      | SearchInterests |
      | [{c-phone 0.25}] |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:Cache-Control | private,no-store |
      | Body | Zweiter privater Suchinhalt |
      | Absent | Erster privater Suchinhalt |
      | Absent | Testmodell |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=phone |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 3 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Testmodell &lt;b&gt; |
      | Absent | Erster privater Suchinhalt |
      | Absent | Zweiter privater Suchinhalt |

  Szenario: Die echte Suchcachefrist wird durch kürzere Pflichtdaten begrenzt
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Search":{"Cards":[{"SKU":"RED-256","Name":"Kurzes Suchangebot","Compliance":{"Context":{"ValidUntil":{"Present":true,"Value":"2026-10-16T12:00:02Z"}}}}]}} |
    Und die Suchaufrufe ab jetzt gezählt werden
      | Fault |
      | |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=phone |
    Und die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=phone |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 1 |
    Angenommen die Suchaufrufe ab jetzt gezählt werden
      | Fault |
      | Search:InternalError |
    Wenn die Fragmentcachefrist verstreicht
      | Milliseconds |
      | 2500 |
    Und die zusammengesetzte HTTPansicht wird angefordert
      | URL |
      | /de-DE/suche?q=phone |
    Dann die Suchportaufrufe der Komposition entsprechen
      | SearchCalls |
      | 1 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 503 |
      | Header:Cache-Control | private,no-store |
      | Absent | Kurzes Suchangebot |
