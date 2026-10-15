# Implementierter Fachkern und konkrete Applicationgrenzen

Der Kern liegt in genau einem Rootmodul. [Typenkatalog](type-model.md), [Katalog](catalog.md), [Handel](commerce.md) und [Ports](ports.md) bleiben fachlich maßgeblich. Die konkreten englischen Go-Signaturen stehen in `internal/domain/{value,catalog,sale}/type.go` und `internal/application/type.go`; technische SDKs, HTTP und SQL gehören ausschließlich in Adapter.

## Öffentliche Fachtypen

| Package | Öffentlicher Vertrag |
| --- | --- |
| `value` | Spezifische Kennungen, `IdentifierSyntax.Validate`, registrierte Locale-/Land-/Währungsparser, exakte `Decimal`- und `Money`-Operationen, `Optional[T].Validate` und `RequiredJSON.Validate`. `Decimal.Compare` und `Scaled` liefern Fehler statt stiller Ersatzwerte. `Code(error)` ist die technische Error-As-Grenze für stabile Fehlercodes. |
| `catalog` | Versionierte Family/Category und `Catalog.Publish`, Product/Variant, typisierte Attribute mit expliziter Präsenz, kanonische exakte Measurements, aufgelöste Herstellerrollen und vollständige kontextgebundene Pflichtdaten. Eine Behauptung vollständiger Daten ersetzt keine strukturierten Pflichtangaben. |
| `sale` | Inventory mit Reservierungsverbrauch, Cart/Checkout mit Revision und Ablauf, positionsbezogene Bewertung/Rabatte/Steuern, Userinteressen und unveränderliche Orderkopien. `Fulfillment` hält Liefermeldungen getrennt vom Kauf. |

`CartEvaluation.MatchesPrices` vergleicht tatsächliche Positionspreise, Mengen, Rabatte, Versand, Steuer und Regelstand. Gegenläufige Preisänderungen mit identischer Gesamtsumme sind eine Änderung. EvaluationID, Auswertungszeit und reine Inventoryrevisionen sind keine Preisänderung. Dieser Vertrag ist auch im echten PurchaseTransaction-Adapter zu verwenden.

## Temporaler Eingang und Activities

Rohe Sessionfähigkeiten gehören weder in Domainwerte noch in Temporal-History oder Outboxpayloads. Die `Service`-Fassaden `Change(ctx, session, command)`, `Evaluate(ctx, session, cartID)`, `Confirm(ctx, session, command)` und `Place(ctx, session, command)` lösen die Fähigkeit auf, prüfen Cartbesitz und überschreiben eine vom Aufrufer behauptete UserID. Erst danach rufen sie den jeweiligen ProcessPort auf. Ein bereits bestellter Cart liest einen erneuten Place-Versuch direkt über ApplyPlace/PurchaseTransaction; ein terminal geschlossener Workflow muss dafür keinen neuen Update annehmen.

| ProcessPort | Application-Activity |
| --- | --- |
| `UpdateCart(ChangeCommand)` | `ApplyChange(ChangeCommand)` |
| `EvaluateCart(EvaluateCommand)` | `ApplyEvaluate(EvaluateCommand)` |
| `ConfirmEvaluation(ConfirmCommand)` | `ApplyConfirm(ConfirmCommand)` |
| `PlaceOrder(PlaceCommand)` | `ApplyPlace(PlaceCommand)` |

Diese Commands enthalten ausschließlich autorisierte Fachkennungen und Eingaben, keine Session. Activities prüfen Cart.UserID erneut gegen Command.UserID. Der Temporaladapter darf Activities über diese Methoden ausführen, die öffentliche HTTP-Grenze muss die Fassaden verwenden. `EnsureCart`/`EnsureOrder` binden Definition und Regelversion; `ReportDelivery` enthält ebenfalls keine Fähigkeit.

`StartCart` beginnt ausdrücklich einen neuen Einkauf unter derselben Gastidentität über CartStarter. CartStarter persistiert Cart und EnsureCart-Outboxauftrag atomar und behandelt CommandID idempotent. Ein terminaler Cart wird dadurch niemals wieder geöffnet. `OrderStore.ForCart` erschließt den Bestellverweis für einen weiterhin lesbaren bestellten Cart.

## Persistenz- und Prozesssemantik

CartWriter speichert Cart, erwartete Revision, CommandID und nötige Prozessaufträge atomar. `CartWrite.Change` beziehungsweise `Confirmation` enthält die ursprüngliche autorisierte Nutzereingabe für den Idempotenzvergleich. Auswertungszeit, neu erzeugte Auswertungskennung und LastActivityAt dürfen dabei nicht als geänderte Nutzereingabe gelten. Ein nackter Evaluate-Aufruf erhält seinen eigenen EvaluationID-basierten Speichervorgang und verlängert keine Aktivität.

`Service.Commands` ist im Runtimebetrieb ein dauerhafter `CartCommandReader`. ApplyChange und ApplyConfirm lesen nach Besitzprüfung mit ReplayChange beziehungsweise ReplayConfirm zunächst den gespeicherten Beleg: identische Originalpayload liefert das ursprüngliche Ergebnis trotz inzwischen höherer Revision, abweichende Payload unter derselben CommandID ergibt Conflict. Die reine Core-Suite kann den Reader bewusst weglassen, um isolierte Domain-/Writergrenzen zu prüfen.

