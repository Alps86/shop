# Öffentliche Application- und Adapterports

Diese Operationen sind der verbindliche semantische Vertrag; SDK-Typen bleiben außerhalb. Technische Methoden nehmen `context.Context`, konkrete Commands/Queries und liefern Ergebnis plus `error`. Kein generisches Repositoryframework. Readoperationen verändern nichts. Domainmethoden bekommen reine Fakten/Werte, niemals Repositorys.

| Port / Operation | Eingabe → Ausgabe / beobachtbare Wirkung |
| --- | --- |
| CatalogReader.Product | ProductID, SKU optional, Context → aufgelöstes ProductDetail mit Variant, Pflichtdaten, aktuellen Inventorywerten; NotFound/InvalidVariant/ComplianceIncomplete |
| CatalogReader.Category | CategoryID, Context → Baum/Breadcrumb/Filterdefinition/Topmerkmale; keine Suchindexabhängigkeit für Navigation |
| GuestService.Open | vorhandene Sessionfähigkeit optional → UserID, neue opaque Sessionfähigkeit falls nötig, CartID optional; Defaultprofil neutral |
| GuestService.SelectProfile | Session, ProfileCode → eigene Userinteressen aktualisiert, Cart unverändert, Bestätigung nicht durch Interessen ungültig |
| CartReader.Get | Session, CartID → eigener CartView einschließlich aktueller Bewertung/zulässiger Aktionen; NotFound/Expired, Read verlängert Aktivität nicht |
| AdminOrderReader.List/Get | Ausschließlich lokaler Verwaltungsadapter, Page oder OrderID → paginierte Orderübersicht bzw. Kauf-/Lieferdetail ohne vorgespielte Gastsession; keine öffentliche Route |
| CartService.Change | Session, CartID, CommandID, ExpectedRevision, Add/SetQuantity/Remove/Shipping/Payment/Address → CartView mit neuer Revision/Bewertung; Versionskonflikt ohne Teilwirkung |
| CartService.Evaluate | Session, CartID → Evaluation mit Positionen/Summen/Inventoryrevisionen; keine Reservierung |
| CartService.Confirm | Session, CartID, EvaluationID, ExpectedRevision → bestätigter Checkoutzustand; exakte aktuelle Bewertung erforderlich |
| CheckoutService.Place | Session, CartID, CommandID, EvaluationID → Accepted mit Prozessreferenz oder vorhandene OrderID; PriceConfirmationRequired/Unavailable/Expired/Conflict |
| OrderReader.Get | Session, OrderID → unveränderliche Kaufansicht + getrennte Lieferansicht; fremde Kennung NotFound |
| SearchPort.Search | Text, CategoryID optional, FilterSet, Sort, Page, Context, interne Interests → SearchResult aus ProductCards, Facets, Suggestions; SearchUnavailable |
| RecommendationPort.Recommend | ProductID, gewählte SKU, Context, interne Interests, Limit → sortierte kompatible verfügbare Angebote; leere Liste + degraded bei technischem Ausfall |
| InventoryService.Change | SKU, ExpectedRevision, neuer Preis optional/neuer Bestand optional, CommandID → neue Revision und Outbox; InvalidMoney/BelowReserved/Conflict |
| FulfillmentService.Report | OrderID, ShipmentID, Status, Lines, CommandID → bestätigter separater Status; InvalidQuantity/InvalidTransition/Conflict |
| PurchaseTransaction.Commit | CartID, UserID, CommandID, bestätigte Evaluation, PaymentAuthorization optional, Now → OrderID; atomare aktuelle Prüfung/Reservierung/Order/Cartabschluss/Outbox |
| ProcessPort.EnsureCart | CartID, gebundene Definition/Regeln → dauerhaft vorhandener Cartprozess; idempotent auf WorkflowID |
| ProcessPort.UpdateCart | CartID, CommandID, Command → bestätigtes Application-Ergebnis oder vorläufig Pending |
| ProcessPort.EnsureOrder | OrderID, gebundene Versionen → dauerhaft vorhandener Orderprozess; idempotent |
| ProcessPort.ReportDelivery | OrderID, CommandID, FulfillmentUpdate → quittierte Aufnahme, tatsächlicher Status über OrderReader |
| ProjectionPort.Apply | vollständiges ProductProjectionEvent → versioniertes Upsert/Tombstone; Duplicate/Stale als erfolgreicher Noop |
| ProjectionPort.Rebuild | neuer Indexname, Snapshotgrenze → vollständiger Index + Catchup/atomarer Aliaswechsel |

