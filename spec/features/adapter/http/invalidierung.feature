# language: de
@http
Funktionalität: Authentifizierte öffentliche Such- und SKUcacheinvalidierung
  Als Laufzeitadapter invalidiere ich öffentliche Suchantworten auch ohne SKU und danach echte gültige SKUkeys.

  Szenariogrundriss: Jeder Aufruf invalidiert die Suche und danach jede echte SKU
    Angenommen die Cacheinvalidierung wird eingerichtet
      | Endpoint | Key | Status |
      | test-server | 00000000000000000000000000000000 | 200 |
    Wenn die SKUcacheinvalidierung ausgeführt wird
      | SKUs |
      | <skus> |
    Dann die Cachegrenze beobachtet
      | Outcome | Calls | Methods | SKUs | Keys | PublicSearch |
      | success | <anzahl> | <methoden> | <gesendete-skus> | <schluessel> | <global> |
    Beispiele:
      | skus | anzahl | methoden | gesendete-skus | schluessel | global |
      | RED-256,BLUE-128 | 3 | PURGE,PURGE,PURGE | ,RED-256,BLUE-128 | 00000000000000000000000000000000,00000000000000000000000000000000,00000000000000000000000000000000 | 1,, |
      | | 1 | PURGE | | 00000000000000000000000000000000 | 1 |

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

  Szenario: Ein nachfolgender SKUfehler wird nach erfolgreichem globalem Ban nicht verschluckt
    Angenommen die Cacheinvalidierung wird eingerichtet
      | Endpoint | Key | Status | GlobalStatus |
      | test-server | 00000000000000000000000000000000 | 503 | 200 |
    Wenn die SKUcacheinvalidierung ausgeführt wird
      | SKUs |
      | RED-256 |
    Dann die Cachegrenze beobachtet
      | Calls | PublicSearch | SKUs | Outcome |
      | 2 | 1, | ,RED-256 | cacheinvalidierung: HTTP 503 |

  Szenario: Auch eine ungültige spätere SKU verhindert sämtliche Kommunikation
    Angenommen die Cacheinvalidierung wird eingerichtet
      | Endpoint | Key | Status |
      | test-server | 00000000000000000000000000000000 | 200 |
    Wenn die SKUcacheinvalidierung ausgeführt wird
      | SKUs |
      | RED-256,bad sku |
    Dann die Cachegrenze beobachtet
      | Calls |
      | 0 |
    Und die Adapterinitialisierung meldet
      | Error |
      | InvalidInput |
