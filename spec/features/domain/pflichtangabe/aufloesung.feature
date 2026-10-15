# language: de
@spec-only @core
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
      | Code | InheritSharedCompliance |
      | <code> | <erben> |
    Beispiele:
      | fall | override | pflicht | geliefert | code | erben |
      | base | absent | Manufacturer | Manufacturer | OK | true |
      | variant | present | Manufacturer,Warnings | Manufacturer | ComplianceIncomplete | false |
      | energy | present | Class,Scale,Label,InformationSheet,Model,EPREL | Class,Scale,Label,InformationSheet,Model,EPREL | OK | false |
      | energy | present | Class,Scale,Label,InformationSheet,Model,EPREL | Class,Label | ComplianceIncomplete | false |
      | chemical | present | VerifiedClassification,Supplier,Pictograms,SignalWord,Statements | unknown | ComplianceIncomplete | false |
      | nonApplicable | absent | RuleReference,Reason | RuleReference,Reason | OK | true |
      | nonApplicable | absent | RuleReference,Reason | unknown | ComplianceIncomplete | true |
      | connected | present | dataActInformation | absent | ComplianceIncomplete | false |
      | electrical | present | DisposalNotice | absent | ComplianceIncomplete | false |
      | battery | present | ApplicableBatteryNotice | absent | ComplianceIncomplete | false |

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
