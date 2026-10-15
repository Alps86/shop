# language: de
@core @sale
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

  Szenariogrundriss: Ungültige Bewertungen werden abgewiesen
    Angenommen die Bewertungsgrundlage wird explizit festgelegt
      | SKU | ProductID | Quantity | UnitPrice | Currency | RuleID | RuleVersion | MainProductID | AccessoryID | RuleAmount | RuleCurrency | Scope |
      | <sku> | p-phone | <menge> | <preis> | <waehrung> | <regel> | <version> | <haupt> | <zubehoer> | <rabatt> | <regelwaehrung> | <bereich> |
    Wenn der Warenkorb bewertet wird
      | ShippingMethod | Profile | TaxRate |
      | <versand> | neutral | 19 |
    Dann lautet der fachliche Fehlercode
      | Code |
      | <code> |
    Beispiele:
      | sku | menge | preis | waehrung | regel | version | haupt | zubehoer | rabatt | regelwaehrung | bereich | versand | code |
      | PHONE | 1 | 100 | EUR | r-a | 1 | p-phone | p-case | 50 | EUR | pair | invalid | InvalidInput |
      | PHONE | 0 | 100 | EUR | r-a | 1 | p-phone | p-case | 50 | EUR | pair | standard | InvalidInput |
      | empty | 1 | 100 | EUR | r-a | 1 | p-phone | p-case | 50 | EUR | pair | standard | InvalidInput |
      | PHONE | 1 | -1 | EUR | r-a | 1 | p-phone | p-case | 50 | EUR | pair | standard | InvalidMoney |
      | PHONE | 2 | 9223372036854775807 | EUR | r-a | 1 | p-phone | p-case | 50 | EUR | pair | standard | AmountOverflow |
      | PHONE | 1 | 9223372036854775807 | EUR | r-a | 1 | p-phone | p-case | 50 | EUR | pair | standard | AmountOverflow |
      | PHONE | 1 | 100 | USD | r-a | 1 | p-phone | p-case | 50 | EUR | pair | standard | InvalidMoney |
      | PHONE | 1 | 100 | EUR | empty | 1 | p-phone | p-case | 50 | EUR | pair | standard | InvalidInput |
      | PHONE | 1 | 100 | EUR | r-a | 0 | p-phone | p-case | 50 | EUR | pair | standard | InvalidInput |
      | PHONE | 1 | 100 | EUR | r-a | 1 | empty | p-case | 50 | EUR | pair | standard | InvalidInput |
      | PHONE | 1 | 100 | EUR | r-a | 1 | p-phone | p-phone | 50 | EUR | pair | standard | InvalidInput |
      | PHONE | 1 | 100 | EUR | r-a | 1 | p-phone | p-case | -1 | EUR | pair | standard | InvalidMoney |
      | PHONE | 1 | 100 | EUR | r-a | 1 | p-phone | p-case | 0 | EUR | pair | standard | InvalidMoney |
      | PHONE | 1 | 100 | EUR | r-a | 1 | p-phone | p-case | 50 | USD | pair | standard | InvalidMoney |
      | PHONE | 1 | 100 | EUR | r-a | 1 | p-phone | p-case | 50 | EUR | invalid | standard | InvalidInput |

  Szenario: Bestellrabatt gilt auch bei mehreren Zubehörvarianten nur einmal
    Angenommen eine synthetische Rabattregel gilt für Testprodukte
      | RuleID | MainProductID | AccessoryID | Scope | Amount | Currency |
      | r-demo | p-phone | p-case | order | 500 | EUR |
    Und die bewertbaren Positionen lauten
      | SKU | ProductID | Quantity | UnitPrice | Currency |
      | PHONE | p-phone | 2 | 10000 | EUR |
      | CASE | p-case | 1 | 2000 | EUR |
      | CASE2 | p-case | 1 | 2000 | EUR |
    Wenn der Warenkorb bewertet wird
      | ShippingMethod | Profile | TaxRate |
      | standard | neutral | 19 |
    Dann lautet die Gesamtbewertung
      | Subtotal | ItemDiscountCASE | AggregateDiscount | ShippingCost | Total | Taxes | Currency |
      | 24000 | 500 | 500 | 490 | 23990 | 3829 | EUR |

  Szenariogrundriss: Preisbestätigung vergleicht Kaufwerte statt Prozessmetadaten
    Angenommen eine synthetische Rabattregel gilt für Testprodukte
      | RuleID | MainProductID | AccessoryID | Scope | Amount | Currency |
      | r-demo | p-phone | p-case | pair | 500 | EUR |
    Und die bewertbaren Positionen lauten
      | SKU | ProductID | Quantity | UnitPrice | Currency |
      | PHONE | p-phone | 1 | 10000 | EUR |
      | CASE | p-case | 1 | 2000 | EUR |
    Wenn der Warenkorb bewertet wird
      | ShippingMethod | Profile | TaxRate |
      | standard | neutral | 19 |
    Und eine Bewertungsänderung fachlich verglichen wird
      | Field | Value |
      | <feld> | <wert> |
    Dann lautet der Preisvergleich
      | MatchesPrices |
      | <gleich> |
    Beispiele:
      | feld | wert | gleich |
      | none | unchanged | true |
      | ID | e-other | true |
      | CartRevision | 99 | true |
      | InventoryRevision | 99 | true |
      | reverseItems | unchanged | true |
      | RuleSetVersion | changed | false |
      | Subtotal.MinorUnits | 0 | false |
      | ShippingCost.MinorUnits | 0 | false |
      | Total.MinorUnits | 0 | false |
      | Taxes.0.Amount.MinorUnits | 0 | false |
      | Taxes.0.Rate | 7 | false |
      | Items | empty | false |
      | Items.0.SKU | OTHER | false |
      | Items.0.Quantity | 2 | false |
      | Items.0.UnitPrice.MinorUnits | 1 | false |
      | Items.1.Discounts.0.Amount.MinorUnits | 1 | false |
      | Items.1.Discounts.0.RuleID | r-other | false |
      | Items.1.Discounts.0.ItemSKUs.0 | OTHER | false |
      | Discounts.0.Amount.MinorUnits | 1 | false |

  Szenario: Mehrere Zubehörvarianten behalten Positionen und ein gemeinsames Regelaggregat
    Angenommen eine synthetische Rabattregel gilt für Testprodukte
      | RuleID | MainProductID | AccessoryID | Scope | Amount | Currency |
      | r-demo | p-phone | p-case | pair | 500 | EUR |
    Und die bewertbaren Positionen lauten
      | SKU | ProductID | Quantity | UnitPrice | Currency |
      | PHONE | p-phone | 2 | 10000 | EUR |
      | CASE | p-case | 1 | 2000 | EUR |
      | CASE2 | p-case | 1 | 2000 | EUR |
    Wenn der Warenkorb bewertet wird
      | ShippingMethod | Profile | TaxRate |
      | standard | neutral | 19 |
    Dann besitzt die Bewertung getrennte Zuordnungen und ein Regelaggregat
      | RuleID | AggregateCount | AggregateAmount | AggregateSKUs | DiscountCASE | DiscountCASE2 | Total |
      | r-demo | 1 | 1000 | CASE,CASE2 | 500 | 500 | 23490 |
