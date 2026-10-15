# language: de
@application @app
Funktionalität: Gastfähigkeit und Interessenprofile
  Als Gast behalte ich meinen eigenen Einkauf über Besuche hinweg.

  Szenariogrundriss: Profilwahl teilt keine Gastidentität
    Angenommen zwei getrennte Browsersessions wurden geöffnet
      | Browser | UserID | CartID | Profile |
      | a | u-a | c-a | neutral |
      | b | u-b | c-b | neutral |
    Wenn das Profil in einem Browser bestätigt wird
      | Browser | Profile |
      | a | <profil> |
    Dann lautet die Gastansicht
      | Browser | UserID | CartID | Profile |
      | a | u-a | c-a | <profil> |
      | b | u-b | c-b | neutral |
    Beispiele:
      | profil |
      | neutral |
      | mobile |
      | home |
      | coffee |

  @spec-only
  Szenariogrundriss: Besitz und zustandsändernde Requests
    Angenommen eine Ressource gehört einem anderen Gast
      | Owner | CartID | OrderID |
      | u-a | c-a | o-a |
    Wenn ein Gastrequest ausgeführt wird
      | SessionUser | Resource | Method | CSRF |
      | <gast> | <ressource> | <methode> | <csrf> |
    Dann lautet die HTTPantwort ohne Seiteneffekt
      | Status | CacheControl | MutationCount |
      | <status> | no-store | 0 |
    Beispiele:
      | gast | ressource | methode | csrf | status |
      | u-b | cart:c-a | GET | absent | 404 |
      | u-b | order:o-a | GET | absent | 404 |
      | u-b | order:missing | GET | absent | 404 |
      | u-a | profile:mobile | GET-confirmation | absent | 200 |
      | u-a | cart:c-a:add | GET | absent | 405 |
      | u-a | cart:c-a:add | POST | invalid | 403 |

  Szenariogrundriss: Wiederaufnahme und getrennte Fristen
    Angenommen eine Gastsession mit aktivem Cart besteht
      | IssuedAt | LastActivityAt |
      | 2026-10-01T00:00:00Z | 2026-10-02T00:00:00Z |
    Wenn der Einkauf erneut gelesen wird
      | At | Cookie |
      | <zeit> | <cookie> |
    Dann lautet die Wiederaufnahme
      | Code | ExistingUser | CartState | LastActivityAt |
      | <code> | <gast> | <zustand> | 2026-10-02T00:00:00Z |
    Beispiele:
      | zeit | cookie | code | gast | zustand |
      | 2026-10-08T23:59:59Z | valid | OK | true | open |
      | 2026-10-09T00:00:00Z | valid | Expired | true | expired |
      | 2026-10-09T00:00:01Z | valid | Expired | true | expired |
      | 2026-10-31T00:00:00Z | expired | NotFound | false | inaccessible |
      | 2026-10-03T00:00:00Z | absent | NotFound | false | inaccessible |

  @spec-only
  Szenario: Cookiegrenzen im echten HTTPadapter
    Wenn ein neuer Gast über HTTP gestartet wird
      | Operation |
      | GuestService.Open |
    Dann besitzt das lokale Democookie die vereinbarten Sicherheitswerte
      | CookieName | HttpOnly | SameSite | Path | TokenBits | StoredToken | MaxAgeDays |
      | shop_session | true | Lax | / | 256 | hash-only | 30 |
