# language: de
@spec-only @runtime
Funktionalität: Echte langlebige Cart- und Orderprozesse
  Als Gast behalte ich den Prozess über Worker-Neustarts hinweg.

  Szenariogrundriss: Eindeutige Workflowidentität und gebundene Konfiguration
    Angenommen ein fachlicher Zustand ist in PostgreSQL gespeichert
      | ResourceID | DefinitionVersion | RuleSetVersion |
      | <kennung> | <definition> | demo-v1 |
    Wenn der echte Temporaladapter den Prozess wiederholt sicherstellt
      | Operation | Attempts | TaskQueue |
      | <operation> | 2 | shop |
    Dann existiert genau die gebundene Prozessinstanz
      | WorkflowID | InstanceCount | DefinitionVersion | RuleSetVersion |
      | <workflow> | 1 | <definition> | demo-v1 |
    Beispiele:
      | kennung | definition | operation | workflow |
      | c-a | cart-v1 | EnsureCart | cart/c-a |
      | o-a | order-v1 | EnsureOrder | order/o-a |

  Szenariogrundriss: Ablaufzeit und Aktivitätswirkung
    Angenommen ein Cartprozess hat eine kontrollierte letzte Aktivität
      | CartID | LastActivityAt | State |
      | c-a | 2026-10-01T00:00:00Z | checkingOut |
    Wenn eine Prozessoperation ausgeführt wird
      | At | Operation | Successful | Duplicate |
      | <zeit> | <operation> | <erfolg> | <duplikat> |
    Dann lautet der fachliche Zustand
      | Code | State | LastActivityAt |
      | <code> | <zustand> | <aktivitaet> |
    Beispiele:
      | zeit | operation | erfolg | duplikat | code | zustand | aktivitaet |
      | 2026-10-07T23:59:59Z | read | true | false | OK | checkingOut | 2026-10-01T00:00:00Z |
      | 2026-10-08T00:00:00Z | place | false | false | Expired | expired | 2026-10-01T00:00:00Z |
      | 2026-10-08T00:00:01Z | timer | true | false | Expired | expired | 2026-10-01T00:00:00Z |
      | 2026-10-07T00:00:00Z | change | true | false | OK | checkingOut | 2026-10-07T00:00:00Z |
      | 2026-10-07T00:00:00Z | change | true | true | OK | checkingOut | 2026-10-01T00:00:00Z |
      | 2026-10-07T00:00:00Z | change | false | false | InvalidInput | checkingOut | 2026-10-01T00:00:00Z |

  Szenariogrundriss: Wiederanlauf und Fehler unterscheiden
    Angenommen ein Cartlauf bindet veröffentlichte Versionen
      | DefinitionVersion | RuleSetVersion | CurrentPublishedRules |
      | cart-v1 | demo-v1 | demo-v2 |
    Wenn eine Prozessstörung auftritt und der Worker neu startet
      | Fault | PriorCommit | Attempts |
      | <stoerung> | <commit> | <versuche> |
    Dann lautet die Prozessbeobachtung
      | BusinessSource | RuleSetVersion | OrderCount | TechnicalStatus | Retryable |
      | PostgreSQL | demo-v1 | <orders> | <status> | <retry> |
    Beispiele:
      | stoerung | commit | versuche | orders | status | retry |
      | workerRestartAfterCommit | true | 1 | 1 | running | false |
      | dependencyTimeout | false | 5 | 0 | needsAttention | true |
      | invalidQuantity | false | 1 | 0 | running | false |
      | replayAfterRulePublication | false | 1 | 0 | running | false |

  Szenario: Veralteter Timer überschreibt neue Aktivität nicht
    Angenommen ein früher Timer wurde eingeplant und spätere Aktivität gespeichert
      | OldDeadline | LastActivityAt | TimerAt |
      | 2026-10-08T00:00:00Z | 2026-10-07T00:00:00Z | 2026-10-08T00:00:00Z |
    Wenn die Ablaufactivity den fachlichen Zustand erneut sperrt und prüft
    Dann lautet die Ablaufentscheidung
      | State | NextDeadline |
      | open | 2026-10-14T00:00:00Z |

  Szenario: Änderung während simulierter Freigabe invalidiert alten Versuch
    Angenommen eine Kartenfreigabe bindet einen Checkoutversuch
      | CartRevision | Total | Authorization | CommandID |
      | 4 | 10490 | approved | cmd-a |
    Wenn vor dem Commit eine neue Cartrevision gespeichert wird
      | CartRevision | Total |
      | 5 | 20490 |
    Dann lautet das Prozessergebnis
      | Code | OrderCount | Reserved | AuthorizationRetained | RequiresConfirmation |
      | Conflict | 0 | 0 | false | true |

  Szenariogrundriss: Prozessannahme und Wiederholungsbudgets sind explizit
    Wenn eine echte Prozessoperation ausgeführt wird
      | Case | ElapsedMilliseconds |
      | <fall> | <dauer> |
    Dann lautet die technische Prozessantwort
      | Status | StartToCloseSeconds | ScheduleToCloseSeconds | MaximumAttempts | InitialRetrySeconds | MaximumRetrySeconds | LocationPresent |
      | <status> | 10 | 120 | 5 | 1 | 10 | <location> |
    Beispiele:
      | fall | dauer | status | location |
      | completedUpdate | 100 | 200 | false |
      | durablyAcceptedStillRunning | 2000 | 202 | true |

  Szenario: Neue Regeln gelten ausschließlich für neue Carts
    Angenommen ein Cart ist an einen alten Regelstand gebunden
      | ExistingCartID | BoundRules | PublishedRules |
      | c-old | demo-v1 | demo-v2 |
    Wenn ein weiterer Cart geöffnet und der alte Prozess replayt wird
      | NewCartID |
      | c-new |
    Dann lauten die gebundenen Versionen
      | CartID | RuleSetVersion |
      | c-old | demo-v1 |
      | c-new | demo-v2 |
