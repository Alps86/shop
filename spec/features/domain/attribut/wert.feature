# language: de
@spec-only @core
Funktionalität: Attributpräsenz und Familientypen
  Als Betreiber unterscheide ich fehlende Angaben von vorhandenen Null- und Falschwerten.

  Szenariogrundriss: Typgerechte Präsenz
    Angenommen eine Attributdefinition ist veröffentlicht
      | Type | Level | Options | Required | Minimum | Maximum |
      | <typ> | variant | red,blue | <pflicht> | <minimum> | <maximum> |
    Wenn ein Attributwert validiert wird
      | Location | Present | Number.Present | Number.Value | Boolean.Present | Boolean.Value | Options | ForeignBranch.Present |
      | <ebene> | <vorhanden> | <zahlpraesenz> | <zahl> | <boolpraesenz> | <bool> | <auswahl> | <fremd> |
    Dann lautet das Attributergebnis
      | Code | CanonicalOptions |
      | <code> | <kanonisch> |
    Beispiele:
      | typ | pflicht | minimum | maximum | ebene | vorhanden | zahlpraesenz | zahl | boolpraesenz | bool | auswahl | fremd | code | kanonisch |
      | number | false | -1 | 1 | variant | true | true | 0 | false | false | [] | false | OK | [] |
      | boolean | false | absent | absent | variant | true | false | 0 | true | false | [] | false | OK | [] |
      | boolean | false | absent | absent | variant | false | false | 0 | false | false | [] | false | OK | [] |
      | boolean | true | absent | absent | variant | false | false | 0 | false | false | [] | false | InvalidAttribute | [] |
      | number | true | 1 | 10 | variant | true | true | 0 | false | false | [] | false | InvalidAttribute | [] |
      | number | false | absent | absent | variant | true | false | 0 | false | false | [] | false | InvalidAttribute | [] |
      | boolean | false | absent | absent | product | true | false | 0 | true | false | [] | false | InvalidAttribute | [] |
      | select | false | absent | absent | variant | true | false | 0 | false | false | red | false | OK | red |
      | select | false | absent | absent | variant | true | false | 0 | false | false | red,blue | false | InvalidAttribute | [] |
      | select | false | absent | absent | variant | true | false | 0 | false | false | [] | false | InvalidAttribute | [] |
      | multiSelect | false | absent | absent | variant | true | false | 0 | false | false | red,blue | false | OK | blue,red |
      | multiSelect | false | absent | absent | variant | true | false | 0 | false | false | red,red | false | InvalidAttribute | [] |
      | multiSelect | false | absent | absent | variant | true | false | 0 | false | false | green | false | InvalidAttribute | [] |
      | multiSelect | false | absent | absent | variant | true | false | 0 | false | false | [] | false | InvalidAttribute | [] |
      | boolean | false | absent | absent | variant | true | false | 0 | true | false | [] | true | InvalidAttribute | [] |

  Szenariogrundriss: Exakte Messwerte und Optionals
    Wenn ein Fachwert validiert wird
      | Type | Input | TargetUnit |
      | <typ> | <eingabe> | <einheit> |
    Dann lautet das Wertergebnis
      | Code | Canonical |
      | <code> | <ergebnis> |
    Beispiele:
      | typ | eingabe | einheit | code | ergebnis |
      | Measurement | 500:ml | l | OK | 0.5:l |
      | Measurement | 1:TB | GB | OK | 1000:GB |
      | Measurement | 1:kg | l | InvalidUnit | absent |
      | Measurement | 0:ml | l | InvalidAttribute | absent |
      | OptionalTime | Present:false,Value:zero | none | OK | absent |
      | OptionalTime | Present:false,Value:2026-10-15T00:00:00Z | none | InvalidInput | absent |
      | OptionalTime | Present:true,Value:2026-10-15T00:00:00Z | none | OK | 2026-10-15T00:00:00Z |
      | RequiredField | null | none | InvalidInput | absent |
      | OptionalReplacement | null | none | OK | absent |

  Szenariogrundriss: Text- und Messzweige entsprechen der Definition
    Wenn ein strukturierter Attributwert validiert wird
      | DefinitionType | Level | CanonicalUnit | Text.Present | Text.de-DE | Measurement.Present | Measurement.Value | Measurement.Unit | Number.Present |
      | <typ> | product | ml | <textpraesenz> | <text> | <messpraesenz> | <wert> | <einheit> | <zahlpraesenz> |
    Dann lautet das strukturierte Attributergebnis
      | Code | ResolvedText | CanonicalMeasurement |
      | <code> | <aufgeloest> | <messwert> |
    Beispiele:
      | typ | textpraesenz | text | messpraesenz | wert | einheit | zahlpraesenz | code | aufgeloest | messwert |
      | text | true | Testbeschreibung | false | 0 | empty | false | OK | Testbeschreibung | absent |
      | text | true | empty | false | 0 | empty | false | InvalidAttribute | absent | absent |
      | text | false | empty | false | 0 | empty | false | InvalidAttribute | absent | absent |
      | measurement | false | empty | true | 0.5 | l | false | OK | absent | 500:ml |
      | measurement | false | empty | false | 0 | empty | false | InvalidAttribute | absent | absent |
      | measurement | false | empty | true | 500 | ml | true | InvalidAttribute | absent | absent |

  Szenariogrundriss: Familydefinition ist selbst gültig und eindeutig
    Wenn eine Attributdefinition publiziert wird
      | Type | Level | Codes | OptionCodes | Minimum | Maximum |
      | <typ> | <ebene> | <codes> | <optionen> | <minimum> | <maximum> |
    Dann lautet die Definitionsvalidierung
      | Code |
      | <code> |
    Beispiele:
      | typ | ebene | codes | optionen | minimum | maximum | code |
      | number | product | weight | [] | 0 | 10 | OK |
      | number | product | weight | [] | 10 | 0 | InvalidAttribute |
      | number | product | weight,weight | [] | 0 | 10 | InvalidAttribute |
      | select | variant | color | red,red | absent | absent | InvalidAttribute |
      | unknown | product | weight | [] | absent | absent | InvalidAttribute |
      | number | unknown | weight | [] | absent | absent | InvalidAttribute |

  Szenario: Marketingfallback ist von Pflichttextfallback getrennt
    Angenommen ein Marketingtext besitzt Übersetzungen
      | de-DE | de | fr |
      | Deutscher Demotext | Basissprachtext | Texte de démonstration |
    Wenn eine Textauflösung angefragt wird
      | RequestedLocale | Kind |
      | de-AT | marketing |
    Dann lautet das Textergebnis
      | Text | SourceLocale |
      | Basissprachtext | de |
