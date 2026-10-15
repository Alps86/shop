# language: de
@spec-only @runtime
Funktionalität: Atomare Outbox und logisches CDC
  Als Betreiber projiziere ich nur bestätigte Änderungen und verliere keine neuere Version.

  Szenariogrundriss: Fachzustand und Outbox teilen den Commit
    Wenn eine Inventoryänderung in PostgreSQL ausgeführt wird
      | SKU | NewPrice | TransactionOutcome |
      | PHONE | 12000 | <ausgang> |
    Dann lautet die beobachtbare Quellen- und Projektionswirkung
      | Price | OutboxEvents | CDCProtocol | PublishedTable | SearchPrice |
      | <preis> | <events> | pgoutput | outbox | <suchpreis> |
    Beispiele:
      | ausgang | preis | events | suchpreis |
      | commit | 12000 | 1 | 12000 |
      | rollback | 10000 | 0 | 10000 |

  Szenariogrundriss: Versionsschutz und Tombstones
    Angenommen die Suchprojektion enthält einen bekannten Versionsstand
      | ProductID | ProjectionVersion | Kind | Price |
      | p-a | 5 | <bestand> | 10000 |
    Wenn ein vollständiges Projektionsereignis angewendet wird
      | EventID | ProductID | ProjectionVersion | Kind | Price | SchemaVersion |
      | e-a | p-a | <version> | <art> | 12000 | 1 |
    Dann lautet die öffentliche Projektion
      | Result | Visible | AppliedVersion | Price |
      | <ergebnis> | <sichtbar> | <angewandt> | <preis> |
    Beispiele:
      | bestand | version | art | ergebnis | sichtbar | angewandt | preis |
      | upsert | 6 | upsert | Applied | true | 6 | 12000 |
      | upsert | 4 | upsert | Stale | true | 5 | 10000 |
      | upsert | 5 | upsert | Duplicate | true | 5 | 10000 |
      | upsert | 6 | delete | Applied | false | 6 | absent |
      | delete | 4 | upsert | Stale | false | 5 | absent |

  Szenariogrundriss: Quittierung und Reindex
    Wenn eine Projektionsstörung an einer kontrollierten Grenze auftritt
      | Case | SearchAcknowledged | CatchupComplete |
      | <fall> | <bestaetigt> | <aufgeholt> |
    Dann lautet die Adapterbeobachtung
      | AdvanceLSN | Retry | Alias | Diagnostic |
      | <lsn> | <retry> | <alias> | <diagnose> |
    Beispiele:
      | fall | bestaetigt | aufgeholt | lsn | retry | alias | diagnose |
      | crashBeforeAcknowledgement | false | false | false | true | old | backlog |
      | acknowledged | true | false | true | false | old | healthy |
      | rebuildCaughtUp | true | true | true | false | new | healthy |
      | rebuildFailed | false | false | false | true | old | rebuildFailed |
      | slotLost | false | false | false | false | old | reindexRequired |
      | invalidSchema | false | false | false | false | old | invalidSchema |

  Szenario: Persistierter Prozessauftrag heilt die Temporalstartlücke
    Angenommen ein Kauf ist committed und der Worker vor dem Prozessstart abgestürzt
      | OrderID | OutboxCommand | Acknowledged |
      | o-a | EnsureOrder | false |
    Wenn der Dispatcher nach Neustart denselben Auftrag erneut zustellt
      | Attempts |
      | 2 |
    Dann lautet die Prozessbeobachtung
      | WorkflowID | InstanceCount | Acknowledged | SearchCheckpointChanged |
      | order/o-a | 1 | true | false |
