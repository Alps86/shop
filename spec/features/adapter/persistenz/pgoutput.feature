# language: de
@runtime @pgoutput
Funktionalität: Echter logischer PostgreSQL-Nachlauf
  Der Reindex liest einen konsistent exportierten Snapshot und verarbeitet danach WAL-Ereignisse.

  Szenario: Ein nach dem Snapshot geschriebener Tombstone erreicht das Projektionsziel
    Angenommen ein exportierter PostgreSQL-Snapshot mit eigenem pgoutput-Slot besteht
    Wenn danach ein vollständiges Löschereignis committed wird
    Und der WAL-Nachlauf bis zur aktuellen Grenze erfolgt
    Dann lautet die tatsächliche Projektionsbeobachtung
      | EventCount | ProductID |
      | 1 | p-phone |

  Szenario: Ein Zielausfall bestätigt kein unverarbeitetes WAL und der Slot setzt fort
    Angenommen ein eigener dauerhafter pgoutput-Verbraucher vorbereitet ist
    Und das Projektionsziel vorübergehend ausfällt
    Wenn danach ein vollständiges Löschereignis committed wird
    Und der Verbraucher wegen des Zielausfalls beendet wird
    Und derselbe Slot mit wieder verfügbarem Ziel neu startet
    Und das Ereignis verarbeitet und der Verbraucher angehalten wird
    Dann lautet die tatsächliche Projektionsbeobachtung
      | EventCount | ProductID |
      | 1 | p-phone |

  Szenariogrundriss: Ein ungültiges Schema hält sichtbar an und quittiert keine LSN
    Angenommen ein eigener dauerhafter pgoutput-Verbraucher vorbereitet ist
    Wenn ein ungültiges Projektionsereignis tatsächlich committed wird
      | SchemaVersion | EventID | Revision | Locale | RawJSON |
      | <schema> | <event> | <revision> | <locale> | <json> |
    Und der Verbraucher seinen Diagnosefehler meldet
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN |
      | <code> | 0/0 |
    Beispiele:
      | schema | event | revision | locale | json | code |
      | 2 | e-bad | 1 | de-DE | | InvalidInput |
      | 1 | | 1 | de-DE | | InvalidInput |
      | 1 | e-bad | 0 | de-DE | | InvalidInput |
      | 1 | e-bad | 1 | invalid | | UnsupportedContext |
      | 1 | e-bad | 1 | de-DE | {"SchemaVersion":"invalid"} | InvalidInput |

  Szenario: Ein abgebrochener Reindex gibt den exportierten Slot idempotent frei
    Angenommen ein exportierter PostgreSQL-Snapshot mit eigenem pgoutput-Slot besteht
    Wenn der exportierte Snapshot wiederholt abgebrochen wird
    Dann ist der temporäre Reindexslot freigegeben

  Szenario: Ein abgebrochener Cursor wird nicht still erneut verwendet
    Angenommen ein exportierter PostgreSQL-Snapshot mit eigenem pgoutput-Slot besteht
    Wenn der exportierte Snapshot wiederholt abgebrochen wird
    Und der WAL-Nachlauf bis zur aktuellen Grenze erfolgt
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN |
      | NotFound | 0/0 |

  Szenario: Ergänzende Decoderfehler werden ohne vorgespielten PostgreSQL-Server geprüft
    Wenn ergänzend folgende Protokollframes durch denselben Decoder laufen
      | FrameHex | Code |
      |  | InvalidInput |
      | 6b00 | InternalError |
      | 7700 | InternalError |
      | 770000000000000000000000000000000000000000000000 | InternalError |
      | 77000000000000000000000000000000000000000000000000 | InvalidInput |
      | 770000000000000000000000000000000000000000000000003f | InternalError |
      | 6b0000000000000000000000000000000000 | OK |
      | 770000000000000000000000000000000000000000000000004d0000000000000000000000000000 | OK |
      | 7700000000000000000000000000000000000000000000000049000000014e0000 | InvalidInput |
      | 7700000000000000000000000000000000000000000000000052000000017075626c6963006f7574626f780064000100656e76656c6f70650000000edaffffffff | OK |
      | 7700000000000000000000000000000000000000000000000049000000014e0000 | InvalidInput |
      | 7700000000000000000000000000000000000000000000000052000000017075626c6963006f746865720064000100656e76656c6f70650000000edaffffffff | OK |
      | 7700000000000000000000000000000000000000000000000049000000014e0000 | OK |

  Szenario: Der laufende Verbraucher wiederholt technische Zielausfälle selbst mit Backoff
    Angenommen ein echter WAL-Verbraucher mit technischem Backoff läuft
    Und das Projektionsziel vorübergehend ausfällt
    Wenn danach ein vollständiges Löschereignis committed wird
    Und das Ziel nach einem beobachteten Fehlversuch wieder erreichbar wird
    Und das Ereignis verarbeitet und der Verbraucher angehalten wird
    Dann lautet die tatsächliche Projektionsbeobachtung
      | EventCount | ProductID |
      | 1 | p-phone |

  Szenario: Ergänzende Metadatenfälle verhindern das Überspringen verlorener WAL-Stände
    Wenn ergänzend Slotmetadatenwerte für eine Fortsetzung geprüft werden
      | Exists | Restart | Confirmed | Checkpoint | WALStatus | Code | Position |
      | true | 10 | 20 | 0 | reserved | OK | 20 |
      | true | 10 | 20 | 15 | extended | OK | 15 |
      | false | 0 | 0 | 15 | | ReindexRequired | 0 |
      | true | 0 | 20 | 15 | lost | ReindexRequired | 0 |
      | true | 16 | 20 | 15 | reserved | ReindexRequired | 0 |

  Szenario: Technische Hülle unterscheidet vollständige Upserts von Tombstones
    Wenn ergänzend die technischen Ereigniswerte validiert werden
      | SchemaVersion | EventID | OccurredAt | Kind | ProductID | Revision | ProductPresent | SnapshotID | InventoryCount | Locale | Code |
      | 1 | e-a | 2026-10-15T00:00:00Z | upsert | p-a | 1 | true | p-a | 1 | de-DE | OK |
      | 1 | e-a | 2026-10-15T00:00:00Z | delete | p-a | 1 | false | | 0 | de-DE | OK |
      | 1 | e-a | | delete | p-a | 1 | false | | 0 | de-DE | InvalidInput |
      | 1 | e-a | 2026-10-15T00:00:00Z | other | p-a | 1 | false | | 0 | de-DE | InvalidInput |
      | 1 | e-a | 2026-10-15T00:00:00Z | delete | invalid id | 1 | false | | 0 | de-DE | InvalidInput |
      | 1 | e-a | 2026-10-15T00:00:00Z | delete | p-a | 1 | false | p-a | 0 | de-DE | InvalidInput |
      | 1 | e-a | 2026-10-15T00:00:00Z | upsert | p-a | 1 | false | | 0 | de-DE | InvalidInput |
      | 1 | e-a | 2026-10-15T00:00:00Z | upsert | p-a | 1 | true | p-b | 0 | de-DE | InvalidInput |
      | 1 | e-a | 2026-10-15T00:00:00Z | delete | p-a | 1 | true | p-a | 0 | de-DE | InvalidInput |
      | 1 | e-a | 2026-10-15T00:00:00Z | delete | p-a | 1 | false | | 1 | de-DE | InvalidInput |

  Szenario: Doppelte EventID innerhalb eines Commits wird genau einmal projiziert
    Angenommen ein eigener dauerhafter pgoutput-Verbraucher vorbereitet ist
    Wenn ein Ereignis innerhalb derselben WAL-Transaktion wiederholt wird
      | Payload |
      | identical |
    Und das Ereignis verarbeitet und der Verbraucher angehalten wird
    Dann lautet die tatsächliche Projektionsbeobachtung
      | EventCount | ProductID |
      | 1 | p-phone |

  Szenario: Abweichende Payload derselben EventID im Commit scheitert vor dem Target
    Angenommen ein eigener dauerhafter pgoutput-Verbraucher vorbereitet ist
    Wenn ein Ereignis innerhalb derselben WAL-Transaktion wiederholt wird
      | Payload |
      | different |
    Und der Verbraucher seinen Diagnosefehler meldet
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN | EventCount |
      | Conflict | 0/0 | 0 |

  Szenario: Bereits quittierte EventID wird nach Wiederanlauf ohne Targetaufruf wiederholt
    Angenommen ein eigener dauerhafter pgoutput-Verbraucher vorbereitet ist
    Wenn danach ein vollständiges Löschereignis committed wird
    Und das Ereignis verarbeitet und der Verbraucher angehalten wird
    Und derselbe Slot mit wieder verfügbarem Ziel neu startet
    Und ein gespeichertes Ereignis erneut veröffentlicht wird
      | Payload |
      | identical |
    Und die Wiederholung ohne weiteren Targetaufruf quittiert wird
    Dann lautet die tatsächliche Projektionsbeobachtung
      | EventCount | ProductID |
      | 1 | p-phone |

  Szenario: Widersprüchliche dauerhafte EventID wird nach Wiederanlauf diagnostiziert
    Angenommen ein eigener dauerhafter pgoutput-Verbraucher vorbereitet ist
    Wenn danach ein vollständiges Löschereignis committed wird
    Und das Ereignis verarbeitet und der Verbraucher angehalten wird
    Und derselbe Slot mit wieder verfügbarem Ziel neu startet
    Und ein gespeichertes Ereignis erneut veröffentlicht wird
      | Payload |
      | different |
    Und der Verbraucher seinen Diagnosefehler meldet
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN | EventCount |
      | Conflict | unchanged | 1 |

  Szenario: Ein verlorener dauerhafter Slot verlangt Reindex statt stiller Neuanlage
    Angenommen ein eigener dauerhafter pgoutput-Verbraucher vorbereitet ist
    Wenn danach ein vollständiges Löschereignis committed wird
    Und das Ereignis verarbeitet und der Verbraucher angehalten wird
    Und der eigene bestätigte Replikationsslot verloren geht
    Und derselbe Verbraucher erneut vorbereitet wird
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN | EventCount |
      | ReindexRequired | unchanged | 1 |

  Szenario: Eine vorhandene Publication auf der falschen Tabelle wird abgelehnt
    Angenommen folgende CDC-SQL-Fixture nur eigene Artefakte verändert
      | SQL |
      | CREATE PUBLICATION $publication FOR TABLE $schema.records WITH (publish='insert') |
    Wenn derselbe Verbraucher erneut vorbereitet wird
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN | EventCount |
      | InvalidInput | 0/0 | 0 |

  Szenario: Eine beschädigte Outboxspalte wird nicht still als erfolgreich konsumiert quittiert
    Angenommen ein eigener dauerhafter pgoutput-Verbraucher vorbereitet ist
    Wenn folgende CDC-SQL-Fixture nur eigene Artefakte verändert
      | SQL |
      | ALTER TABLE $schema.outbox RENAME COLUMN envelope TO broken |
      | INSERT INTO $schema.outbox(broken) VALUES('{}') |
    Und der Verbraucher seinen Diagnosefehler meldet
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN | EventCount |
      | InvalidInput | 0/0 | 0 |

  Szenariogrundriss: Ein Quittierungsfehler rollt Verbraucherbeleg und LSN gemeinsam zurück
    Angenommen ein eigener dauerhafter pgoutput-Verbraucher vorbereitet ist
    Und folgende CDC-SQL-Fixture nur eigene Artefakte verändert
      | SQL |
      | CREATE FUNCTION $schema.fail_ack() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'gezielter Quittierungsfehler'; END $$ |
      | CREATE TRIGGER fail_ack BEFORE INSERT ON $schema.<tabelle> FOR EACH ROW EXECUTE FUNCTION $schema.fail_ack() |
    Wenn danach ein vollständiges Löschereignis committed wird
    Und der Verbraucher seinen Diagnosefehler meldet
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN | EventCount |
      | DependencyUnavailable | 0/0 | 1 |
    Wenn folgende CDC-SQL-Fixture nur eigene Artefakte verändert
      | SQL |
      | DROP TRIGGER fail_ack ON $schema.<tabelle> |
      | DROP FUNCTION $schema.fail_ack() |
    Und derselbe Slot mit wieder verfügbarem Ziel neu startet
    Und das Ereignis verarbeitet und der Verbraucher angehalten wird
    Dann lautet die tatsächliche Projektionsbeobachtung
      | EventCount | ProductID |
      | 2 | p-phone |
    Beispiele:
      | tabelle |
      | cdc_events |
      | cdc_checkpoints |

  Szenario: Ein Slot der falschen Datenbank wird als Reindexbedarf diagnostiziert
    Angenommen derselbe Slotname zu einer anderen eigenen Testdatenbank gehört
    Und ein eigener dauerhafter pgoutput-Verbraucher vorbereitet ist
    Wenn der Verbraucher seinen Diagnosefehler meldet
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN | EventCount |
      | ReindexRequired | 0/0 | 0 |

  Szenariogrundriss: Snapshotfehler bestätigen kein WAL und geben die exportierte Verbindung frei
    Angenommen folgende CDC-SQL-Fixture nur eigene Artefakte verändert
      | SQL |
      | <sql> |
    Wenn ein echter Reindexsnapshot versucht wird
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN | EventCount |
      | <code> | 0/0 | 0 |
    Beispiele:
      | sql | code |
      | DROP TABLE $schema.outbox | InternalError |
      | DROP TABLE $schema.projection_versions CASCADE | InternalError |
      | ALTER TABLE $schema.projection_versions ALTER COLUMN revision TYPE text; INSERT INTO $schema.projection_versions VALUES('p-missing','broken',false) | InternalError |
      | INSERT INTO $schema.projection_versions(product_id,revision,deleted) VALUES('p-missing',1,false) | NotFound |

  Szenario: Ein geschlossener SQL-Pool verhindert die Snapshotübernahme ohne falsche Erfolgsmeldung
    Angenommen der eigene SQL-Verbindungspool geschlossen wird
    Wenn ein echter Reindexsnapshot versucht wird
    Dann meldet die Snapshotgrenze ihren tatsächlichen Fehler
      | Code |
      | InternalError |

  Szenario: Ein Catchup nach geschlossenem SQL-Pool verliert seinen Cursor und bleibt sichtbar fehlgeschlagen
    Angenommen ein exportierter PostgreSQL-Snapshot mit eigenem pgoutput-Slot besteht
    Und der eigene SQL-Verbindungspool geschlossen wird
    Wenn der WAL-Nachlauf bis zur aktuellen Grenze erfolgt
    Dann meldet die Snapshotgrenze ihren tatsächlichen Fehler
      | Code |
      | DependencyUnavailable |

  Szenario: Ein Schemafehler wird auch vom dauerhaften Runner nicht endlos wiederholt
    Angenommen ein echter WAL-Verbraucher mit technischem Backoff läuft
    Wenn ein ungültiges Projektionsereignis tatsächlich committed wird
      | SchemaVersion | EventID | Revision | Locale | RawJSON |
      | 2 | e-bad | 1 | de-DE | |
    Und der Verbraucher seinen Diagnosefehler meldet
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN | EventCount |
      | InvalidInput | 0/0 | 0 |

  Szenario: Kündigung beendet den tatsächlichen Backoff ohne weiteres Projektionsziel
    Angenommen ein echter WAL-Verbraucher mit technischem Backoff läuft
    Und das Projektionsziel vorübergehend ausfällt
    Wenn danach ein vollständiges Löschereignis committed wird
    Und die technische Wiederholung während ihres Backoffs gekündigt wird
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN | EventCount |
      | InternalError | 0/0 | 0 |

  Szenario: Eine beschädigte Checkpointtabelle erzeugt keinen scheinbar neuen Verbrauchsstand
    Angenommen folgende CDC-SQL-Fixture nur eigene Artefakte verändert
      | SQL |
      | DROP TABLE $schema.cdc_checkpoints |
    Wenn derselbe Verbraucher erneut vorbereitet wird
    Dann meldet die WAL-Vorbereitung ihren tatsächlichen Fehler
      | Code |
      | DependencyUnavailable |

  Szenario: Fremde logische Marker beenden den konsistenten Nachlauf nicht vor seiner eigenen Barriere
    Angenommen ein exportierter PostgreSQL-Snapshot mit eigenem pgoutput-Slot besteht
    Und folgende CDC-SQL-Fixture nur eigene Artefakte verändert
      | SQL |
      | SELECT pg_logical_emit_message(false,'unrelated','') |
    Wenn danach ein vollständiges Löschereignis committed wird
    Und der WAL-Nachlauf bis zur aktuellen Grenze erfolgt
    Dann lautet die tatsächliche Projektionsbeobachtung
      | EventCount | ProductID |
      | 1 | p-phone |

  Szenario: Eine fehlende Consumerbelegtabelle quittiert keine ansonsten gültige Projektion
    Angenommen ein eigener dauerhafter pgoutput-Verbraucher vorbereitet ist
    Und folgende CDC-SQL-Fixture nur eigene Artefakte verändert
      | SQL |
      | DROP TABLE $schema.cdc_events |
    Wenn danach ein vollständiges Löschereignis committed wird
    Und der Verbraucher seinen Diagnosefehler meldet
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN | EventCount |
      | DependencyUnavailable | 0/0 | 0 |

  Szenariogrundriss: Fehlende PostgreSQL-Rechte werden an der tatsächlichen Replikationsgrenze sichtbar
    Angenommen eine eigene Testdatenbank ohne folgende Berechtigung besteht
      | Permission |
      | <permission> |
    Wenn derselbe Verbraucher erneut vorbereitet wird
    Dann meldet die WAL-Vorbereitung ihren tatsächlichen Fehler
      | Code |
      | <code> |
    Beispiele:
      | permission | code |
      | SlotMetadata | DependencyUnavailable |
      | Replication | InternalError |
      | PublicationMetadata | InternalError |

  Szenario: Fehlende Slotmetadatenrechte verhindern die Übernahme eines exportierten Snapshots
    Angenommen eine eigene Testdatenbank ohne folgende Berechtigung besteht
      | Permission |
      | SlotMetadata |
    Wenn ein echter Reindexsnapshot versucht wird
    Dann meldet die Snapshotgrenze ihren tatsächlichen Fehler
      | Code |
      | DependencyUnavailable |

  Szenario: Fehlendes Replikationsrecht verhindert schon die exportierte Verbindung
    Angenommen eine eigene Testdatenbank ohne folgende Berechtigung besteht
      | Permission |
      | Replication |
    Wenn ein echter Reindexsnapshot versucht wird
    Dann meldet die Snapshotgrenze ihren tatsächlichen Fehler
      | Code |
      | InternalError |

  Szenario: Eine serverseitig beendete eigene Snapshotverbindung kann keinen Nachlauf vortäuschen
    Angenommen ein exportierter PostgreSQL-Snapshot mit eigenem pgoutput-Slot besteht
    Und die eigene exportierte Replikationsverbindung serverseitig beendet wird
    Wenn der WAL-Nachlauf bis zur aktuellen Grenze erfolgt
    Dann meldet die Snapshotgrenze ihren tatsächlichen Fehler
      | Code |
      | InternalError |

  Szenario: Ein gekündigter Runnerstart scheitert schon an seiner tatsächlichen SQL-Verbindung
    Wenn ein schon gekündigter nativer Runnerstart versucht wird
    Dann meldet die WAL-Vorbereitung ihren tatsächlichen Fehler
      | Code |
      | InternalError |

  Szenario: Der ergänzende Keepalive-Decoder meldet den Fehler einer geschlossenen echten Verbindung
    Wenn ergänzend ein gültiger Keepalive nach geschlossener echter Verbindung Antwort verlangt
    Dann meldet die WAL-Vorbereitung ihren tatsächlichen Fehler
      | Code |
      | InternalError |

  Szenario: Ein ungültiger Slotname wird bei tatsächlicher Slotanlage abgelehnt
    Angenommen ein ungültiger eigener Slotname konfiguriert ist
      | Slot |
      | Invalid-Slot |
    Wenn derselbe Verbraucher erneut vorbereitet wird
    Dann meldet die WAL-Vorbereitung ihren tatsächlichen Fehler
      | Code |
      | InternalError |

  Szenario: Auch die tatsächliche Reindexslotanlage erhält den diagnostischen Consumerpräfix
    Angenommen ein ungültiger eigener Slotname konfiguriert ist
      | Slot |
      | Invalid-Slot |
    Wenn ein echter Reindexsnapshot versucht wird
    Dann meldet die Snapshotgrenze ihren tatsächlichen Fehler
      | Code |
      | InternalError |

  Szenario: Ein maximal langer gültiger Consumername bleibt beim temporären Slot unterstützt
    Angenommen ein langer gültiger eigener Slotname konfiguriert ist
      | Character | Length |
      | s | 63 |
    Und ein exportierter PostgreSQL-Snapshot mit eigenem pgoutput-Slot besteht
    Wenn der exportierte Snapshot wiederholt abgebrochen wird
    Dann ist der temporäre Reindexslot freigegeben

  Szenario: Ein serverseitig gekündigter eigener Replikationsbefehl meldet seinen Serverfehler
    Angenommen ein eigener dauerhafter pgoutput-Verbraucher vorbereitet ist
    Wenn der eigene laufende Replikationsbefehl serverseitig gekündigt wird
    Und der Verbraucher seinen Diagnosefehler meldet
    Dann bleibt der WAL-Checkpoint mit Diagnose unverändert
      | Code | LSN | EventCount |
      | InternalError | 0/0 | 0 |

  Szenario: Der ergänzende Nachrichtendecoder ignoriert eine standardisierte Notice
    Wenn ergänzend eine standardisierte Notice durch denselben Nachrichtendecoder läuft
    Dann meldet die WAL-Vorbereitung ihren tatsächlichen Fehler
      | Code |
      | OK |
