# language: de
@application @app
Funktionalität: Interessen aus der authentifizierten Gastidentität
  Fachliche Profilgewichte werden aus dem gespeicherten User gelesen.

  Hintergrund:
    Angenommen zwei getrennte Browsersessions wurden geöffnet
      | Browser | UserID | CartID | Profile |
      | a | u-a | c-a | mobile |
      | b | u-b | c-b | home |
    Und persistierte Interessen getrennt nach Gast vorliegen
      | UserID | CategoryID | Weight |
      | u-a | mobile | 1 |
      | u-b | home | 0.5 |

  Szenariogrundriss: Die Gastfähigkeit entscheidet über den gelesenen User
    Wenn Interessen über die Gastfähigkeit angefordert werden
      | Session | Fault |
      | <session> | <fehler> |
    Dann lautet die autorisierte Interessensicht
      | Code | Count | CategoryID | Weight | ReadUserID |
      | <code> | <anzahl> | <kategorie> | <gewicht> | <user> |
    Beispiele:
      | session | fehler | code | anzahl | kategorie | gewicht | user |
      | a | none | OK | 1 | mobile | 1 | u-a |
      | b | none | OK | 1 | home | 0.5 | u-b |
      | unknown | none | NotFound | 0 | | | |
      | a | interestRead | DependencyUnavailable | 0 | | | u-a |
