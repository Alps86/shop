# language: de
@application @app
Funktionalität: Reine Zubehörvorschau
  Ein Rabattangebot bewertet eine Wertkopie des Carts ohne gespeicherte Änderung.

  Hintergrund:
    Angenommen ein eigener Gastcart ist vollständig vorbereitet
      | UserID | CartID | SKU | Quantity | UnitPrice | StockQuantity | CartRevision | EvaluationID | Recipient | Street | PostalCode | City | Country |
      | u-a | c-a | PHONE | 1 | 10000 | 5 | 4 | e-a | Demo Gast | Testweg 1 | 12345 | Teststadt | DE |
    Und ein Zubehörangebot mit Rabattregel vorliegt
      | MainProductID | AccessoryID | SKU | UnitPrice | Discount |
      | p-phone | p-case | CASE | 2000 | 500 |

  Szenario: Die Vorschau zeigt den echten Paarpreis und verändert keine gespeicherte Bestätigung
    Wenn eine reine Zubehörvorschau angefordert wird
      | Session | CartID | SKU | Fault |
      | a | c-a | CASE | none |
    Dann lautet die Vorschau ohne gespeicherte Cartänderung
      | Code | Total | Discount | StoredMainQuantity | StoredAccessoryQuantity | StoredRevision | StoredConfirmed | StoredActivityUnchanged |
      | OK | 11990 | 500 | 1 | 0 | 4 | true | true |

  Szenariogrundriss: Fehlende Berechtigung und technische Ausfälle verändern den Cart ebenfalls nicht
    Wenn eine reine Zubehörvorschau angefordert wird
      | Session | CartID | SKU | Fault |
      | <session> | <cart> | <sku> | <fehler> |
    Dann lautet die Vorschau ohne gespeicherte Cartänderung
      | Code | Total | Discount | StoredMainQuantity | StoredAccessoryQuantity | StoredRevision | StoredConfirmed | StoredActivityUnchanged |
      | <code> | 0 | 0 | 1 | 0 | 4 | true | true |
    Beispiele:
      | session | cart | sku | fehler | code |
      | unknown | c-a | CASE | none | NotFound |
      | a | c-missing | CASE | none | NotFound |
      | a | c-a | | none | InvalidInput |
      | a | c-a | CASE | evaluationRead | DependencyUnavailable |
