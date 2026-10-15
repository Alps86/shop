# Typenkatalog

Der folgende Go-nahe Vertrag beschreibt Datenformen, keine bereits vorhandene Implementierung. Alle fachlichen Typen sind pointerfrei. Die Präzisierungen in [Katalog](catalog.md), [Handel](commerce.md) und [Pflichtdaten](compliance.md) sind Bestandteil dieses Vertrags. Primitive benannte Typen werden ausschließlich über validierende Konstruktoren erzeugt. Technische DTOs dürfen davon getrennt bleiben.

```go
type ProductID string
type ProductFamilyID string
type CategoryID string
type CartID string
type UserID string
type OrderID string
type ReservationID string
type DiscountRuleID string
type ShipmentID string

// An absent value is explicit and never represented by a pointer.
type Optional[T any] struct {
    Value   T
    Present bool
}
type SKU string
type Currency string
type Locale string
type CountryCode string
type AttributeCode string
type OptionCode string

type Money struct {
    MinorUnits int64
    Currency   Currency
}

type LocalizedText map[Locale]string

type Product struct {
    ID                     ProductID
    FamilyID               ProductFamilyID
    Name                   LocalizedText
    Description            LocalizedText
    PrimaryCategoryID      CategoryID
    AdditionalCategoryIDs  []CategoryID
    Attributes             map[AttributeCode]AttributeValue
    Images                 []Image
    Variants               []Variant
    Accessories            []AccessoryReference
    Brand                  string
    Manufacturer           EconomicOperator
    ResponsiblePerson      Optional[EconomicOperator]
    Compliance             ProductCompliance
}

type Variant struct {
    SKU         SKU
    GTIN        string
    ManufacturerPartNumber string
    Attributes  map[AttributeCode]AttributeValue
    Images      []Image
    Compliance  Optional[ProductCompliance]
}

type Image struct {
    URL      string
    AltText  LocalizedText
}

type AccessoryReference struct {
    ProductID  ProductID
    // Optional: non-empty only for variant-specific compatibility.
    CompatibleSKUs []SKU
}

type ProductFamily struct {
    ID          ProductFamilyID
    Name        LocalizedText
    Attributes  []AttributeDefinition
}

type AttributeDefinition struct {
    Code      AttributeCode
    Name      LocalizedText
    Type      AttributeType
    Level     AttributeLevel
    Unit      string
    Options   []AttributeOption
}

type AttributeType string
// Proposed supported values: text, number, measurement, select, multiSelect, boolean.
type AttributeLevel string
// Values: product, variant.

type AttributeOption struct {
    Code   OptionCode
    Label  LocalizedText
}

type Decimal string
// Canonical validated decimal value; not an arbitrary unvalidated string.

type Measurement struct {
    Value  Decimal
    Unit   string
}

type AttributeValue struct {
    Type         AttributeType
    Text         Optional[LocalizedText]
    Number       Optional[Decimal]
    Measurement  Optional[Measurement]
    Options      []OptionCode
    Boolean      Optional[bool]
}

type Category struct {
    ID                  CategoryID
    ParentID            Optional[CategoryID]
    Name                LocalizedText
    Description         LocalizedText
    FilterAttributes    []AttributeCode
    TopAttributes       []AttributeCode
}
```

```go
type Inventory struct {
    SKU           SKU
    Price         Money
    StockQuantity int64
    Reservations  []Reservation
    Revision      uint64
}

type Reservation struct {
    ID        ReservationID
    OrderID   OrderID
    Quantity  int64
}

type User struct {
    ID                 UserID
    CategoryInterests  []CategoryInterest
}

type CategoryInterest struct {
    CategoryID  CategoryID
    Weight      Decimal
}

type Cart struct {
    ID             CartID
    UserID         UserID
    Items          []CartItem
    LastActivityAt time.Time
    Evaluation     Optional[CartEvaluation]
}

type CartItem struct {
    SKU       SKU
    Quantity  int64
}

type CartEvaluation struct {
    Items          []EvaluatedCartItem
    Subtotal       Money
    Discounts      []AppliedDiscount
    ShippingCost   Money
    Taxes          []TaxAmount
    Total          Money
}

type EvaluatedCartItem struct {
    SKU        SKU
    Quantity   int64
    UnitPrice  Money
    Discounts  []AppliedDiscount
}

type AppliedDiscount struct {
    RuleID      DiscountRuleID
    Amount      Money
    ItemSKUs    []SKU
}

type TaxAmount struct {
    Rate    Decimal
    Amount  Money
}

type DiscountRule struct {
    ID             DiscountRuleID
    MainProductID  ProductID
    AccessoryID    ProductID
    Amount         Money
    Scope          DiscountScope
}

type DiscountScope string
// Values: order, pair.
```

