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
