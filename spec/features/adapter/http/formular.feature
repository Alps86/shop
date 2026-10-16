# language: de
@http
Funktionalität: Geschützte normale Formulare und HTMX
  Als Gast bediene ich dieselben Applicationfassaden über normale Formulare oder Fragmente.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Cartcommand erhält unveränderte autorisierte Eingabe
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | CSRFSource | Form | HTMX |
      | POST | /de-DE/warenkorb/aktion | test-session-a | /de-DE/warenkorb | cartID=cart-a&commandID=command-a&revision=2&sku=RED-256&quantity=3&action=<aktion> | <htmx> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Header:Vary | HX-Request |
    Und der verwendete Applicationport erhält
      | SKU | Operation | Revision | CommandID | CartID |
      | RED-256 | <aktion> | 2 | command-a | cart-a |
    Beispiele:
      | aktion | htmx | status |
      | add | false | 303 |
      | setQuantity | true | 200 |
      | remove | false | 303 |
      | address | false | 303 |
      | shipping | false | 303 |
      | payment | false | 303 |

  Szenariogrundriss: Neue Bewertung Bestätigung und Abschluss verwenden ihre Fassaden
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | CSRFSource | Form | HTMX |
      | POST | /de-DE/warenkorb/aktion | test-session-a | /de-DE/kasse | cartID=cart-a&commandID=command-a&revision=2&evaluationID=evaluation-a&action=<aktion>&return=checkout | <htmx> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Header:Location | <ziel> |
      | Header:HX-Redirect | <hxziel> |
    Beispiele:
      | aktion | htmx | status | ziel | hxziel |
      | evaluate | false | 303 | /de-DE/kasse | |
      | confirm | true | 200 | | |
      | place | false | 303 | /de-DE/bestellung/order-a | |
      | place | true | 200 | | /de-DE/bestellung/order-a |
      | start | false | 303 | /de-DE/kasse | |

  Szenario: Anonymer Gast beginnt ausdrücklich einen eigenen Einkauf
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | CSRFSource | Form |
      | POST | /de-DE/gast | /de-DE/warenkorb | commandID=start-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 303 |
      | Header:Set-Cookie | shop_session=test-session-a |
      | Header:Set-Cookie | HttpOnly |
      | Header:Set-Cookie | SameSite=Lax |

  Szenario: Profilwechsel bleibt POST und CSRFgebunden
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | CSRFSource | Form |
      | POST | /de-DE/profil | test-session-a | /de-DE/profil | profile=coffee |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 303 |
      | Header:Location | /de-DE/profil |

  Szenariogrundriss: Formularmissbrauch und unlesbare Zahlen scheitern
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | CSRFSource | Form |
      | POST | <url> | test-session-a | <quelle> | <formular> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
    Beispiele:
      | url | quelle | formular | status |
      | /de-DE/warenkorb/aktion | | action=remove | 403 |
      | /de-DE/warenkorb/aktion | /de-DE/warenkorb | revision=invalid | 422 |
      | /de-DE/warenkorb/aktion | /de-DE/warenkorb | revision=2&quantity=invalid | 422 |
      | /de-DE/profil | | profile=coffee | 403 |

  Szenario: Eine HTMXquittung behält den bereits geöffneten Gastcart
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | CSRFSource | Form | HTMX |
      | POST | /de-DE/warenkorb/aktion | test-session-a | /de-DE/warenkorb | cartID=cart-a&commandID=htmx-change&revision=2&sku=RED-256&quantity=3&action=setQuantity | true |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Header:HX-Push-Url | /de-DE/warenkorb |
      | Absent | Eigenen Einkauf starten |
      | Body | Menge ändern |

  Szenariogrundriss: Erfolgreiche private Commands markieren auch Redirectantworten
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Listener | Session | CSRFSource | Form | HTMX |
      | POST | <url> | <listener> | <session> | <source> | <form> | <htmx> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Header:Cache-Control | private,no-store |
      | Header:X-Robots-Tag | noindex,nofollow |
    Beispiele:
      | url | listener | session | source | form | htmx | status |
      | /backend/inventory | backend | | /backend | sku=RED-256&revision=3&commandID=redirect-inventory&price=12500 | false | 303 |
      | /backend/lieferschritt | backend | | /backend | orderID=order-a&shipmentID=shipment-a&commandID=redirect-delivery&sku=RED-256&quantity=1&status=shipped | false | 303 |
      | /de-DE/warenkorb/aktion | public | test-session-a | /de-DE/warenkorb | cartID=cart-a&commandID=redirect-cart&revision=2&sku=RED-256&quantity=1&action=add | false | 303 |
      | /de-DE/profil | public | test-session-a | /de-DE/profil | profile=coffee | false | 303 |
      | /de-DE/warenkorb/aktion | public | test-session-a | /de-DE/kasse | cartID=cart-a&commandID=redirect-order&revision=2&evaluationID=evaluation-a&action=place&return=checkout | false | 303 |
      | /de-DE/warenkorb/aktion | public | test-session-a | /de-DE/kasse | cartID=cart-a&commandID=redirect-order&revision=2&evaluationID=evaluation-a&action=place&return=checkout | true | 200 |
      | /de-DE/warenkorb/aktion | public | test-session-a | /de-DE/warenkorb | cartID=cart-a&commandID=fragment-cart&revision=2&sku=RED-256&quantity=1&action=add | true | 200 |
