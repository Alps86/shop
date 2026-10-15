# language: de
@runtime @postgres
Funktionalität: Maßgebliche PostgreSQL-Persistenz
  Als Gast behalte ich Warenkorb und unveränderliche Käufe in der echten Datenbank.

  Szenario: Ein vollständiger Datensatz wird atomar publiziert
    Angenommen ein kohärenter Demodatensatz in echter PostgreSQL vorbereitet ist
      | DatasetVersion | ProductID | SKU | Name | Price | StockQuantity |
      | demo-v1 | p-phone | PHONE | Synthetisches Testmodell | 10000 | 2 |
    Wenn der Datensatz veröffentlicht wird
    Und die maßgebliche Katalogsicht gelesen wird
    Dann lautet das persistente Importergebnis
      | Code | ProductCount |
      | OK | 1 |

  Szenario: Wiederholter Abschluss erzeugt genau eine Order und Reservierung
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
    Und der Gast den bestätigten Kauf wiederholt abschließt
      | CommandID | Attempts |
      | place-a | 2 |
    Dann lautet die dauerhafte Kaufbeobachtung
      | Code | OrderCount | Reserved | Total | CartState |
      | OK | 1 | 1 | 10490 | ordered |

  Szenario: Zwei echte SQL-Transaktionen konkurrieren um die letzte Einheit
    Angenommen ein kohärenter Demodatensatz in echter PostgreSQL vorbereitet ist
      | DatasetVersion | ProductID | SKU | Name | Price | StockQuantity |
      | demo-v1 | p-phone | PHONE | Synthetisches Testmodell | 10000 | 1 |
    Wenn der Datensatz veröffentlicht wird
    Und zwei getrennte Gäste dieselbe letzte Einheit bestätigt haben
      | SKU | Quantity | Recipient | Street | PostalCode | City | Country |
      | PHONE | 1 | Demo Gast | Testweg 1 | 12345 | Teststadt | DE |
    Und beide Gäste gleichzeitig abschließen
    Dann lautet das echte Konkurrenzresultat
      | OK | Unavailable | Orders |
      | 1 | 1 | 1 |
