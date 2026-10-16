# language: de
@opensearch
Funktionalität: Wiederverwendbare HTTPverbindungen bei paralleler Suche
  Als Betreiber vermeide ich neue Verbindungen bei unveränderten Suchantworten und respektiere eigene Clients.

  Szenariogrundriss: Zwei abgeschlossene Batches erhalten vollständige Antworten und den vorgesehenen Pool
    Angenommen ein frischer Suchorigin mit synchronisierten Antworten bereitsteht
      | Client | Parallel | ResponseFixture | ExpectedFixture |
      | <client> | 10 | spec/fixtures/opensearch/pool-response.json | spec/fixtures/opensearch/pool-expected.json |
    Wenn zwei getrennte parallele Suchbatches ausgeführt werden
    Dann die vollständigen Suchantworten und Verbindungswiederverwendung entsprechen
      | Results | Reused | SentinelCalls |
      | 20 | <reused> | <sentinel> |
    Beispiele:
      | client | reused | sentinel |
      | default | 10 | 0 |
      | explicit | 1 | 0 |
      | foreignDefault | 2 | 20 |
