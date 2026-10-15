# language: de
@application @app
Funktionalität: Häufig zusammen gekauft aus Demostatistik
  Als Gast erhalte ich kompatible Empfehlungen ohne profilspezifische Rabattpreise.

  Hintergrund:
    Angenommen ausdrücklich synthetische abgeschlossene Demowarenkörbe liefern Paarstatistik
      | SourceProductID | TargetProductID | BasketCount | CategoryID | WindowStart | WindowEnd | DatasetVersion | Compatible | Available |
      | p-phone | p-case | 20 | mobile | 2026-07-01T00:00:00Z | 2026-09-29T00:00:00Z | demo-v1 | true | true |
      | p-phone | p-charger | 19 | home | 2026-07-01T00:00:00Z | 2026-09-29T00:00:00Z | demo-v1 | true | true |

  Szenariogrundriss: Begrenzte Profilwirkung
    Wenn Empfehlungen angefragt werden
      | ProductID | SKU | Profile | InterestWeight | Limit |
      | p-phone | PHONE | <profil> | <gewicht> | 3 |
    Dann lautet die öffentliche Angebotsreihenfolge
      | ProductIDs | Label | ExposedMargin | ExposedInterestWeight |
      | <reihenfolge> | Häufig zusammen gekauft | false | false |
    Beispiele:
      | profil | gewicht | reihenfolge |
      | neutral | 0 | p-case,p-charger |
      | mobile | 1 | p-case,p-charger |
      | home | 1 | p-charger,p-case |
      | coffee | 1 | p-case,p-charger |

  Szenariogrundriss: Kompatibilität und Fallback
    Wenn die Empfehlungsquelle verändert und danach abgefragt wird
      | Case | Profile |
      | <fall> | home |
    Dann lautet das Empfehlungsergebnis
      | ProductIDs | Label | Degraded |
      | <angebote> | <label> | <degradiert> |
    Beispiele:
      | fall | angebote | label | degradiert |
      | chargerUnavailable | p-case | Häufig zusammen gekauft | false |
      | chargerIncompatible | p-case | Häufig zusammen gekauft | false |
      | noStatistics | p-case,p-charger | Passendes Zubehör | false |
      | statisticsFailure | p-case,p-charger | Passendes Zubehör | true |
      | noAccessories | [] | absent | false |

  Szenariogrundriss: Statistik und Zielvarianten sind deterministisch
    Angenommen Kandidaten besitzen gleiche Paarhäufigkeit
      | ProductID | SKU | Price | BasketCount | Compatible | Available |
      | p-a | a-2 | 2000 | 5 | true | true |
      | p-a | a-1 | 2000 | 5 | true | true |
      | p-b | b-1 | 1000 | 5 | true | true |
    Wenn die Statistik validiert oder ausgewertet wird
      | Operation | Input |
      | <operation> | <eingabe> |
    Dann lautet das Statistikergebnis
      | Code | ProductIDs | SelectedSKUs |
      | <code> | <produkte> | <skus> |
    Beispiele:
      | operation | eingabe | code | produkte | skus |
      | recommend | neutral | OK | p-a,p-b | a-1,b-1 |
      | import | negativeCount | InvalidInput | [] | [] |
      | import | duplicateUnorderedPair | InvalidInput | [] | [] |
      | import | inconsistentWindow | InvalidInput | [] | [] |
      | import | unknownProduct | InvalidReference | [] | [] |

  Szenariogrundriss: Ungültige interne Interessen werden nicht als Rankinggewicht verwendet
    Wenn ein internes Empfehlungsgewicht geprüft wird
      | Weight |
      | <gewicht> |
    Dann lautet das Empfehlungsergebnis
      | Code | ProductIDs | Label | Degraded |
      | InvalidInput | [] | absent | false |
    Beispiele:
      | gewicht |
      | -0.1 |
      | 1.1 |
      | not-a-decimal |

  Szenario: Technischer Kandidatenausfall blockiert das Angebot nicht
    Wenn die Kandidatenquelle technisch ausfällt
      | Fault |
      | Candidates |
    Dann lautet das Empfehlungsergebnis
      | Code | ProductIDs | Label | Degraded |
      | OK | [] | absent | true |

  Szenario: Die günstigste Zielvariante gewinnt unabhängig von der Quellreihenfolge
    Angenommen Kandidaten besitzen gleiche Paarhäufigkeit
      | ProductID | SKU | Price | BasketCount | Compatible | Available |
      | p-a | a-expensive | 3000 | 5 | true | true |
      | p-a | a-cheap | 1000 | 5 | true | true |
      | p-a | a-middle | 2000 | 5 | true | true |
    Wenn die Statistik validiert oder ausgewertet wird
      | Operation | Input |
      | recommend | neutral |
    Dann lautet das Statistikergebnis
      | Code | ProductIDs | SelectedSKUs |
      | OK | p-a | a-cheap |

  Szenario: Vollständig gültige Paarstatistik ist publizierbar
    Angenommen Kandidaten besitzen gleiche Paarhäufigkeit
      | ProductID | SKU | Price | BasketCount | Compatible | Available |
      | p-a | a-1 | 1000 | 5 | true | true |
      | p-b | b-1 | 2000 | 5 | true | true |
    Wenn die vollständige Paarstatistik validiert wird
      | ProductIDs |
      | p-phone,p-a,p-b |
    Dann lautet das Statistikergebnis
      | Code | ProductIDs | SelectedSKUs |
      | OK | [] | [] |
