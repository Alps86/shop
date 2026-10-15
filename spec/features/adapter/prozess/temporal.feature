# language: de
@temporal
Funktionalität: Kontrollierte Workflowzeit und gesperrte Ablaufprüfung
  Als Gast behalte ich neue Aktivität trotz eines alten Workflowtimers.

  Szenariogrundriss: Der SDK-Zeitlauf führt die echte Ablaufactivity erneut aus
    Angenommen ein kontrollierter SDK-Zeitlauf vorbereitet ist
      | State | NewActivityAfterHours |
      | <StartState> | <Activity> |
    Wenn der Zeitlauf ohne reale Wartezeit ausgeführt wird
    Dann die Ablaufactivity den letzten Datenbankstand respektiert
      | State | ElapsedHours | CheckpointState | NextDeadlineHours |
      | <EndState> | <Elapsed> | <CheckpointState> | <NextDeadlineHours> |
    Beispiele:
      | StartState | Activity | EndState | Elapsed | CheckpointState | NextDeadlineHours |
      | open | 0 | expired | 168 | | |
      | checkingOut | 144 | expired | 312 | checkingOut | 312 |
      | open | 144 | expired | 312 | open | 312 |
      | ordered | 0 | ordered | 0 | | |

  Szenariogrundriss: Kartenfreigabe ist explizit simuliert und an den Versuch gebunden
    Wenn eine ausschließlich simulierte Kartenfreigabe ausgeführt wird
      | Approved | Delay | Deadline | CartID | CartRevision | Total | CommandID |
      | <Approved> | <Delay> | <Deadline> | c-a | 4 | 10490 | cmd-a |
    Dann bindet das Simulationsergebnis ausschließlich den Kaufversuch
      | Code | Approved | CartID | CartRevision | Total | CommandID |
      | <Code> | <Result> | <CartID> | <Revision> | <Total> | <CommandID> |
    Beispiele:
      | Approved | Delay | Deadline | Code | Result | CartID | Revision | Total | CommandID |
      | true | 1ms | 1s | OK | true | c-a | 4 | 10490 | cmd-a |
      | false | 1ms | 1s | OK | false | c-a | 4 | 10490 | cmd-a |
      | true | 1s | 1ms | DeadlineExceeded | false | | 0 | 0 | |

  Szenariogrundriss: SDK-Wiederholungen unterscheiden technische und fachliche Activityfehler
    Angenommen ein kontrollierter SDK-Zeitlauf vorbereitet ist
      | State | NewActivityAfterHours |
      | open | 0 |
    Und die Ablaufabhängigkeit kontrolliert scheitert
      | Code | InitialFailures | ObserveAfter |
      | <Code> | 5 | 2m |
    Wenn der Zeitlauf ohne reale Wartezeit ausgeführt wird
    Dann das ausgeschöpfte Budget als technische Diagnose sichtbar bleibt
      | Attempts | State |
      | <Attempts> | needsAttention |
    Beispiele:
      | Code | Attempts |
      | DependencyUnavailable | 5 |
      | DeadlineExceeded | 5 |
      | NotFound | 1 |

  Szenariogrundriss: SDK-Registrierungsfehler starten keine fachliche Activity
    Angenommen ein kontrollierter SDK-Zeitlauf vorbereitet ist
      | State | NewActivityAfterHours |
      | open | 0 |
    Wenn das SDK eine Handlerregistrierung ablehnt
      | Workflow | Registration |
      | <Workflow> | <Registration> |
    Dann der Workflow ohne Activitywirkung fehlschlägt
      | Failed | ActivityCalls |
      | true | 0 |
    Beispiele:
      | Workflow | Registration |
      | cart | update |
      | cart | query |
      | order | query |

  Szenario: Lieferabhängigkeitsfehler erzeugt keine falsche Bestätigung
    Angenommen ein kontrollierter SDK-Zeitlauf vorbereitet ist
      | State | NewActivityAfterHours |
      | open | 0 |
    Wenn eine Lieferabhängigkeit im kontrollierten SDK-Lauf ausfällt
      | Code |
      | DependencyUnavailable |
    Dann das ausgeschöpfte Budget als technische Diagnose sichtbar bleibt
      | State |
      | needsAttention |

  Szenario: Expliziter erfolgreicher Auftrag setzt einen reparierten Ablauf fort
    Angenommen ein kontrollierter SDK-Zeitlauf vorbereitet ist
      | State | NewActivityAfterHours |
      | open | 0 |
    Und ein erfolgreicher Auftrag nach Reparatur den Prozess ausdrücklich fortsetzt
      | Code | InitialFailures | RetryAfter |
      | DependencyUnavailable | 5 | 90s |
    Wenn der Zeitlauf ohne reale Wartezeit ausgeführt wird
    Dann die Ablaufactivity den letzten Datenbankstand respektiert
      | State | ElapsedHours | ActivityCalls |
      | expired | 168 | 7 |
