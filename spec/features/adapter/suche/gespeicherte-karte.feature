# language: de
@opensearch
Funktionalität: Vollständige Modellkarten aus nativen DocValues
  Als Gast erhalte ich bei inkonsistenten Kartendaten einen ehrlichen Gesamtfehler.

  Szenariogrundriss: Native DocValues erlauben keine unvollständige oder fremde Karte
    Angenommen ein beobachtbarer Suchprotokollorigin vorbereitet ist
      | Body |
      | <Antwort> |
    Wenn die Suche über den Protokollorigin ausgeführt wird
      | Text |
      | A |
    Dann das vollständige Protokollsuchergebnis entspricht
      | Code |
      | SearchUnavailable |
    Beispiele:
      | Fall | Antwort |
      | RootID null | {"hits":{"hits":[{"fields":{"product_id":null},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWEiLCJTS1UiOiJhLXJlZCJ9"]}}]}}}}]}} |
      | RootID leer | {"hits":{"hits":[{"fields":{"product_id":[]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWEiLCJTS1UiOiJhLXJlZCJ9"]}}]}}}}]}} |
      | RootID mehrfach | {"hits":{"hits":[{"fields":{"product_id":["p-a","p-b"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWEiLCJTS1UiOiJhLXJlZCJ9"]}}]}}}}]}} |
      | RootID leerer Wert | {"hits":{"hits":[{"fields":{"product_id":[""]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWEiLCJTS1UiOiJhLXJlZCJ9"]}}]}}}}]}} |
      | RootID ungueltig | {"hits":{"hits":[{"fields":{"product_id":["!"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWEiLCJTS1UiOiJhLXJlZCJ9"]}}]}}}}]}} |
      | Rootfelder fehlen | {"hits":{"hits":[{"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWEiLCJTS1UiOiJhLXJlZCJ9"]}}]}}}}]}} |
      | SKU null | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":null,"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWEiLCJTS1UiOiJhLXJlZCJ9"]}}]}}}}]}} |
      | SKU leer | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":[],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWEiLCJTS1UiOiJhLXJlZCJ9"]}}]}}}}]}} |
      | SKU mehrfach | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red","a-blue"],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWEiLCJTS1UiOiJhLXJlZCJ9"]}}]}}}}]}} |
      | SKU leerer Wert | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":[""],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWEiLCJTS1UiOiJhLXJlZCJ9"]}}]}}}}]}} |
      | SKU ungueltig | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["!"],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWEiLCJTS1UiOiJhLXJlZCJ9"]}}]}}}}]}} |
      | Payload null | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":null}}]}}}}]}} |
      | Payload leer | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":[]}}]}}}}]}} |
      | Payload mehrfach | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["e30=","e30="]}}]}}}}]}} |
      | Payload Base64 ungueltig | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["!"]}}]}}}}]}} |
      | Payload JSONtyp falsch | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["ImludmFsaWQi"]}}]}}}}]}} |
      | Payload JSON ungueltig | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["aW52YWxpZC1qc29u"]}}]}}}}]}} |
      | Payload ProductID abweichend | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWIiLCJTS1UiOiJhLXJlZCJ9"]}}]}}}}]}} |
      | Payload SKU abweichend | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":["eyJQcm9kdWN0SUQiOiJwLWEiLCJTS1UiOiJhLWJsdWUifQ=="]}}]}}}}]}} |
      | Chosen leer | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[]}}}}]}} |
      | Chosen mehrfach | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{}},{"fields":{}}]}}}}]}} |
      | Chosen fehlt | {"hits":{"hits":[{"fields":{"product_id":["p-a"]}}]}} |
      | Payload fehlt | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"]}}]}}}}]}} |
      | Payload leerer Wert | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":[""]}}]}}}}]}} |
      | Payload null Wert | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":[null]}}]}}}}]}} |
      | SKU fehlt | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.card_payload":["e30="]}}]}}}}]}} |
      | Payload falscher Typ | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":false}}]}}}}]}} |

  Szenario: Technischer Kartenpayload spart nur tatsächlich nullwertige Compliancefelder
    Angenommen eine vollständige technische Suchkarte vorbereitet ist
      | Fixture | Compliance |
      | spec/fixtures/opensearch/pool-expected.json | {} |
    Wenn das Projektionsdokument seinen gespeicherten Kartenpayload erzeugt
    Dann der Kartenpayload ausschließlich echte Nullwerte auslässt
      | Field | Present |
      | Context | true |
      | ApplicableRules | true |
      | Safety | false |
      | ChemicalLabel | false |
      | EnergyLabel | false |
      | NetContent | false |
      | UnitPriceBasis | false |
      | Documents | false |
      | Notices | false |
      | RequiredFields | false |
      | NonApplicableReason | false |
    Und der bestehende Suchdecoder die vollständige Karte unverändert rekonstruiert
      | Text | Sort | Page |
      | A | relevance | 1 |

  Szenariogrundriss: Aktive Daten und nichtnullwertige inaktive Werte bleiben vollständig erhalten
    Angenommen eine vollständige technische Suchkarte vorbereitet ist
      | Fixture | Compliance |
      | spec/fixtures/opensearch/pool-expected.json | <Compliance> |
    Wenn das Projektionsdokument seinen gespeicherten Kartenpayload erzeugt
    Dann der Kartenpayload ausschließlich echte Nullwerte auslässt
      | Field | Present |
      | <Feld> | true |
    Und der bestehende Suchdecoder die vollständige Karte unverändert rekonstruiert
      | Text | Sort | Page |
      | A | relevance | 1 |
    Beispiele:
      | Fall | Compliance | Feld |
      | aktive Pflichtdaten | {"Context":{"Market":"DE","Locale":"de-DE","ValidFrom":{"Present":true,"Value":"2026-10-15T00:00:00Z"},"ValidUntil":{"Present":true,"Value":"2027-10-15T00:00:00Z"},"VerificationStatus":"verified"},"ApplicableRules":[{"Code":"Energy","Market":"DE","ValidUntil":{"Present":true,"Value":"2027-10-15T00:00:00Z"}}],"Safety":{"Warnings":[{"Code":"warning","Text":{"de-DE":"Warnung"},"Placement":["listing"]}]},"ChemicalLabel":{"Present":true,"Value":{"ProductIdentifier":"Reiniger","SignalWord":{"de-DE":"Achtung"},"UFI":"UFI-TEST"}},"EnergyLabel":{"Present":true,"Value":{"EfficiencyClass":"A","ScaleMin":"A","ScaleMax":"G","EPRELRegistrationNumber":"123","LabelDocuments":[{"Kind":"energyLabel","URL":"https://example.invalid/label"}]}},"NetContent":{"Present":true,"Value":{"Value":"0.5","Unit":"L"}},"UnitPriceBasis":{"Present":true,"Value":{"Value":"1","Unit":"L"}},"Documents":[{"Kind":"manual","URL":"https://example.invalid/manual"}],"Notices":[{"Code":"notice","Text":{"de-DE":"Hinweis"},"Placement":["listing"]}],"RequiredFields":["Manufacturer"],"NonApplicableReason":"nachgewiesen"} | ChemicalLabel |
      | inaktiv mit erhaltenem Wert | {"ChemicalLabel":{"Present":false,"Value":{"ProductIdentifier":"nicht loeschen","SignalWord":{}}},"EnergyLabel":{"Present":false,"Value":{"EPRELURL":"https://example.invalid/retained"}},"NetContent":{"Present":false,"Value":{"Value":"0","Unit":"L"}},"UnitPriceBasis":{"Present":false,"Value":{"Value":"0","Unit":"L"}}} | ChemicalLabel |
      | angelegte leere Sammlungen | {"Safety":{"Warnings":[],"Instructions":[]},"Documents":[],"Notices":[],"RequiredFields":[]} | Safety |

  Szenario: Ein neuer echter Index hält Karten ausschließlich in nativen DocValues für den Fetch bereit
    Angenommen eine echte OpenSearchverbindung besteht
      | Index | Locale | Currency | Market |
      | shop-docvalues-mapping | de-DE | EUR | DE |
    Und der Suchindex als leere echte Fixture angelegt wird
    Dann die Suchmappingfelder die vereinbarten nativen DocValues bereitstellen
      | Path | JSON |
      | mappings.properties.product_id | {"type":"keyword"} |
      | mappings.properties.variants.properties.sku | {"type":"keyword"} |
      | mappings.properties.variants.properties.card_payload | {"type":"binary","doc_values":true} |

  Szenariogrundriss: Beide kanonischen Base64formen bewahren vollständige Karten aller Restklassen
    Angenommen eine vollständige technische Suchkarte vorbereitet ist
      | Fixture | Compliance |
      | spec/fixtures/opensearch/pool-expected.json | {"Context":{"Market":"DE","Locale":"de-DE","ValidUntil":{"Present":true,"Value":"2027-10-15T00:00:00Z"}},"ApplicableRules":[{"Code":"Energy","ValidUntil":{"Present":true,"Value":"2027-10-15T00:00:00Z"}}]} |
    Wenn das Projektionsdokument seinen gespeicherten Kartenpayload erzeugt
    Und der vollständige Kartenpayload im expliziten Base64format vorliegt
      | Format | Remainder | Mutation |
      | <Format> | <Restklasse> | none |
    Dann der bestehende Suchdecoder die vollständige Karte unverändert rekonstruiert
      | Text | Sort | Page |
      | A | relevance | 1 |
    Beispiele:
      | Format | Restklasse |
      | standard | 0 |
      | standard | 1 |
      | standard | 2 |
      | raw | 0 |
      | raw | 1 |
      | raw | 2 |

  Szenariogrundriss: Nichtkanonische vollständige Kartenpayloads sind trotz dekodierbarer Daten ungültig
    Angenommen eine vollständige technische Suchkarte vorbereitet ist
      | Fixture | Compliance |
      | spec/fixtures/opensearch/pool-expected.json | {} |
    Wenn das Projektionsdokument seinen gespeicherten Kartenpayload erzeugt
    Und der vollständige Kartenpayload im expliziten Base64format vorliegt
      | Format | Remainder | Mutation |
      | <Format> | <Restklasse> | <Mutation> |
    Und die gespeicherte Karte über den öffentlichen Suchdecoder abgefragt wird
      | Text | Sort | Page |
      | A | relevance | 1 |
    Dann das vollständige Protokollsuchergebnis entspricht
      | Code |
      | SearchUnavailable |
    Beispiele:
      | Format | Restklasse | Mutation |
      | standard | 1 | restbits |
      | standard | 2 | restbits |
      | raw | 1 | restbits |
      | raw | 2 | restbits |
      | standard | 0 | cr |
      | standard | 0 | lf |
      | raw | 1 | crlf |

  Szenariogrundriss: Ungültige Base64formen und andere JSONtypen erzeugen keinen Teilerfolg
    Angenommen ein beobachtbarer Suchprotokollorigin vorbereitet ist
      | Body |
      | {"hits":{"hits":[{"fields":{"product_id":["p-a"]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":["a-red"],"variants.card_payload":[<Payload>]}}]}}}}]}} |
    Wenn die Suche über den Protokollorigin ausgeführt wird
      | Text |
      | A |
    Dann das vollständige Protokollsuchergebnis entspricht
      | Code |
      | SearchUnavailable |
    Beispiele:
      | Fall | Payload |
      | falsches Alphabet | "_w" |
      | Länge modulo vier gleich eins | "A" |
      | unvollständiges Padding | "Zg=" |
      | überzähliges Padding | "Zg===" |
      | eingebettetes Padding | "Z=g=" |
      | führendes Padding | "=Zg=" |
      | nichtkanonische Restbits Standard eins | "Zh==" |
      | nichtkanonische Restbits Standard zwei | "Zm9=" |
      | nichtkanonische Restbits Raw eins | "Zh" |
      | nichtkanonische Restbits Raw zwei | "Zm9" |
      | Zahl | 1 |
      | Boolean | false |
      | Objekt | {} |
      | Array | [] |
