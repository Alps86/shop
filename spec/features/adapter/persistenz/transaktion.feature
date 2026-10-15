# language: de
@runtime @postgres
Funktionalität: Aktuelle Daten innerhalb der Kauftransaktion
  Preisbestätigung und Reservierung werden unter echten Datenbanksperren erneut geprüft.

  Hintergrund:
    Angenommen ein kohärenter Demodatensatz in echter PostgreSQL vorbereitet ist
      | DatasetVersion | ProductID | SKU | Name | Price | StockQuantity |
      | demo-v1 | p-phone | PHONE | Synthetisches Testmodell | 10000 | 2 |
    Wenn der Datensatz veröffentlicht wird
    Und ein Gast einen Warenkorb mit einer Position dauerhaft anlegt
      | SKU | Quantity |
      | PHONE | 1 |
    Und der Gast den vollständigen Checkout bestätigt
      | Recipient | Street | PostalCode | City | Country |
      | Demo Gast | Testweg 1 | 12345 | Teststadt | DE |

  Szenariogrundriss: Aktuelle Preise und Bestände entscheiden unabhängig von Produktstammdaten
    Wenn der Bestand atomar geändert wird
      | SKU | CommandID | Revision | SetPrice | Price | SetStock | Stock |
      | PHONE | inventory-a | 1 | <preis> | <betrag> | <bestand> | <anzahl> |
    Und der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved | Stock | ProjectionRevision | ProductRevision | ReturnedCart | Confirmation |
      | <code> | <orders> | <reserviert> | <stock> | <projektion> | 1 | <cart> | <bestaetigt> |
    Beispiele:
      | preis | betrag | bestand | anzahl | code | orders | reserviert | stock | projektion | cart | bestaetigt |
      | true | 11000 | false | 0 | PriceConfirmationRequired | 0 | 0 | 2 | 2 | true | false |
      | false | 0 | true | 3 | OK | 1 | 1 | 3 | 3 | false | false |
      | false | 0 | true | 0 | Unavailable | 0 | 0 | 0 | 2 | true | true |

  Szenariogrundriss: Liefermeldungen verbrauchen reservierte Ware genau einmal
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Und eine bestätigte Lieferung dauerhaft verarbeitet wird
      | ShipmentID | CommandID | Status | SKU | Quantity |
      | shipment-a | ship-a | <status> | PHONE | 1 |
    Und eine bestätigte Lieferung dauerhaft verarbeitet wird
      | ShipmentID | CommandID | Status | SKU | Quantity |
      | shipment-a | ship-a | <status> | PHONE | 1 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved | Stock |
      | OK | 1 | <reserviert> | <bestand> |
    Beispiele:
      | status | reserviert | bestand |
      | packed | 1 | 2 |
      | shipped | 0 | 1 |

  Szenariogrundriss: Bereits gespeicherte Mengenbefehle werden vor der aktuellen Revision gelesen
    Wenn derselbe Mengenbefehl nach dem Commit erneut ausgeführt wird
      | SKU | Quantity |
      | PHONE | <menge> |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved |
      | <code> | 0 | 0 |
    Beispiele:
      | menge | code |
      | 1 | OK |
      | 2 | Conflict |

  Szenario: Prozessaufträge sind von CDC getrennt und bleiben bei Zielausfall offen
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Und die getrennte Prozessoutbox ausgeliefert wird
      | FailureKind |
      | EnsureOrder |
    Dann lautet die dauerhafte Prozessbeobachtung
      | Code | CartCount | OrderCount | Pending | SearchCheckpoints |
      | DependencyUnavailable | 1 | 0 | 5 | 0 |
    Wenn die getrennte Prozessoutbox ausgeliefert wird
      | FailureKind |
      | none |
    Dann lautet die dauerhafte Prozessbeobachtung
      | Code | CartCount | OrderCount | Pending | SearchCheckpoints |
      | OK | 1 | 1 | 0 | 0 |
    Wenn die getrennte Prozessoutbox ausgeliefert wird
      | FailureKind |
      | none |
    Dann lautet die dauerhafte Prozessbeobachtung
      | Code | CartCount | OrderCount | Pending | SearchCheckpoints |
      | OK | 1 | 1 | 0 | 0 |

  Szenariogrundriss: Jeder fehlgeschlagene Transaktionsschritt rollt den Kauf vollständig zurück
    Angenommen ein echter SQL-Schreibfehler gezielt vorbereitet ist
      | Table | Operation | MatchField | MatchValue | SQLState |
      | <tabelle> | <operation> | <feld> | <wert> | <sqlstate> |
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved | Stock | ProjectionRevision |
      | <code> | 0 | 0 | 2 | 1 |
    Beispiele:
      | tabelle | operation | feld | wert | sqlstate | code |
      | records | INSERT | kind | inventory | P0001 | DependencyUnavailable |
      | projection_versions | UPDATE | | | P0001 | DependencyUnavailable |
      | outbox | INSERT | | | P0001 | DependencyUnavailable |
      | orders | INSERT | | | P0001 | DependencyUnavailable |
      | carts | UPDATE | | | P0001 | DependencyUnavailable |
      | process_outbox | INSERT | | | P0001 | DependencyUnavailable |
      | commands | INSERT | | | P0001 | DependencyUnavailable |
      | orders | INSERT | | | 23505 | Conflict |

  Szenariogrundriss: Bestellungen und lokale Verwaltungsansichten lesen denselben unveränderlichen Kauf
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | ID |
      | <operation> | <id> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | OwnOrder |
      | <code> | <eigen> |
    Beispiele:
      | operation | id | code | eigen |
      | order | current | OK | true |
      | order | missing | NotFound | false |
      | adminOrder | current | OK | true |
      | adminOrder | missing | NotFound | false |
      | cartOrder | current | OK | true |
      | cartOrder | missing | NotFound | false |

  Szenariogrundriss: Der Ablauf prüft die echte zuletzt gespeicherte Aktivität erneut
    Wenn der persistente Ablauf unter erneuter Sperre geprüft wird
      | ID | Hours |
      | <id> | <stunden> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | State | Confirmed |
      | <code> | <zustand> | <bestaetigt> |
    Beispiele:
      | id | stunden | code | zustand | bestaetigt |
      | current | 167 | OK | checkingOut | true |
      | current | 168 | OK | expired | false |
      | missing | 168 | NotFound | | false |

  Szenariogrundriss: Ein neuer Einkauf behält User und Fähigkeit
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Und derselbe Gast ausdrücklich einen Einkauf startet
      | CommandID |
      | <command> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | OwnUser | OriginalCart | CartCount | SessionCount |
      | OK | true | <original> | <carts> | 1 |
    Beispiele:
      | command | original | carts |
      | start-a | true | 1 |
      | start-b | false | 2 |

  Szenariogrundriss: Der direkte Kaufport prüft Besitz, Alter und genaue Bestätigungsbindung erneut
    Wenn die direkte Kauftransaktionsgrenze aufgerufen wird
      | User | Cart | Evaluation | Confirmed | ConfirmationRevisionDelta | AgeHours | EmptyCart | Authorization | CommandID |
      | <user> | <cart> | <evaluation> | <bestaetigt> | <delta> | <alter> | <leer> | absent | direct-a |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved |
      | <code> | 0 | 0 |
    Beispiele:
      | user | cart | evaluation | bestaetigt | delta | alter | leer | code |
      | foreign | current | current | true | 0 | 0 | false | NotFound |
      | own | missing | current | true | 0 | 0 | false | NotFound |
      | own | current | other | true | 0 | 0 | false | PriceConfirmationRequired |
      | own | current | current | false | 0 | 0 | false | PriceConfirmationRequired |
      | own | current | current | true | 1 | 0 | false | PriceConfirmationRequired |
      | own | current | current | true | 0 | 168 | false | Expired |
      | own | current | current | true | 0 | 0 | true | InvalidInput |

  Szenariogrundriss: SQL-Kauf akzeptiert nur die an diesen Versuch gebundene Kartenfreigabe
    Wenn die Kartenzahlung im Cart neu bestätigt wird
      | PaymentMethod | Authorization |
      | creditCard | approved |
    Und die direkte Kauftransaktionsgrenze aufgerufen wird
      | User | Cart | Evaluation | Confirmed | ConfirmationRevisionDelta | AgeHours | EmptyCart | Authorization | CommandID |
      | own | current | current | true | 0 | 0 | false | <freigabe> | direct-a |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved |
      | <code> | <orders> | <reserviert> |
    Beispiele:
      | freigabe | code | orders | reserviert |
      | approved | OK | 1 | 1 |
      | declined | PaymentDeclined | 0 | 0 |
      | absent | Conflict | 0 | 0 |
      | wrongCart | Conflict | 0 | 0 |
      | wrongRevision | Conflict | 0 | 0 |
      | wrongCommand | Conflict | 0 | 0 |
      | wrongTotal | Conflict | 0 | 0 |

  Szenario: Eine neue CommandID kann einen bestellten Cart nicht erneut kaufen
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Und der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-b | 1 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved |
      | AlreadyOrdered | 1 | 1 |

  Szenario: Abweichende Kaufpayload unter derselben CommandID ist ein Konflikt
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Und die direkte Kauftransaktionsgrenze aufgerufen wird
      | User | Cart | Evaluation | Confirmed | ConfirmationRevisionDelta | AgeHours | EmptyCart | Authorization | CommandID |
      | own | current | other | true | 0 | 0 | false | absent | place-a |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved |
      | Conflict | 1 | 1 |

  Szenariogrundriss: Inventorybefehle lesen Originalbelege vor einer inzwischen höheren Revision
    Wenn der Bestand atomar geändert wird
      | SKU | CommandID | Revision | SetPrice | Price | SetStock | Stock |
      | PHONE | inventory-a | 1 | false | 0 | true | 3 |
    Und der Bestand atomar geändert wird
      | SKU | CommandID | Revision | SetPrice | Price | SetStock | Stock |
      | PHONE | <command> | 1 | false | 0 | true | <bestand> |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | Stock | ProjectionRevision | ProductRevision |
      | <code> | 3 | 2 | 1 |
    Beispiele:
      | command | bestand | code |
      | inventory-a | 3 | OK |
      | inventory-a | 4 | Conflict |
      | inventory-b | 3 | Conflict |

  Szenariogrundriss: Ungültige Inventoryänderung hinterlässt Preis und Projektion unverändert
    Wenn der Bestand atomar geändert wird
      | SKU | CommandID | Revision | SetPrice | Price | SetStock | Stock |
      | PHONE | inventory-a | 1 | <preis> | <betrag> | <bestand> | <anzahl> |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | Stock | ProjectionRevision |
      | <code> | 2 | 1 |
    Beispiele:
      | preis | betrag | bestand | anzahl | code |
      | true | -1 | false | 0 | InvalidMoney |
      | false | 0 | true | -1 | InvalidInput |

  Szenariogrundriss: Die Liefertransaktion validiert den Fortschritt erneut unter ihrer Sperre
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Und die dauerhafte Liefertransaktionsgrenze direkt aufgerufen wird
      | ShipmentID | CommandID | LookupCommandID | Status | SKU | Quantity |
      | shipment-a | ship-a | <lookup> | <status> | PHONE | <menge> |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved | Stock |
      | <code> | 1 | 1 | 2 |
    Beispiele:
      | lookup | status | menge | code |
      | missing | shipped | 1 | InvalidInput |
      | ship-a | invalid | 1 | InvalidTransition |
      | ship-a | shipped | 3 | InvalidQuantity |

  Szenariogrundriss: Ein SQL-Fehler während der Lieferung rollt Bestandsverbrauch und Fortschritt zurück
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Und ein echter SQL-Schreibfehler gezielt vorbereitet ist
      | Table | Operation | MatchField | MatchValue | SQLState |
      | <tabelle> | <operation> | <feld> | <wert> | P0001 |
    Und die dauerhafte Liefertransaktionsgrenze direkt aufgerufen wird
      | ShipmentID | CommandID | LookupCommandID | Status | SKU | Quantity |
      | shipment-a | ship-a | ship-a | shipped | PHONE | 1 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved | Stock |
      | DependencyUnavailable | 1 | 1 | 2 |
    Beispiele:
      | tabelle | operation | feld | wert |
      | records | INSERT | kind | inventory |
      | outbox | INSERT | | |
      | orders | UPDATE | | |
      | commands | INSERT | | |

  Szenariogrundriss: Der Cartwriter prüft Originalpayload vor der aktuellen Sperrrevision
    Wenn der Cartschreibport eine typisierte Originaländerung erhält
      | CommandID | ExpectedRevision | Quantity |
      | <command> | 1 | <quantity> |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved |
      | <code> | 0 | 0 |
    Beispiele:
      | command | quantity | code |
      | add-a | 1 | OK |
      | add-a | 2 | Conflict |
      | new-a | 1 | Conflict |

  Szenariogrundriss: Eine Bestätigungswiederholung prüft die ursprüngliche Bewertung im dauerhaften Beleg
    Wenn dieselbe Bestätigung nach dem Commit erneut ausgeführt wird
      | CommandID | Evaluation |
      | confirm-a | <evaluation> |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved |
      | <code> | 0 | 0 |
    Beispiele:
      | evaluation | code |
      | original | OK |
      | other | Conflict |

  Szenariogrundriss: Ein fehlgeschlagener Cartwrite rollt Prozessauftrag und Beleg zusammen zurück
    Angenommen ein echter SQL-Schreibfehler gezielt vorbereitet ist
      | Table | Operation | MatchField | MatchValue | SQLState |
      | <tabelle> | <operation> | | | P0001 |
    Wenn der Cartschreibport eine typisierte Originaländerung erhält
      | CommandID | ExpectedRevision | Quantity |
      | new-a | 3 | 1 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved |
      | DependencyUnavailable | 0 | 0 |
    Beispiele:
      | tabelle | operation |
      | carts | UPDATE |
      | process_outbox | INSERT |
      | commands | INSERT |

  Szenariogrundriss: Fehlende aktuelle Kaufabhängigkeiten erzeugen weder Order noch Reservierung
    Angenommen folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | <sql> |
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code |
      | <code> |
    Beispiele:
      | sql | code |
      | DROP TABLE $schema.commands | DependencyUnavailable |
      | DELETE FROM $schema.sku_registry WHERE sku='PHONE' | NotFound |
      | DELETE FROM $schema.records WHERE kind='inventory' | NotFound |
      | DELETE FROM $schema.records WHERE kind='ruleset' | NotFound |
      | UPDATE $schema.records SET payload=jsonb_set(payload,'{Compliance,Context,VerificationStatus}','"pending"') WHERE kind='product' | ComplianceIncomplete |

  Szenario: Der Fehler beim Speichern einer neuen Preisbestätigung bleibt atomar
    Wenn der Bestand atomar geändert wird
      | SKU | CommandID | Revision | SetPrice | Price | SetStock | Stock |
      | PHONE | inventory-a | 1 | true | 12000 | false | 0 |
    Und ein echter SQL-Schreibfehler gezielt vorbereitet ist
      | Table | Operation | MatchField | MatchValue | SQLState |
      | carts | UPDATE | | | P0001 |
    Und der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved | Stock |
      | DependencyUnavailable | 0 | 0 | 2 |

  Szenariogrundriss: Inventoryänderung und Projektionsereignis rollen gemeinsam zurück
    Angenommen ein echter SQL-Schreibfehler gezielt vorbereitet ist
      | Table | Operation | MatchField | MatchValue | SQLState |
      | <tabelle> | <operation> | <feld> | <wert> | P0001 |
    Wenn der Bestand atomar geändert wird
      | SKU | CommandID | Revision | SetPrice | Price | SetStock | Stock |
      | PHONE | inventory-a | 1 | true | 12000 | false | 0 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved | Stock | ProjectionRevision |
      | DependencyUnavailable | 0 | 0 | 2 | 1 |
    Beispiele:
      | tabelle | operation | feld | wert |
      | records | INSERT | kind | inventory |
      | outbox | INSERT | | |

  Szenario: Ein beschädigter Prozessauftrag wird vor dem Versand sichtbar abgelehnt
    Angenommen folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | UPDATE $schema.process_outbox SET envelope=jsonb_set(envelope,'{CartID}','[]') |
    Wenn die getrennte Prozessoutbox ausgeliefert wird
      | FailureKind |
      | none |
    Dann lautet die dauerhafte Prozessbeobachtung
      | Code | CartCount | OrderCount | Pending | SearchCheckpoints |
      | DependencyUnavailable | 0 | 0 | 4 | 0 |

  Szenario: Ein fehlgeschlagenes Prozessack lässt den Auftrag für dieselbe dauerhafte Zielkennung offen
    Angenommen ein echter SQL-Schreibfehler gezielt vorbereitet ist
      | Table | Operation | MatchField | MatchValue | SQLState |
      | process_outbox | UPDATE | | | P0001 |
    Wenn die getrennte Prozessoutbox ausgeliefert wird
      | FailureKind |
      | none |
    Dann lautet die dauerhafte Prozessbeobachtung
      | Code | CartCount | OrderCount | Pending | SearchCheckpoints |
      | DependencyUnavailable | 1 | 0 | 4 | 0 |

  Szenario: Eine beschädigte Bestellzeile wird nicht als gültige Adminliste ausgegeben
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Und folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | UPDATE $schema.orders SET payload=jsonb_set(payload,'{Items}','[{"Quantity":"broken"}]') |
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | Page |
      | adminOrders | 1 |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code |
      | DependencyUnavailable |

  Szenario: Eine ausverkaufte Wunschvariante liefert verfügbare Varianten desselben Produkts zurück
    Angenommen ein Produkt mit gemischten Pflichtdatenvarianten vorbereitet wird
      | MainStatus | AdditionalStatus | AdditionalSKU | AdditionalPrice | AdditionalStock |
      | verified | verified | PHONE-BLUE | 10500 | 3 |
    Wenn der Datensatz veröffentlicht wird
    Und der Bestand atomar geändert wird
      | SKU | CommandID | Revision | SetPrice | Price | SetStock | Stock |
      | PHONE | inventory-a | 1 | false | 0 | true | 0 |
    Und der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | ReturnedCart | UnavailableSKU | Alternatives | AlternativeSKU |
      | Unavailable | 0 | true | PHONE | 1 | PHONE-BLUE |

  Szenario: Ein fehlender Cartbelegspeicher wird innerhalb des Schreibports sichtbar
    Angenommen folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | DROP TABLE $schema.commands |
    Wenn der Cartschreibport eine typisierte Originaländerung erhält
      | CommandID | ExpectedRevision | Quantity |
      | new-a | 3 | 1 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved |
      | DependencyUnavailable | 0 | 0 |

  Szenariogrundriss: Fehlende Lieferabhängigkeiten verursachen keinen Reservierungsverbrauch
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Und folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | <sql> |
    Und die dauerhafte Liefertransaktionsgrenze direkt aufgerufen wird
      | ShipmentID | CommandID | LookupCommandID | Status | SKU | Quantity |
      | shipment-a | ship-a | ship-a | shipped | PHONE | 1 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved | Stock |
      | <code> | <orders> | 1 | 2 |
    Beispiele:
      | sql | code | orders |
      | DROP TABLE $schema.commands | DependencyUnavailable | 1 |
      | DELETE FROM $schema.orders | NotFound | 0 |

  Szenario: Ein Ablaufauftrag kann einen bereits bestellten Cart nicht verändern
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Und der persistente Ablauf unter erneuter Sperre geprüft wird
      | ID | Hours |
      | current | 168 |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | State | Confirmed |
      | AlreadyOrdered | ordered | true |

  Szenario: Eine fehlende SKU-Produktzuordnung verhindert auch die Inventorypublikation
    Angenommen folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | DELETE FROM $schema.sku_registry WHERE sku='PHONE' |
    Wenn der Bestand atomar geändert wird
      | SKU | CommandID | Revision | SetPrice | Price | SetStock | Stock |
      | PHONE | inventory-a | 1 | true | 12000 | false | 0 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved | Stock | ProjectionRevision |
      | NotFound | 0 | 0 | 2 | 1 |

  Szenario: Eine beschädigte Positionsmenge erhält keine Reservierung
    Angenommen folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | UPDATE $schema.carts SET payload=jsonb_set(payload,'{Items,0,Quantity}','0') |
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved | Stock |
      | InvalidInput | 0 | 0 | 2 |

  Szenario: Ein beschädigter Bestellverweis wird nicht durch eine zweite Bestellung repariert
    Wenn der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Und folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | DELETE FROM $schema.orders |
    Und der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-other | 1 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved | Stock |
      | NotFound | 0 | 1 | 2 |

  Szenario: Fehlende Alternativdaten erzeugen bei Ausverkauf kein erfundenes Angebot
    Wenn der Bestand atomar geändert wird
      | SKU | CommandID | Revision | SetPrice | Price | SetStock | Stock |
      | PHONE | inventory-a | 1 | false | 0 | true | 0 |
    Und folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | DELETE FROM $schema.records WHERE kind='category' |
    Und der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 1 |
    Dann lautet die persistente Zustandsbeobachtung
      | Code | OrderCount | Reserved | Stock | Alternatives |
      | NotFound | 0 | 0 | 0 | 0 |
