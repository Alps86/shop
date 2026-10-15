# language: de
@runtime @postgres
Funktionalität: Tatsächliche PostgreSQL-Lesefehler
  Ein Datenbankfehler wird sichtbar und erzeugt keine erfundene erfolgreiche Ansicht.

  Hintergrund:
    Angenommen ein kohärenter Demodatensatz in echter PostgreSQL vorbereitet ist
      | DatasetVersion | ProductID | SKU | Name | Price | StockQuantity |
      | demo-v1 | p-phone | PHONE | Synthetisches Testmodell | 10000 | 2 |
    Wenn der Datensatz veröffentlicht wird
    Und eine neue echte Gastfähigkeit geöffnet wird

  Szenariogrundriss: Technische Lesefehler entstehen an der tatsächlichen SQL-Grenze
    Angenommen folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | <sql> |
    Wenn ein persistenter Lesevertrag aufgerufen wird
      | Operation | ID | Page | User | Session |
      | <operation> | p-phone | 1 | own | own |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code |
      | DependencyUnavailable |
    Beispiele:
      | operation | sql |
      | product | DROP TABLE $schema.records |
      | product | UPDATE $schema.records SET payload=jsonb_set(payload,'{Revision}','"invalid"') WHERE kind='product' |
      | category | DROP TABLE $schema.records |
      | adminOrders | DROP TABLE $schema.orders |
      | adminOrders | ALTER TABLE $schema.orders RENAME COLUMN payload TO broken |
      | resolve | DROP TABLE $schema.carts CASCADE |
      | interests | DROP TABLE $schema.users CASCADE |
      | diagnostics | DROP TABLE $schema.process_outbox |

  Szenariogrundriss: Fehlerhafte Profile verändern die Gastidentität nicht
    Wenn das persistierte Gastprofil gewählt wird
      | Profile | User |
      | <profil> | <user> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code |
      | <code> |
    Beispiele:
      | profil | user | code |
      | invalid | own | InvalidInput |
      | neutral | unknown | NotFound |
      | home | own | InvalidReference |

  Szenario: Eine beschädigte User-Tabelle erzeugt keine erfolgreiche Profilantwort
    Angenommen folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | DROP TABLE $schema.users CASCADE |
    Wenn das persistierte Gastprofil gewählt wird
      | Profile | User |
      | mobile | own |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code |
      | DependencyUnavailable |

  Szenario: Die Diagnose stammt aus einer tatsächlich ausgeführten Datenbankabfrage
    Wenn ein persistenter Lesevertrag aufgerufen wird
      | Operation |
      | diagnostics |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | PostgreSQL | PendingProcesses | ProjectionEvents | Checkpoints | ReplicationSlots |
      | OK | ready | 0 | 1 | {} | {} |

  Szenariogrundriss: Kündigung eines Requests wird über die echte SQL-Grenze erhalten
    Wenn ein bereits gekündigter Kontext die öffentliche SQL-Grenze erreicht
      | Boundary |
      | <grenze> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | Cancelled |
      | <code> | true |
    Beispiele:
      | grenze | code |
      | Migrate | InternalError |
      | Import | DependencyUnavailable |
      | GuestOpen | DependencyUnavailable |
      | StartCart | DependencyUnavailable |
      | SaveCart | DependencyUnavailable |
      | Inventory | DependencyUnavailable |
      | Purchase | DependencyUnavailable |
      | FulfillmentSave | DependencyUnavailable |
      | Expire | DependencyUnavailable |
      | Profile | DependencyUnavailable |
      | Snapshot | InternalError |
      | ProcessDispatch | DependencyUnavailable |
      | Resolve | DependencyUnavailable |
      | CartGet | DependencyUnavailable |
      | Product | DependencyUnavailable |
      | AdminOrders | DependencyUnavailable |
      | Diagnostics | DependencyUnavailable |
      | Checkpoint | DependencyUnavailable |
      | Slot | DependencyUnavailable |
      | Seen | DependencyUnavailable |
      | ReplayChange | DependencyUnavailable |
      | ReplayConfirm | DependencyUnavailable |
      | Variant | DependencyUnavailable |

  Szenariogrundriss: Reale Schemafehler werden auch innerhalb begonnener Transaktionen weitergegeben
    Angenommen folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | <sql> |
    Wenn die öffentliche SQL-Grenze nach dem Datenbankfehler aufgerufen wird
      | Boundary |
      | <grenze> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code |
      | <code> |
    Beispiele:
      | grenze | sql | code |
      | GuestOpen | DROP TABLE $schema.users CASCADE | DependencyUnavailable |
      | GuestOpen | DROP TABLE $schema.sessions | DependencyUnavailable |
      | StartCart | DROP TABLE $schema.commands | DependencyUnavailable |
      | StartCart | DELETE FROM $schema.records WHERE kind='active-ruleset' | NotFound |
      | StartCart | DROP TABLE $schema.carts CASCADE | DependencyUnavailable |
      | StartCart | DROP TABLE $schema.process_outbox | DependencyUnavailable |
      | SaveCart | DROP TABLE $schema.carts CASCADE | DependencyUnavailable |
      | Inventory | DROP TABLE $schema.commands | DependencyUnavailable |
      | Inventory | DROP TABLE $schema.records | DependencyUnavailable |
      | Profile | DROP TABLE $schema.records | DependencyUnavailable |
      | ProcessDispatch | DROP TABLE $schema.process_outbox | DependencyUnavailable |
      | Snapshot | SELECT 1 | InternalError |
      | ReplayChange | DROP TABLE $schema.commands | DependencyUnavailable |
      | ReplayConfirm | DROP TABLE $schema.commands | DependencyUnavailable |

  Szenariogrundriss: Die Preisermittlung verwendet tatsächliche Produkt-, Pflichtdaten-, Bestands- und Regelzeilen
    Angenommen ein Gast einen Warenkorb mit einer Position dauerhaft anlegt
      | SKU | Quantity |
      | PHONE | 1 |
    Und folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | <sql> |
    Wenn die öffentliche SQL-Grenze nach dem Datenbankfehler aufgerufen wird
      | Boundary |
      | Evaluation |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code |
      | <code> |
    Beispiele:
      | sql | code |
      | DELETE FROM $schema.sku_registry WHERE sku='PHONE' | NotFound |
      | DELETE FROM $schema.records WHERE kind='product' | NotFound |
      | DELETE FROM $schema.records WHERE kind='inventory' | NotFound |
      | DELETE FROM $schema.records WHERE kind='ruleset' | NotFound |
      | UPDATE $schema.records SET payload=jsonb_set(payload,'{Compliance,Context,VerificationStatus}','"pending"') WHERE kind='product' | ComplianceIncomplete |

  Szenariogrundriss: Ungültige Verbindung oder ausgefallene Entropie erzeugen keine erfolgreiche Gastfähigkeit
    Wenn die technische Speichergrenze mit folgender Konfiguration aufgerufen wird
      | DSN | Entropy |
      | <dsn> | <entropy> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code |
      | InternalError |
    Beispiele:
      | dsn | entropy |
      | ://invalid | default |
      | current | unavailable |

  Szenario: Der technische Konstruktor akzeptiert das standardmäßige PostgreSQL-Schema ohne Tabellen anzulegen
    Wenn die technische Speichergrenze mit folgender Konfiguration aufgerufen wird
      | DSN | Schema | Operation |
      | current | default | open |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code |
      | OK |

  Szenariogrundriss: Ein konkurrierender Schreibauftrag respektiert sein tatsächliches SQL-Zeitbudget
    Wenn eine echte konkurrierende Schreibsperre das Requestbudget überschreitet
      | Boundary |
      | <boundary> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | Deadline |
      | DependencyUnavailable | true |
    Beispiele:
      | boundary |
      | StartCart |
      | Import |

  Szenario: Ein unbekanntes Produkt erzeugt keinen erfundenen Tombstone
    Wenn die Produktquelle mit versioniertem Tombstone gelöscht werden soll
      | ProductID |
      | p-missing |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code |
      | NotFound |

  Szenario: Ein Fehler beim Löschen der Produktzeile rollt auch ihren neuen Versionsstand zurück
    Angenommen ein echter SQL-Schreibfehler gezielt vorbereitet ist
      | Table | Operation | MatchField | MatchValue | SQLState |
      | records | DELETE | | | P0001 |
    Wenn die Produktquelle mit versioniertem Tombstone gelöscht werden soll
      | ProductID |
      | p-phone |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Stock | ProjectionRevision |
      | DependencyUnavailable | 0 | 2 | 1 |

  Szenario: Eine ungültige Userkennung erzeugt keinen neuen dauerhaften Warenkorb
    Wenn der dauerhafte Cartstarter eine explizite Userkennung erhält
      | UserID | CommandID |
      | invalid user | start-invalid |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code |
      | InvalidInput |
