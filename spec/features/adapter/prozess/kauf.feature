# language: de
@temporal
Funktionalität: Echter Gastkauf mit PostgreSQL und Temporal
  Als Gast schließe ich einen langlebigen Cartprozess mit simuliertem Zahlungsweg ab.

  Szenariogrundriss: Beide Zahlungswege erzeugen eine dauerhafte Order mit Prozessoutbox
    Angenommen ein echter Temporalclient ist verbunden
    Und PostgreSQL den vollständigen Kaufzustand hält
      | Fixture |
      | spec/fixtures/runtime/katalog.json |
    Und ein Gast einen echten Cartprozess öffnet
      | CommandID | Fault |
      | start-a | <Fault> |
    Wenn der Gast ausschließlich Workflowupdates verwendet
      | CommandID | Operation | SKU | Quantity | Recipient | Street | PostalCode | City | Country | ShippingMethod | PaymentMethod |
      | add-a | add | PHONE | 1 | | | | | | | |
      | address-a | address | | 0 | Demo Gast | Testweg 1 | 12345 | Teststadt | DE | | |
      | payment-a | payment | | 0 | | | | | | | <Payment> |
    Und der Worker vollständig neu startet
    Und der Kauf über den Workflow verbindlich abgeschlossen wird
      | ConfirmCommandID | PlaceCommandID |
      | confirm-a | place-a |
    Dann der dauerhafte Kauf und gebundene Orderprozess beobachtet werden
      | Code | OrderCount | Reserved | Total | RuleSetVersion |
      | OK | 1 | 1 | 10490 | demo-v1 |
    Und wiederholtes EnsureOrder dieselbe Ausführung erhält
      | Attempts |
      | 2 |
    Beispiele:
      | Payment | Fault |
      | invoice | none |
      | creditCard | none |
      | invoice | afterCommit |
      | creditCard | afterCommit |

  Szenario: Teillieferungen und terminale Wiederholung erhalten den Kaufsnapshot
    Angenommen ein echter Temporalclient ist verbunden
    Und PostgreSQL den vollständigen Kaufzustand hält
      | Fixture |
      | spec/fixtures/runtime/katalog.json |
    Und ein Gast einen echten Cartprozess öffnet
      | CommandID | Fault |
      | start-a | none |
    Wenn der Gast ausschließlich Workflowupdates verwendet
      | CommandID | Operation | SKU | Quantity | Recipient | Street | PostalCode | City | Country | ShippingMethod | PaymentMethod |
      | add-a | add | PHONE | 2 | | | | | | | |
      | address-a | address | | 0 | Demo Gast | Testweg 1 | 12345 | Teststadt | DE | | |
    Und der Kauf über den Workflow verbindlich abgeschlossen wird
      | ConfirmCommandID | PlaceCommandID |
      | confirm-a | place-a |
    Und Liefermeldungen durch den echten Orderworkflow laufen
      | CommandID | ShipmentID | Status | SKU | Quantity | Code | State | Shipped | Delivered |
      | ship-a | shipment-a | shipped | PHONE | 1 | OK | partiallyShipped | 1 | 0 |
      | ship-a | shipment-a | shipped | PHONE | 1 | OK | partiallyShipped | 1 | 0 |
      | deliver-a | shipment-a | delivered | PHONE | 1 | OK | partiallyShipped | 1 | 1 |
      | ship-b | shipment-b | shipped | PHONE | 1 | OK | shipped | 2 | 1 |
      | deliver-b | shipment-b | delivered | PHONE | 1 | OK | delivered | 2 | 2 |
    Und der Orderworkflow terminal beendet ist
    Dann Liefermeldungen durch den echten Orderworkflow laufen
      | CommandID | ShipmentID | Status | SKU | Quantity | Code | State | Shipped | Delivered |
      | deliver-b | shipment-b | delivered | PHONE | 1 | OK | delivered | 2 | 2 |
      | deliver-b | shipment-b | delivered | PHONE | 2 | Conflict | delivered | 2 | 2 |
    Und Liefermeldungen durch den echten Orderworkflow laufen
      | OrderID | CommandID | ShipmentID | Status | SKU | Quantity | Code | State | Shipped | Delivered |
      | o-missing | missing-a | shipment-missing | shipped | PHONE | 1 | NotFound | delivered | 2 | 2 |

  Szenario: Neue Revision während Kartenfreigabe verwirft den alten Versuch
    Angenommen ein echter Temporalclient ist verbunden
    Und PostgreSQL den vollständigen Kaufzustand hält
      | Fixture |
      | spec/fixtures/runtime/katalog.json |
    Und ein Gast einen echten Cartprozess öffnet
      | CommandID | Fault |
      | start-a | gatedPayment |
    Wenn der Gast ausschließlich Workflowupdates verwendet
      | CommandID | Operation | SKU | Quantity | Recipient | Street | PostalCode | City | Country | ShippingMethod | PaymentMethod |
      | add-a | add | PHONE | 1 | | | | | | | |
      | address-a | address | | 0 | Demo Gast | Testweg 1 | 12345 | Teststadt | DE | | |
      | payment-a | payment | | 0 | | | | | | | creditCard |
    Und eine neue Revision während der simulierten Freigabe entsteht
      | ConfirmCommandID | PlaceCommandID | ChangeCommandID | SKU | NewQuantity |
      | confirm-a | place-a | change-during-payment | PHONE | 2 |
    Dann der alte Kaufversuch keine Bestellung oder Reservierung hinterlässt
      | Code | OrderCount | Reserved | RequiresConfirmation | RetryCode |
      | Conflict | 0 | 0 | true | PriceConfirmationRequired |

  Szenario: Dauerhaft angenommene Kartenfreigabe liefert Pending und läuft weiter
    Angenommen ein echter Temporalclient ist verbunden
    Und PostgreSQL den vollständigen Kaufzustand hält
      | Fixture |
      | spec/fixtures/runtime/katalog.json |
    Und ein Gast einen echten Cartprozess öffnet
      | CommandID | Fault | PaymentDelay |
      | start-a | none | 3s |
    Wenn der Gast ausschließlich Workflowupdates verwendet
      | CommandID | Operation | SKU | Quantity | Recipient | Street | PostalCode | City | Country | ShippingMethod | PaymentMethod |
      | add-a | add | PHONE | 1 | | | | | | | |
      | address-a | address | | 0 | Demo Gast | Testweg 1 | 12345 | Teststadt | DE | | |
      | payment-a | payment | | 0 | | | | | | | creditCard |
    Und der angenommene Kauf nach dem Anfragebudget weiterläuft
      | ConfirmCommandID | PlaceCommandID | Pending |
      | confirm-a | place-a | true |
    Dann der dauerhafte Kauf und gebundene Orderprozess beobachtet werden
      | Code | OrderCount | Reserved | Total | RuleSetVersion |
      | OK | 1 | 1 | 10490 | demo-v1 |

  Szenario: Verlorener noch offener Orderprozess wird nicht als Lieferung quittiert
    Angenommen ein echter Temporalclient ist verbunden
    Und PostgreSQL den vollständigen Kaufzustand hält
      | Fixture |
      | spec/fixtures/runtime/katalog.json |
    Und ein Gast einen echten Cartprozess öffnet
      | CommandID | Fault |
      | start-a | none |
    Wenn der Gast ausschließlich Workflowupdates verwendet
      | CommandID | Operation | SKU | Quantity | Recipient | Street | PostalCode | City | Country | ShippingMethod | PaymentMethod |
      | add-a | add | PHONE | 1 | | | | | | | |
      | address-a | address | | 0 | Demo Gast | Testweg 1 | 12345 | Teststadt | DE | | |
    Und der Kauf über den Workflow verbindlich abgeschlossen wird
      | ConfirmCommandID | PlaceCommandID |
      | confirm-a | place-a |
    Und der noch offene Orderworkflow administrativ beendet wird
    Dann Liefermeldungen durch den echten Orderworkflow laufen
      | CommandID | ShipmentID | Status | SKU | Quantity | Code | State | Shipped | Delivered |
      | ship-a | shipment-a | shipped | PHONE | 1 | NotFound | accepted | 0 | 0 |

  Szenario: Veröffentlichte neue Regeln verändern keine laufende History
    Angenommen ein echter Temporalclient ist verbunden
    Und PostgreSQL den vollständigen Kaufzustand hält
      | Fixture |
      | spec/fixtures/runtime/katalog.json |
    Und ein Gast einen echten Cartprozess öffnet
      | CommandID | Fault |
      | start-a | none |
    Wenn ein neuer Regelstand nur für neue Carts veröffentlicht wird
      | Fixture | PublishedRules | CommandID |
      | spec/fixtures/runtime/katalog.json | demo-v2 | start-new |
    Und der Worker vollständig neu startet
    Dann bestehende und neue Prozesse ihre jeweiligen Versionen behalten
      | ExistingRules | NewRules |
      | demo-v1 | demo-v2 |

  Szenario: Abgelaufene Prozesse liefern terminale Fachfehler statt technischer Nichtverfügbarkeit
    Angenommen ein echter Temporalclient ist verbunden
    Und PostgreSQL den vollständigen Kaufzustand hält
      | Fixture |
      | spec/fixtures/runtime/katalog.json |
    Und ein Gast einen echten Cartprozess öffnet
      | CommandID | Fault | Age |
      | start-a | none | 168h |
    Wenn der abgelaufene Cartworkflow terminal abgeschlossen ist
    Dann die Gastfassade terminale Fachfehler ohne Workflowupdate liefert
      | Operation | Code |
      | change | Expired |
      | evaluate | Expired |
      | confirm | Expired |
      | place | Expired |

  Szenariogrundriss: Zeitgrenze, Reads, Duplikate und Fehler erhalten die richtige Aktivität
    Angenommen ein echter Temporalclient ist verbunden
    Und PostgreSQL den vollständigen Kaufzustand hält
      | Fixture |
      | spec/fixtures/runtime/katalog.json |
    Und ein Gast einen echten Cartprozess öffnet
      | CommandID | Fault |
      | start-a | none |
    Und der Gast ausschließlich Workflowupdates verwendet
      | CommandID | Operation | SKU | Quantity | Recipient | Street | PostalCode | City | Country | ShippingMethod | PaymentMethod |
      | address-a | address | | 0 | Demo Gast | Testweg 1 | 12345 | Teststadt | DE | | |
    Und eine gespeicherte Aktivitätsbasis mit Commandbeleg besteht
      | CommandID |
      | seed-a |
    Wenn eine zeitbezogene Prozessoperation auf PostgreSQL ausgeführt wird
      | Operation | AtOffset | Duplicate | Valid | CommandID |
      | <Operation> | <AtOffset> | <Duplicate> | <Valid> | change-a |
    Dann nur erfolgreiche neue Nutzeränderungen die Aktivität verschieben
      | Code | State | LastActivityOffset |
      | <Code> | <State> | <LastActivityOffset> |
    Beispiele:
      | Operation | AtOffset | Duplicate | Valid | Code | State | LastActivityOffset |
      | read | 167h59m59s | false | true | OK | checkingOut | 0s |
      | place | 168h | false | true | Expired | expired | 0s |
      | timer | 168h1s | false | true | Expired | expired | 0s |
      | change | 144h | false | true | OK | checkingOut | 144h |
      | change | 144h | true | true | OK | checkingOut | 0s |
      | change | 144h | false | false | InvalidInput | checkingOut | 0s |

  Szenario: Workerrestart nach Commit setzt denselben Prozess ohne zweite Order fort
    Angenommen ein echter Temporalclient ist verbunden
    Und PostgreSQL den vollständigen Kaufzustand hält
      | Fixture |
      | spec/fixtures/runtime/katalog.json |
    Und ein Gast einen echten Cartprozess öffnet
      | CommandID | Fault |
      | start-a | restartAfterCommit |
    Wenn der Gast ausschließlich Workflowupdates verwendet
      | CommandID | Operation | SKU | Quantity | Recipient | Street | PostalCode | City | Country | ShippingMethod | PaymentMethod |
      | add-a | add | PHONE | 1 | | | | | | | |
      | address-a | address | | 0 | Demo Gast | Testweg 1 | 12345 | Teststadt | DE | | |
    Und der Worker nach dem Kaufcommit vor der Activityantwort neu startet
      | ConfirmCommandID | PlaceCommandID | AwaitTimeout |
      | confirm-a | place-a | 20s |
    Dann der dauerhafte Kauf und gebundene Orderprozess beobachtet werden
      | Code | OrderCount | Reserved | Total | RuleSetVersion |
      | OK | 1 | 1 | 10490 | demo-v1 |
