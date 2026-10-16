# language: de
@adapter @composition @content-export
Funktionalität: Versionsgebundener Inhaltsaufbau
  Die eigene Shopdatei erzeugt Hugo-Inhalte aus dem tatsächlichen Katalogstand.

  Szenario: Der Inhaltsbefehl ist eine erfolgreiche eigene Prozessrolle
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market | SeedFile |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE | ../../spec/fixtures/runtime/startkatalog.json |
    Und eine abweichende Laufzeitkonfiguration ausgeführt wird
      | Operation | Field | Value |
      | content | Listen | 127.0.0.1:0 |
    Dann lautet das Ergebnis der Laufzeitoperation
      | Code |
      | OK |
    Und die erzeugten Hugo-Inhalte enthalten
      | File | Contains |
      | product-p-phone.md | "title":"Synthetisches Testtelefon" |
      | product-p-phone.md | "url":"/_frame/product/p-phone/1/" |
      | product-p-phone.md | "layout":"dynamic" |
      | product-p-phone.md | "frame_kind":"product" |
      | product-p-phone.md | "root_version":1 |
      | product-p-phone.md | "canonical":"/de-DE/produkt/testmodell--p-phone" |
      | product-p-phone.md | "fragment":"/fragments/de-DE/produkt/testmodell--p-phone" |
      | category-mobile.md | "url":"/_frame/category/mobile/1/" |
      | category-mobile.md | "frame_kind":"category" |
      | category-home.md | "title":"Haushalt" |
      | category-coffee.md | "title":"Kaffee" |
    Wenn die ausführbare Shopdatei gebaut wird
      | Package |
      | ./cmd/shop |
    Und die Shopdatei mit Argumenten beendet wird
      | Arguments |
      | content |
    Dann lautet der Prozessstatus
      | ExitCode |
      | 0 |

  Szenariogrundriss: Dateisystemkonflikte verhindern den Verzeichnisexport sichtbar
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market | SeedFile |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE | ../../spec/fixtures/runtime/startkatalog.json |
    Und der Katalog in ein technisches Inhaltsziel exportiert wird
      | Path | Existing | Body |
      | blocked | <Vorhanden> | <Inhalt> |
    Dann der Inhaltsexport ein Dateisystemergebnis liefert
      | Error |
      | true |

    Beispiele:
      | Vorhanden | Inhalt |
      | file | occupied |
      | directory-entry | product-p-phone.md |

  Szenariogrundriss: Ungültige Oberflächenkonfiguration wird vor dem Dienststart abgelehnt
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE |
    Und eine abweichende Laufzeitkonfiguration ausgeführt wird
      | Operation | Field | Value |
      | open | <Feld> | <Wert> |
    Dann lautet das Ergebnis der Laufzeitoperation
      | Code |
      | InvalidInput |

    Beispiele:
      | Feld | Wert |
      | PublicOrigin | not-a-url |
      | PublicOrigin | ftp://localhost |
      | PublicOrigin | http://localhost/path |
      | ContentDirectory | empty |
      | CacheEndpoint | not-a-url |
      | CacheKey | short |
      | CSRFKey | short |
      | BackendListen | empty |


  Szenario: Eine verlorene Fachverbindung verhindert den Export statt leere Artefakte zu erzeugen
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market | SeedFile |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE | ../../spec/fixtures/runtime/startkatalog.json |
    Und die geöffnete Fachverbindung geschlossen wird
    Und der Katalog in ein technisches Inhaltsziel exportiert wird
      | Path | Existing | Body |
      | generated | absent | |
    Dann der Inhaltsexport ein Dateisystemergebnis liefert
      | Error |
      | true |


  Szenariogrundriss: Der Inhaltsbefehl meldet Startfehler über denselben öffentlichen Vertrag
    Wenn die Laufzeit mit echten Dienstverbindungen geöffnet wird
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | Listen | Locale | Currency | Market | SeedFile |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | 127.0.0.1:0 | de-DE | EUR | DE | ../../spec/fixtures/runtime/startkatalog.json |
    Und eine abweichende Laufzeitkonfiguration ausgeführt wird
      | Operation | Field | Value |
      | content | <Feld> | <Wert> |
    Dann lautet das Ergebnis der Laufzeitoperation
      | Code |
      | <Fehler> |

    Beispiele:
      | Feld | Wert | Fehler |
      | Schema | invalid schema | InvalidInput |
      | DSN | postgres://shop@127.0.0.1:1/postgres?sslmode=disable | InternalError |
      | DSN | ://invalid | InternalError |
      | SeedFile | /nonexistent-shop-content-seed.json | InternalError |
