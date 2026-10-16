# language: de
@application @app
Funktionalität: Interessen aus der authentifizierten Gastidentität
  Fachliche Profilgewichte werden aus dem gespeicherten User gelesen.

  Hintergrund:
    Angenommen zwei getrennte Browsersessions wurden geöffnet
      | Browser | UserID | CartID | Profile |
      | a | u-a | c-a | mobile |
      | b | u-b | c-b | home |
      | c | u-c | c-c | neutral |
    Und persistierte Interessen getrennt nach Gast vorliegen
      | UserID | CategoryID | Weight |
      | u-a | mobile | 1 |
      | u-b | home | 0.5 |

  Szenariogrundriss: Die Gastfähigkeit entscheidet über den gelesenen User
    Wenn Interessen über die Gastfähigkeit angefordert werden
      | Session | Fault |
      | <session> | <fehler> |
    Dann lautet die autorisierte Interessensicht
      | Code | Count | CategoryID | Weight | ReadUserID | ResolveCalls | ReadCalls |
      | <code> | <anzahl> | <kategorie> | <gewicht> | <user> | 0 | 1 |
    Beispiele:
      | session | fehler | code | anzahl | kategorie | gewicht | user |
      | a | none | OK | 1 | mobile | 1 | u-a |
      | b | none | OK | 1 | home | 0.5 | u-b |
      | unknown | none | NotFound | 0 | | | |
      | a | interestRead | DependencyUnavailable | 0 | | | u-a |

  Szenariogrundriss: Ablauf wird vor dem Payloadfehler geprüft
    Angenommen die Interessenfähigkeit folgendes Alter hat
      | Session | Age |
      | a | <alter> |
    Wenn Interessen über die Gastfähigkeit angefordert werden
      | Session | Fault |
      | a | <fehler> |
    Dann lautet die autorisierte Interessensicht
      | Code | Field | Count | CategoryID | Weight | ResolveCalls | ReadCalls |
      | <code> | <feld> | <anzahl> | <kategorie> | <gewicht> | 0 | 1 |
    Beispiele:
      | alter | fehler | code | feld | anzahl | kategorie | gewicht |
      | 719h59m59.999999999s | none | OK | | 1 | mobile | 1 |
      | 720h | none | NotFound | Session | 0 | | |
      | 720h0m0.000000001s | none | NotFound | Session | 0 | | |
      | 719h59m59s | payload | DependencyUnavailable | PostgreSQL | 0 | | |
      | 720h | payload | NotFound | Session | 0 | | |
      | 721h | payload | NotFound | Session | 0 | | |

  Szenario: Das neutrale Profil besitzt keine Interessen
    Wenn Interessen über die Gastfähigkeit angefordert werden
      | Session | Fault |
      | c | none |
    Dann lautet die autorisierte Interessensicht
      | Code | Count | ResolveCalls | ReadCalls |
      | OK | 0 | 0 | 1 |
