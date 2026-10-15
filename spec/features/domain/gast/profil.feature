# language: de
@spec-only @core
Funktionalität: Fachliche Gastinteressen
  Als Gast verwende ich eindeutige gültige Kategorieinteressen.

  Szenariogrundriss: Userinteressen sind eindeutige begrenzte Fachwerte
    Wenn die Kategorieinteressen eines Users validiert werden
      | UserID | CategoryIDs | Weights | ExistingCategories |
      | u-a | <kategorien> | <gewichte> | mobile,home,coffee |
    Dann lautet das Interessenvalidierungsergebnis
      | Code |
      | <code> |
    Beispiele:
      | kategorien | gewichte | code |
      | mobile,home | 0,1 | OK |
      | mobile | -0.1 | InvalidInput |
      | mobile | 1.1 | InvalidInput |
      | mobile,mobile | 0.5,0.5 | InvalidInput |
      | missing | 0.5 | InvalidReference |
