# language: de
@temporal
Funktionalität: Dauerhafte Temporalidentität
  Als Betreiber starte ich denselben Warenkorbprozess idempotent.

  Szenario: Wiederholter Start erhält eine Workflowausführung
    Angenommen ein echter Temporalclient ist verbunden
    Wenn der Cartprozess wird wiederholt gestartet
      | DefinitionVersion | RuleSetVersion | Attempts |
      | cart-v1 | demo-v1 | 2 |
    Dann genau dieselbe Workflowausführung besteht
      | SameRun | TaskQueue |
      | true | shop |

  Szenariogrundriss: Updates verwenden nach Workerrestart dieselbe History und prüfen Besitz
    Angenommen ein echter Temporalclient ist verbunden
    Und ein Worker den gespeicherten Cart übernimmt
      | DefinitionVersion | RuleSetVersion | Fixture |
      | cart-v1 | demo-v1 | spec/fixtures/runtime/warenkorb.json |
    Wenn der Cartprozess wird wiederholt gestartet
      | DefinitionVersion | RuleSetVersion | Attempts |
      | cart-v1 | demo-v1 | 2 |
    Und der Worker vollständig neu startet und seine History wieder bereitsteht
      | ReadinessTimeout |
      | 10s |
    Und ein echtes Workflowupdate gesendet wird
      | UserID | CommandID | Operation |
      | <UserID> | update-a | <Operation> |
    Dann antwortet der weitergeführte Prozess
      | Code | State | RuleSetVersion |
      | <Code> | running | demo-v1 |
    Und die echte History die begrenzten Activitybudgets enthält
      | StartToCloseSeconds | ScheduleToCloseSeconds | MaximumAttempts | InitialRetrySeconds | MaximumRetrySeconds |
      | 10 | 120 | 5 | 1 | 10 |
    Beispiele:
      | UserID | Operation | Code |
      | u-a | shipping | OK |
      | u-foreign | shipping | NotFound |
      | u-a | add | InvalidInput |

  Szenariogrundriss: Jede Workflowoperation quittiert das Applicationergebnis
    Angenommen ein echter Temporalclient ist verbunden
    Und ein Worker den gespeicherten Cart übernimmt
      | DefinitionVersion | RuleSetVersion | Fixture |
      | cart-v1 | demo-v1 | spec/fixtures/runtime/warenkorb.json |
    Wenn der Cartprozess wird wiederholt gestartet
      | DefinitionVersion | RuleSetVersion | Attempts |
      | cart-v1 | demo-v1 | 2 |
    Und eine typisierte Temporaloperation verarbeitet wird
      | Operation | UserID |
      | <Operation> | <UserID> |
    Dann antwortet der weitergeführte Prozess
      | Code | State | RuleSetVersion |
      | <Code> | running | demo-v1 |
    Beispiele:
      | Operation | UserID | Code |
      | evaluate | u-a | OK |
      | evaluate | u-foreign | NotFound |
      | confirm | u-a | PriceConfirmationRequired |
      | confirm | u-foreign | NotFound |
      | place | u-a | PriceConfirmationRequired |
      | place | u-foreign | NotFound |

  Szenariogrundriss: Nicht angenommene Operationen behaupten keinen Prozessfortschritt
    Angenommen ein echter Temporalclient ist verbunden
    Wenn eine echte Temporalgrenze mit begrenztem Kontext aufgerufen wird
      | Operation | Timeout | DefinitionVersion | RuleSetVersion | UserID | OuterCommandID | InnerCommandID |
      | <Operation> | <Timeout> | cart-v1 | demo-v1 | u-a | outer-a | inner-a |
    Dann die technische Grenze einen ehrlichen Fehler meldet
      | Code |
      | <Code> |
    Beispiele:
      | Operation | Timeout | Code |
      | ensure | -1s | DeadlineExceeded |
      | update | -1s | DeadlineExceeded |
      | update | 1s | DependencyUnavailable |
      | report | 1s | InvalidInput |
      | closed | 1s | DependencyUnavailable |

  Szenario: Angenommene Cartänderung liefert Pending mit Prozessreferenz
    Angenommen ein echter Temporalclient ist verbunden
    Und ein Worker den gespeicherten Cart übernimmt
      | DefinitionVersion | RuleSetVersion | Fixture |
      | cart-v1 | demo-v1 | spec/fixtures/runtime/warenkorb.json |
    Und die Application-Schreibantwort kontrolliert verzögert wird
      | Delay |
      | 3s |
    Wenn der Cartprozess wird wiederholt gestartet
      | DefinitionVersion | RuleSetVersion | Attempts |
      | cart-v1 | demo-v1 | 2 |
    Und ein echtes Workflowupdate gesendet wird
      | UserID | CommandID | Operation |
      | u-a | pending-a | shipping |
    Dann antwortet der weitergeführte Prozess
      | Code | State | RuleSetVersion |
      | Pending | running | demo-v1 |
    Und der angenommene Cartbefehl später abgeschlossen wird
      | Revision |
      | 2 |

  Szenario: Ein fremder Besitzer darf keinen fehlenden Cartworkflow starten
    Angenommen ein echter Temporalclient ist verbunden
    Und ein Worker den gespeicherten Cart übernimmt
      | DefinitionVersion | RuleSetVersion | Fixture |
      | cart-v1 | demo-v1 | spec/fixtures/runtime/warenkorb.json |
    Wenn eine typisierte Temporaloperation verarbeitet wird
      | Operation | UserID |
      | evaluate | u-foreign |
    Dann kein neuer Cartworkflow beobachtet wird
      | Code | Exists |
      | NotFound | false |
