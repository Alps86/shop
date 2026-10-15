# language: de
@core @catalog
Funktionalität: Kontextgebundene verifizierte Pflichtangaben
  Als Käufer erhalte ich nur freigegebene Angaben für die gewählte Variante.

  Szenariogrundriss: Freigabe und Gültigkeitsgrenzen synthetischer Prüfpakete
    Angenommen ein ausdrücklich fiktives Pflichtdatenpaket ist vorbereitet
      | ProductID | SKU | Market | Locale | ValidFrom | ValidUntil | VerificationStatus | MissingField |
      | p-test | TEST | DE | de-DE | 2026-10-01T00:00:00Z | 2026-11-01T00:00:00Z | <status> | <fehlend> |
    Wenn das Angebot aufgelöst wird
      | SKU | Market | Locale | At |
      | TEST | <markt> | <sprache> | <zeit> |
    Dann lautet die Pflichtdatenfreigabe
      | Code | Purchasable | RequiredTextFallback |
      | <code> | <kaufbar> | false |
    Beispiele:
      | status | fehlend | markt | sprache | zeit | code | kaufbar |
      | verified | none | DE | de-DE | 2026-10-01T00:00:00Z | OK | true |
      | verified | none | DE | de-DE | 2026-10-31T23:59:59Z | OK | true |
      | verified | none | DE | de-DE | 2026-11-01T00:00:00Z | ComplianceIncomplete | false |
      | verified | none | DE | de-DE | 2026-09-30T23:59:59Z | ComplianceIncomplete | false |
      | pending | none | DE | de-DE | 2026-10-15T00:00:00Z | ComplianceIncomplete | false |
      | verified | Manufacturer.ElectronicAddress | DE | de-DE | 2026-10-15T00:00:00Z | ComplianceIncomplete | false |
      | verified | ResponsiblePerson | DE | de-DE | 2026-10-15T00:00:00Z | ComplianceIncomplete | false |
      | verified | GermanRequiredNotice | DE | de-DE | 2026-10-15T00:00:00Z | ComplianceIncomplete | false |
      | verified | none | FR | fr-FR | 2026-10-15T00:00:00Z | UnsupportedContext | false |

  Szenariogrundriss: Vollständige Variantenpakete und anwendbare Prüfgruppen
    Wenn ein synthetisches Prüfangebot validiert wird
      | Case | VariantOverride | RequiredFields | SuppliedFields |
      | <fall> | <override> | <pflicht> | <geliefert> |
    Dann lautet das Prüfungsergebnis
      | Code | ResolvedRequiredFields |
      | <code> | <resolved> |
    Beispiele:
      | fall | override | pflicht | geliefert | code | resolved |
      | base | absent | Manufacturer | Manufacturer | OK | Manufacturer |
      | variant | present | Manufacturer,Warnings | Manufacturer | ComplianceIncomplete |  |
      | energy | present | Class,Scale,Label,InformationSheet,Model,EPREL | Class,Scale,Label,InformationSheet,Model,EPREL | OK | Class,Scale,Label,InformationSheet,Model,EPREL |
      | energy | present | Class,Scale,Label,InformationSheet,Model,EPREL | Class,Label | ComplianceIncomplete |  |
      | chemical | present | VerifiedClassification,Supplier,Pictograms,SignalWord,Statements | unknown | ComplianceIncomplete |  |
      | nonApplicable | absent | RuleReference,Reason | RuleReference,Reason | OK | RuleReference,Reason |
      | nonApplicable | absent | RuleReference,Reason | unknown | ComplianceIncomplete |  |
      | connected | present | dataActInformation | absent | ComplianceIncomplete |  |
      | electrical | present | DisposalNotice | absent | ComplianceIncomplete |  |
      | battery | present | ApplicableBatteryNotice | absent | ComplianceIncomplete |  |

  Szenariogrundriss: Grundpreis folgt aktuellem Inventorypreis
    Wenn ein mengenbezogenes Testangebot bewertet wird
      | Price | Currency | NetContent | UnitPriceBasis |
      | <preis> | EUR | <fuellmenge> | 1:l |
    Dann lautet das Grundpreisergebnis
      | Code | UnitPrice | Unit |
      | <code> | <grundpreis> | EUR/l |
    Beispiele:
      | preis | fuellmenge | code | grundpreis |
      | 1000 | 500:ml | OK | 2000 |
      | 1250 | 500:ml | OK | 2500 |
      | 1000 | 0:ml | InvalidAttribute | absent |
      | 1000 | 500:g | InvalidUnit | absent |

  Szenariogrundriss: Synthetische Rollen und vollständiger Variantenoverride
    Angenommen ausdrücklich fiktive Lieferantenprüfdaten sind vorbereitet
      | Manufacturer.Name | Manufacturer.Street | Manufacturer.PostalCode | Manufacturer.City | Manufacturer.Country | Manufacturer.ElectronicAddress | ResponsiblePerson.Name | ResponsiblePerson.Country | ResponsiblePerson.ElectronicAddress |
      | Testhersteller außerhalb EU | Testweg 1 | 10000 | Teststadt | US | manufacturer@example.invalid | Testverantwortlicher EU | DE | responsible@example.invalid |
    Und die Variante besitzt optionale eigenständige Rollen
      | Manufacturer.Present | Manufacturer.Name | ResponsiblePerson.Present | ResponsiblePerson.Name |
      | <herstellerpraesenz> | <hersteller> | <verantwortlicherpraesenz> | <verantwortlicher> |
    Wenn die Rollen für das Angebot aufgelöst werden
      | SKU | Market | Locale |
      | TEST | DE | de-DE |
    Dann lauten die getrennten öffentlichen Rollen
      | Manufacturer.Name | ResponsiblePerson.Name |
      | <erwarteterhersteller> | <erwarteterverantwortlicher> |
    Beispiele:
      | herstellerpraesenz | hersteller | verantwortlicherpraesenz | verantwortlicher | erwarteterhersteller | erwarteterverantwortlicher |
      | false | absent | false | absent | Testhersteller außerhalb EU | Testverantwortlicher EU |
      | true | Testvariantenhersteller | true | Testvariantenverantwortlicher | Testvariantenhersteller | Testvariantenverantwortlicher |

  Szenario: Fiktives Energiepaket hat konkrete Skala und modellgebundene Dokumente
    Wenn ein ausdrücklich synthetisches Energieprüfpaket validiert wird
      | SKU | Model | EfficiencyClass | ScaleMin | ScaleMax | EPRELRegistrationNumber | LabelSKU | LabelURL | InformationSheetSKU | InformationSheetURL | DocumentLocale |
      | TEST-ENERGY | TEST-MODEL | B | G | A | synthetic-test-identifier | TEST-ENERGY | https://example.invalid/test-label.pdf | TEST-ENERGY | https://example.invalid/test-sheet.pdf | de-DE |
    Dann lautet der Prüfvertrag ohne reale Klassifizierungsbehauptung
      | Code | Class | Scale | LabelLinked | InformationSheetLinked | ClaimsRealCertification |
      | OK | B | G-A | true | true | false |

  Szenario: Dokument einer anderen SKU wird nicht als Pflichtnachweis übernommen
    Wenn ein ausdrücklich synthetisches Energiedokument zugeordnet wird
      | OfferSKU | DocumentSKU | Model | DocumentModel |
      | TEST-ENERGY | OTHER | TEST-MODEL | OTHER-MODEL |
    Dann lautet das Pflichtdatenergebnis
      | Code | Purchasable |
      | ComplianceIncomplete | false |

  Szenario: Synthetische Sicherheitshinweise bleiben sichtbare Daten
    Angenommen ein ausschließlich fiktives Prüfangebot hat ausdrücklich freigegebene Testhinweise
      | NoticeText | PictogramCode | PictogramURL | AltText | Placement | DataClassification |
      | Fiktiver Testhinweis – kein reales Produkt verwenden | TEST-PIC | https://example.invalid/test-pictogram.svg | Fiktives Testpiktogramm | listing,detail | synthetic-only |
    Wenn das öffentliche Pflichtdatenpaket gelesen wird
    Dann enthält es sichtbare Prüfhinweise unabhängig vom PDF
      | NoticeText | PictogramCode | AltText | VisibleWithoutPDF | ClaimsRealHazardClassification |
      | Fiktiver Testhinweis – kein reales Produkt verwenden | TEST-PIC | Fiktives Testpiktogramm | true | false |

  Szenariogrundriss: Rollen Kontext und Dokumente müssen zum Angebot passen
    Wenn ein öffentliches Pflichtdatenpaket mit einer Änderung geprüft wird
      | Field | Value |
      | <feld> | <wert> |
    Dann lautet das Pflichtdatenergebnis
      | Code | Purchasable |
      | <code> | <kaufbar> |
    Beispiele:
      | feld | wert | code | kaufbar |
      | SKU | OTHER | InvalidReference | false |
      | ResponsiblePerson.Name | empty | ComplianceIncomplete | false |
      | ResponsiblePerson.Country | US | ComplianceIncomplete | false |
      | Manufacturer.Country | FR | OK | true |
      | Context.Market | FR | ComplianceIncomplete | false |
      | Context.Locale | fr-FR | ComplianceIncomplete | false |
      | EnergyLabel | present | OK | true |
      | Energy.Model | empty | ComplianceIncomplete | false |
      | Energy.Class | empty | ComplianceIncomplete | false |
      | Energy.Class | invalid | ComplianceIncomplete | false |
      | Energy.ScaleMin | invalid | ComplianceIncomplete | false |
      | Energy.ScaleMax | invalid | ComplianceIncomplete | false |
      | Energy.Class | Z | ComplianceIncomplete | false |
      | Energy.LabelModel | OTHER | ComplianceIncomplete | false |
      | Energy.LabelLocale | fr-FR | ComplianceIncomplete | false |
      | Energy.LabelURL | http://example.invalid/label.pdf | ComplianceIncomplete | false |
      | Energy.InformationSheet | absent | ComplianceIncomplete | false |

  Szenariogrundriss: Mengenbezug und Preisgrenzen
    Wenn ein mengenbezogenes Testangebot bewertet wird
      | Price | Currency | NetContent | UnitPriceBasis |
      | <preis> | <waehrung> | <inhalt> | <basis> |
    Dann lautet das Grundpreisergebnis
      | Code | UnitPrice | Unit |
      | <code> | <grundpreis> | EUR/l |
    Beispiele:
      | preis | waehrung | inhalt | basis | code | grundpreis |
      | 1000 | EUR | 500:ml | 0:l | InvalidAttribute | absent |
      | 1000 | ZZZ | 500:ml | 1:l | InvalidMoney | absent |
      | 9223372036854775807 | EUR | 500:ml | 1:l | AmountOverflow | absent |

  Szenariogrundriss: Freigabemarkierungen ersetzen keine strukturierten Pflichtdaten
    Wenn ein Pflichtpaket nur eine behauptete Vollständigkeit enthält
      | RequiredFields | SuppliedFields |
      | <pflicht> | <behauptet> |
    Dann lautet das Pflichtdatenergebnis
      | Code | Purchasable |
      | ComplianceIncomplete | false |
    Beispiele:
      | pflicht | behauptet |
      | Class,Scale,Label,InformationSheet,Model,EPREL | Class,Scale,Label,InformationSheet,Model,EPREL |
      | VerifiedClassification,Supplier,Pictograms,SignalWord,Statements | VerifiedClassification,Supplier,Pictograms,SignalWord,Statements |
      | RuleReference,Reason | RuleReference,Reason |

  Szenariogrundriss: Strukturierte Chemie- und Hinweispakete sind vollständig
    Wenn eine strukturierte Pflichtdatengruppe geprüft wird
      | Case | Value |
      | <fall> | <wert> |
    Dann lautet das Pflichtdatenergebnis
      | Code | Purchasable |
      | <code> | <kaufbar> |
    Beispiele:
      | fall | wert | code | kaufbar |
      | chemical | valid | OK | true |
      | supplier | empty | ComplianceIncomplete | false |
      | identifier | empty | ComplianceIncomplete | false |
      | signal | empty | ComplianceIncomplete | false |
      | pictograms | absent | ComplianceIncomplete | false |
      | statements | absent | ComplianceIncomplete | false |
      | pictogramCode | empty | ComplianceIncomplete | false |
      | pictogramURL | http://example.invalid/pic.svg | ComplianceIncomplete | false |
      | pictogramAlt | empty | ComplianceIncomplete | false |
      | statementCode | empty | ComplianceIncomplete | false |
      | statementText | empty | ComplianceIncomplete | false |
      | dataActInformation | valid | OK | true |
      | dataActInformation | fr-FR | ComplianceIncomplete | false |
      | DisposalNotice | valid | OK | true |
      | ApplicableBatteryNotice | valid | OK | true |
      | DisposalNotice | wrongCode | ComplianceIncomplete | false |
      | unknownRequirement | valid | ComplianceIncomplete | false |

  Szenariogrundriss: Die Rollenlesesicht behauptet keine Kaufbarkeit
    Wenn Rollen ohne behauptete Kaufprüfung aufgelöst werden
      | Manufacturer | VerificationStatus | ExistingSKU | SKU |
      | Fiktiver Hersteller | pending | PHONE | <sku> |
    Dann lautet die ungeprüfte Rollenansicht
      | Code | Manufacturer | VerificationStatus |
      | <code> | <hersteller> | <status> |
    Beispiele:
      | sku | code | hersteller | status |
      | PHONE | OK | Fiktiver Hersteller | pending |
      | OTHER | InvalidReference | | |
