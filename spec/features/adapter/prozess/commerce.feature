# language: de
@commerce-runtime
Funktionalität: Unmittelbare Warenkorbnutzung in der vollständigen Laufzeit
  Als Gast ändere ich meinen neuen Warenkorb ohne technisches Warten auf den Prozessdispatcher.

  Szenario: Die erste Änderung folgt unmittelbar auf den dauerhaften Warenkorbstart
    Angenommen die vollständige Commerce-Laufzeit mit echten Diensten läuft
      | Fixture | Namespace | Locale | Currency | Market | OpenSearch |
      | spec/fixtures/runtime/katalog.json | default | de-DE | EUR | DE | ready |
    Wenn ein Gast unmittelbar nach dem Start seinen Warenkorb ändert
      | StartCommandID | ChangeCommandID | Operation | SKU | Quantity |
      | start | add | add | PHONE | 1 |
    Dann lautet die unmittelbare Warenkorbantwort
      | Code | SKU |
      | OK | PHONE |

  Szenario: Ein unbekannter Warenkorb wird beim Prozessstart aus dem dauerhaften Speicher abgewiesen
    Angenommen die vollständige Commerce-Laufzeit mit echten Diensten läuft
      | Fixture | Namespace | Locale | Currency | Market | OpenSearch |
      | spec/fixtures/runtime/katalog.json | default | de-DE | EUR | DE | ready |
    Wenn ein nicht gespeicherter Warenkorb direkt beim Prozessport angefragt wird
      | CartID | UserID |
      | c-never-created | u-foreign |
    Dann lautet die unmittelbare Warenkorbantwort
      | Code |
      | NotFound |
