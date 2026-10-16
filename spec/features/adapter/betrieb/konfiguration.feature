# language: de
@adapter @composition
Funktionalität: Fehlergrenzen der Laufzeitkonfiguration
  Fehlerhafte Startwerte und nicht erreichbare Dienste liefern einen sichtbaren Fehler.

  Grundlage:
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE |

  Szenariogrundriss: Fehler werden vor einem scheinbar erfolgreichen Start sichtbar
    Wenn eine abweichende Laufzeitkonfiguration ausgeführt wird
      | Operation | Field | Value |
      | <Operation> | <Feld> | <Wert> |
    Dann lautet das Ergebnis der Laufzeitoperation
      | Code |
      | <Fehler> |

    Beispiele:
      | Operation | Feld | Wert | Fehler |
      | open | Schema | invalid! | InvalidInput |
      | open | DSN | empty | InvalidInput |
      | open | SearchEndpoint | ://invalid | InvalidInput |
      | open | SearchEndpoint | ftp://localhost | InvalidInput |
      | open | DSN | ://invalid | InternalError |
      | open | DSN | postgres://shop@127.0.0.1:1/postgres?connect_timeout=1&sslmode=disable | InternalError |
      | open | TemporalAddress | 127.0.0.1:1 | InternalError |
      | run | Schema | invalid! | InvalidInput |
      | run | Listen | diagnostic-listener | InternalError |
      | run | BackendListen | diagnostic-listener | InternalError |
      | run | TemporalNamespace | missing-runtime-namespace | InternalError |
      | healthcheck | Listen | invalid | InternalError |
      | healthcheck | Listen | 127.0.0.1:% | InternalError |
      | healthcheck | Listen | 127.0.0.1:1 | InternalError |
      | healthcheck | Listen | diagnostic-listener | DependencyUnavailable |
      | open | SeedFile | ../../spec/fixtures/runtime/missing.json | InternalError |
      | open | SeedFile | ../../go.mod | InternalError |
