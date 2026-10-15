# language: de
@application @app
Funktionalität: Fachliche Activitygrenzen und Portfehler
  Als Gast erhalte ich bei ungültigem Zustand oder Portfehler keine falsche Kaufbestätigung.

  Hintergrund:
    Angenommen ein eigener Gastcart ist vollständig vorbereitet
      | UserID | CartID | SKU | Quantity | UnitPrice | StockQuantity | CartRevision | EvaluationID | Recipient | Street | PostalCode | City | Country |
      | u-a | c-a | PHONE | 1 | 10000 | 2 | 4 | e-a | Demo Gast | Testweg 1 | 12345 | Teststadt | DE |

  Szenariogrundriss: Activityvalidierung und technische Fehler
    Wenn ein autorisierter Activityaufruf verarbeitet wird
      | Operation | UserID | CartID | State | CheckoutPresent | Confirmed | EvaluationID | Quantity | Fault | Payment |
      | <operation> | <gast> | <cart> | <zustand> | <checkout> | <bestaetigt> | <bewertung> | <menge> | <fehler> | <zahlung> |
    Dann lautet das Activityergebnis
      | Code |
      | <code> |
    Beispiele:
      | operation | gast | cart | zustand | checkout | bestaetigt | bewertung | menge | fehler | zahlung | code |
      | add | u-a | c-a | checkingOut | true | true | e-a | 1 | none | invoice | OK |
      | add | u-a | c-a | checkingOut | true | true | e-a | 1 | evaluationRead | invoice | DependencyUnavailable |
      | add | u-a | c-a | checkingOut | true | true | e-a | 1 | evaluationRevision | invoice | Conflict |
      | remove | u-a | c-a | checkingOut | true | true | e-a | 1 | none | invoice | OK |
      | unknown | u-a | c-a | checkingOut | true | true | e-a | 1 | none | invoice | InvalidInput |
      | setQuantity | u-a | c-a | checkingOut | true | true | e-a | 0 | none | invoice | InvalidInput |
      | add | u-a | missing | checkingOut | true | true | e-a | 1 | none | invoice | NotFound |
      | add | u-b | c-a | checkingOut | true | true | e-a | 1 | none | invoice | NotFound |
      | address | u-a | c-a | checkingOut | false | false | e-a | 1 | none | invoice | OK |
      | evaluate | u-a | c-a | checkingOut | true | true | e-a | 1 | none | invoice | OK |
      | evaluate | u-b | c-a | checkingOut | true | true | e-a | 1 | none | invoice | NotFound |
      | evaluate | u-a | c-a | ordered | true | true | e-a | 1 | none | invoice | AlreadyOrdered |
      | evaluate | u-a | c-a | checkingOut | true | true | e-a | 1 | evaluationRead | invoice | DependencyUnavailable |
      | evaluate | u-a | c-a | checkingOut | true | true | e-a | 1 | evaluationRevision | invoice | Conflict |
      | evaluate | u-a | c-a | checkingOut | true | true | e-a | 1 | cartWrite | invoice | DependencyUnavailable |
      | confirm | u-a | c-a | checkingOut | true | true | e-a | 1 | none | invoice | OK |
      | confirm | u-b | c-a | checkingOut | true | true | e-a | 1 | none | invoice | NotFound |
      | confirm | u-a | c-a | checkingOut | true | true | e-b | 1 | none | invoice | PriceConfirmationRequired |
      | confirm | u-a | c-a | checkingOut | true | true | e-a | 1 | cartWrite | invoice | DependencyUnavailable |
      | place | u-b | c-a | checkingOut | true | true | e-a | 1 | none | invoice | NotFound |
      | place | u-a | c-a | expired | true | true | e-a | 1 | none | invoice | Expired |
      | place | u-a | c-a | checkingOut | false | false | e-a | 1 | none | invoice | PriceConfirmationRequired |
      | place | u-a | c-a | checkingOut | true | false | e-a | 1 | none | invoice | PriceConfirmationRequired |
      | place | u-a | c-a | checkingOut | true | true | e-b | 1 | none | invoice | PriceConfirmationRequired |
      | place | u-a | c-a | checkingOut | true | true | e-a | 1 | technical | creditCard | DependencyUnavailable |
      | place | u-a | c-a | checkingOut | true | true | e-a | 1 | mismatch | creditCard | Conflict |

  Szenariogrundriss: Lieferfehler erzeugen keine erfolgreiche Quittung
    Angenommen eine angenommene Bestellung wurde noch nicht versandt
      | OrderID | SKU | Purchased | Reserved | Shipped | Delivered |
      | o-a | PHONE | 2 | 2 | 0 | 0 |
    Wenn eine Liefermeldung an einer technischen Grenze scheitert
      | Fault | ShipmentID | CommandID | Status | Quantity |
      | <fehler> | s-a | cmd-a | shipped | 1 |
    Dann lautet die getrennte Lieferansicht
      | Code | State | Shipped | Delivered | Reserved | PurchaseSnapshotChanged |
      | DependencyUnavailable | accepted | 0 | 0 | 2 | false |
    Beispiele:
      | fehler |
      | fulfillmentRead |
      | fulfillmentWrite |

  Szenariogrundriss: Die Warenkorbsicht nennt erlaubte nächste Aktionen
    Wenn eine Warenkorbsicht für den aktuellen Zustand erstellt wird
      | State | Confirmed |
      | <zustand> | <bestaetigt> |
    Dann nennt die Warenkorbsicht die erlaubten Aktionen
      | Actions |
      | <aktionen> |
    Beispiele:
      | zustand | bestaetigt | aktionen |
      | open | false | change,evaluate |
      | checkingOut | false | change,evaluate,confirm |
      | checkingOut | true | change,evaluate,place |
      | ordered | true | viewOrder,startCart |
      | expired | false | startCart |
