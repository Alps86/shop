# language: de
@projection
Funktionalität: Vollständige Suchprojektion über echte PostgreSQL-WAL-Daten
  Als Betreiber behalte ich bestätigte Preis- und Bestandsänderungen auch beim Indexwechsel.

  Szenario: Import und atomare Inventoryänderung erreichen die echte öffentliche Suche
    Angenommen PostgreSQL und OpenSearch über echten WAL-CDC verbunden sind
      | Locale | Currency | Market |
      | de-DE | EUR | DE |
    Wenn ein veröffentlichter Produktdatensatz committed wird
      | Fixture |
      | spec/fixtures/runtime/katalog.json |
    Dann die öffentliche Suche den committed Zustand erhält
      | TotalProducts | SKU | Price | Available | CheckpointAdvanced |
      | 1 | PHONE | 10000 | 10 | true |
    Und der echte Replikationspfad dauerhaft quittiert ist
      | Plugin | Persistent | PublishedTables | CheckpointPresent |
      | pgoutput | true | outbox | true |
    Wenn Preis und Bestand atomar geändert werden
      | SKU | ExpectedRevision | CommandID | Price | Stock |
      | PHONE | 1 | inventory-a | 12000 | 8 |
    Dann die öffentliche Suche den committed Zustand erhält
      | TotalProducts | SKU | Price | Available | CheckpointAdvanced |
      | 1 | PHONE | 12000 | 8 | true |

  Szenariogrundriss: Snapshotnachlauf und Commit nach Catchup gehen beim Aliaswechsel nicht verloren
    Angenommen PostgreSQL und OpenSearch über echten WAL-CDC verbunden sind
      | Locale | Currency | Market |
      | de-DE | EUR | DE |
    Wenn ein veröffentlichter Produktdatensatz committed wird
      | Fixture |
      | spec/fixtures/runtime/katalog.json |
    Dann die öffentliche Suche den committed Zustand erhält
      | TotalProducts | SKU | Price | Available | CheckpointAdvanced |
      | 1 | PHONE | 10000 | 10 | true |
    Wenn ein echter Rebuild an einer kontrollierten WAL-Grenze wartet
      | Boundary | IndexSuffix |
      | <Boundary> | v2 |
    Und Preis und Bestand atomar geändert werden
      | SKU | ExpectedRevision | CommandID | Price | Stock |
      | PHONE | 1 | concurrent-a | 13000 | 7 |
    Dann die öffentliche Suche den committed Zustand erhält
      | TotalProducts | SKU | Price | Available | CheckpointAdvanced |
      | 1 | PHONE | 13000 | 7 | true |
    Wenn Snapshot und WAL-Nachlauf vor dem Aliaswechsel freigegeben werden
    Dann die öffentliche Suche den committed Zustand erhält
      | TotalProducts | SKU | Price | Available |
      | 1 | PHONE | 13000 | 7 |
    Und der Suchalias genau den neuen physischen Index adressiert
      | IndexSuffix | IndexCount |
      | v2 | 1 |
    Beispiele:
      | Boundary |
      | snapshot |
      | afterCatchup |


  Szenario: Derselbe Slot setzt fort und ältere Daten beleben keinen Tombstone wieder
    Angenommen PostgreSQL und OpenSearch über echten WAL-CDC verbunden sind
      | Locale | Currency | Market |
      | de-DE | EUR | DE |
    Wenn ein veröffentlichter Produktdatensatz committed wird
      | Fixture |
      | spec/fixtures/runtime/katalog.json |
    Dann die öffentliche Suche den committed Zustand erhält
      | TotalProducts | SKU | Price | Available | CheckpointAdvanced |
      | 1 | PHONE | 10000 | 10 | true |
    Wenn ein früher bestätigter Projektionsstand aus PostgreSQL gesichert wird
      | ProductID |
      | p-phone |
    Und derselbe frühere Stand erneut an OpenSearch zugestellt wird
      | Attempts |
      | 2 |
    Dann die öffentliche Suche den committed Zustand erhält
      | TotalProducts | SKU | Price | Available |
      | 1 | PHONE | 10000 | 10 |
    Wenn der dauerhafte WAL-Verbraucher vorübergehend angehalten wird
    Und Preis und Bestand atomar geändert werden
      | SKU | ExpectedRevision | CommandID | Price | Stock |
      | PHONE | 1 | offline-a | 12000 | 8 |
    Und derselbe dauerhafte WAL-Slot mit neuem Verbraucher fortgesetzt wird
    Dann die öffentliche Suche den committed Zustand erhält
      | TotalProducts | SKU | Price | Available | CheckpointAdvanced |
      | 1 | PHONE | 12000 | 8 | true |
    Wenn derselbe frühere Stand erneut an OpenSearch zugestellt wird
      | Attempts |
      | 2 |
    Dann die öffentliche Suche den committed Zustand erhält
      | TotalProducts | SKU | Price | Available |
      | 1 | PHONE | 12000 | 8 |
    Wenn das Produkt mit einem dauerhaften Tombstone gelöscht wird
      | ProductID |
      | p-phone |
    Dann die öffentliche Suche den committed Zustand erhält
      | TotalProducts | CheckpointAdvanced |
      | 0 | true |
    Wenn ein vollständiger echter Rebuild durchgeführt wird
      | IndexSuffix |
      | v2 |
    Dann der Suchalias genau den neuen physischen Index adressiert
      | IndexSuffix | IndexCount |
      | v2 | 1 |
    Wenn derselbe frühere Stand erneut an OpenSearch zugestellt wird
      | Attempts |
      | 2 |
    Dann die öffentliche Suche den committed Zustand erhält
      | TotalProducts |
      | 0 |
