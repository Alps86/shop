# language: de
@spec-only @frontend
Funktionalität: Schlichte Shopseiten nach EP-Grundkonzept
  Als Gast navigiere ich durch einen eigenen Shop mit vertrauter Seitenstruktur.

  Szenariogrundriss: Seitenrenderer nutzt explizite öffentliche DTOs
    Angenommen der Renderer erhält ein vorbereitetes Ansichtsobjekt
      | DTO | State | Origin |
      | <dto> | <zustand> | synthetic-test-data |
    Wenn die Seite als vollständiges HTML gerendert wird
      | Route |
      | <route> |
    Dann sind die erwarteten Komponenten vorhanden
      | Components | AbsentComponents | Reference |
      | <komponenten> | dealerPicker,accountLogin,realPayment | <referenz> |
    Beispiele:
      | dto | zustand | route | komponenten | referenz |
      | HomeView | normal | /de-DE/ | logo,navigation,search,cartLink,campaign,categoryLinks,offers,brands,service,footer | E1 |
      | CategoryView | normal | /de-DE/kategorie/smartphones--c-phone | breadcrumb,categoryTree,intro,filters,sort,gridToggle,pagination,cards | E2,E4 |
      | ProductDetail | normal | /de-DE/produkt/test--p-test | gallery,brand,name,sku,colorSelector,storageSelector,grossPrice,shippingInfo,buyAction,technicalDetails,manufacturer,accessories | E3,E5 |
      | CartView | empty | /de-DE/warenkorb | emptyCart,shoppingLink,simulationBanner | E6 |
      | CartView | populated | /de-DE/warenkorb | positions,quantityForm,removeAction,evaluation,accessoryDecision,checkoutLink,simulationBanner | own-contract |
      | CheckoutView | normal | /de-DE/kasse | addressForm,shippingChoice,paymentChoice,evaluation,confirmAction,demoPlaceAction,simulationBanner | own-contract |
      | OrderView | accepted | /de-DE/bestellung/o-a | purchaseSnapshot,deliveryProgress,simulationBanner | own-contract |
      | BackendCatalogView | normal | /backend | catalogReadOnly,priceAction,stockAction,simulationBanner | own-contract |
      | BackendOrderView | normal | /backend/bestellung/o-a | snapshot,shipmentForm,partialShipmentAction,simulationBanner | own-contract |
      | EditorialView | normal | /de-DE/service | markdownContent,contact,footer | E1 |

  Szenariogrundriss: Klare Leer- Lade- und Fehlerzustände
    Angenommen der Renderer erhält eine konkrete Zustandsantwort
      | DTO | State | PreservedInput |
      | <dto> | <zustand> | <eingabe> |
    Wenn die HTTPansicht ausgegeben wird
    Dann lautet der sichtbare Nutzungsvertrag
      | Status | Visible | Action | BuyEnabled |
      | <status> | <sichtbar> | <aktion> | <kaufen> |
    Beispiele:
      | dto | zustand | eingabe | status | sichtbar | aktion | kaufen |
      | HomeView | offersUnavailable | none | 200 | categoryLinks,errorHint | retry | false |
      | SearchResult | empty | brand=Alpha | 200 | zeroResults,selectedChips | resetFilters | false |
      | SearchResult | loading | brand=Alpha | 200 | aria-busy,selectedChips | wait | false |
      | SearchResult | unavailable | q=phone | 503 | searchError,q=phone | retry,categoryNavigation | false |
      | ProductDetail | unknown | none | 404 | notFound | categoryNavigation | false |
      | ProductDetail | removed | none | 410 | removed | categoryNavigation | false |
      | ProductDetail | unavailable | sku=PHONE | 200 | unavailable,variantOptions | chooseVariant | false |
      | ProductDetail | compliancePending | sku=PHONE | 200 | Pflichtangaben werden geprüft | retry | false |
      | CartView | invalidQuantity | quantity=0 | 422 | fieldError,quantity=0 | correctQuantity | false |
      | CartView | expired | none | 410 | Warenkorb abgelaufen | newCart | false |
      | CheckoutView | priceChanged | address=preserved | 409 | newEvaluation,address=preserved | confirmNewTotal | false |
      | CheckoutView | declined | address=preserved | 422 | Simulierte Freigabe abgelehnt | choosePayment,retry | false |
      | OrderView | processStarting | none | 200 | Bestellung angenommen,Status wird aktualisiert | refresh | false |
      | OrderView | foreign | none | 404 | notFound | shoppingLink | false |
      | BackendOrderView | empty | none | 200 | emptyOrderList | catalogLink | false |
      | BackendOrderView | conflict | command=preserved | 409 | conflict | reload | false |
      | BackendOrderView | invalidShipmentQuantity | quantity=3 | 422 | fieldError | correctQuantity | false |
