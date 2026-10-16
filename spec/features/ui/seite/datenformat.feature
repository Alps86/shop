# language: de
@http
Funktionalität: Typisierte Produktdaten werden unverfälscht dargestellt
  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Unterschiedliche Attributtypen erscheinen aus dem DTO
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Search":{"Cards":[{"Name":"Testmodell","SKU":"RED-256","TopAttributes":{"test":<attribute>}}]}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/suche |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | <text> |
    Beispiele:
      | attribute | text |
      | {"Text":{"Present":true,"Value":{"de-DE":"Textwert"}}} | Textwert |
      | {"Measurement":{"Present":true,"Value":{"Value":"12.5","Unit":"kg"}}} | 12.5 kg |
      | {"Boolean":{"Present":true,"Value":false}} | false |

  Szenariogrundriss: Grundpreis verwendet die öffentliche Maßberechnung
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Search":{"Cards":[{"Name":"Testmodell","SKU":"RED-256","UnitPrice":{"MinorUnits":1000,"Currency":"EUR"},"Compliance":{"NetContent":{"Present":true,"Value":{"Value":"0.5","Unit":"l"}},"UnitPriceBasis":<basis>}}]}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/suche |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Body | <text> |
    Beispiele:
      | basis | status | text |
      | {"Present":true,"Value":{"Value":"1","Unit":"l"}} | 200 | 20,00 EUR / 1 l |
      | {"Present":false} | 200 | Testmodell |
      | {"Present":true,"Value":{"Value":"1","Unit":"kg"}} | 503 | |

  Szenario: Fehlerhafte vom Port gewählte SKU erzeugt eine klare Fehleransicht
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Detail":{"ChosenSKU":"MISSING"}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/produkt/testmodell--p-phone |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 422 |

  Szenario: Energieinformationen zeigen Klasse vollständige Skala und beide Dokumente
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Detail":{"Cards":[{"SKU":"BLUE-128","Compliance":{"Context":{"VerificationStatus":"verified"},"EnergyLabel":{"Present":true,"Value":{"EfficiencyClass":"B","ScaleMin":"A","ScaleMax":"G","LabelDocuments":[{"URL":"/synthetic-label.pdf"}],"InformationSheets":[{"URL":"/synthetic-sheet.pdf"}]}}}}]}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/produkt/testmodell--p-phone |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | BodyAll | Energieeffizienzklasse <strong>B</strong>,Skala A bis G,href="/synthetic-label.pdf",Energielabel,href="/synthetic-sheet.pdf",Produktdatenblatt |

  Szenario: Pendingkarten behaupten auch im Listing keine Freigabe
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Search":{"Cards":[{"Name":"Ungeprüftes Demoprodukt","SKU":"PENDING","AvailableQuantity":0,"Compliance":{"Context":{"VerificationStatus":"pending"}}}]}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/suche |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Pflichtangaben werden geprüft |
      | Absent | Rechtskonformität bestätigt |