```go
type Checkout struct {
    CartID               CartID
    ShippingAddress      Address
    ShippingMethod       ShippingMethod
    PaymentMethod        PaymentMethod
    ConfirmedEvaluation  Optional[CartEvaluation]
}

type Address struct {
    Recipient    string
    Street       string
    PostalCode   string
    City         string
    Country      CountryCode
}

type ShippingMethod string
// Values: standard, express.
type PaymentMethod string
// Values: creditCard, invoice.

type Order struct {
    ID               OrderID
    UserID           UserID
    PlacedAt         time.Time
    Items            []OrderItem
    ShippingAddress  Address
    ShippingMethod   ShippingMethod
    PaymentMethod    PaymentMethod
    Subtotal         Money
    Discounts        []AppliedDiscount
    ShippingCost     Money
    Taxes            []TaxAmount
    Total            Money
}

type OrderItem struct {
    SKU                SKU
    ProductName        string
    VariantAttributes  map[AttributeCode]AttributeValue
    Quantity           int64
    UnitPrice          Money
    Discounts          []AppliedDiscount
}

type FulfillmentUpdate struct {
    OrderID      OrderID
    ShipmentID   ShipmentID
    Lines        []FulfillmentLine
    Status       DeliveryStatus
}

type FulfillmentLine struct {
    SKU       SKU
    Quantity  int64
}

type DeliveryStatus string
// Initial simulation: packed, shipped, delivered.
```

```go
type EconomicOperator struct {
    Name               string
    TradeName          string
    PostalAddress      BusinessAddress
    ElectronicAddress  string
    Phone              string
}

type BusinessAddress struct {
    Street       string
    PostalCode   string
    City         string
    Country      CountryCode
}

type ProductCompliance struct {
    ApplicableRules   []RuleReference
    Safety            SafetyInformation
    ChemicalLabel     Optional[ChemicalLabel]
    EnergyLabel       Optional[EnergyLabel]
    NetContent        Optional[Measurement]
    UnitPriceBasis    Optional[Measurement]
    Documents         []ProductDocument
    Notices           []RequiredNotice
}

type RuleReference struct {
    Code        string
    Market      CountryCode
    ValidFrom   Optional[time.Time]
    ValidUntil  Optional[time.Time]
}

type SafetyInformation struct {
    Warnings      []RequiredNotice
    Instructions  []ProductDocument
}

type RequiredNotice struct {
    Code          string
    Rules         []RuleReference
    Text          LocalizedText
    Placement     []DisplayContext
}

type DisplayContext string
// Examples: teaser, listing, detail, cart, checkout.
// Mandatory placement follows the applicable rule, not an arbitrary editor choice.

type ChemicalLabel struct {
    ProductIdentifier      string
    Supplier               EconomicOperator
    NominalQuantity        Optional[Measurement]
    Pictograms             []HazardPictogram
    SignalWord             LocalizedText
    HazardStatements       []CodedStatement
    PrecautionaryStatements []CodedStatement
    SupplementalStatements []CodedStatement
    UFI                    string
}

type HazardPictogram struct {
    Code      string
    ImageURL  string
    AltText   LocalizedText
}

type CodedStatement struct {
    Code  string
    Text  LocalizedText
}

type EnergyLabel struct {
    SupplierModelIdentifier string
    EPRELRegistrationNumber string
    EPRELURL                string
    EfficiencyClass         string
    ScaleMin                string
    ScaleMax                string
    LabelDocuments          []ProductDocument
    InformationSheets       []ProductDocument
}

type ProductDocument struct {
    Kind      DocumentKind
    Locale    Locale
    URL       string
    Version   string
    IssuedAt  Optional[time.Time]
}

type DocumentKind string
// Examples: energyLabel, productInformationSheet,
// safetyDataSheet, instructions, dataActInformation.
```
