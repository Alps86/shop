# language: de
@opensearch
Funktionalität: Cachefähige Completion neben unverändertem Hauptsuchkontext
  Als Gast erhalte ich vollständige Suchergebnisse oder einen ehrlichen Gesamtfehler innerhalb desselben Suchaufrufs.

  Szenario: Completion verwendet einen aliasgebundenen Multirequest
    Angenommen ein beobachtbarer Suchprotokollorigin vorbereitet ist
      | ResponseFixture |
      | spec/fixtures/opensearch/msearch-response.json |
    Wenn die Suche über den Protokollorigin ausgeführt wird
      | Text |
      | ph |
    Dann der Suchdrahtvertrag entspricht
      | Path | ContentType | Calls | Lines | FinalNewline |
      | /protocol/_msearch | application/x-ndjson | 1 | 4 | true |
    Und das vollständige Protokollsuchergebnis entspricht
      | Code | ExpectedFixture |
      | OK | spec/fixtures/opensearch/pool-expected.json |
    Und die NDJSON-Teilabfragen entsprechen
      | Line | Path | JSON |
      | 0 | request_cache | true |
      | 2 | request_cache | true |
      | 1 | _source | false |
      | 1 | stored_fields | ["product_id"] |
      | 1 | size | 12 |
      | 1 | suggest | null |
      | 3 | _source | false |
      | 3 | stored_fields | ["_none_"] |
      | 3 | size | 0 |
      | 3 | query | {"match_all":{}} |
      | 1 | query.bool.filter | [{"term":{"deleted":false}},{"term":{"published":true}}] |
      | 1 | sort | [{"_score":"desc"},{"product_id":"asc"}] |
      | 3 | suggest.catalog.prefix | "ph" |
      | 3 | suggest.catalog.completion.contexts.visibility | ["public"] |
      | 3 | suggest.catalog.completion.fuzzy | {"fuzziness":1,"min_length":4,"prefix_length":2,"transpositions":true} |
      | 3 | suggest.catalog.completion.skip_duplicates | true |
      | 3 | suggest.catalog.completion.size | 25 |

  Szenariogrundriss: Kurze Texte behalten den Einzelpfad
    Angenommen ein beobachtbarer Suchprotokollorigin vorbereitet ist
      | ResponseFixture |
      | spec/fixtures/opensearch/pool-response-single.json |
    Wenn die Suche über den Protokollorigin ausgeführt wird
      | Text |
      | <Text> |
    Dann der Suchdrahtvertrag entspricht
      | Path | ContentType | Calls | Lines | FinalNewline |
      | /protocol/_search?request_cache=true | application/json | 1 | 1 | false |
    Und das vollständige Protokollsuchergebnis entspricht
      | Code | ExpectedFixture |
      | OK | spec/fixtures/opensearch/pool-expected-single.json |
    Beispiele:
      | Text |
      | |
      | p |
      |  p  |

  Szenariogrundriss: Fehler in der Hülle oder einem Teil liefern keinerlei Teilergebnis
    Angenommen ein beobachtbarer Suchprotokollorigin vorbereitet ist
      | Status | Body |
      | <Status> | <Antwort> |
    Wenn die Suche über den Protokollorigin ausgeführt wird
      | Text |
      | ph |
    Dann das vollständige Protokollsuchergebnis entspricht
      | Code |
      | SearchUnavailable |
    Beispiele:
      | Status | Antwort |
      | 503 | {} |
      | 200 | invalid-json |
      | 200 | {} |
      | 200 | {"responses":[]} |
      | 200 | {"responses":null} |
      | 200 | {"responses":false} |
      | 200 | {"responses":{}} |
      | 200 | {"responses":[{"status":"200"},{"status":200}]} |
      | 200 | {"responses":[{"status":200},{"status":true}]} |
      | 200 | {"responses":[{"status":200,"timed_out":"false"},{"status":200}]} |
      | 200 | {"responses":[{"status":200},{"status":200,"timed_out":0}]} |
      | 200 | {"responses":[{"status":200,"_shards":{"failed":"0"}},{"status":200}]} |
      | 200 | {"responses":[{"status":200},{"status":200,"_shards":{"failed":false}}]} |
      | 200 | {"responses":[{"status":200}]} |
      | 200 | {"responses":[{"status":200},{"status":200},{"status":200}]} |
      | 200 | {"responses":[{"status":400},{"status":200}]} |
      | 200 | {"responses":[{"status":200},{"status":500}]} |
      | 200 | {"responses":[{}, {"status":200}]} |
      | 200 | {"responses":[{"status":200},{}]} |
      | 200 | {"responses":[{"status":200,"error":{}},{"status":200}]} |
      | 200 | {"responses":[{"status":200},{"status":200,"error":{}}]} |
      | 200 | {"responses":[{"status":200,"timed_out":true},{"status":200}]} |
      | 200 | {"responses":[{"status":200},{"status":200,"timed_out":true}]} |
      | 200 | {"responses":[{"status":200,"_shards":{"failed":1}},{"status":200}]} |
      | 200 | {"responses":[{"status":200},{"status":200,"_shards":{"failed":1}}]} |
      | 200 | {"responses":[{"status":200,"aggregations":{"facet_brand":false}},{"status":200}]} |
      | 200 | {"responses":[{"status":200},{"status":200,"suggest":false}]} |

  Szenario: Leere Suchergebnisse und fehlende Vorschläge bleiben gültig
    Angenommen ein beobachtbarer Suchprotokollorigin vorbereitet ist
      | Body |
      | {"responses":[{"status":200},{"status":200}]} |
    Wenn die Suche über den Protokollorigin ausgeführt wird
      | Text |
      | ph |
    Dann das vollständige Protokollsuchergebnis entspricht
      | Code | ExpectedFixture |
      | OK | spec/fixtures/opensearch/msearch-empty-expected.json |

  Szenariogrundriss: Ein gemeinsames Budget bricht Warten auf Header und Body ohne Retry ab
    Angenommen ein beobachtbarer Suchprotokollorigin vorbereitet ist
      | ResponseFixture | DelayMilliseconds | Phase |
      | spec/fixtures/opensearch/msearch-response.json | 800 | <Phase> |
    Wenn die Suche über den Protokollorigin ausgeführt wird
      | Text |
      | ph |
    Dann das vollständige Protokollsuchergebnis entspricht
      | Code |
      | SearchUnavailable |
    Und bleibt die öffentliche Suchantwort innerhalb des Fehlerbudgets
      | MaximumMilliseconds |
      | 750 |
    Und der Suchdrahtvertrag entspricht
      | Calls | FirstPartSent |
      | 1 | <ErsterTeil> |
    Beispiele:
      | Phase | ErsterTeil |
      | headers | false |
      | body | true |

  Szenariogrundriss: Hauptdaten und unverbindliche Vorschläge dürfen unterschiedliche Livegenerationen zeigen
    Angenommen ein beobachtbarer Suchprotokollorigin vorbereitet ist
      | ResponseFixture |
      | spec/fixtures/opensearch/msearch-<Generation>-response.json |
    Wenn die Suche über den Protokollorigin ausgeführt wird
      | Text |
      | ph |
    Dann das vollständige Protokollsuchergebnis entspricht
      | Code | ExpectedFixture |
      | OK | spec/fixtures/opensearch/msearch-<Generation>-expected.json |
    Beispiele:
      | Generation |
      | old |
      | new |

  Szenariogrundriss: Unicode und getrimmte Texte behalten die bestehende Completionregel
    Angenommen ein beobachtbarer Suchprotokollorigin vorbereitet ist
      | Body |
      | {"responses":[{"status":200},{"status":200}]} |
    Wenn die Suche über den Protokollorigin ausgeführt wird
      | Text |
      | <Text> |
    Dann der Suchdrahtvertrag entspricht
      | Path | Calls |
      | /protocol/_msearch | 1 |
    Und die NDJSON-Teilabfragen entsprechen
      | Line | Path | JSON |
      | 3 | suggest.catalog.prefix | <Prefix> |
      | 3 | suggest.catalog.completion.fuzzy.fuzziness | <Fuzziness> |
    Und das vollständige Protokollsuchergebnis entspricht
      | Code | ExpectedFixture |
      | OK | spec/fixtures/opensearch/msearch-empty-expected.json |
    Beispiele:
      | Text | Prefix | Fuzziness |
      |  ph  | "ph" | 1 |
      | äö | "äö" | 1 |
      | abcdefg | "abcdefg" | 1 |
      | äbcdefgh | "äbcdefgh" | 2 |

  Szenario: Eine vollständige Hauptkarte verdeckt keinen Fehler der Vorschläge
    Angenommen ein beobachtbarer Suchprotokollorigin vorbereitet ist
      | ResponseFixture |
      | spec/fixtures/opensearch/msearch-partial-error.json |
    Wenn die Suche über den Protokollorigin ausgeführt wird
      | Text |
      | ph |
    Dann das vollständige Protokollsuchergebnis entspricht
      | Code |
      | SearchUnavailable |

  Szenariogrundriss: Das begrenzte Suchbudget lässt eine vollständige Antwort nach 350 Millisekunden zu
    Angenommen ein beobachtbarer Suchprotokollorigin vorbereitet ist
      | ResponseFixture | DelayMilliseconds | Phase |
      | spec/fixtures/opensearch/msearch-response.json | 350 | <Phase> |
    Wenn die Suche über den Protokollorigin ausgeführt wird
      | Text |
      | ph |
    Dann das vollständige Protokollsuchergebnis entspricht
      | Code | ExpectedFixture |
      | OK | spec/fixtures/opensearch/pool-expected.json |
    Und bleibt die öffentliche Suchantwort innerhalb des Fehlerbudgets
      | MaximumMilliseconds |
      | 750 |
    Beispiele:
      | Phase |
      | headers |
      | body |

  Szenariogrundriss: Ursprüngliche Transportfehler bleiben sicher klassifiziert und respektieren den Parent
    Angenommen ein beobachtbarer Suchprotokollorigin vorbereitet ist
      | ResponseFixture | DelayMilliseconds | Phase | CaptureDiagnostics | EndpointSuffix |
      | spec/fixtures/opensearch/msearch-response.json | 800 | headers | true | /private-path?private-query=secret |
    Wenn die Suche über den Protokollorigin ausgeführt wird
      | Text | ParentMilliseconds | Parent | CloseOrigin |
      | private-query | <ParentMilliseconds> | <Parent> | <CloseOrigin> |
    Dann das vollständige Protokollsuchergebnis entspricht
      | Code |
      | SearchUnavailable |
    Und bleibt die öffentliche Suchantwort innerhalb des Fehlerbudgets
      | MaximumMilliseconds |
      | <Maximum> |
    Und die ursprüngliche Transportdiagnose ausschließlich sichere Felder enthält
      | Calls | phase | dynamicType | unwrappedType | deadline | canceled | ctxErrCode |
      | 1 | transport | *url.Error | <UnwrappedType> | <Deadline> | <Canceled> | <Context> |
    Beispiele:
      | ParentMilliseconds | Parent | CloseOrigin | Maximum | UnwrappedType | Deadline | Canceled | Context |
      | | | false | 750 | context.deadlineExceededError | true | false | deadline |
      | 50 | | false | 300 | context.deadlineExceededError | true | false | deadline |
      | | canceled | false | 300 | *errors.errorString | false | true | canceled |
      | | | true | 300 | *net.OpError | false | false | none |