ApplyChange validiert nach diesem Beleglookup die Domainänderung, bewertet einen nichtleeren Cart neu und speichert erst danach. Bei einer fehlgeschlagenen Bewertung entsteht keine Teiländerung. Ein leerer Cart besitzt keine kaufbare Bewertung. CartWriter gibt `NewCartView` mit erlaubten Aktionen zurück. Confirm bindet die genaue Bewertung, Place prüft die Bestätigung und bindet eine simulierte Kartenfreigabe an CartID, Revision, Total und CommandID.

NewPurchase erzeugt Order und bestellten Cart nach einmaliger fachlicher Prüfung als zusammengehörige Werte. PurchaseTransaction bleibt die einzige atomare Abschlussgrenze: aktuelle Preise/Pflichtdaten/Bestände, Reservierungen, Orderkopie, Cartabschluss, Idempotenz und Outbox gemeinsam. `PurchaseResult.Cart` enthält bei neuer Preisbestätigung die aktualisierte Bewertung und ungültige alte Bestätigung. `UnavailableSKU` und `Alternatives` liefern ausschließlich verfügbare Alternativen desselben Produkts; Auswahl erfolgt weiterhin aktiv durch den Gast. Fehlerergebnisse werden mit ihren DTOs unverändert weitergegeben. Nach erfolgreichem Commit gehören Retry und bestehende Order zur Transaktionsgrenze, ohne erneute simulierte Freigabe.

FulfillmentStore.Save persistiert validierten Fortschritt und den zugehörigen Reservierungs-/Bestandsverbrauch atomar. `packed` verbraucht nichts, `shipped` verbraucht die reservierte Menge, `delivered` setzt zuvor versandte Mengen als geliefert. Der Kauf bleibt unverändert. Dieser IO-Vertrag benötigt im Runtimepaket zusätzlich echte PostgreSQL-/Temporalnachweise.

## Sicht- und Suchverträge

`NewCartView` liefert zustandsabhängig `change`, `evaluate`, `confirm`, `place`, `viewOrder` und `startCart`; abgelaufene Carts geben Expired mit einer Sicht für den expliziten Neustart zurück. OrderView trennt Purchase und Fulfillment. CatalogReader liefert ProductDetail und CategoryView einschließlich Breadcrumb, Kindern und kuratierten Definitionen. VariantReader.ProductForSKU löst eine SKU über die maßgebliche Produktzuordnung auf und liefert dieselbe ProductDetail-Sicht mit expliziter Auswahl. Product.ResolveRoles löst Rollen und geerbte Pflichtdaten auch für pending-Sichten auf; ResolveCompliance prüft weiterhin den Kontext und die Freigabe. Nicht freigegebene Varianten bleiben sichtbar, besitzen aber verfügbare Menge null. AdminCatalogReader liefert den kleinen maßgeblichen Demokatalog einschließlich Family/Category/Regeln/Inventory; AdminOrderReader paginiert Bestellungen. Diese lokalen Verwaltungsports benötigen keine vorgespielte Gastsession und lesen keine Stammdaten aus OpenSearch.

SearchService validiert RequestContext, Seite/Sortierung, kuratierte SearchSchema-Attribute, eindeutige gültige Optionen und konsistente exakte Zahlen-/Preisintervalle. Beliebige Operatorstrings sind nicht Teil des typisierten Intervallvertrags. SearchPort führt erst anschließend die echte Suche aus; deren Ausfall wird SearchUnavailable. Das öffentliche Resultat besitzt zwölf Modelle pro Seite und keine internen Margen oder Interessen. `Service.GetInterests(ctx, session)` authentifiziert die Gastfähigkeit und liest über `InterestStore.ForUser(ctx, UserID)` ausschließlich die gespeicherten CategoryInterests dieses Users. HTTP bildet aus ProfileCode keine Gewichte und nimmt keine fremde UserID für diese Abfrage entgegen.

RecommendationSource liefert kompatibilitätsprüfbare, publizierte und pflichtdatenfreigegebene Kandidaten. RecommendationService filtert unverfügbare Kandidaten, wählt die günstigste Variante und ordnet anhand der synthetischen Paarstatistik mit begrenztem Profilgewicht. Quellenfehler ergeben eine leere degradierte Empfehlung beziehungsweise den deterministischen Zubehörfallback. Service.PreviewAccessory authentifiziert Session und Cartbesitz, prüft Ablauf und fügt die vorgeschlagene SKU einmal auf einer unabhängigen Cartwertkopie hinzu. Die anschließende echte Evaluatorbewertung persistiert weder Cart noch Aktivität oder Bestätigung; eine Annahme verwendet erst den normalen Change-Command. Es gibt keine produktive In-Memory-Suche oder Ersatzpersistenz.

## Nachweisgrenzen

Die rekursiven Strict-Godog-Suites für `value`, `catalog`, `sale` und `app` führen die aktiven öffentlichen Verträge aus; der zentrale Gherkinparser liest zusätzlich alle noch offenen Runtime-/Frontendverträge. Explizite Testdoubles prüfen Application- und Transaktionsportsemantik. Sie belegen keine echte Datenbanksperre, dauerhafte Outbox, CDC, OpenSearch-Query oder Temporal-History. Diese zusätzlichen Nachweise bleiben den tatsächlichen Adaptern und der Gesamtdemo vorbehalten. Ausführungsberichte und aktuelle PR-Head-Nachweise gehören in Issue und PR, nicht in den Architekturvertrag.
