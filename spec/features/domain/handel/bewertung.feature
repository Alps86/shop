# language: de
@spec-only @core
Funktionalität: Positionen Rabatt und Steuern
  Als Käufer erhalte ich eine nachvollziehbare Bewertung ohne doppelten Rabattabzug.

  Szenariogrundriss: Paarmengen und Versand
    Angenommen eine synthetische Rabattregel gilt für Testprodukte
      | RuleID | MainProductID | AccessoryID | Scope | Amount | Currency |
      | r-demo | p-phone | p-case | <bereich> | 500 | EUR |
    Und die bewertbaren Positionen lauten
      | SKU | ProductID | Quantity | UnitPrice | Currency |
      | PHONE | p-phone | <hauptmenge> | 10000 | EUR |
      | CASE | p-case | <zubehoermenge> | 2000 | EUR |
    Wenn der Warenkorb bewertet wird
      | ShippingMethod | Profile | TaxRate |
      | <versand> | <profil> | 19 |
    Dann lautet die Gesamtbewertung
      | Subtotal | ItemDiscountCASE | AggregateDiscount | ShippingCost | Total | Taxes | Currency |
      | <zwischen> | <rabatt> | <rabatt> | <kosten> | <gesamt> | <steuer> | EUR |
    Beispiele:
      | bereich | hauptmenge | zubehoermenge | versand | profil | zwischen | rabatt | kosten | gesamt | steuer |
      | pair | 1 | 1 | standard | neutral | 12000 | 500 | 490 | 11990 | 1914 |
      | pair | 1 | 1 | standard | mobile | 12000 | 500 | 490 | 11990 | 1914 |
      | pair | 1 | 1 | express | home | 12000 | 500 | 1490 | 12990 | 2074 |
      | pair | 2 | 1 | standard | coffee | 22000 | 500 | 490 | 21990 | 3510 |
      | pair | 2 | 2 | standard | neutral | 24000 | 1000 | 490 | 23490 | 3750 |
      | order | 2 | 2 | standard | neutral | 24000 | 500 | 490 | 23990 | 3830 |

  Szenariogrundriss: Zubehörentscheidung und Rabattgrenzen
    Angenommen eine synthetische Kombination ist bewertbar
      | MainQuantity | AccessoryQuantity | AccessoryUnitPrice | PairDiscount | Currency |
      | 1 | 1 | 2000 | <regelbetrag> | EUR |
    Wenn eine Kombinationsaktion ausgeführt wird
      | Action |
      | <aktion> |
    Dann lautet die Kombinationsbewertung
      | MainQuantity | AccessoryQuantity | Discount | AccessoryGrossAfterDiscount |
      | <haupt> | <zubehoer> | <rabatt> | <rest> |
    Beispiele:
      | regelbetrag | aktion | haupt | zubehoer | rabatt | rest |
      | 500 | removeMain | 0 | 1 | 0 | 2000 |
      | 500 | removeAccessory | 1 | 0 | 0 | 0 |
      | 500 | declineRecommendation | 1 | 1 | 500 | 1500 |
      | 3000 | evaluate | 1 | 1 | 2000 | 0 |

  Szenario: Deterministische Regelzuordnung
    Angenommen passende Rabattregeln liegen vor
      | RuleID | Scope | Amount | MainQuantity | AccessoryQuantity |
      | r-a | pair | 300 | 1 | 1 |
      | r-b | pair | 500 | 1 | 1 |
    Wenn die gemeinsame Zubehörposition bewertet wird
    Dann lautet die Rabattzuordnung
      | RuleID | AppliedAmount | AssignedAccessoryUnits |
      | r-a | 300 | 1 |
      | r-b | 0 | 0 |
