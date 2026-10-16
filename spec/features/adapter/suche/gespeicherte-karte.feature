# language: de
@opensearch
Funktionalität: Vollständige Modellkarten aus gespeicherten Suchfeldern
  Als Gast erhalte ich bei inkonsistent gespeicherten Kartendaten einen ehrlichen Gesamtfehler.

  Szenariogrundriss: Gespeicherte Felder erlauben keine unvollständige oder fremde Karte
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
