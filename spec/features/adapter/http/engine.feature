# language: de
@http @composition
Funktionalität: Echte Standardenginekomposition
  Als Gast erhalte ich von Hugo NginxSSI und VarnishESI zusammengesetztes HTML.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |
    Und die Standardengines werden mit Hugo gestartet
      | Nginx | Varnish | CacheKey |
      | deploy/nginx/nginx.conf | deploy/varnish/default.vcl | synthetic-composition-test-key-32-bytes |

  Szenariogrundriss: Hugo und Go werden tatsächlich über SSI und ESI zusammengesetzt
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | <url> | <cookie> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | <inhalt> |
      | Absent | <!--# |
      | Absent | <esi:include |
      | Body | data-composition="nginx-ssi" |
    Beispiele:
      | url | cookie | inhalt |
      | /de-DE/ | | Technik, die zu deinem Alltag passt |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | | data-sku="RED-256" |
      | /de-DE/service | | Technik verstehen. In Ruhe entscheiden. |
      | /de-DE/kategorie/smartphones--c-phone | | data-frame="hugo" |

  Szenario: Persönlicher Cart umgeht den gemeinsamen Fragmentcache
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | /de-DE/warenkorb | shop_session=test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:Cache-Control | private,no-store |
      | Body | RED-256 |
      | Absent | <esi:include |

  Szenario: Bereits vor Parentheadern erkannter Pflichtdatenfehler ist HTTP503
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | Product |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 503 |
      | Header:Retry-After | 2 |
      | Absent | In den Warenkorb |

  Szenario: Erst beim ESIabruf erkannter Portfehler zeigt sicheren200Fallback
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | Product#2:InternalError |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Daten vorübergehend nicht verfügbar |
      | Absent | In den Warenkorb |
      | Absent | 120,00 EUR |

  Szenario: Abgebrochener Subrequest belässt den benachbarten Ladehinweis
    Angenommen die HTTPkonfiguration verwendet
      | FragmentFault |
      | connectionClose |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Aktuelle Produktdaten sind nur bei vollständig geladenem Inhalt verfügbar |
      | Absent | In den Warenkorb |
      | Absent | 120,00 EUR |

  Szenario: Persönlicher Zustand und direkte Fragmentrouten sind am Einstieg gesperrt
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | /fragments/de-DE/produkt/testmodell--p-phone?sku=RED-256 | |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 404 |

  Szenario: Authentifizierter SKUban ersetzt ein wirklich gecachtes Preisfragment
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Body | 120,00 EUR |
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Detail":{"Cards":[{"SKU":"RED-256","UnitPrice":{"MinorUnits":13000,"Currency":"EUR"},"AvailableQuantity":7,"Compliance":{"Context":{"VerificationStatus":"verified"}}}]}} |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Body | 120,00 EUR |
    Wenn der komponierte Cache invalidiert wird
      | Key | SKUs |
      | synthetic-composition-test-key-32-bytes | RED-256 |
    Und die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | 130,00 EUR |
      | Absent | 120,00 EUR |

  Szenario: Standardenginelogs enthalten weder Gastcookie noch Startfähigkeit
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | /de-DE/warenkorb | shop_session=test-session-a |
    Und die zusammengesetzte HTTPansicht wird angefordert
      | URL | Cookie |
      | /de-DE/warenkorb | |
    Dann die Standardenginelogs enthalten keine Fähigkeiten
      | Forbidden |
      | test-session-a |
      | shop_session |
      | shop_start |
      | Set-Cookie |

  Szenario: Zwei echte Cookiejars teilen denselben neutralen Rahmen und getrennte private Badges
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Guests":{"session-first":{"UserID":"first","Session":"session-first","CartID":{"Present":true,"Value":"cart-first"}},"session-second":{"UserID":"second","Session":"session-second","CartID":{"Present":true,"Value":"cart-second"}}},"Carts":{"cart-first":{"Cart":{"ID":"cart-first","UserID":"first","Items":[{"SKU":"RED-256","Quantity":2}]}},"cart-second":{"Cart":{"ID":"cart-second","UserID":"second","Items":[{"SKU":"BLUE-128","Quantity":7}]}}}} |
    Und getrennte Gastcookiejars werden eingerichtet
      | Jar | Session |
      | first | session-first |
      | second | session-second |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Jar |
      | /de-DE/ | first |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | data-composition="nginx-ssi" |
      | Absent | Warenkorb (2) |
      | Absent | Warenkorb (7) |
    Und die öffentliche Rahmenantwort wird verglichen
      | Mode |
      | remember |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Jar |
      | /de-DE/ | second |
    Dann die öffentliche Rahmenantwort wird verglichen
      | Mode |
      | identical |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Jar |
      | /de-DE/warenkorb/badge | first |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Warenkorb (2) |
      | Absent | Warenkorb (7) |
      | Header:Cache-Control | private,no-store |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Jar |
      | /de-DE/warenkorb/badge | second |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Warenkorb (7) |
      | Absent | Warenkorb (2) |
      | Header:Cache-Control | private,no-store |
    Wenn die zusammengesetzte HTTPansicht wird angefordert
      | URL | Jar |
      | /de-DE/warenkorb/badge | first |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Body | Warenkorb (2) |
      | Header:Cache-Control | private,no-store |
