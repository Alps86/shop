# language: de
@performance-capture
Funktionalität: Ursprüngliche Messantworten eindeutig und ohne private Header zuordnen
  Die Abschlussreihenfolge paralleler curl-Aufrufe bestimmt nicht die Antwortdatei.
  Fehlerdiagnosen übernehmen ausschließlich die vereinbarten öffentlichen Header.

  Szenario: Umgekehrte Abschlussreihenfolge bewahrt Dateizuordnung und tatsächlichen Status
    Dann kalte curl-Antworten unabhängig von ihrer Abschlussreihenfolge zugeordnet werden
      | JSON                                                                                                                                                                              | File | Error              |
      | {"urlnum":1,"filename_effective":"/tmp/capture-contract/1","url_effective":"http://127.0.0.1:8080/de-DE/kategorie/test","http_code":503}                                              | 1    | Messroute HTTP 503 |
      | {"urlnum":0,"filename_effective":"/tmp/capture-contract/0","url_effective":"http://127.0.0.1:8080/de-DE/kategorie/test","http_code":200}                                              | 0    | <nil>              |

  Szenario: Private Wireheader gelangen nicht in Diagnoseartefakte
    Dann Diagnoseheader ausschließlich erlaubte öffentliche Werte enthalten
      | Name          | Value             |
      | Content-Type  | text/html         |
      | Cache-Control | no-store          |
      | Retry-After   | 1                 |
      | Cookie        | private-session   |
      | Authorization | private-token     |
      | Set-Cookie    | private-response  |

  Szenario: Originalfehler und Body bleiben nach temporärem Cleanup bestehen
    Dann ursprüngliche kalte Fehlerantworten vor dem Cleanup dauerhaft erhalten bleiben
      | Status | Body                       |
      | 503    | <html>unavailable</html>   |
      | 200    | <html>missing</html>       |
