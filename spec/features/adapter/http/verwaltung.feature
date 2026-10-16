# language: de
@http
Funktionalität: Verwaltungscommands über Applicationports
  Als lokaler Betreiber ändere ich ausschließlich freigegebene Demoaktionen.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Preis und Bestand verwenden Revision und CommandID
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Listener | CSRFSource | Form |
      | <methode> | /backend/inventory | backend | <quelle> | <formular> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
    Beispiele:
      | methode | quelle | formular | status |
      | GET | | | 405 |
      | POST | | sku=RED-256 | 403 |
      | POST | /backend | sku=RED-256&revision=3&commandID=inventory-a&price=12500 | 303 |
      | POST | /backend | sku=RED-256&revision=3&commandID=inventory-a&stock=5 | 303 |
      | POST | /backend | sku=RED-256&revision=3&commandID=inventory-a&price=12500&stock=5 | 303 |
      | POST | /backend | sku=RED-256&revision=invalid&price=12500 | 422 |
      | POST | /backend | sku=RED-256&revision=3&price=invalid | 422 |
      | POST | /backend | sku=RED-256&revision=3&stock=invalid | 422 |

  Szenariogrundriss: Liefermeldung umfasst explizite SKU und Teilmenge
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Listener | CSRFSource | Form |
      | <methode> | /backend/lieferschritt | backend | <quelle> | <formular> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
    Beispiele:
      | methode | quelle | formular | status |
      | GET | | | 405 |
      | POST | | orderID=order-a | 403 |
      | POST | /backend | orderID=order-a&shipmentID=shipment-a&commandID=delivery-a&sku=RED-256&quantity=1&status=shipped | 303 |
      | POST | /backend | orderID=order-a&shipmentID=shipment-a&commandID=delivery-a&sku=RED-256&quantity=invalid&status=shipped | 422 |

  Szenariogrundriss: Reale Commandfehler bleiben unverändert sichtbar
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | <fehler> |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Listener | CSRFSource | Form |
      | POST | <url> | backend | /backend | <formular> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Header:Cache-Control | private,no-store |
    Beispiele:
      | fehler | url | formular | status |
      | Inventory:Conflict | /backend/inventory | sku=RED-256&revision=3&price=12500&commandID=inventory-a | 409 |
      | ReportDelivery:InvalidInput | /backend/lieferschritt | orderID=order-a&sku=RED-256&quantity=3&status=shipped&commandID=delivery-a | 422 |
      | ReportDelivery:InternalError | /backend/lieferschritt | orderID=order-a&sku=RED-256&quantity=1&status=shipped&commandID=delivery-a | 503 |

  Szenariogrundriss: Verwaltungslistener bedient keine Shoprouten
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Listener |
      | <methode> | <url> | backend |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
    Beispiele:
      | methode | url | status |
      | GET | /de-DE/ | 404 |
      | POST | /backend | 405 |
      | POST | /backend/bestellung/order-a | 405 |

  Szenario: Inventarformular reicht den vollständigen Originalauftrag weiter
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Listener | CSRFSource | Form |
      | POST | /backend/inventory | backend | /backend | sku=RED-256&revision=3&commandID=inventory-original&price=12500&stock=5 |
    Dann der verwendete Applicationport erhält
      | InventorySKU | InventoryRevision | InventoryCommandID | InventoryPrice | InventoryStock |
      | RED-256 | 3 | inventory-original | 12500 | 5 |

  Szenario: Lieferformular behält Order und CommandID
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Listener | CSRFSource | Form |
      | POST | /backend/lieferschritt | backend | /backend | orderID=order-a&shipmentID=shipment-a&commandID=delivery-original&sku=RED-256&quantity=1&status=shipped |
    Dann der verwendete Applicationport erhält
      | DeliveryOrder | DeliveryCommandID |
      | order-a | delivery-original |
