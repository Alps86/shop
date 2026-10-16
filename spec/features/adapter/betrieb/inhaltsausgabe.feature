# language: de
@adapter @hugo-io
Funktionalität: Fehlertransparente Hugo-Ausgabe
  Fehler im Ausgabestrom brechen den Artefaktaufbau sichtbar ab.

  Szenariogrundriss: Ein unterbrochener Strom liefert den ursprünglichen IO-Fehler
    Angenommen eine öffentliche Hugo-Seite vorbereitet wird
      | title | description | url | layout | frame_kind | canonical | fragment |
      | Testmodell | Eigener kurzer Inhalt | /_frame/product/p-test/1/ | dynamic | product | /de-DE/produkt/test--p-test | /fragments/de-DE/produkt/test--p-test |
    Wenn die Hugo-Seite an einen abbrechenden Ausgabestrom geschrieben wird
      | Writer | SuccessfulWrites |
      | <Strom> | <Schreibvorgaenge> |
    Dann der Schreibfehler erhalten bleibt
      | Error |
      | <Fehler> |

    Beispiele:
      | Strom | Schreibvorgaenge | Fehler |
      | closed-file | 0 | os.ErrClosed |
      | interrupted | 1 | io.ErrClosedPipe |
