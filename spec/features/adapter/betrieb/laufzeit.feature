# language: de
@adapter @composition
Funktionalität: Laufzeitkomposition der echten Adapter
  Die Laufzeit bindet echte Dienste und liefert gemessene Zustände über HTTP.

  Szenario: Die Datenbankdiagnose stammt aus der tatsächlich geöffneten PostgreSQL-Verbindung
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE |
    Und die Laufzeitdiagnose per HTTP abgefragt wird
      | Path |
      | /health |
    Dann liefert die Laufzeitdiagnose
      | Status | Contains |
      | 200 | ready |

  Szenario: Die vollständige Laufzeit startet Worker und WAL-Projektion neben dem HTTP-Dienst
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE |
    Und Worker Projektor und HTTP-Dienst gestartet werden
    Und die vollständige Bereitschaft begrenzt abgewartet wird
      | Path | TimeoutSeconds | Contains |
      | /ready | 10 | CDC:initialized |
    Dann liefert die Laufzeitdiagnose
      | Status | Contains |
      | 200 | CDC:initialized |

  Szenario: Die echte Shopdatei startet Dienste prüft Bereitschaft und stoppt geordnet
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE |
    Und die ausführbare Shopdatei gebaut wird
      | Package |
      | ./cmd/shop |
    Und die Shopdatei als eigener Dienst gestartet wird
    Und die vollständige Bereitschaft begrenzt abgewartet wird
      | Path | TimeoutSeconds | Contains |
      | /ready | 10 | CDC:initialized |
    Dann liefert die Laufzeitdiagnose
      | Status | Contains |
      | 200 | CDC:initialized |
    Wenn die Shopdatei mit Argumenten beendet wird
      | Arguments |
      | healthcheck |
    Dann lautet der Prozessstatus
      | ExitCode |
      | 0 |
    Wenn die Shopdatei mit Argumenten beendet wird
      | Arguments |
      | unknown |
    Dann lautet der Prozessstatus
      | ExitCode |
      | 1 |

    Wenn die Shopdatei mit Argumenten beendet wird
      | Arguments |
      | healthcheck,extra |
    Dann lautet der Prozessstatus
      | ExitCode |
      | 1 |

  Szenario: Eine geschlossene Fachverbindung wird als echte fehlende Bereitschaft gemeldet
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE |
    Und die geöffnete Fachverbindung geschlossen wird
    Und die Laufzeitdiagnose per HTTP abgefragt wird
      | Path |
      | /health |
    Dann liefert die Laufzeitdiagnose
      | Status | Contains |
      | 503 | DependencyUnavailable |
    Wenn die Laufzeitdiagnose per HTTP abgefragt wird
      | Path |
      | /ready |
    Dann liefert die Laufzeitdiagnose
      | Status | Contains |
      | 503 | DependencyUnavailable |

  Szenario: Ein verlorener dauerhafter WAL-Slot sperrt die Kaufbereitschaft nicht
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE |
    Und ein gespeicherter WAL-Checkpoint ohne seinen Slot besteht
    Und Worker Projektor und HTTP-Dienst gestartet werden
    Und die vollständige Bereitschaft begrenzt abgewartet wird
      | Path | TimeoutSeconds | Contains |
      | /ready | 10 | CDC:ReindexRequired |
    Dann liefert die Laufzeitdiagnose
      | Status | Contains |
      | 200 | CDC:ReindexRequired |

  Szenario: Ein expliziter synthetischer Startkatalog wird vor den Adapterbindungen importiert
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market | SeedFile |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE | ../../spec/fixtures/runtime/startkatalog.json |
    Dann die veröffentlichte Laufzeitquelle enthält
      | Products | Families | Categories |
      | 1 | 1 | 3 |
    Wenn Worker Projektor und HTTP-Dienst gestartet werden
    Und die vollständige Bereitschaft begrenzt abgewartet wird
      | Path | TimeoutSeconds | Contains |
      | /ready | 10 | CDC:initialized |
    Dann liefert die Laufzeitdiagnose
      | Status | Contains |
      | 200 | CDC:initialized |

  Szenario: Ein technisch geschlossener TCP-Listener wird als Startfehler zurückgegeben
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE |
    Und der HTTP-Dienst einen bereits geschlossenen TCP-Listener erhält
    Dann lautet das Ergebnis der Laufzeitoperation
      | Code |
      | InternalError |

  Szenario: Der Projektor bleibt nach einem echten Verbindungsverlust über sein Retryintervall hinweg erreichbar
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE |
    Und die geöffnete Fachverbindung geschlossen wird
    Und Worker Projektor und HTTP-Dienst gestartet werden
    Und die Bereitschaft nach dem technischen Retryintervall erneut abgefragt wird
      | Path | Milliseconds |
      | /ready | 1100 |
    Dann liefert die Laufzeitdiagnose
      | Status | Contains |
      | 503 | DependencyUnavailable |

  Szenario: Fehlende Katalogleserechte verhindern die Adapterbindung vor dem Workerstart
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE |
    Und eine neue Laufzeit ohne Leseberechtigung für den Katalog geöffnet wird
    Dann lautet das Ergebnis der Laufzeitoperation
      | Code |
      | DependencyUnavailable |
