# language: de
@http
Funktionalität: Authentifizierte SKUcacheinvalidierung
  Als Laufzeitadapter invalidiere ich nur gültige SKUkeys über den internen Cachevertrag.

  Szenariogrundriss: Jede SKU erhält einen authentifizierten PURGErequest
    Angenommen die Cacheinvalidierung wird eingerichtet
      | Endpoint | Key | Status |
      | test-server | 00000000000000000000000000000000 | 200 |
    Wenn die SKUcacheinvalidierung ausgeführt wird
      | SKUs |
      | <skus> |
    Dann die Cachegrenze beobachtet
      | Outcome | Calls | Methods | SKUs | Keys |
      | success | <anzahl> | <methoden> | <skus> | <schluessel> |
    Beispiele:
      | skus | anzahl | methoden | schluessel |
      | RED-256,BLUE-128 | 2 | PURGE,PURGE | 00000000000000000000000000000000,00000000000000000000000000000000 |
      | | 0 | | |

  Szenariogrundriss: Ungültige Cachekonfiguration startet keine Kommunikation
    Angenommen die Cacheinvalidierung wird eingerichtet
      | Endpoint | Key | Status |
      | <endpoint> | <schluessel> | 200 |
    Dann die Adapterinitialisierung meldet
      | Error |
      | <fehler> |
    Beispiele:
      | endpoint | schluessel | fehler |
      | % | 00000000000000000000000000000000 | cacheendpoint |
      | https://cache.example | 00000000000000000000000000000000 | internes HTTP |
      | http:/missing | 00000000000000000000000000000000 | internen Host |
      | test-server | short | mindestens 32 Bytes |

  Szenariogrundriss: Fehler werden dem Dispatcher übergeben statt als Purgeerfolg gemeldet
    Angenommen die Cacheinvalidierung wird eingerichtet
      | Endpoint | Key | Status |
      | test-server | 00000000000000000000000000000000 | <status> |
    Wenn die SKUcacheinvalidierung ausgeführt wird
      | SKUs |
      | <sku> |
    Dann die Adapterinitialisierung meldet
      | Error |
      | <fehler> |
    Beispiele:
      | status | sku | fehler |
      | 503 | RED-256 | HTTP 503 |
      | 200 | bad sku | InvalidInput |

  Szenario: Abgebrochene Invalidierung wird dem Aufrufer als Fehler geliefert
    Angenommen die Cacheinvalidierung wird eingerichtet
      | Endpoint | Key | Status |
      | test-server | synthetic-composition-test-key-32-bytes | 200 |
    Wenn die SKUcacheinvalidierung ausgeführt wird
      | SKUs | Context |
      | RED-256 | cancelled |
    Dann die Cachegrenze beobachtet
      | Outcome | Calls |
      | context canceled | 0 |
