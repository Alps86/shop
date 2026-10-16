# language: de
@http @functional-budget
Funktionalität: Begrenztes Fragmentbudget für die funktionale Demo
  Vollständige Daten dürfen innerhalb der technischen Frist ankommen.
  Der technische Abbruchvertrag ersetzt kein Leistungsziel für vollständige Seiten.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Direkte Fragmentantwort trennt verzögerte Daten von sicherem Ersatz
    Angenommen die HTTPkonfiguration verwendet
      | SearchDelayMilliseconds |
      | <Delay> |
    Wenn die Fragmentansicht mit ihrer Laufzeit angefordert wird
      | Method | URL |
      | GET | /fragments/de-DE/suche?q=phone |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | <Body> |
      | Absent | <Absent> |
    Und die gemessene Antwort ihre technische Abbruchgrenze einhält
      | MaximumMilliseconds |
      | 1000 |
    Beispiele:
      | Delay | Body | Absent |
      | 350 | RED-256 | Daten vorübergehend nicht verfügbar |
      | 900 | Daten vorübergehend nicht verfügbar | RED-256 |

  Szenariogrundriss: Varnish wartet innerhalb der technischen Grenze auf das Kategoriefragment
    Angenommen die Standardengines werden mit Hugo gestartet
      | Nginx | Varnish | CacheKey |
      | deploy/nginx/nginx.conf | deploy/varnish/default.vcl | synthetic-composition-test-key-32-bytes |
    Und die HTTPkonfiguration verwendet
      | SearchDelayMilliseconds |
      | <Delay> |
    Wenn die zusammengesetzte Ansicht mit ihrer Laufzeit angefordert wird
      | URL | Cookie |
      | /de-DE/kategorie/smartphones--c-phone | |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | <Body> |
      | Body | data-composition="nginx-ssi" |
      | Absent | <Absent> |
      | Absent | <esi:include |
      | Absent | <!--# |

    Und die gemessene Antwort ihre technische Abbruchgrenze einhält
      | MaximumMilliseconds |
      | 1500 |
    Beispiele:
      | Delay | Body | Absent |
      | 350 | RED-256 | Daten vorübergehend nicht verfügbar |
      | 900 | Daten vorübergehend nicht verfügbar | RED-256 |
