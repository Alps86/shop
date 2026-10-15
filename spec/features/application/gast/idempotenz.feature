# language: de
@application @app
Funktionalität: Dauerhafte Fachbefehle nach Activity-Neustart
  Die Application liest gespeicherte Befehlsbelege vor der fachlichen Revisionsprüfung.

  Hintergrund:
    Angenommen ein eigener Gastcart ist vollständig vorbereitet
      | UserID | CartID | SKU | Quantity | UnitPrice | StockQuantity | CartRevision | EvaluationID | Recipient | Street | PostalCode | City | Country |
      | u-a | c-a | PHONE | 1 | 10000 | 5 | 4 | e-a | Demo Gast | Testweg 1 | 12345 | Teststadt | DE |

  Szenariogrundriss: Ein abgeschlossener Mengenbefehl wird exakt wiederholt
    Wenn ein bereits gespeicherter Fachbefehl erneut als Activity ausgeführt wird
      | UserID | CartID | CommandID | SKU | RetryQuantity |
      | u-a | c-a | change-a | PHONE | <menge> |
    Dann lautet das wiederholte Activityergebnis
      | Code | Quantity |
      | <code> | 2 |
    Beispiele:
      | menge | code |
      | 2 | OK |
      | 3 | Conflict |

  Szenariogrundriss: Bestätigungen lesen ihren gespeicherten Beleg
    Wenn eine gespeicherte Bestätigung erneut als Activity ausgeführt wird
      | UserID | CartID | CommandID | EvaluationID | RetryEvaluationID | Fault |
      | u-a | c-a | confirm-a | e-a | <evaluation> | <fehler> |
    Dann lautet das wiederholte Activityergebnis
      | Code | Quantity |
      | <code> | 1 |
    Beispiele:
      | evaluation | fehler | code |
      | e-a | none | OK |
      | e-b | none | Conflict |
      | e-a | commandRead | DependencyUnavailable |
