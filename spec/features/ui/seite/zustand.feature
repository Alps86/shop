# language: de
@http
Funktionalität: Sichtbare Zustände statt stiller Erfolge
  Als Gast erkenne ich portseitige Fehler und behalte einen passenden nächsten Schritt.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Fehlgeschlagener Cartcommand liefert persönliche Korrekturansicht
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | <fehler> |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | CSRFSource | Form |
      | POST | /de-DE/warenkorb/aktion | test-session-a | /de-DE/warenkorb | cartID=cart-a&commandID=command-a&revision=2&quantity=0&action=<aktion>&return=<ziel> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Body | <hinweis> |
      | Header:Cache-Control | private,no-store |
    Beispiele:
      | fehler | aktion | ziel | status | hinweis |
      | Change:InvalidInput | setQuantity | cart | 422 | Eingabe prüfen |
      | Change:Conflict | address | checkout | 409 | Daten geändert |
      | Change:Pending | shipping | checkout | 202 | Status wird aktualisiert |
      | Confirm:Conflict | confirm | checkout | 409 | Daten geändert |
      | Evaluate:InternalError | evaluate | cart | 503 | Daten vorübergehend nicht verfügbar |
      | StartCart:InternalError | start | cart | 503 | Daten vorübergehend nicht verfügbar |
      | Place:PaymentDeclined | place | checkout | 422 | Simulierte Freigabe abgelehnt |

  Szenariogrundriss: Gaststart und Profilfehler sind keine Erfolgsmeldung
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | <fehler> |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | CSRFSource | Form |
      | POST | <url> | <session> | <quelle> | <formular> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 503 |
    Beispiele:
      | fehler | url | session | quelle | formular |
      | Open | /de-DE/gast | | /de-DE/warenkorb | commandID=start-a |
      | StartCart | /de-DE/gast | | /de-DE/warenkorb | commandID=start-a |
      | SelectProfile | /de-DE/profil | test-session-a | /de-DE/profil | profile=coffee |

  Szenariogrundriss: Fremde Gastfähigkeit öffnet keine persönlichen Daten
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | <url> | foreign |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 404 |
      | Absent | Unveränderlicher Kaufname |
    Beispiele:
      | url |
      | /de-DE/warenkorb |
      | /de-DE/profil |
      | /de-DE/bestellung/order-a |
      | /de-DE/warenkorb/badge |

  Szenario: Fehlende Gastidentität führt beim Checkout zum eigenen Cart
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/kasse |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 303 |
      | Header:Location | /de-DE/warenkorb |

  Szenario: Pending ist eine angenommene Verarbeitung ohne erfundene Order
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Purchase":{"Pending":true,"OrderID":"","ProcessReference":"cart/cart-a"}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | CSRFSource | Form |
      | POST | /de-DE/warenkorb/aktion | test-session-a | /de-DE/kasse | action=place&cartID=cart-a&revision=2&commandID=place-a&evaluationID=evaluation-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 202 |
      | Body | Bestellung angenommen, Status wird aktualisiert |
      | Header:Location | /de-DE/kasse |
      | Header:Retry-After | 2 |
      | Absent | /de-DE/bestellung/ |
