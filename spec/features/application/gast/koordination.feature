# language: de
@application @app
Funktionalität: Autorisierte Gastfassade und Prozessgrenze
  Szenariogrundriss: Gastzugriffe bewahren Besitz und melden Abhängigkeiten
    Angenommen ein Gastzugriff mit folgenden Grenzen besteht
      | Owner | State | IssuedAt | Fault |
      | <owner> | <state> | <issued> | <fault> |
    Wenn der öffentliche Gastvertrag ausgeführt wird
      | Operation | Session | Argument | UserID | EvaluationID | CommandID | ExpectedRevision | Change | SKU | Quantity |
      | <operation> | <session> | <argument> | u-forged | e-a | request-a | 0 | add | PHONE | 1 |
    Dann ist das beobachtete Zugriffsergebnis
      | Code | ProcessCalls | StoredProfile |
      | <code> | 0 | <profile> |
    Beispiele:
      | operation | session | argument | owner | state | issued | fault | code | profile |
      | open | | | u-a | open | 2026-10-15T00:00:00Z | none | OK | neutral |
      | open | a | | u-a | open | 2026-10-15T00:00:00Z | none | OK | neutral |
      | open | missing | | u-a | open | 2026-10-15T00:00:00Z | none | NotFound | neutral |
      | open | a | | u-a | open | 2026-09-15T00:00:00Z | none | NotFound | neutral |
      | open | | | u-a | open | 2026-10-15T00:00:00Z | guestOpen | DependencyUnavailable | neutral |
      | open | a | | u-a | open | 2026-10-15T00:00:00Z | guestResolve | DependencyUnavailable | neutral |
      | profile | a | coffee | u-a | open | 2026-10-15T00:00:00Z | none | OK | coffee |
      | profile | a | invalid | u-a | open | 2026-10-15T00:00:00Z | none | InvalidInput | neutral |
      | profile | missing | coffee | u-a | open | 2026-10-15T00:00:00Z | none | NotFound | neutral |
      | profile | a | coffee | u-a | open | 2026-10-15T00:00:00Z | guestProfile | DependencyUnavailable | neutral |
      | cart | a | c-a | u-a | open | 2026-10-15T00:00:00Z | none | OK | neutral |
      | cart | a | missing | u-a | open | 2026-10-15T00:00:00Z | none | NotFound | neutral |
      | cart | a | c-a | u-b | open | 2026-10-15T00:00:00Z | none | NotFound | neutral |
      | cart | a | c-a | u-a | open | 2026-10-15T00:00:00Z | cartRead | DependencyUnavailable | neutral |
      | cart | a | c-a | u-a | open | 2026-10-07T00:00:00Z | none | Expired | neutral |
      | cart | a | c-a | u-a | ordered | 2026-10-07T00:00:00Z | none | OK | neutral |
      | order | a | o-a | u-a | open | 2026-10-15T00:00:00Z | none | OK | neutral |
      | order | a | missing | u-a | open | 2026-10-15T00:00:00Z | none | NotFound | neutral |
      | order | a | o-a | u-b | open | 2026-10-15T00:00:00Z | none | NotFound | neutral |
      | order | missing | o-a | u-a | open | 2026-10-15T00:00:00Z | none | NotFound | neutral |
      | order | a | o-a | u-a | open | 2026-10-15T00:00:00Z | orderRead | DependencyUnavailable | neutral |
      | order | a | o-a | u-a | open | 2026-10-15T00:00:00Z | fulfillmentRead | DependencyUnavailable | neutral |
      | start | a | start-a | u-a | open | 2026-10-15T00:00:00Z | none | OK | neutral |
      | start | missing | start-a | u-a | open | 2026-10-15T00:00:00Z | none | NotFound | neutral |
      | start | a | start-a | u-a | open | 2026-10-15T00:00:00Z | cartStart | DependencyUnavailable | neutral |

  Szenariogrundriss: Unberechtigte Befehle erreichen keinen Prozess
    Angenommen ein Gastzugriff mit folgenden Grenzen besteht
      | Owner | State | IssuedAt | Fault |
      | u-b | open | 2026-10-15T00:00:00Z | none |
    Wenn der öffentliche Gastvertrag ausgeführt wird
      | Operation | Session | Argument | UserID | EvaluationID | CommandID | ExpectedRevision | Change | SKU | Quantity |
      | <operation> | a | c-a | u-forged | e-a | request-a | 0 | add | PHONE | 1 |
    Dann ist das beobachtete Zugriffsergebnis
      | Code | ProcessCalls | ContainsSession |
      | NotFound | 0 | false |
    Beispiele:
      | operation |
      | evaluate |
      | confirm |
      | change |
      | place |

  Szenario: Auswertung übergibt nur die aufgelöste Identität
    Angenommen ein Gastzugriff mit folgenden Grenzen besteht
      | Owner | State | IssuedAt | Fault |
      | u-a | open | 2026-10-15T00:00:00Z | none |
    Wenn der öffentliche Gastvertrag ausgeführt wird
      | Operation | Session | Argument | UserID | EvaluationID | CommandID | ExpectedRevision | Change | SKU | Quantity |
      | evaluate | a | c-a | u-forged | e-a | request-a | 0 | add | PHONE | 1 |
    Dann ist das beobachtete Zugriffsergebnis
      | Code | ProcessCalls | ProcessUser | ContainsSession |
      | OK | 1 | u-a | false |

  Szenariogrundriss: Autorisierte Änderung und Bestätigung gelangen durch den Prozesseingang
    Angenommen ein eigener Gastcart ist vollständig vorbereitet
      | UserID | CartID | SKU | Quantity | UnitPrice | StockQuantity | CartRevision | EvaluationID | Recipient | Street | PostalCode | City | Country |
      | u-a | c-a | PHONE | 1 | 10000 | 2 | 4 | e-a | Demo Gast | Testweg 1 | 12345 | Teststadt | DE |
    Wenn der öffentliche Gastvertrag ausgeführt wird
      | Operation | Session | Argument | UserID | EvaluationID | CommandID | ExpectedRevision | Change | SKU | Quantity |
      | <operation> | a | c-a | u-forged | e-a | request-a | 4 | add | PHONE | 1 |
    Dann ist das beobachtete Zugriffsergebnis
      | Code | ProcessCalls | ProcessUser | ContainsSession | CartID | CartUser | CartQuantity | Confirmed |
      | OK | 1 | u-a | false | c-a | u-a | <quantity> | <confirmed> |
    Beispiele:
      | operation | quantity | confirmed |
      | change | 2 | false |
      | confirm | 1 | true |

  Szenario: Neue Gastidentität erhält ein eigenes neutrales Profil
    Angenommen ein Gastzugriff mit folgenden Grenzen besteht
      | Owner | State | IssuedAt | Fault |
      | u-a | open | 2026-10-15T00:00:00Z | none |
    Wenn der öffentliche Gastvertrag ausgeführt wird
      | Operation | Session | Argument |
      | open | | |
    Dann ist das beobachtete Zugriffsergebnis
      | Code | GuestUser | GuestSession | Profile | SessionCount |
      | OK | u-new | new-session | neutral | 2 |

  Szenario: Neuer Warenkorb bewahrt die Gastidentität
    Angenommen ein Gastzugriff mit folgenden Grenzen besteht
      | Owner | State | IssuedAt | Fault |
      | u-a | ordered | 2026-10-15T00:00:00Z | none |
    Wenn der öffentliche Gastvertrag ausgeführt wird
      | Operation | Session | Argument |
      | start | a | start-a |
    Dann ist das beobachtete Zugriffsergebnis
      | Code | CartID | CartUser | SessionCount |
      | OK | c-new | u-a | 1 |


  Szenario: Fehlender Bestellverweis wird nicht als lesbarer Kauf ausgegeben
    Angenommen ein Gastzugriff mit folgenden Grenzen besteht
      | Owner | State | IssuedAt | Fault |
      | u-a | ordered | 2026-10-15T00:00:00Z | orderForCart |
    Wenn der öffentliche Gastvertrag ausgeführt wird
      | Operation | Session | Argument |
      | cart | a | c-a |
    Dann ist das beobachtete Zugriffsergebnis
      | Code | ProcessCalls |
      | DependencyUnavailable | 0 |
