# language: de
@http
Funktionalität: Geschützte Gastcookies und Profilbestätigung
  Als Gast teile ich über öffentliche Profile keine Identität oder Zugriffsrechte.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Cookieattribute entsprechen dem expliziten lokalen oder HTTPSbetrieb
    Angenommen die HTTPkonfiguration verwendet
      | SecureCookie |
      | <secure> |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | CSRFSource | Form |
      | POST | /de-DE/gast | /de-DE/warenkorb | commandID=start-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 303 |
      | Header:Set-Cookie | shop_session=test-session-a |
      | Header:Set-Cookie | Path=/ |
      | Header:Set-Cookie | Max-Age=2592000 |
      | Header:Set-Cookie | HttpOnly |
      | Header:Set-Cookie | SameSite=Lax |
      | <securefeld> | Secure |
    Beispiele:
      | secure | securefeld |
      | false | AbsentHeader:Set-Cookie |
      | true | Header:Set-Cookie |

  Szenario: CSRFtoken eines anderen Gasts autorisiert keinen Profilcommand
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/profil | test-session-a |
    Und der Shop liefert eine HTTPansicht
      | Method | URL | Session | CSRF | Form |
      | POST | /de-DE/profil | test-session-b | previous | profile=coffee |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 403 |
    Und der verwendete Applicationport erhält
      | Calls |
      | Open |

  Szenariogrundriss: Öffentliche Profillinks zeigen nur eine Bestätigung
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/profil?profile=<profil> | test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | value="<profil>" selected |
      | Header:Cache-Control | private,no-store |
    Und der verwendete Applicationport erhält
      | Calls |
      | Open |
    Beispiele:
      | profil |
      | neutral |
      | mobile |
      | home |
      | coffee |

  Szenario: Gaststart erhält eine ausdrückliche spätere Profilbestätigung
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | CSRFSource | Form |
      | POST | /de-DE/gast | /de-DE/profil?profile=coffee | commandID=start-a&return=profile&profile=coffee |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 303 |
      | Header:Location | /de-DE/profil?profile=coffee |

  Szenario: Gaststart erhält die aktiv ausgewählte Produktvariante
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | CSRFSource | Form |
      | POST | /de-DE/gast | /de-DE/warenkorb?add=RED-256 | commandID=start-a&add=RED-256 |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 303 |
      | Header:Location | /de-DE/warenkorb?add=RED-256 |

  Szenariogrundriss: Ungültige Profilanfragen verändern keine Interessen
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | <methode> | <url> | test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
    Beispiele:
      | methode | url | status |
      | GET | /de-DE/profil?profile=unknown | 422 |
      | PUT | /de-DE/profil | 405 |

  Szenario: Gespeicherte Interessen werden nur als Querydaten weitergegeben
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Interests":[{"CategoryID":"c-phone","Weight":"0.75"}]} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/kategorie/smartphones--c-phone | test-session-a |
    Dann der verwendete Applicationport erhält
      | SearchInterests |
      | [{c-phone 0.75}] |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/produkt/testmodell--p-phone | test-session-a |
    Dann der verwendete Applicationport erhält
      | RecommendationInterests |
      | [{c-phone 0.75}] |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Absent | 0.75 |
      | Header:Cache-Control | private,no-store |