## Gemeinsame Datenformen

`RequestContext{Locale Locale, Currency Currency, Market CountryCode}` trennt Anzeige und Absatzmarkt. `CommandID` und `EvaluationID` eigene benannte ID-Typen nach [Kennungsregeln](catalog.md). `ExpectedRevision uint64` schützt Lost Updates. Sessionfähigkeit ist ausschließlich technisches Zugangstoken, kein Domainfeld/öffentlicher URL-Parameter. Cart-/OrderViews geben weder Token noch interne Marge/Interessen aus.

`ProductCard` enthält ProductID, URL, Name, gewählte SKU, Bild/Alttext, UnitPrice, Availability, VariantCount, TopAttributes und erforderliche ComplianceSummary. `ProductDetail` ergänzt vollständige Variantenoptionen, Galerie, Attribute, Compliance und Zubehör. `CartView` enthält State/Revision/Items/Evaluation und erlaubte nächste Aktionen. `SearchResult` enthält TotalProducts, Cards, Facets mit Productzählungen, Page/PageSize, Suggestions und Degraded. `OrderView` enthält Kauf-Snapshot und separat FulfillmentView mit versandten/gelieferten Mengen.

`FilterSet` erlaubt Brand[]/AttributeOptions[]/numerische Intervalle/PriceRange; validiert gegen Category.FilterAttributes. `Sort` ist `relevance`, `priceAsc`, `priceDesc`, `nameAsc`. Page ab 1, PageSize fest 12, maximal Page 100; ungültige Operatoren/Werte InvalidFilter. Transportfragen wie JSON/HTML sind Adapterverantwortung.

## Fehler und Transaktionen

Fachlich: InvalidInput, InvalidReference, DuplicateSKU, InvalidAttribute, InvalidMoney, UnsupportedContext, ComplianceIncomplete, Conflict, PriceConfirmationRequired, Unavailable, Expired, AlreadyOrdered. Technisch: DependencyUnavailable, DeadlineExceeded. Errors liefern stabile Codes/Feldpfade, keine SQL-/SDK-Details oder Stringvergleichslogik. Unbekannte Ressource und unzulässiger Gastzugriff öffentlich beide NotFound.

HTTP ordnet Eingabefehler 422, NotFound 404, Konflikt/Preisänderung 409, Expired 410, technische Nichtverfügbarkeit 503 zu; erfolgreiche Erstellung 201, synchrone Änderung 200, persistierte asynchrone Annahme 202 mit Location. Retry nur bei technischen Fehlern und derselben CommandID. Gleiches ID+Payload liefert dasselbe Ergebnis, andere Payload zur selben ID Conflict. Idempotenzdatensätze bleiben mindestens für Lebenszeit des Cart bzw. der Order erhalten. Fachliche Transaktionsgrenzen stehen unter [Abschluss](../06-laufzeit/checkout.md) und [Projektion](../06-laufzeit/projection.md).

Alle IO-Ports respektieren Kontextabbruch. Öffentliche Readrequests Budget 500 ms, Suche 250 ms, interne Fragmentaufrufe 150 ms; Commands 2 s bis dauerhafter Annahme. Ein HTTP-Timeout widerruft keinen bereits committed Command: UI liest Status und wiederholt nur dieselbe CommandID.
