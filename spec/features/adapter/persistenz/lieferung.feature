# language: de
@runtime @postgres
Funktionalität: Atomare Lieferungen mehrerer gekaufter Varianten
  Reservierungsverbrauch und Lieferstatus bleiben gemeinsam dauerhaft, der Kauf unverändert.

  Hintergrund:
    Angenommen ein kohärenter Demodatensatz in echter PostgreSQL vorbereitet ist
      | DatasetVersion | ProductID | SKU | Name | Price | StockQuantity |
      | demo-v1 | p-phone | PHONE | Synthetisches Testmodell | 10000 | 2 |
    Und ein Produkt mit gemischten Pflichtdatenvarianten vorbereitet wird
      | MainStatus | AdditionalStatus | AdditionalSKU | AdditionalPrice | AdditionalStock |
      | verified | verified | PHONE-BLUE | 10500 | 3 |
    Wenn der Datensatz veröffentlicht wird
    Und ein Gast einen Warenkorb mit einer Position dauerhaft anlegt
      | SKU | Quantity |
      | PHONE | 1 |
    Und eine weitere Position tatsächlich hinzugefügt wird
      | SKU | Quantity | CommandID |
      | PHONE-BLUE | 1 | add-blue |
    Und der Gast den vollständigen Checkout bestätigt
      | Recipient | Street | PostalCode | City | Country |
      | Demo Gast | Testweg 1 | 12345 | Teststadt | DE |
    Und der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Und der öffentliche Kaufsnapshot vor der Lieferung gesichert wird

  Szenario: Eine gemeinsame wiederholte Lieferung verbraucht jede reservierte Variante genau einmal
    Wenn eine gemeinsame Lieferung für mehrere Positionen verarbeitet wird
      | ShipmentID | CommandID | Status | SKU | Quantity |
      | shipment-a | ship-a | shipped | PHONE-BLUE | 1 |
      | shipment-a | ship-a | shipped | PHONE | 1 |
    Und eine gemeinsame Lieferung für mehrere Positionen verarbeitet wird
      | ShipmentID | CommandID | Status | SKU | Quantity |
      | shipment-a | ship-a | shipped | PHONE-BLUE | 1 |
      | shipment-a | ship-a | shipped | PHONE | 1 |
    Dann Einzelbestände und unveränderlicher Kauf übereinstimmen
      | Code | SKU | Stock | Reserved | OrderUnchanged |
      | OK | PHONE | 1 | 0 | true |
      | OK | PHONE-BLUE | 2 | 0 | true |

  Szenario: Eine unvollständige Produktprojektion rollt auch die Inventoryänderung zurück
    Angenommen folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | DELETE FROM $schema.records WHERE kind='inventory' AND id='PHONE-BLUE' |
    Wenn der Bestand atomar geändert wird
      | SKU | CommandID | Revision | SetPrice | Price | SetStock | Stock |
      | PHONE | inventory-a | 2 | true | 12000 | false | 0 |
    Dann Einzelbestände und unveränderlicher Kauf übereinstimmen
      | Code | SKU | Stock | Reserved | OrderUnchanged |
      | NotFound | PHONE | 2 | 1 | true |

  Szenario: Ein fehlender Reservierungsbeleg erzeugt keinen erfolgreichen Versand
    Angenommen folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | UPDATE $schema.records SET payload=jsonb_set(payload,'{Reservations}','[]') WHERE kind='inventory' AND id='PHONE' |
    Wenn eine gemeinsame Lieferung für mehrere Positionen verarbeitet wird
      | ShipmentID | CommandID | Status | SKU | Quantity |
      | shipment-a | ship-a | shipped | PHONE-BLUE | 1 |
      | shipment-a | ship-a | shipped | PHONE | 1 |
    Dann Einzelbestände und unveränderlicher Kauf übereinstimmen
      | Code | SKU | Stock | Reserved | OrderUnchanged |
      | InvalidQuantity | PHONE | 2 | 0 | true |
      | InvalidQuantity | PHONE-BLUE | 3 | 1 | true |

  Szenario: Eine fehlende Inventoryzeile verbraucht auch keine andere Position
    Angenommen folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | DELETE FROM $schema.records WHERE kind='inventory' AND id='PHONE' |
    Wenn eine gemeinsame Lieferung für mehrere Positionen verarbeitet wird
      | ShipmentID | CommandID | Status | SKU | Quantity |
      | shipment-a | ship-a | shipped | PHONE-BLUE | 1 |
      | shipment-a | ship-a | shipped | PHONE | 1 |
    Dann Einzelbestände und unveränderlicher Kauf übereinstimmen
      | Code | SKU | Stock | Reserved | OrderUnchanged |
      | NotFound | PHONE-BLUE | 3 | 1 | true |
